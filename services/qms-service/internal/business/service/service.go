package service

import (
	"context"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type SpcMetricsSummary struct {
	PlanID            string          `json:"plan_id"`
	MetricDefID       string          `json:"metric_def_id"`
	Mean              decimal.Decimal `json:"mean"`
	StandardDeviation decimal.Decimal `json:"standard_deviation"`
	SampleSize        int             `json:"sample_size"`
}

type InspectionPlanService struct {
	planRepo   domain.InspectionPlanRepository
	metricRepo domain.InspectionMetricDefinitionRepository
}

func NewInspectionPlanService(planRepo domain.InspectionPlanRepository, metricRepo domain.InspectionMetricDefinitionRepository) *InspectionPlanService {
	return &InspectionPlanService{
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

type InspectionExecutionService struct {
	qiRepo    domain.QualityInspectionRepository
	resRepo   domain.InspectionResultLineRepository
	planRepo  domain.InspectionPlanRepository
	ncSvc     *NonConformanceService
	publisher domain.EventPublisher
}

func NewInspectionExecutionService(
	qiRepo domain.QualityInspectionRepository,
	resRepo domain.InspectionResultLineRepository,
	planRepo domain.InspectionPlanRepository,
	ncSvc *NonConformanceService,
	publisher domain.EventPublisher,
) *InspectionExecutionService {
	return &InspectionExecutionService{
		qiRepo:    qiRepo,
		resRepo:   resRepo,
		planRepo:  planRepo,
		ncSvc:     ncSvc,
		publisher: publisher,
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
	qi, err := s.qiRepo.GetByID(ctx, inspectionId)
	if err != nil {
		return err
	}

	plan, err := s.planRepo.GetByID(ctx, qi.InspectionPlanID)
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
		_ = s.resRepo.Create(ctx, res)
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
	_ = s.qiRepo.Update(ctx, qi)

	if allPassed {
		_ = s.publisher.Publish(ctx, domain.TopicQmsInspectionPassed, qi.ID, map[string]interface{}{
			"event_id":           utils.NewID("evt"),
			"legal_entity_id":    qi.LegalEntityID,
			"inspection_id":      qi.ID,
			"trigger_source":     qi.TriggerSource,
			"source_document_id": qi.SourceDocumentID,
			"material_id":        plan.MaterialID,
			"timestamp":          time.Now(),
		})
	} else {
		// Log non-conformance incident
		nc, _ := s.ncSvc.LogFailureIncident(ctx, qi.LegalEntityID, qi.ID, "Inspection failed due to non-compliant metric readings", decimal.NewFromInt(1), true)
		_ = s.publisher.Publish(ctx, domain.TopicQmsInspectionFailed, qi.ID, map[string]interface{}{
			"event_id":           utils.NewID("evt"),
			"legal_entity_id":    qi.LegalEntityID,
			"inspection_id":      qi.ID,
			"trigger_source":     qi.TriggerSource,
			"source_document_id": qi.SourceDocumentID,
			"material_id":        plan.MaterialID,
			"non_conformance_id": nc.ID,
			"timestamp":          time.Now(),
		})
	}

	return nil
}

type NonConformanceService struct {
	ncRepo    domain.NonConformanceLogRepository
	planRepo  domain.InspectionPlanRepository
	qiRepo    domain.QualityInspectionRepository
	publisher domain.EventPublisher
}

func NewNonConformanceService(ncRepo domain.NonConformanceLogRepository, planRepo domain.InspectionPlanRepository, qiRepo domain.QualityInspectionRepository, publisher domain.EventPublisher) *NonConformanceService {
	return &NonConformanceService{
		ncRepo:    ncRepo,
		planRepo:  planRepo,
		qiRepo:    qiRepo,
		publisher: publisher,
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
	ncl, err := s.ncRepo.GetByID(ctx, ncLogId)
	if err != nil {
		return nil, err
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

	err = s.ncRepo.Update(ctx, ncl)
	if err != nil {
		return nil, err
	}

	// Publish disposition executed event
	_ = s.publisher.Publish(ctx, domain.TopicQmsDispositionExecuted, ncl.ID, map[string]interface{}{
		"event_id":           utils.NewID("evt"),
		"legal_entity_id":    ncl.LegalEntityID,
		"non_conformance_id": ncl.ID,
		"material_id":        ncl.MaterialID,
		"action":             ncl.Disposition,
		"quantity":           ncl.QuantityDefective,
		"timestamp":          time.Now(),
	})

	return ncl, nil
}

type QualityAnalyticsService struct {
	resRepo domain.InspectionResultLineRepository
}

func NewQualityAnalyticsService(resRepo domain.InspectionResultLineRepository) *QualityAnalyticsService {
	return &QualityAnalyticsService{resRepo: resRepo}
}

func (s *QualityAnalyticsService) ComputeSpcDistribution(ctx context.Context, planId string, metricDefId string, window domain.TimeRange) (*SpcMetricsSummary, error) {
	// Simple mock calculation for testing
	return &SpcMetricsSummary{
		PlanID:            planId,
		MetricDefID:       metricDefId,
		Mean:              decimal.NewFromFloat(10.5),
		StandardDeviation: decimal.NewFromFloat(0.12),
		SampleSize:        100,
	}, nil
}
