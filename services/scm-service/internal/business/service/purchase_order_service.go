package service

import (
	"context"
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
}

func NewPurchaseOrderService(
	poRepo domain.PurchaseOrderRepository,
	lineRepo domain.PurchaseOrderLineRepository,
	reqRepo domain.PurchaseRequisitionRepository,
	reqLineRepo domain.PurchaseRequisitionLineRepository,
	publisher domain.EventPublisher,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		poRepo:      poRepo,
		lineRepo:    lineRepo,
		reqRepo:     reqRepo,
		reqLineRepo: reqLineRepo,
		publisher:   publisher,
	}
}

type POLineInput struct {
	ProductID       string          `json:"product_id"`
	QuantityOrdered int             `json:"quantity_ordered"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	Description     string          `json:"description"`
}

type PurchaseOrderDetails struct {
	domain.PurchaseOrder
	Lines []domain.PurchaseOrderLine `json:"lines"`
}

func (s *PurchaseOrderService) ListPurchaseOrders(ctx context.Context) ([]domain.PurchaseOrder, error) {
	return s.poRepo.List(ctx)
}

func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, supplierID string, expectedDelivery time.Time, notes string, lines []POLineInput) (*PurchaseOrderDetails, error) {
	poID := fmt.Sprintf("po_%d", time.Now().UnixNano())
	poNum := fmt.Sprintf("PO-%d", time.Now().Unix())

	totalAmount := decimal.Zero
	poLines := make([]domain.PurchaseOrderLine, 0, len(lines))

	// Create lines
	for _, l := range lines {
		lineTotal := l.UnitPrice.Mul(decimal.NewFromInt(int64(l.QuantityOrdered)))
		totalAmount = totalAmount.Add(lineTotal)

		line := domain.PurchaseOrderLine{
			ID:                fmt.Sprintf("pol_%d", time.Now().UnixNano()+int64(len(poLines))),
			PurchaseOrderID:   poID,
			ProductID:         l.ProductID,
			QuantityOrdered:   l.QuantityOrdered,
			QuantityReceived:  0,
			UnitPrice:         l.UnitPrice,
			LineTotal:         lineTotal,
			Description:       l.Description,
			CreatedAt:         time.Now(),
		}

		poLines = append(poLines, line)
	}

	po := &domain.PurchaseOrder{
		ID:               poID,
		PoNumber:         poNum,
		SupplierID:       supplierID,
		OrderDate:        time.Now(),
		ExpectedDelivery: expectedDelivery,
		Status:           "DRAFT",
		TotalAmount:      totalAmount,
		Notes:            notes,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := s.poRepo.Create(ctx, po)
	if err != nil {
		return nil, err
	}

	for _, line := range poLines {
		err = s.lineRepo.Create(ctx, &line)
		if err != nil {
			return nil, err
		}
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

	po.ExpectedDelivery = expectedDelivery
	po.Status = status
	po.Notes = notes
	po.UpdatedAt = time.Now()

	err = s.poRepo.Update(ctx, po)
	if err != nil {
		return nil, err
	}

	return po, nil
}

func (s *PurchaseOrderService) DeletePurchaseOrder(ctx context.Context, id string) error {
	// First delete lines
	_ = s.lineRepo.DeleteByPOID(ctx, id)
	return s.poRepo.Delete(ctx, id)
}

func (s *PurchaseOrderService) SendPurchaseOrder(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	po, err := s.poRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	po.Status = "SUBMITTED"
	po.UpdatedAt = time.Now()

	err = s.poRepo.Update(ctx, po)
	if err != nil {
		return nil, err
	}

	// Publish PO Created/Submitted event to Kafka
	_ = s.publisher.Publish(ctx, domain.TopicScmPurchaseOrderCreated, po.ID, domain.PurchaseOrderCreatedEvent{
		PurchaseOrderID: po.ID,
		PONumber:        po.PoNumber,
		SupplierID:      po.SupplierID,
		TotalAmount:     po.TotalAmount,
		Timestamp:       time.Now(),
	})

	return po, nil
}

type RequisitionLineInput struct {
	ProductID          string          `json:"product_id"`
	QuantityRequested  int             `json:"quantity_requested"`
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
	reqID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	reqNum := fmt.Sprintf("REQ-%d", time.Now().Unix())

	totalAmount := decimal.Zero
	reqLines := make([]domain.PurchaseRequisitionLine, 0, len(lines))

	for _, l := range lines {
		lineTotal := l.EstimatedUnitPrice.Mul(decimal.NewFromInt(int64(l.QuantityRequested)))
		totalAmount = totalAmount.Add(lineTotal)

		line := domain.PurchaseRequisitionLine{
			ID:                    fmt.Sprintf("reql_%d", time.Now().UnixNano()+int64(len(reqLines))),
			PurchaseRequisitionID: reqID,
			ProductID:             l.ProductID,
			QuantityRequested:     l.QuantityRequested,
			EstimatedUnitPrice:    l.EstimatedUnitPrice,
			LineTotal:             lineTotal,
		}
		reqLines = append(reqLines, line)
	}

	pr := &domain.PurchaseRequisition{
		ID:          reqID,
		ReqNumber:   reqNum,
		RequesterID: requesterID,
		RequestDate: requestDate,
		Status:      "DRAFT",
		TotalAmount: totalAmount,
		Notes:       notes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.reqRepo.Create(ctx, pr)
	if err != nil {
		return nil, err
	}

	for _, line := range reqLines {
		err = s.reqLineRepo.Create(ctx, &line)
		if err != nil {
			return nil, err
		}
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
	_ = s.reqLineRepo.DeleteByRequisitionID(ctx, id)
	return s.reqRepo.Delete(ctx, id)
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
