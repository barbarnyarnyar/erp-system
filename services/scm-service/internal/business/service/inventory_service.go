package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type InventoryService struct {
	invRepo   domain.InventoryItemRepository
	moveRepo  domain.InventoryMovementRepository
	publisher domain.EventPublisher
}

func NewInventoryService(
	invRepo domain.InventoryItemRepository,
	moveRepo domain.InventoryMovementRepository,
	publisher domain.EventPublisher,
) *InventoryService {
	return &InventoryService{
		invRepo:   invRepo,
		moveRepo:  moveRepo,
		publisher: publisher,
	}
}

func (s *InventoryService) ListInventory(ctx context.Context) ([]domain.InventoryItem, error) {
	return s.invRepo.List(ctx)
}

func (s *InventoryService) CreateInventoryItem(ctx context.Context, productID, locationID string, qtyOnHand, reorderPoint, maxStock int, cost decimal.Decimal) (*domain.InventoryItem, error) {
	id := fmt.Sprintf("inv_%d", time.Now().UnixNano())

	ii := &domain.InventoryItem{
		ID:                id,
		ProductID:         productID,
		LocationID:        locationID,
		QuantityOnHand:    qtyOnHand,
		QuantityReserved:  0,
		QuantityAvailable: qtyOnHand,
		ReorderPoint:      reorderPoint,
		MaximumStock:      maxStock,
		UnitCost:          cost,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err := s.invRepo.Create(ctx, ii)
	if err != nil {
		return nil, err
	}

	s.publishValuation(ctx, ii)

	return ii, nil
}

func (s *InventoryService) GetInventoryItem(ctx context.Context, id string) (*domain.InventoryItem, error) {
	return s.invRepo.GetByID(ctx, id)
}

func (s *InventoryService) UpdateInventoryItem(ctx context.Context, id string, qtyOnHand, qtyReserved, reorderPoint, maxStock int, cost decimal.Decimal) (*domain.InventoryItem, error) {
	ii, err := s.invRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ii.QuantityOnHand = qtyOnHand
	ii.QuantityReserved = qtyReserved
	ii.QuantityAvailable = qtyOnHand - qtyReserved
	ii.ReorderPoint = reorderPoint
	ii.MaximumStock = maxStock
	ii.UnitCost = cost
	ii.UpdatedAt = time.Now()

	err = s.invRepo.Update(ctx, ii)
	if err != nil {
		return nil, err
	}

	s.publishValuation(ctx, ii)

	return ii, nil
}

func (s *InventoryService) AdjustInventory(ctx context.Context, productID, locationID string, qty int, movementType string, notes string) (*domain.InventoryItem, error) {
	ii, err := s.invRepo.GetByProductAndLocation(ctx, productID, locationID)
	if err != nil {
		// Not found, initialize a new inventory item
		ii, err = s.CreateInventoryItem(ctx, productID, locationID, 0, 10, 1000, decimal.Zero)
		if err != nil {
			return nil, err
		}
	}

	switch movementType {
	case "RECEIPT", "ADJUSTMENT_ADD":
		ii.QuantityOnHand += qty
		ii.QuantityAvailable += qty
	case "ISSUE", "ADJUSTMENT_SUB":
		if ii.QuantityOnHand < qty {
			return nil, errors.New("insufficient inventory on hand to perform issue")
		}
		ii.QuantityOnHand -= qty
		ii.QuantityAvailable -= qty
	default:
		return nil, fmt.Errorf("unknown inventory movement type: %s", movementType)
	}

	ii.UpdatedAt = time.Now()
	err = s.invRepo.Update(ctx, ii)
	if err != nil {
		return nil, err
	}

	// Create movement log
	move := &domain.InventoryMovement{
		ID:            fmt.Sprintf("move_%d", time.Now().UnixNano()),
		ProductID:     productID,
		LocationID:    locationID,
		MovementType:  movementType,
		Quantity:      qty,
		UnitCost:      ii.UnitCost,
		ReferenceType: "MANUAL_ADJUSTMENT",
		ReferenceID:   ii.ID,
		Notes:         notes,
		CreatedAt:     time.Now(),
	}
	_ = s.moveRepo.Create(ctx, move)

	s.publishValuation(ctx, ii)

	return ii, nil
}

func (s *InventoryService) publishValuation(ctx context.Context, ii *domain.InventoryItem) {
	totalVal := ii.UnitCost.Mul(decimal.NewFromInt(int64(ii.QuantityOnHand)))

	_ = s.publisher.Publish(ctx, domain.TopicScmInventoryValued, ii.ID, domain.InventoryValuedEvent{
		InventoryItemID: ii.ID,
		ProductID:       ii.ProductID,
		LocationID:      ii.LocationID,
		QuantityOnHand:  ii.QuantityOnHand,
		UnitCost:        ii.UnitCost,
		TotalValuation:  totalVal,
		Timestamp:       time.Now(),
	})
}
