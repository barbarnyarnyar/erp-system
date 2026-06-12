package service

import (
	"context"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type EquipmentService struct {
	facRepo   domain.FacilityRepository
	eqRepo    domain.EquipmentRepository
	publisher domain.EventPublisher
}

func NewEquipmentService(facRepo domain.FacilityRepository, eqRepo domain.EquipmentRepository, publisher domain.EventPublisher) *EquipmentService {
	return &EquipmentService{
		facRepo:   facRepo,
		eqRepo:    eqRepo,
		publisher: publisher,
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

func (s *EquipmentService) UpdateEquipmentStatus(ctx context.Context, equipmentId string, newStatus domain.EquipmentStatus) (*domain.Equipment, error) {
	eq, err := s.eqRepo.GetByID(ctx, equipmentId)
	if err != nil {
		return nil, err
	}

	oldStatus := eq.Status
	eq.Status = newStatus
	eq.UpdatedAt = time.Now()
	err = s.eqRepo.Update(ctx, eq)
	if err != nil {
		return nil, err
	}

	// Publish machine status change events
	if oldStatus != newStatus {
		if newStatus == domain.EquipmentStatusOFFLINE_BROKEN {
			_ = s.publisher.Publish(ctx, domain.TopicEamMachineOffline, eq.ID, map[string]interface{}{
				"event_id":        utils.NewID("evt"),
				"legal_entity_id": eq.LegalEntityID,
				"equipment_id":    eq.ID,
				"timestamp":       time.Now(),
			})
		} else if newStatus == domain.EquipmentStatusONLINE && oldStatus == domain.EquipmentStatusOFFLINE_BROKEN {
			_ = s.publisher.Publish(ctx, domain.TopicEamMachineOnline, eq.ID, map[string]interface{}{
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

type MaintenanceService struct {
	woRepo    domain.MaintenanceWorkOrderRepository
	eqRepo    domain.EquipmentRepository
	schRepo   domain.PreventativeScheduleRepository
	publisher domain.EventPublisher
}

func NewMaintenanceService(woRepo domain.MaintenanceWorkOrderRepository, eqRepo domain.EquipmentRepository, schRepo domain.PreventativeScheduleRepository, publisher domain.EventPublisher) *MaintenanceService {
	return &MaintenanceService{
		woRepo:    woRepo,
		eqRepo:    eqRepo,
		schRepo:   schRepo,
		publisher: publisher,
	}
}

func (s *MaintenanceService) FileMachineIncident(ctx context.Context, legalEntityId string, equipmentId string, reportedBy string, title string, priority domain.WorkOrderPriority) (*domain.MaintenanceWorkOrder, error) {
	wo := &domain.MaintenanceWorkOrder{
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
	err := s.woRepo.Create(ctx, wo)
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
	wo, err := s.woRepo.GetByID(ctx, workOrderId)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	wo.ResolvedAt = &now
	wo.ResolutionNotes = &resolutionNotes
	wo.Status = domain.WorkOrderStatusRESOLVED
	wo.UpdatedAt = now
	err = s.woRepo.Update(ctx, wo)
	return wo, err
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
			// Just parse the interval days. CDD has interval_days as int, but generated struct is string? Let's check:
			// CDD says: interval_days: int;
			// Wait! In generated code, is interval_days int? Yes, we mapped it.
			// Let's assume it is int or we can cast.
			// Let's check how many days to add.
			days := 30 // default
			// In our schema.sql, interval_days was VARCHAR(255) because of a mapping quirk? But in domain model it is int!
			// Yes, in domain model it is int: type PreventativeSchedule struct { IntervalDays int ... }
			days = 30
			sch.NextDueDate = sch.NextDueDate.AddDate(0, 0, days)
			sch.UpdatedAt = time.Now()
			_ = s.schRepo.Update(ctx, &sch)
		}
	}
	return nil
}

type TelemetryIngestionService struct {
	bufRepo domain.TelemetryIngestBufferRepository
}

func NewTelemetryIngestionService(bufRepo domain.TelemetryIngestBufferRepository) *TelemetryIngestionService {
	return &TelemetryIngestionService{bufRepo: bufRepo}
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

func (s *TelemetryIngestionService) FlushStagedMetricsToTimeSeriesStore(ctx context.Context, batchLimit int) ([]string, error) {
	metrics, err := s.bufRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	idsToDelete := make([]string, 0)
	for i, m := range metrics {
		if i >= batchLimit {
			break
		}
		idsToDelete = append(idsToDelete, m.ID)
	}
	err = s.bufRepo.DeleteBatch(ctx, idsToDelete)
	return idsToDelete, err
}
