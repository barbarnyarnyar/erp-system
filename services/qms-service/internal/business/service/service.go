package service

import (
	"context"
	"math"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const txKey = "gorm_tx"

func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}

type SpcMetricsSummary struct {
	PlanID            string          `json:"plan_id"`
	MetricDefID       string          `json:"metric_def_id"`
	Mean              decimal.Decimal `json:"mean"`
	StandardDeviation decimal.Decimal `json:"standard_deviation"`
	SampleSize        int             `json:"sample_size"`
}

// ==========================================
// InspectionPlanService Implementation
// ==========================================

type InspectionPlanService struct {
	db         *gorm.DB
	planRepo   domain.InspectionPlanRepository
	metricRepo domain.InspectionMetricDefinitionRepository
}

func NewInspectionPlanService(db *gorm.DB, planRepo domain.InspectionPlanRepository, metricRepo domain.InspectionMetricDefinitionRepository) *InspectionPlanService {
	return &InspectionPlanService{
		db:         db,
		planRepo:   planRepo,
		metricRepo: metricRepo,
	}
}

func (s *InspectionPlanService) ConfigurePlan(ctx context.Context, legalEntityId string, materialId string, name string) (*domain.InspectionPlan, error) {
	ip := &domain.InspectionPlan{
		ID:            utils.NewID("plan"),
		LegalEntityID: legalEntityId,
		MaterialID:    materialId,
		PlanName:      name,
		IsActive:      true,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := s.planRepo.Create(ctx, ip)
	return ip, err
}

func (s *InspectionPlanService) GetPlanByMaterial(ctx context.Context, legalEntityId, materialId string) (*domain.InspectionPlan, error) {
	return s.planRepo.GetByMaterial(ctx, legalEntityId, materialId)
}


func (s *InspectionPlanService) RegisterPlanMetric(ctx context.Context, planId string, key string, displayName string, dataType domain.MetricDataType, minLim, maxLim *decimal.Decimal) (*domain.InspectionMetricDefinition, error) {
	imd := &domain.InspectionMetricDefinition{
		ID:                utils.NewID("met"),
		InspectionPlanID:  planId,
		MetricKey:         key,
		DisplayName:       displayName,
		DataType:          dataType,
		MinToleranceLimit: minLim,
		MaxToleranceLimit: maxLim,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	err := s.metricRepo.Create(ctx, imd)
	return imd, err
}

// ==========================================
// InspectionExecutionService Implementation
// ==========================================

type InspectionExecutionService struct {
	db          *gorm.DB
	qiRepo      domain.QualityInspectionRepository
	resRepo     domain.InspectionResultLineRepository
	planRepo    domain.InspectionPlanRepository
	ncSvc       *NonConformanceService
	reliableSvc ReliableMessagingService
}

func NewInspectionExecutionService(
	db *gorm.DB,
	qiRepo domain.QualityInspectionRepository,
	resRepo domain.InspectionResultLineRepository,
	planRepo domain.InspectionPlanRepository,
	ncSvc *NonConformanceService,
	reliableSvc ReliableMessagingService,
) *InspectionExecutionService {
	return &InspectionExecutionService{
		db:          db,
		qiRepo:      qiRepo,
		resRepo:     resRepo,
		planRepo:    planRepo,
		ncSvc:       ncSvc,
		reliableSvc: reliableSvc,
	}
}

func (s *InspectionExecutionService) StageInspection(ctx context.Context, legalEntityId string, planId string, trigger domain.InspectionTriggerType, sourceDocId string) (*domain.QualityInspection, error) {
	qi := &domain.QualityInspection{
		ID:               utils.NewID("insp"),
		LegalEntityID:    legalEntityId,
		InspectionPlanID: planId,
		InspectionNumber: "INSP-" + utils.NewID("num")[:8],
		TriggerSource:    trigger,
		SourceDocumentID: sourceDocId,
		Status:           domain.InspectionStatusPENDING,
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	err := s.qiRepo.Create(ctx, qi)
	return qi, err
}

func (s *InspectionExecutionService) AssignInspector(ctx context.Context, inspectionId string, inspectorHrId string) (*domain.QualityInspection, error) {
	qi, err := s.qiRepo.GetByID(ctx, inspectionId)
	if err != nil {
		return nil, err
	}
	qi.InspectorHrID = &inspectorHrId
	qi.Status = domain.InspectionStatusIN_PROGRESS
	qi.Version++
	qi.UpdatedAt = time.Now()
	err = s.qiRepo.Update(ctx, qi)
	return qi, err
}

func (s *InspectionExecutionService) RecordBulkMeasurements(ctx context.Context, inspectionId string, samples []domain.MetricSubmissionInput) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		qi, err := s.qiRepo.GetByID(txCtx, inspectionId)
		if err != nil {
			return err
		}

		plan, err := s.planRepo.GetByID(txCtx, qi.InspectionPlanID)
		if err != nil {
			return err
		}

		allPassed := true
		for _, sample := range samples {
			res := &domain.InspectionResultLine{
				ID:                   utils.NewID("res"),
				InspectionID:         inspectionId,
				MetricDefinitionID:   sample.MetricDefinitionID,
				SampleSequence:       sample.SampleSequence,
				MeasuredNumericValue: sample.NumericValue,
				MeasuredBooleanValue: sample.BooleanValue,
				IsCompliant:          sample.IsCompliant,
				CreatedAt:            time.Now(),
			}
			if err := s.resRepo.Create(txCtx, res); err != nil {
				return err
			}
			if !sample.IsCompliant {
				allPassed = false
			}
		}

		if allPassed {
			qi.Status = domain.InspectionStatusPASSED
		} else {
			qi.Status = domain.InspectionStatusFAILED
		}
		qi.Version++
		qi.UpdatedAt = time.Now()
		if err := s.qiRepo.Update(txCtx, qi); err != nil {
			return err
		}

		triggerTypeStr := ""
		if t, ok := qi.TriggerSource.(domain.InspectionTriggerType); ok {
			triggerTypeStr = string(t)
		} else if t, ok := qi.TriggerSource.(string); ok {
			triggerTypeStr = t
		}

		if allPassed {
			err = s.reliableSvc.PushToOutbox(txCtx, tx, domain.TopicQmsInspectionPassed, qi.ID, map[string]interface{}{
				"event_id":           utils.NewID("evt"),
				"legal_entity_id":    qi.LegalEntityID,
				"inspection_id":      qi.ID,
				"trigger_source":     triggerTypeStr,
				"source_document_id": qi.SourceDocumentID,
				"material_id":        plan.MaterialID,
				"timestamp":          time.Now(),
			})
			if err != nil {
				return err
			}
		} else {
			// Log non-conformance incident
			nc, err := s.ncSvc.LogFailureIncident(txCtx, qi.LegalEntityID, qi.ID, "Inspection failed due to non-compliant metric readings", decimal.NewFromInt(1), true)
			if err != nil {
				return err
			}
			err = s.reliableSvc.PushToOutbox(txCtx, tx, domain.TopicQmsInspectionFailed, qi.ID, map[string]interface{}{
				"event_id":           utils.NewID("evt"),
				"legal_entity_id":    qi.LegalEntityID,
				"inspection_id":      qi.ID,
				"trigger_source":     triggerTypeStr,
				"source_document_id": qi.SourceDocumentID,
				"material_id":        plan.MaterialID,
				"non_conformance_id": nc.ID,
				"timestamp":          time.Now(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// ==========================================
// NonConformanceService Implementation
// ==========================================

type NonConformanceService struct {
	db          *gorm.DB
	ncRepo      domain.NonConformanceLogRepository
	planRepo    domain.InspectionPlanRepository
	qiRepo      domain.QualityInspectionRepository
	reliableSvc ReliableMessagingService
}

func NewNonConformanceService(
	db *gorm.DB,
	ncRepo domain.NonConformanceLogRepository,
	planRepo domain.InspectionPlanRepository,
	qiRepo domain.QualityInspectionRepository,
	reliableSvc ReliableMessagingService,
) *NonConformanceService {
	return &NonConformanceService{
		db:          db,
		ncRepo:      ncRepo,
		planRepo:    planRepo,
		qiRepo:      qiRepo,
		reliableSvc: reliableSvc,
	}
}

func (s *NonConformanceService) LogFailureIncident(ctx context.Context, legalEntityId string, inspectionId string, description string, qty decimal.Decimal, autoQuarantine bool) (*domain.NonConformanceLog, error) {
	qi, err := s.qiRepo.GetByID(ctx, inspectionId)
	if err != nil {
		return nil, err
	}
	plan, err := s.planRepo.GetByID(ctx, qi.InspectionPlanID)
	if err != nil {
		return nil, err
	}

	ncl := &domain.NonConformanceLog{
		ID:                utils.NewID("nc"),
		LegalEntityID:     legalEntityId,
		InspectionID:      inspectionId,
		NcNumber:          "NC-" + utils.NewID("num")[:8],
		MaterialID:        plan.MaterialID,
		DefectDescription: description,
		QuantityDefective: qty,
		IsQuarantined:     autoQuarantine,
		Version:           1,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	err = s.ncRepo.Create(ctx, ncl)
	return ncl, err
}

func (s *NonConformanceService) ExecuteDisposition(ctx context.Context, ncLogId string, action domain.DispositionAction, notes string, resolverHrId string) (*domain.NonConformanceLog, error) {
	var ncl *domain.NonConformanceLog
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		var err error
		ncl, err = s.ncRepo.GetByID(txCtx, ncLogId)
		if err != nil {
			return err
		}

		var actionVal interface{} = action
		ncl.Disposition = &actionVal
		ncl.DispositionNotes = &notes
		ncl.ResolvedByHrID = &resolverHrId
		now := time.Now()
		ncl.ResolvedAt = &now
		ncl.IsQuarantined = false // Released from quarantine upon resolution
		ncl.Version++
		ncl.UpdatedAt = now

		err = s.ncRepo.Update(txCtx, ncl)
		if err != nil {
			return err
		}

		// Publish disposition executed event
		err = s.reliableSvc.PushToOutbox(txCtx, tx, domain.TopicQmsDispositionExecuted, ncl.ID, map[string]interface{}{
			"event_id":           utils.NewID("evt"),
			"legal_entity_id":    ncl.LegalEntityID,
			"non_conformance_id": ncl.ID,
			"material_id":        ncl.MaterialID,
			"action":             string(action),
			"quantity":           ncl.QuantityDefective,
			"timestamp":          time.Now(),
		})
		return err
	})
	return ncl, err
}

// ==========================================
// QualityAnalyticsService Implementation
// ==========================================

type QualityAnalyticsService struct {
	db      *gorm.DB
	resRepo domain.InspectionResultLineRepository
}

func NewQualityAnalyticsService(db *gorm.DB, resRepo domain.InspectionResultLineRepository) *QualityAnalyticsService {
	return &QualityAnalyticsService{
		db:      db,
		resRepo: resRepo,
	}
}

func (s *QualityAnalyticsService) ComputeSpcDistribution(ctx context.Context, planId string, metricDefId string, window domain.TimeRange) (*SpcMetricsSummary, error) {
	lines, err := s.resRepo.ListByMetricAndDateRange(ctx, metricDefId, window.StartDate, window.EndDate)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return &SpcMetricsSummary{
			PlanID:            planId,
			MetricDefID:       metricDefId,
			Mean:              decimal.Zero,
			StandardDeviation: decimal.Zero,
			SampleSize:        0,
		}, nil
	}

	n := len(lines)
	sum := decimal.Zero
	numericVals := make([]decimal.Decimal, 0, n)
	for _, l := range lines {
		if l.MeasuredNumericValue != nil {
			sum = sum.Add(*l.MeasuredNumericValue)
			numericVals = append(numericVals, *l.MeasuredNumericValue)
		}
	}

	nNumeric := len(numericVals)
	if nNumeric == 0 {
		return &SpcMetricsSummary{
			PlanID:            planId,
			MetricDefID:       metricDefId,
			Mean:              decimal.Zero,
			StandardDeviation: decimal.Zero,
			SampleSize:        0,
		}, nil
	}

	mean := sum.Div(decimal.NewFromInt(int64(nNumeric)))

	varianceSum := decimal.Zero
	for _, val := range numericVals {
		diff := val.Sub(mean)
		varianceSum = varianceSum.Add(diff.Mul(diff))
	}
	variance := varianceSum.Div(decimal.NewFromInt(int64(nNumeric)))
	
	var stdDev decimal.Decimal
	varF, _ := variance.Float64()
	if varF > 0 {
		stdDev = decimal.NewFromFloat(math.Sqrt(varF))
	} else {
		stdDev = decimal.Zero
	}

	return &SpcMetricsSummary{
		PlanID:            planId,
		MetricDefID:       metricDefId,
		Mean:              mean.Round(4),
		StandardDeviation: stdDev.Round(4),
		SampleSize:        nNumeric,
	}, nil
}

// ==========================================
// OutboxRelayWorker Implementation
// ==========================================

type OutboxRelayWorker interface {
	GetUnsentMessages(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error)
	UpdateOutboxStatus(ctx context.Context, tx *gorm.DB, outboxID string, status domain.OutboxStatus) error
}

type OutboxRelayWorkerImpl struct {
	db         *gorm.DB
	outboxRepo domain.TransactionalOutboxRepository
}

func NewOutboxRelayWorker(db *gorm.DB, outboxRepo domain.TransactionalOutboxRepository) OutboxRelayWorker {
	return &OutboxRelayWorkerImpl{
		db:         db,
		outboxRepo: outboxRepo,
	}
}

func (s *OutboxRelayWorkerImpl) GetUnsentMessages(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	return s.outboxRepo.GetUnsent(ctx, limit)
}

func (s *OutboxRelayWorkerImpl) UpdateOutboxStatus(ctx context.Context, tx *gorm.DB, outboxID string, status domain.OutboxStatus) error {
	txCtx := ctx
	if tx != nil {
		txCtx = context.WithValue(ctx, txKey, tx)
	}
	return s.outboxRepo.UpdateStatus(txCtx, outboxID, status)
}

// ==========================================
// ReliableMessagingService Implementation
// ==========================================

type ReliableMessagingService interface {
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
	CommitInboundEvent(ctx context.Context, eventID string, eventType string, payload interface{}) error
	PushToOutbox(ctx context.Context, tx *gorm.DB, eventType string, aggregateID string, payload interface{}) error
	ExecuteIdempotentTransaction(ctx context.Context, eventID string, eventType string, payload interface{}, businessRoutine func(ctx context.Context) error) error
}

type ReliableMessagingServiceImpl struct {
	db         *gorm.DB
	inboxRepo  domain.KafkaEventInboxRepository
	outboxRepo domain.TransactionalOutboxRepository
}

func NewReliableMessagingService(db *gorm.DB, inboxRepo domain.KafkaEventInboxRepository, outboxRepo domain.TransactionalOutboxRepository) ReliableMessagingService {
	return &ReliableMessagingServiceImpl{
		db:         db,
		inboxRepo:  inboxRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *ReliableMessagingServiceImpl) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	return s.inboxRepo.Exists(ctx, eventID)
}

func (s *ReliableMessagingServiceImpl) CommitInboundEvent(ctx context.Context, eventID string, eventType string, payload interface{}) error {
	inbox := &domain.KafkaEventInbox{
		EventID:          eventID,
		EventType:        eventType,
		ProcessedAt:      time.Now(),
		ProcessingStatus: domain.EventProcessingStatusSUCCESS,
		Payload:          payload,
	}
	return s.inboxRepo.Create(ctx, inbox)
}

func (s *ReliableMessagingServiceImpl) PushToOutbox(ctx context.Context, tx *gorm.DB, eventType string, aggregateID string, payload interface{}) error {
	outbox := &domain.TransactionalOutbox{
		ID:          utils.NewID("out"),
		EventType:   eventType,
		AggregateID: aggregateID,
		Payload:     payload,
		Status:      domain.OutboxStatusPENDING,
		CreatedAt:   time.Now(),
	}
	txCtx := ctx
	if tx != nil {
		txCtx = context.WithValue(ctx, txKey, tx)
	}
	return s.outboxRepo.Create(txCtx, outbox)
}

func (s *ReliableMessagingServiceImpl) ExecuteIdempotentTransaction(
	ctx context.Context,
	eventID string,
	eventType string,
	payload interface{},
	businessRoutine func(ctx context.Context) error,
) error {
	// First check if the event was already successfully processed or sent to DLQ
	msg, err := s.inboxRepo.GetByID(ctx, eventID)
	if err == nil && msg != nil {
		if msg.ProcessingStatus == domain.EventProcessingStatusSUCCESS || msg.ProcessingStatus == "FAILED_DLQ" {
			return nil
		}
	}

	attempts := 0
	if msg != nil {
		attempts = msg.AttemptCount
	}
	attempts++

	// If attempts >= 5, route to DLQ
	if attempts >= 5 {
		inboxEntry := &domain.KafkaEventInbox{
			EventID:          eventID,
			EventType:        eventType,
			ProcessedAt:      time.Now(),
			ProcessingStatus: "FAILED_DLQ",
			Payload:          payload,
			AttemptCount:     attempts,
		}

		if msg != nil {
			_ = s.inboxRepo.Update(ctx, inboxEntry)
		} else {
			_ = s.inboxRepo.Create(ctx, inboxEntry)
		}

		// Central DLQ topic
		dlqTopic := "erp.system.dlq"
		dlqMsg := map[string]interface{}{
			"event_id":       eventID,
			"original_topic": eventType,
			"payload":        payload,
			"error":          "Max retries reached (5 attempts)",
			"failed_at":      time.Now(),
		}

		if pub, ok := ctx.Value("publisher").(interface {
			Publish(ctx context.Context, topic string, key string, payload interface{}) error
		}); ok && pub != nil {
			_ = pub.Publish(ctx, dlqTopic, eventID, dlqMsg)
		}

		return nil // Return nil so consumer commits offset and doesn't loop
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		if err := businessRoutine(txCtx); err != nil {
			inboxEntry := &domain.KafkaEventInbox{
				EventID:          eventID,
				EventType:        eventType,
				ProcessedAt:      time.Now(),
				ProcessingStatus: domain.EventProcessingStatusFAILED,
				Payload:          payload,
				AttemptCount:     attempts,
			}
			if msg != nil {
				_ = s.inboxRepo.Update(txCtx, inboxEntry)
			} else {
				_ = s.inboxRepo.Create(txCtx, inboxEntry)
			}
			return err
		}

		inboxEntry := &domain.KafkaEventInbox{
			EventID:          eventID,
			EventType:        eventType,
			ProcessedAt:      time.Now(),
			ProcessingStatus: domain.EventProcessingStatusSUCCESS,
			Payload:          payload,
			AttemptCount:     attempts,
		}
		if msg != nil {
			return s.inboxRepo.Update(txCtx, inboxEntry)
		} else {
			return s.inboxRepo.Create(txCtx, inboxEntry)
		}
	})
}
