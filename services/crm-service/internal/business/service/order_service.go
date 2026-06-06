package service

import (
	"log"
	"context"
	"fmt"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type SalesOrderItemInput struct {
	ProductID string          `json:"product_id"`
	Quantity  int             `json:"quantity"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	Discount  decimal.Decimal `json:"discount"`
}

type SalesOrderService struct {
	orderRepo     domain.SalesOrderRepository
	orderItemRepo domain.SalesOrderItemRepository
	customerRepo  domain.CustomerRepository
	publisher     domain.EventPublisher
}

func NewSalesOrderService(
	orderRepo domain.SalesOrderRepository,
	orderItemRepo domain.SalesOrderItemRepository,
	customerRepo domain.CustomerRepository,
	publisher domain.EventPublisher,
) *SalesOrderService {
	return &SalesOrderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		customerRepo:  customerRepo,
		publisher:     publisher,
	}
}

func (s *SalesOrderService) CreateSalesOrder(ctx context.Context, customerID string, items []SalesOrderItemInput) (*domain.SalesOrder, error) {
	orderID := fmt.Sprintf("so_%d", time.Now().UnixNano())
	total := decimal.Zero

	for _, it := range items {
		subtotal := decimal.NewFromInt(int64(it.Quantity)).Mul(it.UnitPrice).Sub(it.Discount)
		total = total.Add(subtotal)
	}

	order := &domain.SalesOrder{
		ID:          orderID,
		CustomerID:  customerID,
		OrderDate:   time.Now(),
		Status:      "DRAFT",
		TotalAmount: total,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		itemID := fmt.Sprintf("soi_%d", time.Now().UnixNano())
		item := &domain.SalesOrderItem{
			ID:           itemID,
			SalesOrderID: orderID,
			ProductID:    it.ProductID,
			Quantity:     it.Quantity,
			UnitPrice:    it.UnitPrice,
			Discount:     it.Discount,
		}
		_ = s.orderItemRepo.Create(ctx, item)
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderCreated, orderID, domain.SalesOrderCreatedEvent{
		SalesOrderID: orderID,
		CustomerID:   customerID,
		TotalAmount:  total,
		Timestamp:    time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderCreated, err)
	}

	return order, nil
}

func (s *SalesOrderService) GetSalesOrder(ctx context.Context, id string) (*domain.SalesOrder, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *SalesOrderService) ListSalesOrders(ctx context.Context) ([]domain.SalesOrder, error) {
	return s.orderRepo.List(ctx)
}

func (s *SalesOrderService) UpdateSalesOrder(ctx context.Context, id string, status string) (*domain.SalesOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	err = s.orderRepo.Update(ctx, order)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderUpdated, id, domain.SalesOrderUpdatedEvent{
		SalesOrderID: id,
		Status:       status,
		TotalAmount:  order.TotalAmount,
		Timestamp:    time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderUpdated, err)
	}

	switch status {
	case "CONFIRMED":
		if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderConfirmed, id, domain.SalesOrderConfirmedEvent{
			SalesOrderID: id,
			CustomerID:   order.CustomerID,
			TotalAmount:  order.TotalAmount,
			Timestamp:    time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderConfirmed, err)
		}
	case "SHIPPED":
		if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderShipped, id, domain.SalesOrderShippedEvent{
			SalesOrderID: id,
			Timestamp:    time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderShipped, err)
		}
	case "DELIVERED":
		if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderDelivered, id, domain.SalesOrderDeliveredEvent{
			SalesOrderID: id,
			Timestamp:    time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderDelivered, err)
		}
	case "CANCELLED":
		if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderCancelled, id, domain.SalesOrderCancelledEvent{
			SalesOrderID: id,
			Reason:       "Administrative status update",
			Timestamp:    time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderCancelled, err)
		}
	}

	return order, nil
}

func (s *SalesOrderService) ConfirmSalesOrder(ctx context.Context, id string) (*domain.SalesOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}

	if !order.CanConfirm() {
		return nil, domain.ErrOrderNotConfirmable
	}

	customer, err := s.customerRepo.GetByID(ctx, order.CustomerID)
	if err != nil {
		return nil, domain.ErrCustomerNotFound
	}
	if customer.Status != domain.CustomerStatusActive {
		return nil, domain.ErrCustomerNotActive
	}

	items, err := s.orderItemRepo.ListByOrderID(ctx, order.ID)
	if err != nil || len(items) == 0 {
		return nil, domain.ErrOrderHasNoItems
	}
	for _, it := range items {
		if it.Quantity <= 0 {
			return nil, domain.ErrInvalidItemQuantity
		}
	}

	order.MarkConfirmed(time.Now())
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderConfirmed, order.ID, domain.SalesOrderConfirmedEvent{
		SalesOrderID: order.ID,
		CustomerID:   order.CustomerID,
		TotalAmount:  order.TotalAmount,
		Timestamp:    time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderConfirmed, err)
	}

	return order, nil
}

func (s *SalesOrderService) DeleteSalesOrder(ctx context.Context, id string) error {
	if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderCancelled, id, domain.SalesOrderCancelledEvent{
		SalesOrderID: id,
		Reason:       "Manual cancellation request",
		Timestamp:    time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderCancelled, err)
	}
	return s.orderRepo.Delete(ctx, id)
}

func (s *SalesOrderService) ReceiveSalesOrder(ctx context.Context, salesOrderID, customerID string, totalAmount decimal.Decimal) error {
	if err := s.publisher.Publish(ctx, domain.TopicCrmSalesOrderReceived, salesOrderID, domain.SalesOrderReceivedEvent{
		SalesOrderID: salesOrderID,
		CustomerID:   customerID,
		TotalAmount:  totalAmount,
		Timestamp:    time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicCrmSalesOrderReceived, err)
		return err
	}
	return nil
}
