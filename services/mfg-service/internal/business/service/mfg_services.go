package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"erp-system/shared/utils"
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

// ==========================================
// FloorConfigurationService Implementation
// ==========================================

type FloorConfigurationService interface {
	EstablishWorkCenter(ctx context.Context, legalEntityID, code, name string) (*domain.WorkCenter, error)
	AppendStationToCenter(ctx context.Context, workCenterID, routingCode string, stationType domain.StationType, equipmentID *string, setupTime, runTime int) (*domain.RoutingStation, error)
}

type FloorConfigurationServiceImpl struct {
	wcRepo      domain.WorkCenterRepository
	stationRepo domain.RoutingStationRepository
}

func NewFloorConfigurationService(wcRepo domain.WorkCenterRepository, stationRepo domain.RoutingStationRepository) FloorConfigurationService {
	return &FloorConfigurationServiceImpl{
		wcRepo:      wcRepo,
		stationRepo: stationRepo,
	}
}

func (s *FloorConfigurationServiceImpl) EstablishWorkCenter(ctx context.Context, legalEntityID, code, name string) (*domain.WorkCenter, error) {
	wc, err := s.wcRepo.GetByCode(ctx, legalEntityID, code)
	if err == nil && wc != nil {
		return wc, nil
	}

	id := utils.NewID("wc")
	newWc := &domain.WorkCenter{
		ID:             id,
		LegalEntityID:  legalEntityID,
		WorkCenterCode: code,
		Name:           name,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.wcRepo.Create(ctx, newWc); err != nil {
		return nil, err
	}
	return newWc, nil
}

func (s *FloorConfigurationServiceImpl) AppendStationToCenter(ctx context.Context, workCenterID, routingCode string, stationType domain.StationType, equipmentID *string, setupTime, runTime int) (*domain.RoutingStation, error) {
	_, err := s.wcRepo.GetByID(ctx, workCenterID)
	if err != nil {
		return nil, fmt.Errorf("work center not found: %w", err)
	}

	existing, err := s.stationRepo.GetByCode(ctx, workCenterID, routingCode)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("routing station with code %s already exists in work center", routingCode)
	}

	id := utils.NewID("rs")
	newStation := &domain.RoutingStation{
		ID:                    id,
		WorkCenterID:          workCenterID,
		RoutingCode:           routingCode,
		StationType:           stationType,
		EquipmentID:           equipmentID,
		StandardSetupTimeMins: setupTime,
		StandardRunTimeMins:   runTime,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.stationRepo.Create(ctx, newStation); err != nil {
		return nil, err
	}
	return newStation, nil
}

// ==========================================
// WorkOrderExecutionService Implementation
// ==========================================

type WorkOrderExecutionService interface {
	InstantiateWorkOrder(ctx context.Context, legalEntityID, materialID, bomHeaderID string, qtyTarget decimal.Decimal, start, end time.Time) (*domain.WorkOrder, error)
	TransitionWorkOrderState(ctx context.Context, workOrderID string, currentState, targetState domain.WorkOrderState) (*domain.WorkOrder, error)
	RerouteWorkOrderStation(ctx context.Context, workOrderID, currentStationID, targetStationID string, isRework bool) error
	FreezeObsoleteWorkOrders(ctx context.Context, materialID string, newBomHeaderID string) error
}

type WorkOrderExecutionServiceImpl struct {
	db          *gorm.DB
	woRepo      domain.WorkOrderRepository
	stateRepo   domain.WorkOrderRoutingStateRepository
	stationRepo domain.RoutingStationRepository
	outboxRepo  domain.TransactionalOutboxRepository
}

func NewWorkOrderExecutionService(
	db *gorm.DB,
	woRepo domain.WorkOrderRepository,
	stateRepo domain.WorkOrderRoutingStateRepository,
	stationRepo domain.RoutingStationRepository,
	outboxRepo domain.TransactionalOutboxRepository,
) WorkOrderExecutionService {
	return &WorkOrderExecutionServiceImpl{
		db:          db,
		woRepo:      woRepo,
		stateRepo:   stateRepo,
		stationRepo: stationRepo,
		outboxRepo:  outboxRepo,
	}
}

func (s *WorkOrderExecutionServiceImpl) emitEvent(ctx context.Context, eventType, aggregateID string, payload interface{}) error {
	outbox := &domain.TransactionalOutbox{
		ID:          utils.NewID("outbox"),
		EventType:   eventType,
		AggregateID: aggregateID,
		Payload:     payload,
		Status:      domain.OutboxStatusPENDING,
		RetryCount:  0,
		CreatedAt:   time.Now(),
	}
	return s.outboxRepo.Create(ctx, outbox)
}

func (s *WorkOrderExecutionServiceImpl) InstantiateWorkOrder(ctx context.Context, legalEntityID, materialID, bomHeaderID string, qtyTarget decimal.Decimal, start, end time.Time) (*domain.WorkOrder, error) {
	id := utils.NewID("wo")
	woNum := "WO-" + id[:8]

	newWo := &domain.WorkOrder{
		ID:               id,
		LegalEntityID:    legalEntityID,
		MaterialID:       materialID,
		BomHeaderID:      bomHeaderID,
		WorkOrderNumber:  woNum,
		QuantityTarget:   qtyTarget,
		QuantityProduced: decimal.Zero,
		Status:           domain.WorkOrderStateSTAGED,
		ScheduledStart:   start,
		ScheduledEnd:     end,
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.woRepo.Create(ctx, newWo); err != nil {
		return nil, err
	}

	return newWo, nil
}

func (s *WorkOrderExecutionServiceImpl) TransitionWorkOrderState(ctx context.Context, workOrderID string, currentState, targetState domain.WorkOrderState) (*domain.WorkOrder, error) {
	var wo *domain.WorkOrder
	var err error

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		wo, err = s.woRepo.GetByID(txCtx, workOrderID)
		if err != nil {
			return fmt.Errorf("work order not found: %w", err)
		}

		if wo.Status != currentState {
			return fmt.Errorf("invalid current state: expected %s, got %s", currentState, wo.Status)
		}

		wo.Status = targetState
		wo.UpdatedAt = time.Now()

		if err := s.woRepo.Update(txCtx, wo); err != nil {
			return err
		}

		if targetState == domain.WorkOrderStateIN_PROGRESS {
			evt := domain.MfgProductionStartedEvent{
				EventID:       utils.NewID("evt"),
				LegalEntityID: wo.LegalEntityID,
				WorkOrderID:   wo.ID,
				MaterialID:    wo.MaterialID,
				Timestamp:     time.Now(),
			}
			if err := s.emitEvent(txCtx, domain.TopicMfgProductionStarted, wo.ID, evt); err != nil {
				return err
			}
		} else if targetState == domain.WorkOrderStateCOMPLETED {
			evt := domain.MfgWorkOrderCompletedEvent{
				EventID:          utils.NewID("evt"),
				LegalEntityID:    wo.LegalEntityID,
				WorkOrderID:      wo.ID,
				MaterialID:       wo.MaterialID,
				QuantityProduced: wo.QuantityProduced,
				Timestamp:        time.Now(),
			}
			if err := s.emitEvent(txCtx, domain.TopicMfgWorkOrderCompleted, wo.ID, evt); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return wo, nil
}

func (s *WorkOrderExecutionServiceImpl) RerouteWorkOrderStation(ctx context.Context, workOrderID, currentStationID, targetStationID string, isRework bool) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		_, err := s.woRepo.GetByID(txCtx, workOrderID)
		if err != nil {
			return fmt.Errorf("work order not found: %w", err)
		}

		_, err = s.stationRepo.GetByID(txCtx, targetStationID)
		if err != nil {
			return fmt.Errorf("target station not found: %w", err)
		}

		activeState, err := s.stateRepo.GetActiveByWorkOrderID(txCtx, workOrderID)
		if err == nil && activeState != nil {
			now := time.Now()
			activeState.ExitedAt = &now
			if err := s.stateRepo.Update(txCtx, activeState); err != nil {
				return err
			}
		}

		newState := &domain.WorkOrderRoutingState{
			ID:                     utils.NewID("wors"),
			WorkOrderID:            workOrderID,
			CurrentStationID:       targetStationID,
			NextSuggestedStationID: nil,
			IsReworkLoop:           isRework,
			EnteredAt:              time.Now(),
			ExitedAt:               nil,
		}

		return s.stateRepo.Create(txCtx, newState)
	})
}

func (s *WorkOrderExecutionServiceImpl) FreezeObsoleteWorkOrders(ctx context.Context, materialID string, newBomHeaderID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		wos, err := s.woRepo.List(txCtx)
		if err != nil {
			return err
		}

		for _, wo := range wos {
			if wo.MaterialID == materialID && wo.BomHeaderID != newBomHeaderID &&
				(wo.Status == domain.WorkOrderStateSTAGED ||
					wo.Status == domain.WorkOrderStateRELEASED ||
					wo.Status == domain.WorkOrderStateIN_PROGRESS) {

				log.Printf("[Obsolete BOM Shield] Freezing Work Order %s (Material %s, Old BOM %s) due to release of new BOM %s",
					wo.ID, materialID, wo.BomHeaderID, newBomHeaderID)

				wo.Status = domain.WorkOrderStateON_HOLD
				wo.UpdatedAt = time.Now()

				if err := s.woRepo.Update(txCtx, &wo); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// ==========================================
// ShopFloorTelemetryService Implementation
// ==========================================

type ShopFloorTelemetryService interface {
	RecordBulkMaterialConsumption(ctx context.Context, legalEntityID, workOrderID string, lines []domain.ConsumptionSubmissionInput) error
	CommitProductionYield(ctx context.Context, legalEntityID, workOrderID, stationID string, qtyGood, qtyScrap decimal.Decimal, operatorHrID string) error
}

type ShopFloorTelemetryServiceImpl struct {
	db          *gorm.DB
	woRepo      domain.WorkOrderRepository
	stationRepo domain.RoutingStationRepository
	consumeRepo domain.MaterialConsumptionLogRepository
	yieldRepo   domain.ProductionYieldLogRepository
	outboxRepo  domain.TransactionalOutboxRepository
}

func NewShopFloorTelemetryService(
	db *gorm.DB,
	woRepo domain.WorkOrderRepository,
	stationRepo domain.RoutingStationRepository,
	consumeRepo domain.MaterialConsumptionLogRepository,
	yieldRepo domain.ProductionYieldLogRepository,
	outboxRepo domain.TransactionalOutboxRepository,
) ShopFloorTelemetryService {
	return &ShopFloorTelemetryServiceImpl{
		db:          db,
		woRepo:      woRepo,
		stationRepo: stationRepo,
		consumeRepo: consumeRepo,
		yieldRepo:   yieldRepo,
		outboxRepo:  outboxRepo,
	}
}

func (s *ShopFloorTelemetryServiceImpl) emitEvent(ctx context.Context, eventType, aggregateID string, payload interface{}) error {
	outbox := &domain.TransactionalOutbox{
		ID:          utils.NewID("outbox"),
		EventType:   eventType,
		AggregateID: aggregateID,
		Payload:     payload,
		Status:      domain.OutboxStatusPENDING,
		RetryCount:  0,
		CreatedAt:   time.Now(),
	}
	return s.outboxRepo.Create(ctx, outbox)
}

func (s *ShopFloorTelemetryServiceImpl) RecordBulkMaterialConsumption(ctx context.Context, legalEntityID, workOrderID string, lines []domain.ConsumptionSubmissionInput) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		_, err := s.woRepo.GetByID(txCtx, workOrderID)
		if err != nil {
			return fmt.Errorf("work order not found: %w", err)
		}

		payloadItems := make([]domain.ConsumedItemPayload, 0, len(lines))

		for _, line := range lines {
			_, err = s.stationRepo.GetByID(txCtx, line.RoutingStationID)
			if err != nil {
				return fmt.Errorf("routing station not found for ID %s: %w", line.RoutingStationID, err)
			}

			log := &domain.MaterialConsumptionLog{
				ID:               utils.NewID("mcl"),
				LegalEntityID:    legalEntityID,
				WorkOrderID:      workOrderID,
				MaterialID:       line.MaterialID,
				RoutingStationID: line.RoutingStationID,
				QuantityConsumed: line.QuantityConsumed,
				WarehouseID:      line.WarehouseID,
				OperatorHrID:     "system_operator",
				ConsumedAt:       time.Now(),
			}

			if err := s.consumeRepo.Create(txCtx, log); err != nil {
				return err
			}

			payloadItems = append(payloadItems, domain.ConsumedItemPayload{
				MaterialID:       line.MaterialID,
				QuantityDeducted: line.QuantityConsumed,
				WarehouseID:      line.WarehouseID,
			})
		}

		evt := domain.MfgMaterialConsumedEvent{
			EventID:       utils.NewID("evt"),
			LegalEntityID: legalEntityID,
			WorkOrderID:   workOrderID,
			Items:         payloadItems,
			Timestamp:     time.Now(),
		}

		return s.emitEvent(txCtx, domain.TopicMfgMaterialConsumed, workOrderID, evt)
	})
}

func (s *ShopFloorTelemetryServiceImpl) CommitProductionYield(ctx context.Context, legalEntityID, workOrderID, stationID string, qtyGood, qtyScrap decimal.Decimal, operatorHrID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		wo, err := s.woRepo.GetByID(txCtx, workOrderID)
		if err != nil {
			return fmt.Errorf("work order not found: %w", err)
		}

		_, err = s.stationRepo.GetByID(txCtx, stationID)
		if err != nil {
			return fmt.Errorf("routing station not found: %w", err)
		}

		log := &domain.ProductionYieldLog{
			ID:               utils.NewID("pyl"),
			LegalEntityID:    legalEntityID,
			WorkOrderID:      workOrderID,
			RoutingStationID: stationID,
			QuantityGood:     qtyGood,
			QuantityScrap:    qtyScrap,
			OperatorHrID:     operatorHrID,
			RecordedAt:       time.Now(),
		}

		if err := s.yieldRepo.Create(txCtx, log); err != nil {
			return err
		}

		wo.QuantityProduced = wo.QuantityProduced.Add(qtyGood)
		wo.UpdatedAt = time.Now()

		if err := s.woRepo.Update(txCtx, wo); err != nil {
			return err
		}

		evt := domain.MfgYieldProducedEvent{
			EventID:          utils.NewID("evt"),
			LegalEntityID:    legalEntityID,
			WorkOrderID:      workOrderID,
			RoutingStationID: stationID,
			QuantityGood:     qtyGood,
			QuantityScrap:    qtyScrap,
			OperatorHrID:     operatorHrID,
			Timestamp:        time.Now(),
		}

		return s.emitEvent(txCtx, domain.TopicMfgYieldProduced, workOrderID, evt)
	})
}

// ==========================================
// OutboxRelayWorker Implementation
// ==========================================

type OutboxRelayWorker interface {
	GetUnsentMessages(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error)
	LogProcessingAttempt(ctx context.Context, outboxID string, currentRetries int, errorNotes string) error
	UpdateOutboxStatus(ctx context.Context, outboxID string, status domain.OutboxStatus) error
}

type OutboxRelayWorkerImpl struct {
	outboxRepo domain.TransactionalOutboxRepository
}

func NewOutboxRelayWorker(outboxRepo domain.TransactionalOutboxRepository) OutboxRelayWorker {
	return &OutboxRelayWorkerImpl{outboxRepo: outboxRepo}
}

func (s *OutboxRelayWorkerImpl) GetUnsentMessages(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	return s.outboxRepo.GetUnsent(ctx, limit)
}

func (s *OutboxRelayWorkerImpl) LogProcessingAttempt(ctx context.Context, outboxID string, currentRetries int, errorNotes string) error {
	msg, err := s.outboxRepo.GetByID(ctx, outboxID)
	if err != nil {
		return err
	}
	msg.RetryCount = currentRetries + 1
	if msg.RetryCount >= 5 {
		msg.Status = domain.OutboxStatusFAILED
	}
	return s.outboxRepo.Update(ctx, msg)
}

func (s *OutboxRelayWorkerImpl) UpdateOutboxStatus(ctx context.Context, outboxID string, status domain.OutboxStatus) error {
	msg, err := s.outboxRepo.GetByID(ctx, outboxID)
	if err != nil {
		return err
	}
	msg.Status = status
	return s.outboxRepo.Update(ctx, msg)
}

// ==========================================
// ReliableMessagingService Implementation
// ==========================================

type ReliableMessagingService interface {
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
	ExecuteIdempotentTransaction(ctx context.Context, eventID string, eventType string, payload interface{}, businessRoutine func(ctx context.Context) error) error
}

type ReliableMessagingServiceImpl struct {
	db        *gorm.DB
	inboxRepo domain.KafkaEventInboxRepository
}

func NewReliableMessagingService(db *gorm.DB, inboxRepo domain.KafkaEventInboxRepository) ReliableMessagingService {
	return &ReliableMessagingServiceImpl{
		db:        db,
		inboxRepo: inboxRepo,
	}
}

func (s *ReliableMessagingServiceImpl) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	msg, err := s.inboxRepo.GetByID(ctx, eventID)
	if err == nil && msg != nil {
		return msg.ProcessingStatus == domain.EventProcessingStatusSUCCESS, nil
	}
	return false, nil
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
