package service

import (
	"context"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const txKey = "gorm_tx"

// ==========================================
// EquipmentService Implementation
// ==========================================

type EquipmentService struct {
	db          *gorm.DB
	facRepo     domain.FacilityRepository
	eqRepo      domain.EquipmentRepository
	reliableSvc ReliableMessagingService
}

func NewEquipmentService(db *gorm.DB, facRepo domain.FacilityRepository, eqRepo domain.EquipmentRepository, reliableSvc ReliableMessagingService) *EquipmentService {
	return &EquipmentService{
		db:          db,
		facRepo:     facRepo,
		eqRepo:      eqRepo,
		reliableSvc: reliableSvc,
	}
}

func (s *EquipmentService) CreateFacility(ctx context.Context, legalEntityId string, name string, address string) (*domain.Facility, error) {
	f := &domain.Facility{
		ID:              utils.NewID("fac"),
		LegalEntityID:   legalEntityId,
		Name:            name,
		PhysicalAddress: address,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := s.facRepo.Create(ctx, f)
	return f, err
}

func (s *EquipmentService) RegisterEquipment(ctx context.Context, legalEntityId string, facilityId string, assetTag string, name string, serialNumber string) (*domain.Equipment, error) {
	eq := &domain.Equipment{
		ID:            utils.NewID("eq"),
		LegalEntityID: legalEntityId,
		FacilityID:    facilityId,
		AssetTag:      assetTag,
		Name:          name,
		SerialNumber:  serialNumber,
		Status:        domain.EquipmentStatusONLINE,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := s.eqRepo.Create(ctx, eq)
	return eq, err
}

func (s *EquipmentService) UpdateEquipmentStatus(ctx context.Context, tx *gorm.DB, equipmentId string, newStatus domain.EquipmentStatus) (*domain.Equipment, error) {
	txCtx := ctx
	if tx != nil {
		txCtx = context.WithValue(ctx, txKey, tx)
	}

	eq, err := s.eqRepo.GetByID(txCtx, equipmentId)
	if err != nil {
		return nil, err
	}

	oldStatus := eq.Status
	eq.Status = newStatus
	eq.UpdatedAt = time.Now()
	err = s.eqRepo.Update(txCtx, eq)
	if err != nil {
		return nil, err
	}

	// Publish machine status change events via outbox
	if oldStatus != newStatus {
		if newStatus == domain.EquipmentStatusOFFLINE_BROKEN {
			_ = s.reliableSvc.PushToOutbox(txCtx, tx, domain.TopicEamMachineOffline, eq.ID, map[string]interface{}{
				"event_id":        utils.NewID("evt"),
				"legal_entity_id": eq.LegalEntityID,
				"equipment_id":    eq.ID,
				"timestamp":       time.Now(),
			})
		} else if newStatus == domain.EquipmentStatusONLINE && oldStatus == domain.EquipmentStatusOFFLINE_BROKEN {
			_ = s.reliableSvc.PushToOutbox(txCtx, tx, domain.TopicEamMachineOnline, eq.ID, map[string]interface{}{
				"event_id":        utils.NewID("evt"),
				"legal_entity_id": eq.LegalEntityID,
				"equipment_id":    eq.ID,
				"timestamp":       time.Now(),
			})
		}
	}

	return eq, nil
}

func (s *EquipmentService) AssociateFinancialAsset(ctx context.Context, equipmentId string, financialAssetId string) (*domain.Equipment, error) {
	eq, err := s.eqRepo.GetByID(ctx, equipmentId)
	if err != nil {
		return nil, err
	}
	eq.FinancialAssetID = &financialAssetId
	eq.UpdatedAt = time.Now()
	err = s.eqRepo.Update(ctx, eq)
	return eq, err
}

func (s *EquipmentService) FetchTargetTenantAssets(ctx context.Context, legalEntityId string, status domain.EquipmentStatus) ([]domain.Equipment, error) {
	all, err := s.eqRepo.ListByTenant(ctx, legalEntityId)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Equipment, 0)
	for _, eq := range all {
		if eq.Status == status {
			res = append(res, eq)
		}
	}
	return res, nil
}

// ==========================================
// MaintenanceService Implementation
// ==========================================

type MaintenanceService struct {
	db          *gorm.DB
	woRepo      domain.MaintenanceWorkOrderRepository
	eqRepo      domain.EquipmentRepository
	schRepo     domain.PreventativeScheduleRepository
	reliableSvc ReliableMessagingService
}

func NewMaintenanceService(db *gorm.DB, woRepo domain.MaintenanceWorkOrderRepository, eqRepo domain.EquipmentRepository, schRepo domain.PreventativeScheduleRepository, reliableSvc ReliableMessagingService) *MaintenanceService {
	return &MaintenanceService{
		db:          db,
		woRepo:      woRepo,
		eqRepo:      eqRepo,
		schRepo:     schRepo,
		reliableSvc: reliableSvc,
	}
}

func (s *MaintenanceService) FileMachineIncident(ctx context.Context, legalEntityId string, equipmentId string, reportedBy string, title string, priority domain.WorkOrderPriority) (*domain.MaintenanceWorkOrder, error) {
	var wo *domain.MaintenanceWorkOrder
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		wo = &domain.MaintenanceWorkOrder{
			ID:             utils.NewID("wo"),
			LegalEntityID:  legalEntityId,
			EquipmentID:    equipmentId,
			TicketNumber:   "WO-" + utils.NewID("num")[:8],
			Title:          title,
			Category:       domain.MaintenanceCategoryREACTIVE,
			Priority:       priority,
			Status:         domain.WorkOrderStatusOPEN,
			ReportedByHrID: reportedBy,
			ReportedAt:     time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := s.woRepo.Create(txCtx, wo); err != nil {
			return err
		}

		// If priority is critical, mark equipment offline/broken
		if priority == domain.WorkOrderPriorityCRITICAL || priority == domain.WorkOrderPriorityHIGH {
			eq, err := s.eqRepo.GetByID(txCtx, equipmentId)
			if err == nil && eq != nil {
				eq.Status = domain.EquipmentStatusOFFLINE_BROKEN
				eq.UpdatedAt = time.Now()
				_ = s.eqRepo.Update(txCtx, eq)

				// Publish machine offline event
				_ = s.reliableSvc.PushToOutbox(txCtx, tx, domain.TopicEamMachineOffline, eq.ID, map[string]interface{}{
					"event_id":        utils.NewID("evt"),
					"legal_entity_id": eq.LegalEntityID,
					"equipment_id":    eq.ID,
					"work_order_id":   wo.ID,
					"priority":        string(priority),
					"timestamp":       time.Now(),
				})
			}
		}
		return nil
	})
	return wo, err
}

func (s *MaintenanceService) RouteToTechnician(ctx context.Context, workOrderId string, techHrId string) (*domain.MaintenanceWorkOrder, error) {
	wo, err := s.woRepo.GetByID(ctx, workOrderId)
	if err != nil {
		return nil, err
	}
	wo.AssignedTechHrID = &techHrId
	wo.Status = domain.WorkOrderStatusASSIGNED
	wo.UpdatedAt = time.Now()
	err = s.woRepo.Update(ctx, wo)
	return wo, err
}

func (s *MaintenanceService) TransitionToActiveState(ctx context.Context, workOrderId string) (*domain.MaintenanceWorkOrder, error) {
	wo, err := s.woRepo.GetByID(ctx, workOrderId)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	wo.StartedAt = &now
	wo.Status = domain.WorkOrderStatusIN_PROGRESS
	wo.UpdatedAt = now
	err = s.woRepo.Update(ctx, wo)
	return wo, err
}

func (s *MaintenanceService) FinalizeResolution(ctx context.Context, workOrderId string, resolutionNotes string) (*domain.MaintenanceWorkOrder, error) {
	var wo *domain.MaintenanceWorkOrder
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		var err error
		wo, err = s.woRepo.GetByID(txCtx, workOrderId)
		if err != nil {
			return err
		}
		now := time.Now()
		wo.ResolvedAt = &now
		wo.ResolutionNotes = &resolutionNotes
		wo.Status = domain.WorkOrderStatusRESOLVED
		wo.UpdatedAt = now
		err = s.woRepo.Update(txCtx, wo)
		if err != nil {
			return err
		}

		// Set equipment status to ONLINE when resolved
		eq, err := s.eqRepo.GetByID(txCtx, wo.EquipmentID)
		if err == nil && eq != nil {
			eq.Status = domain.EquipmentStatusONLINE
			eq.UpdatedAt = now
			_ = s.eqRepo.Update(txCtx, eq)

			// Publish machine online event
			_ = s.reliableSvc.PushToOutbox(txCtx, tx, domain.TopicEamMachineOnline, eq.ID, map[string]interface{}{
				"event_id":        utils.NewID("evt"),
				"legal_entity_id": eq.LegalEntityID,
				"equipment_id":    eq.ID,
				"work_order_id":   wo.ID,
				"timestamp":       now,
			})
		}
		return nil
	})
	return wo, err
}

func (s *MaintenanceService) RequestSpares(ctx context.Context, workOrderId string, componentDetails interface{}) error {
	wo, err := s.woRepo.GetByID(ctx, workOrderId)
	if err != nil {
		return err
	}
	return s.reliableSvc.PushToOutbox(ctx, nil, domain.TopicEamWorkorderSparesRequested, wo.ID, map[string]interface{}{
		"event_id":          utils.NewID("evt"),
		"legal_entity_id":   wo.LegalEntityID,
		"work_order_id":     wo.ID,
		"component_details": componentDetails,
		"timestamp":         time.Now(),
	})
}

func (s *MaintenanceService) ProcessCronSchedulerLookups(ctx context.Context, targetDate time.Time) error {
	schedules, err := s.schRepo.List(ctx)
	if err != nil {
		return err
	}
	for _, sch := range schedules {
		if sch.IsActive && (sch.NextDueDate.Before(targetDate) || sch.NextDueDate.Equal(targetDate)) {
			// Trigger a PM Work Order
			wo := &domain.MaintenanceWorkOrder{
				ID:             utils.NewID("wo"),
				LegalEntityID:  sch.LegalEntityID,
				EquipmentID:    sch.EquipmentID,
				TicketNumber:   "PM-" + utils.NewID("num")[:8],
				Title:          "Preventative Maintenance: " + sch.Title,
				Description:    sch.InstructionSet,
				Category:       domain.MaintenanceCategoryPREVENTATIVE,
				Priority:       domain.WorkOrderPriorityMEDIUM,
				Status:         domain.WorkOrderStatusOPEN,
				ReportedByHrID: "system",
				ReportedAt:     time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			_ = s.woRepo.Create(ctx, wo)

			// Update last executed and next due date
			now := time.Now()
			sch.LastExecutedAt = &now
			days := sch.IntervalDays
			if days <= 0 {
				days = 30
			}
			sch.NextDueDate = sch.NextDueDate.AddDate(0, 0, days)
			sch.UpdatedAt = time.Now()
			_ = s.schRepo.Update(ctx, &sch)
		}
	}
	return nil
}

// ==========================================
// TelemetryIngestionService Implementation
// ==========================================

type TelemetryIngestionService struct {
	db      *gorm.DB
	bufRepo domain.TelemetryIngestBufferRepository
}

func NewTelemetryIngestionService(db *gorm.DB, bufRepo domain.TelemetryIngestBufferRepository) *TelemetryIngestionService {
	return &TelemetryIngestionService{
		db:      db,
		bufRepo: bufRepo,
	}
}

func (s *TelemetryIngestionService) QueueSensorMetrics(ctx context.Context, legalEntityId string, equipmentId string, sensorKey string, value decimal.Decimal) error {
	tb := &domain.TelemetryIngestBuffer{
		ID:            utils.NewID("tel"),
		LegalEntityID: legalEntityId,
		EquipmentID:   equipmentId,
		SensorKey:     sensorKey,
		ReadingValue:  value,
		RecordedAt:    time.Now(),
	}
	return s.bufRepo.Create(ctx, tb)
}

func (s *TelemetryIngestionService) FlushStagedMetricsToTimeSeriesStore(ctx context.Context, tx *gorm.DB, batchLimit int) ([]string, error) {
	var idsToDelete []string

	execute := func(activeTx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, activeTx)
		metrics, err := s.bufRepo.LockAndList(txCtx, batchLimit)
		if err != nil {
			return err
		}
		if len(metrics) == 0 {
			return nil
		}
		idsToDelete = make([]string, 0, len(metrics))
		for _, m := range metrics {
			idsToDelete = append(idsToDelete, m.ID)
		}
		return s.bufRepo.DeleteBatch(txCtx, idsToDelete)
	}

	var err error
	if tx != nil {
		err = execute(tx)
	} else {
		err = s.db.Transaction(func(activeTx *gorm.DB) error {
			return execute(activeTx)
		})
	}
	return idsToDelete, err
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
	msg, err := s.outboxRepo.GetByID(txCtx, outboxID)
	if err != nil {
		return err
	}
	msg.Status = status
	return s.outboxRepo.Update(txCtx, msg)
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
	msg, err := s.inboxRepo.GetByID(ctx, eventID)
	if err == nil && msg != nil {
		return msg.ProcessingStatus == domain.EventProcessingStatusSUCCESS, nil
	}
	return false, nil
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

func (s *ReliableMessagingServiceImpl) ExecuteIdempotentTransaction(ctx context.Context, eventID string, eventType string, payload interface{}, businessRoutine func(ctx context.Context) error) error {
	processed, err := s.IsEventProcessed(ctx, eventID)
	if err != nil {
		return err
	}
	if processed {
		return nil
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
			}
			_ = s.inboxRepo.Create(txCtx, inboxEntry)
			return err
		}

		inboxEntry := &domain.KafkaEventInbox{
			EventID:          eventID,
			EventType:        eventType,
			ProcessedAt:      time.Now(),
			ProcessingStatus: domain.EventProcessingStatusSUCCESS,
			Payload:          payload,
		}
		return s.inboxRepo.Create(txCtx, inboxEntry)
	})
}
