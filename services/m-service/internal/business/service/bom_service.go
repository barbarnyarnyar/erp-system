package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type BOMComponentInput struct {
	ComponentProductID string          `json:"component_product_id"`
	Quantity           decimal.Decimal `json:"quantity"`
	WasteFactor        decimal.Decimal `json:"waste_factor"`
}

type BOMService struct {
	bomRepo     domain.BillOfMaterialsRepository
	compRepo    domain.BOMComponentRepository
	wcRepo      domain.WorkCenterRepository
	routingRepo domain.RoutingOperationRepository
	publisher   domain.EventPublisher
}

func NewBOMService(
	bomRepo domain.BillOfMaterialsRepository,
	compRepo domain.BOMComponentRepository,
	wcRepo domain.WorkCenterRepository,
	routingRepo domain.RoutingOperationRepository,
	publisher domain.EventPublisher,
) *BOMService {
	return &BOMService{
		bomRepo:     bomRepo,
		compRepo:    compRepo,
		wcRepo:      wcRepo,
		routingRepo: routingRepo,
		publisher:   publisher,
	}
}

func (s *BOMService) CreateBillOfMaterials(ctx context.Context, productID string, version string, description string, components []BOMComponentInput) (*domain.BillOfMaterials, error) {
	bomID := fmt.Sprintf("bom_%d", time.Now().UnixNano())
	bom := &domain.BillOfMaterials{
		ID:          bomID,
		ProductID:   productID,
		Version:     version,
		Status:      "ACTIVE",
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.bomRepo.Create(ctx, bom)
	if err != nil {
		return nil, err
	}

	for _, c := range components {
		compID := fmt.Sprintf("bomc_%d", time.Now().UnixNano())
		comp := &domain.BOMComponent{
			ID:                 compID,
			BomID:              bomID,
			ComponentProductID: c.ComponentProductID,
			Quantity:           c.Quantity,
			WasteFactor:        c.WasteFactor,
		}
		_ = s.compRepo.Create(ctx, comp)
	}

	return bom, nil
}

func (s *BOMService) GetBillOfMaterials(ctx context.Context, id string) (*domain.BillOfMaterials, error) {
	return s.bomRepo.GetByID(ctx, id)
}

func (s *BOMService) ListBOMs(ctx context.Context) ([]domain.BillOfMaterials, error) {
	return s.bomRepo.List(ctx)
}

func (s *BOMService) UpdateBillOfMaterials(ctx context.Context, id string, productID string, version string, description string) (*domain.BillOfMaterials, error) {
	bom, err := s.bomRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	bom.ProductID = productID
	bom.Version = version
	bom.Description = description
	bom.UpdatedAt = time.Now()

	err = s.bomRepo.Update(ctx, bom)
	if err != nil {
		return nil, err
	}
	return bom, nil
}

func (s *BOMService) DeleteBillOfMaterials(ctx context.Context, id string) error {
	return s.bomRepo.Delete(ctx, id)
}

func (s *BOMService) CreateWorkCenter(ctx context.Context, code string, name string, capacity decimal.Decimal, hourlyRate decimal.Decimal) (*domain.WorkCenter, error) {
	id := fmt.Sprintf("wc_%d", time.Now().UnixNano())
	wc := &domain.WorkCenter{
		ID:             id,
		Code:           code,
		Name:           name,
		Status:         "ACTIVE",
		CapacityHours:  capacity,
		HourlyRate:     hourlyRate,
	}
	err := s.wcRepo.Create(ctx, wc)
	if err != nil {
		return nil, err
	}
	return wc, nil
}

func (s *BOMService) GetWorkCenter(ctx context.Context, id string) (*domain.WorkCenter, error) {
	return s.wcRepo.GetByID(ctx, id)
}

func (s *BOMService) ListWorkCenters(ctx context.Context) ([]domain.WorkCenter, error) {
	return s.wcRepo.List(ctx)
}

func (s *BOMService) UpdateWorkCenter(ctx context.Context, id string, code string, name string, status string, capacity decimal.Decimal, hourlyRate decimal.Decimal) (*domain.WorkCenter, error) {
	wc, err := s.wcRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	wc.Code = code
	wc.Name = name
	wc.Status = status
	wc.CapacityHours = capacity
	wc.HourlyRate = hourlyRate

	err = s.wcRepo.Update(ctx, wc)
	if err != nil {
		return nil, err
	}
	return wc, nil
}

func (s *BOMService) DeleteWorkCenter(ctx context.Context, id string) error {
	return s.wcRepo.Delete(ctx, id)
}

func (s *BOMService) CreateRoutingOperation(ctx context.Context, bomID string, sequenceNum int, workCenterID string, name string, setupTime, runTime decimal.Decimal) (*domain.RoutingOperation, error) {
	id := fmt.Sprintf("route_%d", time.Now().UnixNano())
	op := &domain.RoutingOperation{
		ID:             id,
		BomID:          bomID,
		SequenceNumber: sequenceNum,
		WorkCenterID:   workCenterID,
		OperationName:  name,
		SetupTime:      setupTime,
		RunTime:        runTime,
	}
	err := s.routingRepo.Create(ctx, op)
	if err != nil {
		return nil, err
	}
	return op, nil
}

func (s *BOMService) GetRouting(ctx context.Context, bomID string) ([]domain.RoutingOperation, error) {
	return s.routingRepo.ListByBOMID(ctx, bomID)
}

func (s *BOMService) ListRoutings(ctx context.Context) ([]domain.RoutingOperation, error) {
	return s.routingRepo.List(ctx)
}

func (s *BOMService) GetRoutingByID(ctx context.Context, id string) (*domain.RoutingOperation, error) {
	return s.routingRepo.GetByID(ctx, id)
}

func (s *BOMService) UpdateRouting(ctx context.Context, id string, bomID string, sequenceNum int, workCenterID string, name string, setupTime, runTime decimal.Decimal) (*domain.RoutingOperation, error) {
	op, err := s.routingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	op.BomID = bomID
	op.SequenceNumber = sequenceNum
	op.WorkCenterID = workCenterID
	op.OperationName = name
	op.SetupTime = setupTime
	op.RunTime = runTime

	err = s.routingRepo.Update(ctx, op)
	if err != nil {
		return nil, err
	}
	return op, nil
}

func (s *BOMService) DeleteRouting(ctx context.Context, id string) error {
	return s.routingRepo.Delete(ctx, id)
}
