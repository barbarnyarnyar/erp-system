package service

import (
	"context"
	"erp-system/shared/utils"
	"fmt"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type PurchaseOrderService struct {
	poRepo      domain.PurchaseOrderRepository
	lineRepo    domain.PurchaseOrderLineRepository
	reqRepo     domain.PurchaseRequisitionRepository
	reqLineRepo domain.PurchaseRequisitionLineRepository
	publisher   domain.EventPublisher
	tm          domain.TransactionManager
}

func NewPurchaseOrderService(
	poRepo domain.PurchaseOrderRepository,
	lineRepo domain.PurchaseOrderLineRepository,
	reqRepo domain.PurchaseRequisitionRepository,
	reqLineRepo domain.PurchaseRequisitionLineRepository,
	publisher domain.EventPublisher,
	tm domain.TransactionManager,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		poRepo:      poRepo,
		lineRepo:    lineRepo,
		reqRepo:     reqRepo,
		reqLineRepo: reqLineRepo,
		publisher:   publisher,
		tm:          tm,
	}
}

type POLineInput struct {
	MaterialID      string          `json:"material_id"`
	QuantityOrdered decimal.Decimal `json:"quantity_ordered"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
}

type PurchaseOrderDetails struct {
	domain.PurchaseOrder
	Lines []domain.PurchaseOrderLine `json:"lines"`
}

func (s *PurchaseOrderService) ListPurchaseOrders(ctx context.Context) ([]domain.PurchaseOrder, error) {
	return s.poRepo.List(ctx)
}

func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, supplierID string, expectedDelivery time.Time, notes string, lines []POLineInput) (*PurchaseOrderDetails, error) {
	poID := utils.NewID("po")
	poNum := fmt.Sprintf("PO-%d", time.Now().Unix())

	totalAmount := decimal.Zero
	poLines := make([]domain.PurchaseOrderLine, 0, len(lines))

	// Create lines
	for _, l := range lines {
		lineTotal := l.UnitPrice.Mul(l.QuantityOrdered)
		totalAmount = totalAmount.Add(lineTotal)

		line := domain.PurchaseOrderLine{
			ID:               utils.NewID("po-line"),
			PurchaseOrderID:  poID,
			MaterialID:       l.MaterialID,
			QuantityOrdered:  l.QuantityOrdered,
			QuantityReceived: decimal.Zero,
			UnitPrice:        l.UnitPrice,
			LineTotal:        lineTotal,
			CreatedAt:        time.Now(),
		}

		poLines = append(poLines, line)
	}

	po := &domain.PurchaseOrder{
		ID:               poID,
		LegalEntityID:    "00000000-0000-0000-0000-000000000000",
		PoNumber:         poNum,
		SupplierID:       supplierID,
		OrderDate:        time.Now(),
		ExpectedDelivery: expectedDelivery,
		Status:           domain.PurchaseOrderStatusDRAFT,
		TotalAmount:      totalAmount,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		err := s.poRepo.Create(txCtx, po)
		if err != nil {
			return err
		}

		for _, line := range poLines {
			err = s.lineRepo.Create(txCtx, &line)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &PurchaseOrderDetails{
		PurchaseOrder: *po,
		Lines:         poLines,
	}, nil
}

func (s *PurchaseOrderService) GetPurchaseOrder(ctx context.Context, id string) (*PurchaseOrderDetails, error) {
	po, err := s.poRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lines, err := s.lineRepo.ListByPOID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &PurchaseOrderDetails{
		PurchaseOrder: *po,
		Lines:         lines,
	}, nil
}

func (s *PurchaseOrderService) UpdatePurchaseOrder(ctx context.Context, id string, expectedDelivery time.Time, status, notes string) (*domain.PurchaseOrder, error) {
	po, err := s.poRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := po.Status
	po.ExpectedDelivery = expectedDelivery
	po.Status = domain.PurchaseOrderStatus(status)
	po.UpdatedAt = time.Now()

	err = s.poRepo.Update(ctx, po)
	if err != nil {
		return nil, err
	}

	oldStatusStr := string(oldStatus)
	if (status == "DELIVERED" || status == "RECEIVED") && oldStatusStr != "DELIVERED" && oldStatusStr != "RECEIVED" {
		if err := s.publisher.Publish(ctx, domain.TopicScmPurchaseOrderReceived, po.ID, domain.PurchaseOrderReceivedEvent{
			PurchaseOrderID: po.ID,
			PONumber:        po.PoNumber,
			ReceivedDate:    time.Now(),
			Timestamp:       time.Now(),
		}); err != nil {
			utils.LogPublishErr("scm-service", domain.TopicScmPurchaseOrderReceived, err)
		}
	}

	if status == "CANCELLED" && oldStatus != domain.PurchaseOrderStatusCANCELLED {
		if err := s.publisher.Publish(ctx, domain.TopicScmPurchaseOrderCancelled, po.ID, domain.PurchaseOrderCancelledEvent{
			PurchaseOrderID: po.ID,
			PONumber:        po.PoNumber,
			Reason:          notes,
			Timestamp:       time.Now(),
		}); err != nil {
			utils.LogPublishErr("scm-service", domain.TopicScmPurchaseOrderCancelled, err)
		}
	}

	return po, nil
}

func (s *PurchaseOrderService) DeletePurchaseOrder(ctx context.Context, id string) error {
	return s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.lineRepo.DeleteByPOID(txCtx, id); err != nil {
			return err
		}
		return s.poRepo.Delete(txCtx, id)
	})
}

func (s *PurchaseOrderService) SendPurchaseOrder(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	po, err := s.poRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	po.Status = domain.PurchaseOrderStatus("SUBMITTED")
	po.UpdatedAt = time.Now()

	err = s.poRepo.Update(ctx, po)
	if err != nil {
		return nil, err
	}

	// Publish PO Created/Submitted event to Kafka
	if err := s.publisher.Publish(ctx, domain.TopicScmPurchaseOrderCreated, po.ID, domain.PurchaseOrderCreatedEvent{
		PurchaseOrderID: po.ID,
		PONumber:        po.PoNumber,
		SupplierID:      po.SupplierID,
		TotalAmount:     po.TotalAmount,
		Timestamp:       time.Now(),
	}); err != nil {
		utils.LogPublishErr("scm-service", domain.TopicScmPurchaseOrderCreated, err)
	}

	if err := s.publisher.Publish(ctx, domain.TopicScmPurchaseOrderSent, po.ID, domain.PurchaseOrderSentEvent{
		PurchaseOrderID: po.ID,
		PONumber:        po.PoNumber,
		SupplierID:      po.SupplierID,
		Timestamp:       time.Now(),
	}); err != nil {
		utils.LogPublishErr("scm-service", domain.TopicScmPurchaseOrderSent, err)
	}

	return po, nil
}

type RequisitionLineInput struct {
	MaterialID         string          `json:"material_id"`
	QuantityRequested  decimal.Decimal `json:"quantity_requested"`
	EstimatedUnitPrice decimal.Decimal `json:"estimated_unit_price"`
}

type PurchaseRequisitionDetails struct {
	domain.PurchaseRequisition
	Lines []domain.PurchaseRequisitionLine `json:"lines"`
}

func (s *PurchaseOrderService) ListPurchaseRequisitions(ctx context.Context) ([]domain.PurchaseRequisition, error) {
	return s.reqRepo.List(ctx)
}

func (s *PurchaseOrderService) CreatePurchaseRequisition(ctx context.Context, requesterID string, requestDate time.Time, notes string, lines []RequisitionLineInput) (*PurchaseRequisitionDetails, error) {
	reqID := utils.NewID("req")
	reqNum := fmt.Sprintf("REQ-%d", time.Now().Unix())

	totalAmount := decimal.Zero
	reqLines := make([]domain.PurchaseRequisitionLine, 0, len(lines))

	for _, l := range lines {
		lineTotal := l.EstimatedUnitPrice.Mul(l.QuantityRequested)
		totalAmount = totalAmount.Add(lineTotal)

		line := domain.PurchaseRequisitionLine{
			ID:                    utils.NewID("req-line"),
			PurchaseRequisitionID: reqID,
			MaterialID:            l.MaterialID,
			QuantityRequested:     l.QuantityRequested,
			EstimatedUnitPrice:    l.EstimatedUnitPrice,
			LineTotal:             lineTotal,
		}
		reqLines = append(reqLines, line)
	}

	pr := &domain.PurchaseRequisition{
		ID:            reqID,
		LegalEntityID: "00000000-0000-0000-0000-000000000000",
		ReqNumber:     reqNum,
		RequesterID:   requesterID,
		RequestDate:   requestDate,
		Status:        "DRAFT",
		TotalAmount:   totalAmount,
		Notes:         notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		err := s.reqRepo.Create(txCtx, pr)
		if err != nil {
			return err
		}

		for _, line := range reqLines {
			err = s.reqLineRepo.Create(txCtx, &line)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &PurchaseRequisitionDetails{
		PurchaseRequisition: *pr,
		Lines:               reqLines,
	}, nil
}

func (s *PurchaseOrderService) GetPurchaseRequisition(ctx context.Context, id string) (*PurchaseRequisitionDetails, error) {
	pr, err := s.reqRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lines, err := s.reqLineRepo.ListByRequisitionID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &PurchaseRequisitionDetails{
		PurchaseRequisition: *pr,
		Lines:               lines,
	}, nil
}

func (s *PurchaseOrderService) UpdatePurchaseRequisition(ctx context.Context, id string, requestDate time.Time, status, notes string) (*domain.PurchaseRequisition, error) {
	pr, err := s.reqRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pr.RequestDate = requestDate
	pr.Status = status
	pr.Notes = notes
	pr.UpdatedAt = time.Now()

	err = s.reqRepo.Update(ctx, pr)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PurchaseOrderService) DeletePurchaseRequisition(ctx context.Context, id string) error {
	return s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.reqLineRepo.DeleteByRequisitionID(txCtx, id); err != nil {
			return err
		}
		return s.reqRepo.Delete(txCtx, id)
	})
}

func (s *PurchaseOrderService) ApprovePurchaseRequisition(ctx context.Context, id string) (*domain.PurchaseRequisition, error) {
	pr, err := s.reqRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pr.Status = "APPROVED"
	pr.UpdatedAt = time.Now()

	err = s.reqRepo.Update(ctx, pr)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PurchaseOrderService) RejectPurchaseRequisition(ctx context.Context, id string) (*domain.PurchaseRequisition, error) {
	pr, err := s.reqRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pr.Status = "REJECTED"
	pr.UpdatedAt = time.Now()

	err = s.reqRepo.Update(ctx, pr)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PurchaseOrderService) ListPurchaseOrderLines(ctx context.Context, poID string) ([]domain.PurchaseOrderLine, error) {
	return s.lineRepo.ListByPOID(ctx, poID)
}

func (s *PurchaseOrderService) ListPurchaseRequisitionLines(ctx context.Context, reqID string) ([]domain.PurchaseRequisitionLine, error) {
	return s.reqLineRepo.ListByRequisitionID(ctx, reqID)
}
