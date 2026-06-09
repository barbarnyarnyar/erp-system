package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
)

type MaintenanceService struct {
	machineRepo domain.MachineLogRepository
	equipRepo   domain.EquipmentRepository
	maintRepo   domain.MaintenanceOrderRepository
	publisher   domain.EventPublisher
}

func NewMaintenanceService(
	machineRepo domain.MachineLogRepository,
	equipRepo domain.EquipmentRepository,
	maintRepo domain.MaintenanceOrderRepository,
	publisher domain.EventPublisher,
) *MaintenanceService {
	return &MaintenanceService{
		machineRepo: machineRepo,
		equipRepo:   equipRepo,
		maintRepo:   maintRepo,
		publisher:   publisher,
	}
}

func (s *MaintenanceService) LogMachineStatus(ctx context.Context, workCenterID string, statusCode string, message string, severity string) (*domain.MachineLog, error) {
	id := utils.NewID("ml")
	ml := &domain.MachineLog{
		ID:           id,
		WorkCenterID: workCenterID,
		StatusCode:   statusCode,
		Message:      message,
		Severity:     severity,
		Timestamp:    time.Now(),
	}

	err := s.machineRepo.Create(ctx, ml)
	if err != nil {
		return nil, err
	}
	return ml, nil
}

func (s *MaintenanceService) CreateEquipment(ctx context.Context, workCenterID string, name string) (*domain.Equipment, error) {
	id := utils.NewID("eq")
	eq := &domain.Equipment{
		ID:           id,
		WorkCenterID: workCenterID,
		Name:         name,
		Status:       "OPERATIONAL",
	}

	err := s.equipRepo.Create(ctx, eq)
	if err != nil {
		return nil, err
	}
	return eq, nil
}

func (s *MaintenanceService) ScheduleMaintenance(ctx context.Context, equipmentID string, description string, maintType string) (*domain.MaintenanceOrder, error) {
	id := utils.NewID("mo")
	mo := &domain.MaintenanceOrder{
		ID:              id,
		EquipmentID:     equipmentID,
		Description:     description,
		Status:          "SCHEDULED",
		MaintenanceType: maintType,
	}

	err := s.maintRepo.Create(ctx, mo)
	if err != nil {
		return nil, err
	}

	eq, err := s.equipRepo.GetByID(ctx, equipmentID)
	if err == nil {
		eq.Status = "UNDER_MAINTENANCE"
		_ = s.equipRepo.Update(ctx, eq)
	}

	// Publish Maintenance Scheduled Event
	if err := s.publisher.Publish(ctx, domain.TopicMfgMaintenanceScheduled, mo.ID, domain.MaintenanceScheduledEvent{
		MaintenanceOrderID: mo.ID,
		EquipmentID:        equipmentID,
		ScheduledDate:      time.Now(), // Mock scheduled date
		Timestamp:          time.Now(),
	}); err != nil {
		utils.LogPublishErr("m-service", domain.TopicMfgMaintenanceScheduled, err)
	}

	// Publish Equipment Down Event
	if eq != nil {
		if err := s.publisher.Publish(ctx, domain.TopicMfgEquipmentDown, equipmentID, domain.EquipmentDownEvent{
			EquipmentID:  equipmentID,
			WorkCenterID: eq.WorkCenterID,
			Reason:       description,
			Timestamp:    time.Now(),
		}); err != nil {
			utils.LogPublishErr("m-service", domain.TopicMfgEquipmentDown, err)
		}
	}

	return mo, nil
}

func (s *MaintenanceService) CompleteMaintenance(ctx context.Context, id string) (*domain.MaintenanceOrder, error) {
	mo, err := s.maintRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	mo.Status = "COMPLETED"
	mo.CompletedAt = &now
	_ = s.maintRepo.Update(ctx, mo)

	eq, err := s.equipRepo.GetByID(ctx, mo.EquipmentID)
	if err == nil {
		eq.Status = "OPERATIONAL"
		eq.LastMaintenance = &now
		nextM := now.AddDate(0, 3, 0)
		eq.NextMaintenance = &nextM
		_ = s.equipRepo.Update(ctx, eq)
	}

	// Publish Maintenance Completed Event
	if err := s.publisher.Publish(ctx, domain.TopicMfgMaintenanceCompleted, mo.ID, domain.MaintenanceCompletedEvent{
		MaintenanceOrderID: mo.ID,
		EquipmentID:        mo.EquipmentID,
		Timestamp:          now,
	}); err != nil {
		utils.LogPublishErr("m-service", domain.TopicMfgMaintenanceCompleted, err)
	}

	// Publish Equipment Up Event
	if eq != nil {
		if err := s.publisher.Publish(ctx, domain.TopicMfgEquipmentUp, mo.EquipmentID, domain.EquipmentUpEvent{
			EquipmentID:  mo.EquipmentID,
			WorkCenterID: eq.WorkCenterID,
			Timestamp:    now,
		}); err != nil {
			utils.LogPublishErr("m-service", domain.TopicMfgEquipmentUp, err)
		}
	}

	return mo, nil
}

func (s *MaintenanceService) ListMaintenanceSchedules(ctx context.Context) ([]domain.MaintenanceOrder, error) {
	return s.maintRepo.List(ctx)
}

func (s *MaintenanceService) GetMaintenanceSchedule(ctx context.Context, id string) (*domain.MaintenanceOrder, error) {
	return s.maintRepo.GetByID(ctx, id)
}

func (s *MaintenanceService) UpdateMaintenanceSchedule(ctx context.Context, id string, status string, completedAt *time.Time) (*domain.MaintenanceOrder, error) {
	mo, err := s.maintRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	mo.Status = status
	mo.CompletedAt = completedAt

	err = s.maintRepo.Update(ctx, mo)
	if err != nil {
		return nil, err
	}
	return mo, nil
}
