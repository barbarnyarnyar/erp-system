package service

import (
	"context"
	"erp-system/shared/utils"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type stockReservation struct {
	materialID string
	locationID string
	quantity   decimal.Decimal
}

// assertInventoryInvariant validates the SCM inventory invariant:
//
//	quantity_available = quantity_on_hand - quantity_reserved
//
// with all three fields non-negative. Returns nil when satisfied, or a
// descriptive error if violated. Called at the end of every mutation site
// to catch logic bugs that would silently corrupt inventory state.
func assertInventoryInvariant(sb *domain.StockBalance) error {
	if sb.QuantityOnHand.LessThan(decimal.Zero) {
		return fmt.Errorf("inventory invariant violated: quantity_on_hand=%s (must be >= 0)", sb.QuantityOnHand)
	}
	if sb.QuantityReserved.LessThan(decimal.Zero) {
		return fmt.Errorf("inventory invariant violated: quantity_reserved=%s (must be >= 0)", sb.QuantityReserved)
	}
	if sb.QuantityAvailable.LessThan(decimal.Zero) {
		return fmt.Errorf("inventory invariant violated: quantity_available=%s (must be >= 0)", sb.QuantityAvailable)
	}
	expected := sb.QuantityOnHand.Sub(sb.QuantityReserved)
	if !sb.QuantityAvailable.Equal(expected) {
		return fmt.Errorf("inventory invariant violated: quantity_available=%s != quantity_on_hand(%s) - quantity_reserved(%s) = %s",
			sb.QuantityAvailable, sb.QuantityOnHand, sb.QuantityReserved, expected)
	}
	return nil
}

type InventoryService struct {
	invRepo      domain.StockBalanceRepository
	moveRepo     domain.InventoryMovementRepository
	transferRepo domain.StockTransferRepository
	publisher    domain.EventPublisher
	tm           domain.TransactionManager

	mu           sync.RWMutex
	reservations map[string]stockReservation
}

func NewInventoryService(
	invRepo domain.StockBalanceRepository,
	moveRepo domain.InventoryMovementRepository,
	transferRepo domain.StockTransferRepository,
	publisher domain.EventPublisher,
	tm domain.TransactionManager,
) *InventoryService {
	return &InventoryService{
		invRepo:      invRepo,
		moveRepo:     moveRepo,
		transferRepo: transferRepo,
		publisher:    publisher,
		tm:           tm,
		reservations: make(map[string]stockReservation),
	}
}

func (s *InventoryService) ListInventory(ctx context.Context) ([]domain.StockBalance, error) {
	return s.invRepo.List(ctx)
}

func (s *InventoryService) CreateStockBalance(ctx context.Context, materialID, locationID string, qtyOnHand decimal.Decimal) (*domain.StockBalance, error) {
	id := utils.NewID("sb")

	sb := &domain.StockBalance{
		ID:                id,
		LegalEntityID:     "00000000-0000-0000-0000-000000000000",
		LocationID:        locationID,
		MaterialID:        materialID,
		QuantityOnHand:    qtyOnHand,
		QuantityReserved:  decimal.Zero,
		QuantityAvailable: qtyOnHand,
		Version:           0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err := s.invRepo.Create(ctx, sb)
	if err != nil {
		return nil, err
	}

	if err := assertInventoryInvariant(sb); err != nil {
		return nil, err
	}

	s.publishValuation(ctx, sb)

	return sb, nil
}

func (s *InventoryService) GetStockBalance(ctx context.Context, id string) (*domain.StockBalance, error) {
	return s.invRepo.GetByID(ctx, id)
}

func (s *InventoryService) UpdateStockBalance(ctx context.Context, id string, qtyOnHand, qtyReserved decimal.Decimal) (*domain.StockBalance, error) {
	sb, err := s.invRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sb.QuantityOnHand = qtyOnHand
	sb.QuantityReserved = qtyReserved
	sb.QuantityAvailable = qtyOnHand.Sub(qtyReserved)
	sb.UpdatedAt = time.Now()

	if err := assertInventoryInvariant(sb); err != nil {
		return nil, err
	}

	err = s.invRepo.Update(ctx, sb)
	if err != nil {
		return nil, err
	}

	s.publishValuation(ctx, sb)

	return sb, nil
}

func (s *InventoryService) AdjustInventory(ctx context.Context, materialID, locationID string, qty decimal.Decimal, movementType string, notes string) (*domain.StockBalance, error) {
	var result *domain.StockBalance
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		sb, err := s.invRepo.GetByMaterialAndLocation(txCtx, materialID, locationID)
		if err != nil {
			// Not found, initialize a new stock balance
			sb, err = s.CreateStockBalance(txCtx, materialID, locationID, decimal.Zero)
			if err != nil {
				return err
			}
		}

		switch movementType {
		case "RECEIPT", "ADJUSTMENT_ADD":
			sb.QuantityOnHand = sb.QuantityOnHand.Add(qty)
		case "ISSUE", "ADJUSTMENT_SUB":
			if sb.QuantityOnHand.LessThan(qty) {
				return errors.New("insufficient inventory on hand to perform issue")
			}
			sb.QuantityOnHand = sb.QuantityOnHand.Sub(qty)
		default:
			return fmt.Errorf("unknown inventory movement type: %s", movementType)
		}
		
		sb.QuantityAvailable = sb.QuantityOnHand.Sub(sb.QuantityReserved)
		sb.UpdatedAt = time.Now()
		
		if err := assertInventoryInvariant(sb); err != nil {
			return err
		}
		err = s.invRepo.Update(txCtx, sb)
		if err != nil {
			return err
		}

		// Create movement log
		move := &domain.InventoryMovement{
			ID:            utils.NewID("move"),
			LegalEntityID: sb.LegalEntityID,
			MaterialID:    materialID,
			LocationID:    locationID,
			MovementType:  movementType,
			Quantity:      qty,
			ReferenceType: "MANUAL_ADJUSTMENT",
			ReferenceID:   sb.ID,
			CreatedAt:     time.Now(),
		}
		err = s.moveRepo.Create(txCtx, move)
		if err != nil {
			return err
		}

		result = sb
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish specific events outside the transaction block
	if utils.IsAny(movementType, "RECEIPT", "ADJUSTMENT_ADD") {
		if err := s.publisher.Publish(ctx, domain.TopicScmInventoryReceived, result.ID, domain.InventoryReceivedEvent{
			InventoryItemID: result.ID,
			ProductID:       result.MaterialID,
			LocationID:      result.LocationID,
			Quantity:        int(qty.IntPart()),
			Timestamp:       time.Now(),
		}); err != nil {
			utils.LogPublishErr("scm-service", domain.TopicScmInventoryReceived, err)
		}
	} else if utils.IsAny(movementType, "ISSUE", "ADJUSTMENT_SUB") {
		if err := s.publisher.Publish(ctx, domain.TopicScmInventoryShipped, result.ID, domain.InventoryShippedEvent{
			InventoryItemID: result.ID,
			ProductID:       result.MaterialID,
			LocationID:      result.LocationID,
			Quantity:        int(qty.IntPart()),
			Timestamp:       time.Now(),
		}); err != nil {
			utils.LogPublishErr("scm-service", domain.TopicScmInventoryShipped, err)
		}
	}

	// Always publish adjusted
	qtyChange := qty
	if utils.IsAny(movementType, "ISSUE", "ADJUSTMENT_SUB") {
		qtyChange = qty.Neg()
	}
	if err := s.publisher.Publish(ctx, domain.TopicScmInventoryAdjusted, result.ID, domain.InventoryAdjustedEvent{
		InventoryItemID: result.ID,
		ProductID:       result.MaterialID,
		LocationID:      result.LocationID,
		QuantityChange:  int(qtyChange.IntPart()),
		NewQuantity:     int(result.QuantityOnHand.IntPart()),
		Reason:          notes,
		Timestamp:       time.Now(),
	}); err != nil {
		utils.LogPublishErr("scm-service", domain.TopicScmInventoryAdjusted, err)
	}

	// Check low stock / out of stock
	if result.QuantityOnHand.IsZero() {
		if err := s.publisher.Publish(ctx, domain.TopicScmInventoryOutOfStock, result.MaterialID, domain.InventoryOutOfStockEvent{
			ProductID:  result.MaterialID,
			LocationID: result.LocationID,
			Timestamp:  time.Now(),
		}); err != nil {
			utils.LogPublishErr("scm-service", domain.TopicScmInventoryOutOfStock, err)
		}
	}

	s.publishValuation(ctx, result)

	return result, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, materialID, locationID string, quantity decimal.Decimal, referenceID string) error {
	sb, err := s.invRepo.GetByMaterialAndLocation(ctx, materialID, locationID)
	if err != nil {
		return fmt.Errorf("stock balance not found: %w", err)
	}

	if sb.QuantityAvailable.LessThan(quantity) {
		return fmt.Errorf("insufficient available inventory (have %s, requested %s)", sb.QuantityAvailable, quantity)
	}

	sb.QuantityReserved = sb.QuantityReserved.Add(quantity)
	sb.QuantityAvailable = sb.QuantityOnHand.Sub(sb.QuantityReserved)
	sb.UpdatedAt = time.Now()

	if err := assertInventoryInvariant(sb); err != nil {
		return err
	}
	err = s.invRepo.Update(ctx, sb)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.reservations[referenceID] = stockReservation{
		materialID:  materialID,
		locationID: locationID,
		quantity:   quantity,
	}
	s.mu.Unlock()

	s.publishValuation(ctx, sb)
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

	sb, err := s.invRepo.GetByMaterialAndLocation(ctx, res.materialID, res.locationID)
	if err != nil {
		return fmt.Errorf("stock balance not found for released reservation: %w", err)
	}

	sb.QuantityReserved = sb.QuantityReserved.Sub(res.quantity)
	if sb.QuantityReserved.LessThan(decimal.Zero) {
		sb.QuantityReserved = decimal.Zero
	}
	sb.QuantityAvailable = sb.QuantityOnHand.Sub(sb.QuantityReserved)
	sb.UpdatedAt = time.Now()

	if err := assertInventoryInvariant(sb); err != nil {
		return err
	}
	err = s.invRepo.Update(ctx, sb)
	if err != nil {
		return err
	}

	s.publishValuation(ctx, sb)
	return nil
}

func (s *InventoryService) CreateStockTransfer(ctx context.Context, fromLocationID, toLocationID, materialID string, quantity int) (*domain.StockTransfer, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	ii, err := s.invRepo.GetByMaterialAndLocation(ctx, materialID, fromLocationID)
	if err != nil {
		return nil, fmt.Errorf("source stock balance not found: %w", err)
	}

	qtyDec := decimal.NewFromInt(int64(quantity))

	if ii.QuantityAvailable.LessThan(qtyDec) {
		return nil, fmt.Errorf("insufficient source inventory available (have %s, requested %d)", ii.QuantityAvailable, quantity)
	}

	id := utils.NewID("st")
	st := &domain.StockTransfer{
		ID:             id,
		FromLocationID: fromLocationID,
		ToLocationID:   toLocationID,
		MaterialID:     materialID,
		Quantity:       qtyDec,
		Status:         "PENDING",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		err = s.transferRepo.Create(txCtx, st)
		if err != nil {
			return err
		}

		err = s.ReserveStock(txCtx, materialID, fromLocationID, qtyDec, id)
		if err != nil {
			return err
		}
		return nil
	})

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
	var st *domain.StockTransfer
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		st, err = s.transferRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		if st.Status != "PENDING" {
			return fmt.Errorf("stock transfer %s is not pending (status: %s)", id, st.Status)
		}

		err = s.ReleaseReservation(txCtx, id)
		if err != nil {
			return err
		}

		qtyDec := st.Quantity

		fromItem, err := s.invRepo.GetByMaterialAndLocation(txCtx, st.MaterialID, st.FromLocationID)
		if err != nil {
			return err
		}
		fromItem.QuantityOnHand = fromItem.QuantityOnHand.Sub(qtyDec)
		fromItem.QuantityAvailable = fromItem.QuantityOnHand.Sub(fromItem.QuantityReserved)
		fromItem.UpdatedAt = time.Now()
		if err := assertInventoryInvariant(fromItem); err != nil {
			return err
		}
		err = s.invRepo.Update(txCtx, fromItem)
		if err != nil {
			return err
		}

		fromMove := &domain.InventoryMovement{
			ID:            utils.NewID("move-from"),
			LegalEntityID: fromItem.LegalEntityID,
			MaterialID:    st.MaterialID,
			LocationID:    st.FromLocationID,
			MovementType:  "ISSUE",
			Quantity:      qtyDec,
			ReferenceType: "STOCK_TRANSFER",
			ReferenceID:   st.ID,
			CreatedAt:     time.Now(),
		}
		err = s.moveRepo.Create(txCtx, fromMove)
		if err != nil {
			return err
		}

		toItem, err := s.invRepo.GetByMaterialAndLocation(txCtx, st.MaterialID, st.ToLocationID)
		if err != nil {
			toItem, err = s.CreateStockBalance(txCtx, st.MaterialID, st.ToLocationID, qtyDec)
			if err != nil {
				return err
			}
		} else {
			toItem.QuantityOnHand = toItem.QuantityOnHand.Add(qtyDec)
			toItem.QuantityAvailable = toItem.QuantityOnHand.Sub(toItem.QuantityReserved)
			toItem.UpdatedAt = time.Now()
			if err := assertInventoryInvariant(toItem); err != nil {
				return err
			}
			err = s.invRepo.Update(txCtx, toItem)
			if err != nil {
				return err
			}
		}

		toMove := &domain.InventoryMovement{
			ID:            utils.NewID("move-to"),
			LegalEntityID: toItem.LegalEntityID,
			MaterialID:    st.MaterialID,
			LocationID:    st.ToLocationID,
			MovementType:  "RECEIPT",
			Quantity:      qtyDec,
			ReferenceType: "STOCK_TRANSFER",
			ReferenceID:   st.ID,
			CreatedAt:     time.Now(),
		}
		err = s.moveRepo.Create(txCtx, toMove)
		if err != nil {
			return err
		}

		now := time.Now()
		st.Status = "TRANSFERRED"
		st.TransferredAt = &now
		st.UpdatedAt = now

		err = s.transferRepo.Update(txCtx, st)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish valuations outside the transaction
	if fromItem, err := s.invRepo.GetByMaterialAndLocation(ctx, st.MaterialID, st.FromLocationID); err == nil {
		s.publishValuation(ctx, fromItem)
	}
	if toItem, err := s.invRepo.GetByMaterialAndLocation(ctx, st.MaterialID, st.ToLocationID); err == nil {
		s.publishValuation(ctx, toItem)
	}

	return st, nil
}

func (s *InventoryService) publishValuation(ctx context.Context, sb *domain.StockBalance) {
	if err := s.publisher.Publish(ctx, domain.TopicScmInventoryValued, sb.ID, domain.InventoryValuedEvent{
		InventoryItemID: sb.ID,
		ProductID:       sb.MaterialID,
		LocationID:      sb.LocationID,
		QuantityOnHand:  int(sb.QuantityOnHand.IntPart()),
		UnitCost:        decimal.Zero,
		TotalValuation:  decimal.Zero,
		Timestamp:       time.Now(),
	}); err != nil {
		utils.LogPublishErr("scm-service", domain.TopicScmInventoryValued, err)
	}
}

func (s *InventoryService) ListMovements(ctx context.Context) ([]domain.InventoryMovement, error) {
	return s.moveRepo.List(ctx)
}
