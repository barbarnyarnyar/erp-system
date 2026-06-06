package service

import (
	"log"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type stockReservation struct {
	productID  string
	locationID string
	quantity   int
}

// assertInventoryInvariant validates the SCM inventory invariant:
//   quantity_available = quantity_on_hand - quantity_reserved
// with all three fields non-negative. Returns nil when satisfied, or a
// descriptive error if violated. Called at the end of every mutation site
// to catch logic bugs that would silently corrupt inventory state.
func assertInventoryInvariant(ii *domain.InventoryItem) error {
	if ii.QuantityOnHand < 0 {
		return fmt.Errorf("inventory invariant violated: quantity_on_hand=%d (must be >= 0)", ii.QuantityOnHand)
	}
	if ii.QuantityReserved < 0 {
		return fmt.Errorf("inventory invariant violated: quantity_reserved=%d (must be >= 0)", ii.QuantityReserved)
	}
	if ii.QuantityAvailable < 0 {
		return fmt.Errorf("inventory invariant violated: quantity_available=%d (must be >= 0)", ii.QuantityAvailable)
	}
	expected := ii.QuantityOnHand - ii.QuantityReserved
	if ii.QuantityAvailable != expected {
		return fmt.Errorf("inventory invariant violated: quantity_available=%d != quantity_on_hand(%d) - quantity_reserved(%d) = %d",
			ii.QuantityAvailable, ii.QuantityOnHand, ii.QuantityReserved, expected)
	}
	return nil
}

type InventoryService struct {
	invRepo      domain.InventoryItemRepository
	moveRepo     domain.InventoryMovementRepository
	transferRepo domain.StockTransferRepository
	publisher    domain.EventPublisher

	mu           sync.RWMutex
	reservations map[string]stockReservation
}

func NewInventoryService(
	invRepo domain.InventoryItemRepository,
	moveRepo domain.InventoryMovementRepository,
	transferRepo domain.StockTransferRepository,
	publisher domain.EventPublisher,
) *InventoryService {
	return &InventoryService{
		invRepo:      invRepo,
		moveRepo:     moveRepo,
		transferRepo: transferRepo,
		publisher:    publisher,
		reservations: make(map[string]stockReservation),
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

	if err := assertInventoryInvariant(ii); err != nil {
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

	if err := assertInventoryInvariant(ii); err != nil {
		return nil, err
	}

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
	case "ISSUE", "ADJUSTMENT_SUB":
		if ii.QuantityOnHand < qty {
			return nil, errors.New("insufficient inventory on hand to perform issue")
		}
		ii.QuantityOnHand -= qty
	default:
		return nil, fmt.Errorf("unknown inventory movement type: %s", movementType)
	}
	// Always recompute available from the formula; never mutate it by a delta.
	// This preserves the invariant `available = on_hand - reserved` even
	// when `reserved > 0`.
	ii.QuantityAvailable = ii.QuantityOnHand - ii.QuantityReserved

	ii.UpdatedAt = time.Now()
	if err := assertInventoryInvariant(ii); err != nil {
		return nil, err
	}
	err = s.invRepo.Update(ctx, ii)
	if err != nil {
		return nil, err
	}

	// Publish specific events
	if movementType == "RECEIPT" || movementType == "ADJUSTMENT_ADD" {
		if err := s.publisher.Publish(ctx, domain.TopicScmInventoryReceived, ii.ID, domain.InventoryReceivedEvent{
			InventoryItemID: ii.ID,
			ProductID:       ii.ProductID,
			LocationID:      ii.LocationID,
			Quantity:        qty,
			Timestamp:       time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmInventoryReceived, err)
		}
	} else if movementType == "ISSUE" || movementType == "ADJUSTMENT_SUB" {
		if err := s.publisher.Publish(ctx, domain.TopicScmInventoryShipped, ii.ID, domain.InventoryShippedEvent{
			InventoryItemID: ii.ID,
			ProductID:       ii.ProductID,
			LocationID:      ii.LocationID,
			Quantity:        qty,
			Timestamp:       time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmInventoryShipped, err)
		}
	}

	// Always publish adjusted
	qtyChange := qty
	if movementType == "ISSUE" || movementType == "ADJUSTMENT_SUB" {
		qtyChange = -qty
	}
	if err := s.publisher.Publish(ctx, domain.TopicScmInventoryAdjusted, ii.ID, domain.InventoryAdjustedEvent{
		InventoryItemID: ii.ID,
		ProductID:       ii.ProductID,
		LocationID:      ii.LocationID,
		QuantityChange:  qtyChange,
		NewQuantity:     ii.QuantityOnHand,
		Reason:          notes,
		Timestamp:       time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmInventoryAdjusted, err)
	}

	// Check low stock / out of stock
	if ii.QuantityOnHand == 0 {
		if err := s.publisher.Publish(ctx, domain.TopicScmInventoryOutOfStock, ii.ProductID, domain.InventoryOutOfStockEvent{
			ProductID:  ii.ProductID,
			LocationID: ii.LocationID,
			Timestamp:  time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmInventoryOutOfStock, err)
		}
	} else if ii.QuantityOnHand < ii.ReorderPoint {
		if err := s.publisher.Publish(ctx, domain.TopicScmInventoryLowStock, ii.ProductID, domain.InventoryLowStockEvent{
			ProductID:      ii.ProductID,
			LocationID:     ii.LocationID,
			QuantityOnHand: ii.QuantityOnHand,
			ReorderPoint:   ii.ReorderPoint,
			Timestamp:      time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmInventoryLowStock, err)
		}
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

func (s *InventoryService) ReserveStock(ctx context.Context, productID, locationID string, quantity int, referenceID string) error {
	ii, err := s.invRepo.GetByProductAndLocation(ctx, productID, locationID)
	if err != nil {
		return fmt.Errorf("inventory item not found: %w", err)
	}

	if ii.QuantityAvailable < quantity {
		return fmt.Errorf("insufficient available inventory (have %d, requested %d)", ii.QuantityAvailable, quantity)
	}

	ii.QuantityReserved += quantity
	ii.QuantityAvailable = ii.QuantityOnHand - ii.QuantityReserved
	ii.UpdatedAt = time.Now()

	if err := assertInventoryInvariant(ii); err != nil {
		return err
	}
	err = s.invRepo.Update(ctx, ii)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.reservations[referenceID] = stockReservation{
		productID:  productID,
		locationID: locationID,
		quantity:   quantity,
	}
	s.mu.Unlock()

	s.publishValuation(ctx, ii)
	return nil
}

func (s *InventoryService) ReleaseReservation(ctx context.Context, referenceID string) error {
	s.mu.Lock()
	res, ok := s.reservations[referenceID]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("reservation %s not found", referenceID)
	}
	delete(s.reservations, referenceID)
	s.mu.Unlock()

	ii, err := s.invRepo.GetByProductAndLocation(ctx, res.productID, res.locationID)
	if err != nil {
		return fmt.Errorf("inventory item not found for released reservation: %w", err)
	}

	ii.QuantityReserved -= res.quantity
	if ii.QuantityReserved < 0 {
		ii.QuantityReserved = 0
	}
	ii.QuantityAvailable = ii.QuantityOnHand - ii.QuantityReserved
	ii.UpdatedAt = time.Now()

	if err := assertInventoryInvariant(ii); err != nil {
		return err
	}
	err = s.invRepo.Update(ctx, ii)
	if err != nil {
		return err
	}

	s.publishValuation(ctx, ii)
	return nil
}

func (s *InventoryService) CreateStockTransfer(ctx context.Context, fromLocationID, toLocationID, productID string, quantity int) (*domain.StockTransfer, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	ii, err := s.invRepo.GetByProductAndLocation(ctx, productID, fromLocationID)
	if err != nil {
		return nil, fmt.Errorf("source inventory item not found: %w", err)
	}

	if ii.QuantityAvailable < quantity {
		return nil, fmt.Errorf("insufficient source inventory available (have %d, requested %d)", ii.QuantityAvailable, quantity)
	}

	id := fmt.Sprintf("st_%d", time.Now().UnixNano())
	st := &domain.StockTransfer{
		ID:             id,
		FromLocationID: fromLocationID,
		ToLocationID:   toLocationID,
		ProductID:      productID,
		Quantity:       quantity,
		Status:         "PENDING",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = s.transferRepo.Create(ctx, st)
	if err != nil {
		return nil, err
	}

	err = s.ReserveStock(ctx, productID, fromLocationID, quantity, id)
	if err != nil {
		return nil, err
	}

	return st, nil
}

func (s *InventoryService) GetStockTransfer(ctx context.Context, id string) (*domain.StockTransfer, error) {
	return s.transferRepo.GetByID(ctx, id)
}

func (s *InventoryService) ListStockTransfers(ctx context.Context) ([]domain.StockTransfer, error) {
	return s.transferRepo.List(ctx)
}

func (s *InventoryService) ExecuteStockTransfer(ctx context.Context, id string) (*domain.StockTransfer, error) {
	st, err := s.transferRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if st.Status != "PENDING" {
		return nil, fmt.Errorf("stock transfer %s is not pending (status: %s)", id, st.Status)
	}

	err = s.ReleaseReservation(ctx, id)
	if err != nil {
		return nil, err
	}

	fromItem, err := s.invRepo.GetByProductAndLocation(ctx, st.ProductID, st.FromLocationID)
	if err != nil {
		return nil, err
	}
	fromItem.QuantityOnHand -= st.Quantity
	fromItem.QuantityAvailable = fromItem.QuantityOnHand - fromItem.QuantityReserved
	fromItem.UpdatedAt = time.Now()
	if err := assertInventoryInvariant(fromItem); err != nil {
		return nil, err
	}
	err = s.invRepo.Update(ctx, fromItem)
	if err != nil {
		return nil, err
	}
	s.publishValuation(ctx, fromItem)

	fromMove := &domain.InventoryMovement{
		ID:            fmt.Sprintf("move_%d_from", time.Now().UnixNano()),
		ProductID:     st.ProductID,
		LocationID:    st.FromLocationID,
		MovementType:  "ISSUE",
		Quantity:      st.Quantity,
		UnitCost:      fromItem.UnitCost,
		ReferenceType: "STOCK_TRANSFER",
		ReferenceID:   st.ID,
		Notes:         fmt.Sprintf("Stock transfer to %s", st.ToLocationID),
		CreatedAt:     time.Now(),
	}
	_ = s.moveRepo.Create(ctx, fromMove)

	toItem, err := s.invRepo.GetByProductAndLocation(ctx, st.ProductID, st.ToLocationID)
	if err != nil {
		toItem, err = s.CreateInventoryItem(ctx, st.ProductID, st.ToLocationID, 0, 10, 1000, fromItem.UnitCost)
		if err != nil {
			return nil, err
		}
	}
	toItem.QuantityOnHand += st.Quantity
	toItem.QuantityAvailable = toItem.QuantityOnHand - toItem.QuantityReserved
	toItem.UpdatedAt = time.Now()
	if err := assertInventoryInvariant(toItem); err != nil {
		return nil, err
	}
	err = s.invRepo.Update(ctx, toItem)
	if err != nil {
		return nil, err
	}
	s.publishValuation(ctx, toItem)

	toMove := &domain.InventoryMovement{
		ID:            fmt.Sprintf("move_%d_to", time.Now().UnixNano()),
		ProductID:     st.ProductID,
		LocationID:    st.ToLocationID,
		MovementType:  "RECEIPT",
		Quantity:      st.Quantity,
		UnitCost:      fromItem.UnitCost,
		ReferenceType: "STOCK_TRANSFER",
		ReferenceID:   st.ID,
		Notes:         fmt.Sprintf("Stock transfer from %s", st.FromLocationID),
		CreatedAt:     time.Now(),
	}
	_ = s.moveRepo.Create(ctx, toMove)

	now := time.Now()
	st.Status = "TRANSFERRED"
	st.TransferredAt = &now
	st.UpdatedAt = now

	err = s.transferRepo.Update(ctx, st)
	if err != nil {
		return nil, err
	}

	return st, nil
}

func (s *InventoryService) publishValuation(ctx context.Context, ii *domain.InventoryItem) {
	totalVal := ii.UnitCost.Mul(decimal.NewFromInt(int64(ii.QuantityOnHand)))

	if err := s.publisher.Publish(ctx, domain.TopicScmInventoryValued, ii.ID, domain.InventoryValuedEvent{
		InventoryItemID: ii.ID,
		ProductID:       ii.ProductID,
		LocationID:      ii.LocationID,
		QuantityOnHand:  ii.QuantityOnHand,
		UnitCost:        ii.UnitCost,
		TotalValuation:  totalVal,
		Timestamp:       time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicScmInventoryValued, err)
	}
}

func (s *InventoryService) ListMovements(ctx context.Context) ([]domain.InventoryMovement, error) {
	return s.moveRepo.List(ctx)
}
