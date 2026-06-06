package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
)

type WarehouseService struct {
	recRepo    domain.ReceiptRepository
	recLRepo   domain.ReceiptLineRepository
	shipRepo   domain.ShipmentRepository
	shipLRepo  domain.ShipmentLineRepository
	poRepo     domain.PurchaseOrderRepository
	poLRepo    domain.PurchaseOrderLineRepository
	invService *InventoryService
}

func NewWarehouseService(
	recRepo domain.ReceiptRepository,
	recLRepo domain.ReceiptLineRepository,
	shipRepo domain.ShipmentRepository,
	shipLRepo domain.ShipmentLineRepository,
	poRepo domain.PurchaseOrderRepository,
	poLRepo domain.PurchaseOrderLineRepository,
	invService *InventoryService,
) *WarehouseService {
	return &WarehouseService{
		recRepo:    recRepo,
		recLRepo:   recLRepo,
		shipRepo:   shipRepo,
		shipLRepo:  shipLRepo,
		poRepo:     poRepo,
		poLRepo:    poLRepo,
		invService: invService,
	}
}

type ReceiptLineInput struct {
	ProductID        string `json:"product_id"`
	QuantityReceived int    `json:"quantity_received"`
	LocationID       string `json:"location_id"`
}

type ReceiptDetails struct {
	domain.Receipt
	Lines []domain.ReceiptLine `json:"lines"`
}

type ShipmentLineInput struct {
	ProductID       string `json:"product_id"`
	QuantityShipped int    `json:"quantity_shipped"`
	LocationID      string `json:"location_id"`
}

type ShipmentDetails struct {
	domain.Shipment
	Lines []domain.ShipmentLine `json:"lines"`
}

// ============================================================================
// RECEIPTS LOGIC
// ============================================================================

func (s *WarehouseService) ListReceipts(ctx context.Context) ([]domain.Receipt, error) {
	return s.recRepo.List(ctx)
}

func (s *WarehouseService) CreateReceipt(ctx context.Context, poID string, notes string, lines []ReceiptLineInput) (*ReceiptDetails, error) {
	recID := fmt.Sprintf("rec_%d", time.Now().UnixNano())
	recNum := fmt.Sprintf("REC-%d", time.Now().Unix())

	rec := &domain.Receipt{
		ID:            recID,
		ReceiptNumber: recNum,
		ReceivedDate:  time.Now(),
		Status:        "RECEIVED",
		Notes:         notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if poID != "" {
		rec.PurchaseOrderID = &poID
	}

	err := s.recRepo.Create(ctx, rec)
	if err != nil {
		return nil, err
	}

	savedLines := make([]domain.ReceiptLine, 0, len(lines))

	// Look up purchase order lines to match quantities if poID is provided
	var poLines []domain.PurchaseOrderLine
	if poID != "" {
		poLines, _ = s.poLRepo.ListByPOID(ctx, poID)
	}

	for _, l := range lines {
		line := domain.ReceiptLine{
			ID:                fmt.Sprintf("recl_%d", time.Now().UnixNano()+int64(len(savedLines))),
			ReceiptID:         recID,
			ProductID:         l.ProductID,
			QuantityReceived:  l.QuantityReceived,
			CreatedAt:         time.Now(),
		}

		err = s.recLRepo.Create(ctx, &line)
		if err != nil {
			return nil, err
		}
		savedLines = append(savedLines, line)

		// Increment PO received quantity if matching
		if poID != "" {
			for _, pol := range poLines {
				if pol.ProductID == l.ProductID {
					pol.QuantityReceived += l.QuantityReceived
					_ = s.poLRepo.Create(ctx, &pol) // Save back/update line in in-memory repo
					break
				}
			}
		}

		// Adjust stock levels
		locationID := l.LocationID
		if locationID == "" {
			locationID = "loc_default" // default warehouse
		}
		_, _ = s.invService.AdjustInventory(ctx, l.ProductID, locationID, l.QuantityReceived, "RECEIPT", "Received stock via "+recNum)
	}

	// If all items received, update PO status to DELIVERED
	if poID != "" {
		po, err := s.poRepo.GetByID(ctx, poID)
		if err == nil {
			updatedPOLines, _ := s.poLRepo.ListByPOID(ctx, poID)
			allReceived := true
			for _, pol := range updatedPOLines {
				if pol.QuantityReceived < pol.QuantityOrdered {
					allReceived = false
					break
				}
			}
			if allReceived {
				po.Status = "DELIVERED"
			} else {
				po.Status = "PARTIALLY_DELIVERED"
			}
			po.UpdatedAt = time.Now()
			_ = s.poRepo.Update(ctx, po)
		}
	}

	return &ReceiptDetails{
		Receipt: *rec,
		Lines:   savedLines,
	}, nil
}

func (s *WarehouseService) GetReceipt(ctx context.Context, id string) (*ReceiptDetails, error) {
	rec, err := s.recRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lines, err := s.recLRepo.ListByReceiptID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &ReceiptDetails{
		Receipt: *rec,
		Lines:   lines,
	}, nil
}

func (s *WarehouseService) UpdateReceipt(ctx context.Context, id, status, notes string) (*domain.Receipt, error) {
	rec, err := s.recRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rec.Status = status
	rec.Notes = notes
	rec.UpdatedAt = time.Now()

	err = s.recRepo.Update(ctx, rec)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

// ============================================================================
// SHIPMENTS LOGIC
// ============================================================================

func (s *WarehouseService) ListShipments(ctx context.Context) ([]domain.Shipment, error) {
	return s.shipRepo.List(ctx)
}

func (s *WarehouseService) CreateShipment(ctx context.Context, carrier, trackingNum string, estDelivery time.Time, notes string, lines []ShipmentLineInput) (*ShipmentDetails, error) {
	shipID := fmt.Sprintf("ship_%d", time.Now().UnixNano())
	shipNum := fmt.Sprintf("SHP-%d", time.Now().Unix())

	ship := &domain.Shipment{
		ID:                 shipID,
		ShipmentNumber:     shipNum,
		Carrier:            carrier,
		TrackingNumber:     trackingNum,
		ShippedDate:        time.Now(),
		EstimatedDelivery:  estDelivery,
		Status:             "SHIPPED",
		Notes:              notes,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := s.shipRepo.Create(ctx, ship)
	if err != nil {
		return nil, err
	}

	savedLines := make([]domain.ShipmentLine, 0, len(lines))

	for _, l := range lines {
		line := domain.ShipmentLine{
			ID:               fmt.Sprintf("shipl_%d", time.Now().UnixNano()+int64(len(savedLines))),
			ShipmentID:       shipID,
			ProductID:        l.ProductID,
			QuantityShipped:  l.QuantityShipped,
			CreatedAt:        time.Now(),
		}

		err = s.shipLRepo.Create(ctx, &line)
		if err != nil {
			return nil, err
		}
		savedLines = append(savedLines, line)

		// Deduct stock levels (Issue)
		locationID := l.LocationID
		if locationID == "" {
			locationID = "loc_default"
		}
		_, _ = s.invService.AdjustInventory(ctx, l.ProductID, locationID, l.QuantityShipped, "ISSUE", "Shipped stock out via "+shipNum)
	}

	return &ShipmentDetails{
		Shipment: *ship,
		Lines:    savedLines,
	}, nil
}

func (s *WarehouseService) GetShipment(ctx context.Context, id string) (*ShipmentDetails, error) {
	ship, err := s.shipRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lines, err := s.shipLRepo.ListByShipmentID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &ShipmentDetails{
		Shipment: *ship,
		Lines:    lines,
	}, nil
}

func (s *WarehouseService) UpdateShipment(ctx context.Context, id, status, notes string) (*domain.Shipment, error) {
	ship, err := s.shipRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ship.Status = status
	ship.Notes = notes
	ship.UpdatedAt = time.Now()

	err = s.shipRepo.Update(ctx, ship)
	if err != nil {
		return nil, err
	}

	return ship, nil
}

func (s *WarehouseService) ListReceiptLines(ctx context.Context, receiptID string) ([]domain.ReceiptLine, error) {
	return s.recLRepo.ListByReceiptID(ctx, receiptID)
}

func (s *WarehouseService) ListShipmentLines(ctx context.Context, shipmentID string) ([]domain.ShipmentLine, error) {
	return s.shipLRepo.ListByShipmentID(ctx, shipmentID)
}
