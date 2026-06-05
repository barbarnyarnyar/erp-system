package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type ProductionService struct {
	poRepo      domain.ProductionOrderRepository
	woRepo      domain.WorkOrderRepository
	bomRepo     domain.BillOfMaterialsRepository
	compRepo    domain.BOMComponentRepository
	routingRepo domain.RoutingOperationRepository
	wcRepo      domain.WorkCenterRepository
	laborRepo   domain.LaborReportRepository
	machineRepo domain.MachineLogRepository
	qualityRepo domain.QualityInspectionRepository
	nonConfRepo domain.NonConformanceRepository
	equipRepo   domain.EquipmentRepository
	maintRepo   domain.MaintenanceOrderRepository
	costRepo    domain.CostingRecordRepository
	publisher   domain.EventPublisher
}

func NewProductionService(
	poRepo domain.ProductionOrderRepository,
	woRepo domain.WorkOrderRepository,
	bomRepo domain.BillOfMaterialsRepository,
	compRepo domain.BOMComponentRepository,
	routingRepo domain.RoutingOperationRepository,
	wcRepo domain.WorkCenterRepository,
	laborRepo domain.LaborReportRepository,
	machineRepo domain.MachineLogRepository,
	qualityRepo domain.QualityInspectionRepository,
	nonConfRepo domain.NonConformanceRepository,
	equipRepo domain.EquipmentRepository,
	maintRepo domain.MaintenanceOrderRepository,
	costRepo domain.CostingRecordRepository,
	publisher domain.EventPublisher,
) *ProductionService {
	return &ProductionService{
		poRepo:      poRepo,
		woRepo:      woRepo,
		bomRepo:     bomRepo,
		compRepo:    compRepo,
		routingRepo: routingRepo,
		wcRepo:      wcRepo,
		laborRepo:   laborRepo,
		machineRepo: machineRepo,
		qualityRepo: qualityRepo,
		nonConfRepo: nonConfRepo,
		equipRepo:   equipRepo,
		maintRepo:   maintRepo,
		costRepo:    costRepo,
		publisher:   publisher,
	}
}

func (s *ProductionService) CreateProductionOrder(ctx context.Context, bomID string, quantity int, scheduledDate time.Time) (*domain.ProductionOrder, error) {
	bom, err := s.bomRepo.GetByID(ctx, bomID)
	if err != nil {
		return nil, err
	}

	poID := fmt.Sprintf("po_%d", time.Now().UnixNano())
	po := &domain.ProductionOrder{
		ID:            poID,
		BomID:         bomID,
		ProductID:     bom.ProductID,
		Quantity:      quantity,
		Status:        "PLANNED",
		ScheduledDate: scheduledDate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.poRepo.Create(ctx, po)
	if err != nil {
		return nil, err
	}

	// 1. MRP: Generate Material Required Events
	components, _ := s.compRepo.ListByBOMID(ctx, bomID)
	for _, comp := range components {
		qtyNeeded := decimal.NewFromInt(int64(quantity)).Mul(comp.Quantity).Mul(decimal.NewFromFloat(1.0).Add(comp.WasteFactor))
		_ = s.publisher.Publish(ctx, domain.TopicMfgMaterialRequired, comp.ComponentProductID, domain.MaterialRequiredEvent{
			ProductID:  comp.ComponentProductID,
			Quantity:   qtyNeeded,
			RequiredBy: scheduledDate,
			Timestamp:  time.Now(),
		})
	}

	// 2. Capacity Planning & Work Order Scheduling
	operations, _ := s.routingRepo.ListByBOMID(ctx, bomID)
	for _, op := range operations {
		woID := fmt.Sprintf("wo_%d", time.Now().UnixNano())
		wo := &domain.WorkOrder{
			ID:                  woID,
			ProductionOrderID:   poID,
			SequenceNumber:      op.SequenceNumber,
			WorkCenterID:        op.WorkCenterID,
			ScheduledStart:      scheduledDate,
			ScheduledEnd:        scheduledDate.Add(2 * time.Hour),
			Status:              "PENDING",
			LaborHours:          &decimal.Zero,
			MachineHours:        &decimal.Zero,
		}
		_ = s.woRepo.Create(ctx, wo)

		// Publish Work Order Created Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgWorkOrderCreated, wo.ID, domain.WorkOrderCreatedEvent{
			WorkOrderID:       wo.ID,
			ProductionOrderID: poID,
			WorkCenterID:      op.WorkCenterID,
			Timestamp:         time.Now(),
		})
	}

	// 3. Publish Production Scheduled Event
	_ = s.publisher.Publish(ctx, domain.TopicMfgProductionScheduled, po.ID, domain.ProductionScheduledEvent{
		ProductionOrderID: po.ID,
		ProductID:         po.ProductID,
		Quantity:          po.Quantity,
		ScheduledDate:     po.ScheduledDate,
		Timestamp:         time.Now(),
	})

	return po, nil
}

func (s *ProductionService) StartWorkOrder(ctx context.Context, id string) (*domain.WorkOrder, error) {
	wo, err := s.woRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	wo.Status = "IN_PROGRESS"
	wo.ActualStart = &now
	_ = s.woRepo.Update(ctx, wo)

	// Publish Work Order Started Event
	_ = s.publisher.Publish(ctx, domain.TopicMfgWorkOrderStarted, wo.ID, domain.WorkOrderStartedEvent{
		WorkOrderID: wo.ID,
		Timestamp:   now,
	})

	po, err := s.poRepo.GetByID(ctx, wo.ProductionOrderID)
	if err == nil && po.Status == "PLANNED" {
		po.Status = "IN_PROGRESS"
		po.StartDate = &now
		po.UpdatedAt = now
		_ = s.poRepo.Update(ctx, po)

		// Publish Production Started Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgProductionStarted, po.ID, domain.ProductionStartedEvent{
			ProductionOrderID: po.ID,
			ProductID:         po.ProductID,
			Timestamp:         now,
		})
	}

	return wo, nil
}

func (s *ProductionService) ReportLabor(ctx context.Context, workOrderID string, employeeID string, hours decimal.Decimal) (*domain.LaborReport, error) {
	wo, err := s.woRepo.GetByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("lab_%d", time.Now().UnixNano())
	lr := &domain.LaborReport{
		ID:           id,
		WorkOrderID:  workOrderID,
		EmployeeID:   employeeID,
		HoursWorked:  hours,
		Date:         time.Now(),
	}

	err = s.laborRepo.Create(ctx, lr)
	if err != nil {
		return nil, err
	}

	currHours := decimal.Zero
	if wo.LaborHours != nil {
		currHours = *wo.LaborHours
	}
	newHours := currHours.Add(hours)
	wo.LaborHours = &newHours
	_ = s.woRepo.Update(ctx, wo)

	return lr, nil
}

func (s *ProductionService) LogMachineStatus(ctx context.Context, workCenterID string, statusCode string, message string) (*domain.MachineLog, error) {
	id := fmt.Sprintf("ml_%d", time.Now().UnixNano())
	ml := &domain.MachineLog{
		ID:             id,
		WorkCenterID:   workCenterID,
		StatusCode:     statusCode,
		Message:        message,
		Timestamp:      time.Now(),
	}

	err := s.machineRepo.Create(ctx, ml)
	if err != nil {
		return nil, err
	}
	return ml, nil
}

func (s *ProductionService) CompleteWorkOrder(ctx context.Context, id string) (*domain.WorkOrder, error) {
	wo, err := s.woRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	wo.Status = "COMPLETED"
	wo.ActualEnd = &now
	_ = s.woRepo.Update(ctx, wo)

	// Publish Work Order Completed Event
	_ = s.publisher.Publish(ctx, domain.TopicMfgWorkOrderCompleted, wo.ID, domain.WorkOrderCompletedEvent{
		WorkOrderID: wo.ID,
		Timestamp:   now,
	})

	return wo, nil
}

func (s *ProductionService) RecordQualityInspection(ctx context.Context, workOrderID string, inspectorID string, result string, remarks string) (*domain.QualityInspection, error) {
	id := fmt.Sprintf("qi_%d", time.Now().UnixNano())
	qi := &domain.QualityInspection{
		ID:            id,
		WorkOrderID:   workOrderID,
		InspectorID:   inspectorID,
		Result:        result,
		Remarks:       remarks,
		InspectedAt:   time.Now(),
	}

	err := s.qualityRepo.Create(ctx, qi)
	if err != nil {
		return nil, err
	}

	if result == "FAIL" {
		ncID := fmt.Sprintf("nc_%d", time.Now().UnixNano())
		nc := &domain.NonConformance{
			ID:            ncID,
			InspectionID:  id,
			Description:   fmt.Sprintf("Failed quality inspection: %s", remarks),
			Severity:      "MEDIUM",
			Status:        "OPEN",
		}
		_ = s.nonConfRepo.Create(ctx, nc)

		// Publish Quality Inspection Failed Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgQualityInspectionFailed, id, domain.QualityInspectionFailedEvent{
			InspectionID: id,
			WorkOrderID:  workOrderID,
			InspectorID:  inspectorID,
			Remarks:      remarks,
			Timestamp:    time.Now(),
		})

		// Publish Non-Conformance Detected Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgQualityNonConformanceDetected, ncID, domain.QualityNonConformanceDetectedEvent{
			NonConformanceID: ncID,
			InspectionID:     id,
			Severity:         "MEDIUM",
			Description:      nc.Description,
			Timestamp:        time.Now(),
		})

		// Publish Material Wasted Event (using mock quantities/reasons)
		wo, _ := s.woRepo.GetByID(ctx, workOrderID)
		_ = s.publisher.Publish(ctx, domain.TopicMfgMaterialWasted, workOrderID, domain.MaterialWastedEvent{
			ProductionOrderID: wo.ProductionOrderID,
			ProductID:         "component_wasted",
			Quantity:          decimal.NewFromFloat(1.0), // Mock standard wasted quantity
			Reason:            fmt.Sprintf("Inspection Fail: %s", remarks),
			Timestamp:         time.Now(),
		})

	} else if result == "PASS" {
		// Publish Quality Inspection Passed Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgQualityInspectionPassed, id, domain.QualityInspectionPassedEvent{
			InspectionID: id,
			WorkOrderID:  workOrderID,
			InspectorID:  inspectorID,
			Timestamp:    time.Now(),
		})

		wo, _ := s.woRepo.GetByID(ctx, workOrderID)
		workOrders, _ := s.woRepo.ListByProductionOrderID(ctx, wo.ProductionOrderID)
		allComplete := true
		for _, w := range workOrders {
			if w.Status != "COMPLETED" {
				allComplete = false
				break
			}
			ins, err := s.qualityRepo.GetByWorkOrderID(ctx, w.ID)
			if err != nil || ins.Result != "PASS" {
				allComplete = false
				break
			}
		}

		if allComplete {
			_, _ = s.CompleteProductionOrder(ctx, wo.ProductionOrderID)
		}
	}

	return qi, nil
}

func (s *ProductionService) CompleteProductionOrder(ctx context.Context, id string) (*domain.ProductionOrder, error) {
	po, err := s.poRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	po.Status = "COMPLETED"
	po.EndDate = &now
	po.UpdatedAt = now

	err = s.poRepo.Update(ctx, po)
	if err != nil {
		return nil, err
	}

	stdMatCost := decimal.NewFromFloat(50.0).Mul(decimal.NewFromInt(int64(po.Quantity)))
	actualMatCost := stdMatCost

	stdLaborCost := decimal.Zero
	operations, _ := s.routingRepo.ListByBOMID(ctx, po.BomID)
	for _, op := range operations {
		wc, err := s.wcRepo.GetByID(ctx, op.WorkCenterID)
		if err == nil {
			stdMinutes := op.SetupTime.Add(op.RunTime.Mul(decimal.NewFromInt(int64(po.Quantity))))
			stdHours := stdMinutes.Div(decimal.NewFromInt(60))
			stdLaborCost = stdLaborCost.Add(stdHours.Mul(wc.HourlyRate))
		}
	}

	actualLaborCost := decimal.Zero
	workOrders, _ := s.woRepo.ListByProductionOrderID(ctx, po.ID)
	for _, wo := range workOrders {
		wc, err := s.wcRepo.GetByID(ctx, wo.WorkCenterID)
		if err == nil && wo.LaborHours != nil {
			actualLaborCost = actualLaborCost.Add(wo.LaborHours.Mul(wc.HourlyRate))
		}
	}

	laborVar := actualLaborCost.Sub(stdLaborCost)
	matVar := actualMatCost.Sub(stdMatCost)

	costRecordID := fmt.Sprintf("cost_%d", time.Now().UnixNano())
	cr := &domain.CostingRecord{
		ID:                   costRecordID,
		ProductionOrderID:    po.ID,
		StandardMaterialCost: stdMatCost,
		ActualMaterialCost:   actualMatCost,
		StandardLaborCost:    stdLaborCost,
		ActualLaborCost:      actualLaborCost,
		MaterialVariance:     matVar,
		LaborVariance:        laborVar,
	}
	_ = s.costRepo.Create(ctx, cr)

	_ = s.publisher.Publish(ctx, domain.TopicMfgMaterialConsumed, po.ID, domain.MaterialConsumedEvent{
		ProductionOrderID: po.ID,
		ProductID:         po.ProductID,
		Quantity:          decimal.NewFromInt(int64(po.Quantity)),
		Timestamp:         time.Now(),
	})

	_ = s.publisher.Publish(ctx, domain.TopicMfgProductionCompleted, po.ID, domain.ProductionCompletedEvent{
		ProductionOrderID: po.ID,
		ProductID:         po.ProductID,
		Quantity:          po.Quantity,
		Timestamp:         time.Now(),
	})

	return po, nil
}

func (s *ProductionService) CreateEquipment(ctx context.Context, workCenterID string, name string) (*domain.Equipment, error) {
	id := fmt.Sprintf("eq_%d", time.Now().UnixNano())
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

func (s *ProductionService) ScheduleMaintenance(ctx context.Context, equipmentID string, description string, maintType string) (*domain.MaintenanceOrder, error) {
	id := fmt.Sprintf("mo_%d", time.Now().UnixNano())
	mo := &domain.MaintenanceOrder{
		ID:               id,
		EquipmentID:      equipmentID,
		Description:      description,
		Status:           "SCHEDULED",
		MaintenanceType:  maintType,
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
	_ = s.publisher.Publish(ctx, domain.TopicMfgMaintenanceScheduled, mo.ID, domain.MaintenanceScheduledEvent{
		MaintenanceOrderID: mo.ID,
		EquipmentID:        equipmentID,
		ScheduledDate:      time.Now(), // Mock scheduled date
		Timestamp:          time.Now(),
	})

	// Publish Equipment Down Event
	if eq != nil {
		_ = s.publisher.Publish(ctx, domain.TopicMfgEquipmentDown, equipmentID, domain.EquipmentDownEvent{
			EquipmentID:  equipmentID,
			WorkCenterID: eq.WorkCenterID,
			Reason:       description,
			Timestamp:    time.Now(),
		})
	}

	return mo, nil
}

func (s *ProductionService) CompleteMaintenance(ctx context.Context, id string) (*domain.MaintenanceOrder, error) {
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
	_ = s.publisher.Publish(ctx, domain.TopicMfgMaintenanceCompleted, mo.ID, domain.MaintenanceCompletedEvent{
		MaintenanceOrderID: mo.ID,
		EquipmentID:        mo.EquipmentID,
		Timestamp:          now,
	})

	// Publish Equipment Up Event
	if eq != nil {
		_ = s.publisher.Publish(ctx, domain.TopicMfgEquipmentUp, mo.EquipmentID, domain.EquipmentUpEvent{
			EquipmentID:  mo.EquipmentID,
			WorkCenterID: eq.WorkCenterID,
			Timestamp:    now,
		})
	}

	return mo, nil
}

func (s *ProductionService) GetCostingRecord(ctx context.Context, poID string) (*domain.CostingRecord, error) {
	return s.costRepo.GetByProductionOrderID(ctx, poID)
}

func (s *ProductionService) ListProductionPlans(ctx context.Context) ([]domain.ProductionOrder, error) {
	return s.poRepo.List(ctx)
}

func (s *ProductionService) GetProductionPlan(ctx context.Context, id string) (*domain.ProductionOrder, error) {
	return s.poRepo.GetByID(ctx, id)
}

func (s *ProductionService) UpdateProductionPlan(ctx context.Context, id string, quantity int, scheduledDate time.Time, status string) (*domain.ProductionOrder, error) {
	po, err := s.poRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldDate := po.ScheduledDate
	po.Quantity = quantity
	po.ScheduledDate = scheduledDate
	po.Status = status
	po.UpdatedAt = time.Now()

	err = s.poRepo.Update(ctx, po)
	if err != nil {
		return nil, err
	}

	if status == "DELAYED" || scheduledDate.After(oldDate) {
		_ = s.publisher.Publish(ctx, domain.TopicMfgProductionDelayed, po.ID, domain.ProductionDelayedEvent{
			ProductionOrderID: po.ID,
			Reason:            fmt.Sprintf("Schedule updated to %s (Status: %s)", scheduledDate.Format(time.RFC3339), status),
			NewScheduledDate:  scheduledDate,
			Timestamp:         time.Now(),
		})
	}

	return po, nil
}

func (s *ProductionService) DeleteProductionPlan(ctx context.Context, id string) error {
	return s.poRepo.Delete(ctx, id)
}

func (s *ProductionService) RunMRP(ctx context.Context) error {
	orders, err := s.poRepo.List(ctx)
	if err != nil {
		return err
	}

	for _, po := range orders {
		if po.Status == "PLANNED" {
			components, _ := s.compRepo.ListByBOMID(ctx, po.BomID)
			for _, comp := range components {
				qtyNeeded := decimal.NewFromInt(int64(po.Quantity)).Mul(comp.Quantity).Mul(decimal.NewFromFloat(1.0).Add(comp.WasteFactor))
				_ = s.publisher.Publish(ctx, domain.TopicMfgMaterialRequired, comp.ComponentProductID, domain.MaterialRequiredEvent{
					ProductID:  comp.ComponentProductID,
					Quantity:   qtyNeeded,
					RequiredBy: po.ScheduledDate,
					Timestamp:  time.Now(),
				})
			}
		}
	}
	return nil
}

func (s *ProductionService) ListWorkOrders(ctx context.Context) ([]domain.WorkOrder, error) {
	return s.woRepo.List(ctx)
}

func (s *ProductionService) CreateWorkOrder(ctx context.Context, poID string, seqNum int, workCenterID string, start, end time.Time) (*domain.WorkOrder, error) {
	id := fmt.Sprintf("wo_%d", time.Now().UnixNano())
	wo := &domain.WorkOrder{
		ID:                  id,
		ProductionOrderID:   poID,
		SequenceNumber:      seqNum,
		WorkCenterID:        workCenterID,
		ScheduledStart:      start,
		ScheduledEnd:        end,
		Status:              "PENDING",
		LaborHours:          &decimal.Zero,
		MachineHours:        &decimal.Zero,
	}
	err := s.woRepo.Create(ctx, wo)
	if err != nil {
		return nil, err
	}
	return wo, nil
}

func (s *ProductionService) GetWorkOrder(ctx context.Context, id string) (*domain.WorkOrder, error) {
	return s.woRepo.GetByID(ctx, id)
}

func (s *ProductionService) UpdateWorkOrder(ctx context.Context, id string, status string, start, end time.Time, actStart, actEnd *time.Time) (*domain.WorkOrder, error) {
	wo, err := s.woRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	wo.Status = status
	wo.ScheduledStart = start
	wo.ScheduledEnd = end
	wo.ActualStart = actStart
	wo.ActualEnd = actEnd

	err = s.woRepo.Update(ctx, wo)
	if err != nil {
		return nil, err
	}
	return wo, nil
}

func (s *ProductionService) DeleteWorkOrder(ctx context.Context, id string) error {
	_ = s.publisher.Publish(ctx, domain.TopicMfgWorkOrderCancelled, id, domain.WorkOrderCancelledEvent{
		WorkOrderID: id,
		Reason:      "Manual deletion request",
		Timestamp:   time.Now(),
	})
	return s.woRepo.Delete(ctx, id)
}

func (s *ProductionService) ListQualityInspections(ctx context.Context) ([]domain.QualityInspection, error) {
	return s.qualityRepo.List(ctx)
}

func (s *ProductionService) GetQualityInspection(ctx context.Context, id string) (*domain.QualityInspection, error) {
	return s.qualityRepo.GetByID(ctx, id)
}

func (s *ProductionService) UpdateQualityInspection(ctx context.Context, id string, result string, remarks string) (*domain.QualityInspection, error) {
	qi, err := s.qualityRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	qi.Result = result
	qi.Remarks = remarks

	err = s.qualityRepo.Update(ctx, qi)
	if err != nil {
		return nil, err
	}
	return qi, nil
}

func (s *ProductionService) ListMaintenanceSchedules(ctx context.Context) ([]domain.MaintenanceOrder, error) {
	return s.maintRepo.List(ctx)
}

func (s *ProductionService) GetMaintenanceSchedule(ctx context.Context, id string) (*domain.MaintenanceOrder, error) {
	return s.maintRepo.GetByID(ctx, id)
}

func (s *ProductionService) UpdateMaintenanceSchedule(ctx context.Context, id string, status string, completedAt *time.Time) (*domain.MaintenanceOrder, error) {
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
