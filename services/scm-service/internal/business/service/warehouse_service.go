package service

import (
	"context"
	"erp-system/shared/utils"
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
	publisher  domain.EventPublisher
}

func NewWarehouseService(
	recRepo domain.ReceiptRepository,
	recLRepo domain.ReceiptLineRepository,
	shipRepo domain.ShipmentRepository,
	shipLRepo domain.ShipmentLineRepository,
	poRepo domain.PurchaseOrderRepository,
	poLRepo domain.PurchaseOrderLineRepository,
	invService *InventoryService,
	publisher domain.EventPublisher,
) *WarehouseService {
	return &WarehouseService{
		recRepo:    recRepo,
		recLRepo:   recLRepo,
		shipRepo:   shipRepo,
		shipLRepo:  shipLRepo,
		poRepo:     poRepo,
		poLRepo:    poLRepo,
		invService: invService,
		publisher:  publisher,
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
	recID := utils.NewID("rec")
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
			ID:               utils.NewID("receipt-line"),
			ReceiptID:        recID,
			ProductID:        l.ProductID,
			QuantityReceived: l.QuantityReceived,
			CreatedAt:        time.Now(),
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

	// Publish material delivered event for each line
	for _, l := range savedLines {
		if err := s.publisher.Publish(ctx, domain.TopicScmMaterialDelivered, l.ProductID, domain.MaterialDeliveredEvent{
			ProjectID:    "",
			TaskID:       "",
			ShipmentID:   rec.ID,
			DeliveryDate: rec.ReceivedDate,
			Timestamp:    time.Now(),
		}); err != nil {
			utils.LogPublishErr("scm-service", domain.TopicScmMaterialDelivered, err)
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
	shipID := utils.NewID("ship")
	shipNum := fmt.Sprintf("SHP-%d", time.Now().Unix())

	ship := &domain.Shipment{
		ID:                shipID,
		ShipmentNumber:    shipNum,
		Carrier:           carrier,
		TrackingNumber:    trackingNum,
		ShippedDate:       time.Now(),
		EstimatedDelivery: estDelivery,
		Status:            "SHIPPED",
		Notes:             notes,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err := s.shipRepo.Create(ctx, ship)
	if err != nil {
		return nil, err
	}

	savedLines := make([]domain.ShipmentLine, 0, len(lines))

	for _, l := range lines {
		line := domain.ShipmentLine{
			ID:              utils.NewID("shipment-line"),
			ShipmentID:      shipID,
			ProductID:       l.ProductID,
			QuantityShipped: l.QuantityShipped,
			CreatedAt:       time.Now(),
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

	if err := s.publisher.Publish(ctx, domain.TopicScmShipmentCreated, ship.ID, domain.ShipmentCreatedEvent{
		ShipmentID:     ship.ID,
		ShipmentNumber: ship.ShipmentNumber,
		Carrier:        ship.Carrier,
		TrackingNumber: ship.TrackingNumber,
		Timestamp:      time.Now(),
	}); err != nil {
		utils.LogPublishErr("scm-service", domain.TopicScmShipmentCreated, err)
	}

	if err := s.publisher.Publish(ctx, domain.TopicScmShipmentDispatched, ship.ID, domain.ShipmentDispatchedEvent{
		ShipmentID:   ship.ID,
		DispatchedAt: ship.ShippedDate,
		Timestamp:    time.Now(),
	}); err != nil {
		utils.LogPublishErr("scm-service", domain.TopicScmShipmentDispatched, err)
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

	oldStatus := ship.Status
	ship.Status = status
	ship.Notes = notes
	ship.UpdatedAt = time.Now()

	err = s.shipRepo.Update(ctx, ship)
	if err != nil {
		return nil, err
	}

	if status == "DELIVERED" && oldStatus != "DELIVERED" {
		if err := s.publisher.Publish(ctx, domain.TopicScmShipmentDelivered, ship.ID, domain.ShipmentDeliveredEvent{
			ShipmentID:  ship.ID,
			DeliveredAt: time.Now(),
			Timestamp:   time.Now(),
		}); err != nil {
			utils.LogPublishErr("scm-service", domain.TopicScmShipmentDelivered, err)
		}
	}

	if status == "DELAYED" && oldStatus != "DELAYED" {
		if err := s.publisher.Publish(ctx, domain.TopicScmShipmentDelayed, ship.ID, domain.ShipmentDelayedEvent{
			ShipmentID:        ship.ID,
			NewEstimatedDeliv: ship.EstimatedDelivery.AddDate(0, 0, 2), // estimate 2 days delay
			Reason:            notes,
			Timestamp:         time.Now(),
		}); err != nil {
			utils.LogPublishErr("scm-service", domain.TopicScmShipmentDelayed, err)
		}
	}

	return ship, nil
}

func (s *WarehouseService) ListReceiptLines(ctx context.Context, receiptID string) ([]domain.ReceiptLine, error) {
	return s.recLRepo.ListByReceiptID(ctx, receiptID)
}

func (s *WarehouseService) ListShipmentLines(ctx context.Context, shipmentID string) ([]domain.ShipmentLine, error) {
	return s.shipLRepo.ListByShipmentID(ctx, shipmentID)
}

func (s *WarehouseService) TriggerTrainingRequired(ctx context.Context, deptID string, topic string, deadline time.Time) error {
	if err := s.publisher.Publish(ctx, domain.TopicScmTrainingRequired, deptID, domain.SCMTrainingRequiredEvent{
		DepartmentID: deptID,
		Topic:        topic,
		Deadline:     deadline,
		Timestamp:    time.Now(),
	}); err != nil {
		utils.LogPublishErr("scm-service", domain.TopicScmTrainingRequired, err)
	}
	return nil
}
