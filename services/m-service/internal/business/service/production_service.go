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
		costRepo:    costRepo,
		publisher:   publisher,
	}
}

func (s *ProductionService) CreateProductionOrder(ctx context.Context, bomID string, quantity int, scheduledDate time.Time, salesOrderID string) (*domain.ProductionOrder, error) {
	bom, err := s.bomRepo.GetByID(ctx, bomID)
	if err != nil {
		return nil, err
	}

	var salesOrderPtr *string
	if salesOrderID != "" {
		salesOrderPtr = &salesOrderID
	}

	poID := fmt.Sprintf("po_%d", time.Now().UnixNano())
	po := &domain.ProductionOrder{
		ID:            poID,
		BomID:         bomID,
		ProductID:     bom.ProductID,
		Quantity:      quantity,
		Status:        "PLANNED",
		ScheduledDate: scheduledDate,
		SalesOrderID:  salesOrderPtr,
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

		// Consume materials from SCM inventory when production starts
		_ = s.ConsumeMaterials(ctx, po.ID)

		// Publish Production Started Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgProductionStarted, po.ID, domain.ProductionStartedEvent{
			ProductionOrderID: po.ID,
			ProductID:         po.ProductID,
			Timestamp:         now,
		})
	}

	return wo, nil
}

func (s *ProductionService) ConsumeMaterials(ctx context.Context, productionOrderID string) error {
	po, err := s.poRepo.GetByID(ctx, productionOrderID)
	if err != nil {
		return err
	}

	components, err := s.compRepo.ListByBOMID(ctx, po.BomID)
	if err != nil {
		return err
	}

	for _, comp := range components {
		qtyNeeded := decimal.NewFromInt(int64(po.Quantity)).Mul(comp.Quantity).Mul(decimal.NewFromFloat(1.0).Add(comp.WasteFactor))
		
		err = s.publisher.Publish(ctx, domain.TopicMfgMaterialConsumed, po.ID, domain.MaterialConsumedEvent{
			ProductionOrderID: po.ID,
			ProductID:         comp.ComponentProductID,
			Quantity:          qtyNeeded,
			Timestamp:         time.Now(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ProductionService) ReceiveFinishedGoods(ctx context.Context, productionOrderID string) error {
	po, err := s.poRepo.GetByID(ctx, productionOrderID)
	if err != nil {
		return err
	}

	err = s.publisher.Publish(ctx, domain.TopicMfgProductionCompleted, po.ID, domain.ProductionCompletedEvent{
		ProductionOrderID: po.ID,
		ProductID:         po.ProductID,
		Quantity:          po.Quantity,
		Timestamp:         time.Now(),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *ProductionService) ReportLabor(ctx context.Context, workOrderID string, employeeID string, hours decimal.Decimal) (*domain.LaborReport, error) {
	wo, err := s.woRepo.GetByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	// Mock employee hourly rate from HR (snapshotting rate at time of report)
	hourlyRate := decimal.NewFromFloat(25.00)
	if len(employeeID) > 0 {
		hash := 0
		for _, char := range employeeID {
			hash += int(char)
		}
		// Generate varied rates between $20 and $59 based on employee ID hash
		hourlyRate = decimal.NewFromFloat(20.0 + float64(hash%40))
	}
	totalCost := hourlyRate.Mul(hours)

	id := fmt.Sprintf("lab_%d", time.Now().UnixNano())
	lr := &domain.LaborReport{
		ID:           id,
		WorkOrderID:  workOrderID,
		EmployeeID:   employeeID,
		HoursWorked:  hours,
		HourlyRate:   hourlyRate,
		TotalCost:    totalCost,
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

	// Costing calculations
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
		reports, err := s.laborRepo.ListByWorkOrderID(ctx, wo.ID)
		if err == nil {
			for _, lr := range reports {
				actualLaborCost = actualLaborCost.Add(lr.TotalCost)
			}
		}
	}

	laborVar := actualLaborCost.Sub(stdLaborCost)
	matVar := actualMatCost.Sub(stdMatCost)

	// Calculate Overhead costs (overhead is typically factory rent, utilities, depreciation)
	// Standard Overhead is calculated as 25% of standard labor cost
	stdOverheadCost := stdLaborCost.Mul(decimal.NewFromFloat(0.25))
	// Actual Overhead is calculated as 25% of actual labor cost
	actualOverheadCost := actualLaborCost.Mul(decimal.NewFromFloat(0.25))
	overheadVar := actualOverheadCost.Sub(stdOverheadCost)

	costRecordID := fmt.Sprintf("cost_%d", time.Now().UnixNano())
	cr := &domain.CostingRecord{
		ID:                   costRecordID,
		ProductionOrderID:    po.ID,
		StandardMaterialCost: stdMatCost,
		ActualMaterialCost:   actualMatCost,
		StandardLaborCost:    stdLaborCost,
		ActualLaborCost:      actualLaborCost,
		OverheadCost:         actualOverheadCost,
		MaterialVariance:     matVar,
		LaborVariance:        laborVar,
		OverheadVariance:     overheadVar,
	}
	_ = s.costRepo.Create(ctx, cr)

	// Trigger finished goods receipt into SCM inventory
	_ = s.ReceiveFinishedGoods(ctx, po.ID)

	return po, nil
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
