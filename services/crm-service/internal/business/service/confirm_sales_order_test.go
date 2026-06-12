package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type failingCustomerRepo struct {
	domain.CustomerRepository
	err error
}

func (f *failingCustomerRepo) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	return nil, f.err
}

func setupConfirmFixtures(t *testing.T) (svc *service.SalesOrderService, orderRepo domain.SalesOrderRepository, orderItemRepo domain.SalesOrderItemRepository, custRepo domain.CustomerRepository, pub *sharedtesting.MockPublisher) {
	t.Helper()
	orderRepo = memory.NewSalesOrderRepository()
	orderItemRepo = memory.NewSalesOrderItemRepository()
	custRepo = memory.NewCustomerRepository()
	pub = &sharedtesting.MockPublisher{}
	svc = service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, pub)
	return
}

func seedDraftOrderWithCustomer(t *testing.T, orderRepo domain.SalesOrderRepository, orderItemRepo domain.SalesOrderItemRepository, custRepo domain.CustomerRepository) (orderID, customerID string) {
	t.Helper()
	ctx := context.Background()
	customer := &domain.Customer{
		ID:          "cust_1",
		CompanyName: "Acme",
		ContactName: "John",
		Email:       "john@acme.com",
		Status:      domain.CustomerStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := custRepo.Create(ctx, customer); err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	order := &domain.SalesOrder{
		ID:          "so_1",
		CustomerID:  customer.ID,
		OrderDate:   time.Now(),
		Status:      string(domain.SalesOrderStatusDraft),
		TotalAmount: decimal.NewFromInt(100),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	item := &domain.SalesOrderItem{
		ID:           "soi_1",
		SalesOrderID: order.ID,
		ProductID:    "prod_1",
		Quantity:     2,
		UnitPrice:    decimal.NewFromInt(50),
	}
	if err := orderItemRepo.Create(ctx, item); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	return order.ID, customer.ID
}

func TestConfirmSalesOrder_Success(t *testing.T) {
	svc, orderRepo, orderItemRepo, custRepo, pub := setupConfirmFixtures(t)
	orderID, _ := seedDraftOrderWithCustomer(t, orderRepo, orderItemRepo, custRepo)

	got, err := svc.ConfirmSalesOrder(context.Background(), orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != string(domain.SalesOrderStatusConfirmed) {
		t.Errorf("status = %q, want %q", got.Status, domain.SalesOrderStatusConfirmed)
	}

	foundConfirmed := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmSalesOrderConfirmed {
			foundConfirmed = true
		}
	}
	if !foundConfirmed {
		t.Errorf("expected %s event to be fired", domain.TopicCrmSalesOrderConfirmed)
	}
}

func TestConfirmSalesOrder_NotFound(t *testing.T) {
	svc, _, _, _, _ := setupConfirmFixtures(t)
	_, err := svc.ConfirmSalesOrder(context.Background(), "missing")
	if !errors.Is(err, domain.ErrOrderNotFound) {
		t.Errorf("err = %v, want ErrOrderNotFound", err)
	}
}

func TestConfirmSalesOrder_NotDraft(t *testing.T) {
	svc, orderRepo, orderItemRepo, custRepo, _ := setupConfirmFixtures(t)
	orderID, _ := seedDraftOrderWithCustomer(t, orderRepo, orderItemRepo, custRepo)

	order, _ := orderRepo.GetByID(context.Background(), orderID)
	order.Status = string(domain.SalesOrderStatusConfirmed)
	if err := orderRepo.Update(context.Background(), order); err != nil {
		t.Fatalf("set confirmed: %v", err)
	}

	_, err := svc.ConfirmSalesOrder(context.Background(), orderID)
	if !errors.Is(err, domain.ErrOrderNotConfirmable) {
		t.Errorf("err = %v, want ErrOrderNotConfirmable", err)
	}
}

func TestConfirmSalesOrder_CustomerInactive(t *testing.T) {
	svc, orderRepo, orderItemRepo, custRepo, _ := setupConfirmFixtures(t)
	orderID, customerID := seedDraftOrderWithCustomer(t, orderRepo, orderItemRepo, custRepo)

	cust, _ := custRepo.GetByID(context.Background(), customerID)
	cust.Status = domain.CustomerStatusInactive
	if err := custRepo.Update(context.Background(), cust); err != nil {
		t.Fatalf("deactivate customer: %v", err)
	}

	_, err := svc.ConfirmSalesOrder(context.Background(), orderID)
	if !errors.Is(err, domain.ErrCustomerNotActive) {
		t.Errorf("err = %v, want ErrCustomerNotActive", err)
	}
}

func TestConfirmSalesOrder_CustomerNotFound(t *testing.T) {
	orderRepo := memory.NewSalesOrderRepository()
	orderItemRepo := memory.NewSalesOrderItemRepository()
	custRepo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, pub)

	ctx := context.Background()
	order := &domain.SalesOrder{
		ID:          "so_x",
		CustomerID:  "ghost",
		OrderDate:   time.Now(),
		Status:      string(domain.SalesOrderStatusDraft),
		TotalAmount: decimal.NewFromInt(10),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = orderRepo.Create(ctx, order)
	item := &domain.SalesOrderItem{ID: "soi_x", SalesOrderID: order.ID, ProductID: "p", Quantity: 1, UnitPrice: decimal.NewFromInt(10)}
	_ = orderItemRepo.Create(ctx, item)

	_, err := svc.ConfirmSalesOrder(ctx, order.ID)
	if !errors.Is(err, domain.ErrCustomerNotFound) {
		t.Errorf("err = %v, want ErrCustomerNotFound", err)
	}
}

func TestConfirmSalesOrder_EmptyItems(t *testing.T) {
	orderRepo := memory.NewSalesOrderRepository()
	orderItemRepo := memory.NewSalesOrderItemRepository()
	custRepo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, pub)

	ctx := context.Background()
	cust := &domain.Customer{ID: "c1", Status: domain.CustomerStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = custRepo.Create(ctx, cust)
	order := &domain.SalesOrder{
		ID: "so_e", CustomerID: "c1", OrderDate: time.Now(),
		Status: string(domain.SalesOrderStatusDraft), TotalAmount: decimal.Zero,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_ = orderRepo.Create(ctx, order)

	_, err := svc.ConfirmSalesOrder(ctx, order.ID)
	if !errors.Is(err, domain.ErrOrderHasNoItems) {
		t.Errorf("err = %v, want ErrOrderHasNoItems", err)
	}
}

func TestConfirmSalesOrder_InvalidItemQuantity(t *testing.T) {
	orderRepo := memory.NewSalesOrderRepository()
	orderItemRepo := memory.NewSalesOrderItemRepository()
	custRepo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, pub)

	ctx := context.Background()
	cust := &domain.Customer{ID: "c1", Status: domain.CustomerStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = custRepo.Create(ctx, cust)
	order := &domain.SalesOrder{
		ID: "so_q", CustomerID: "c1", OrderDate: time.Now(),
		Status: string(domain.SalesOrderStatusDraft), TotalAmount: decimal.Zero,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_ = orderRepo.Create(ctx, order)
	item := &domain.SalesOrderItem{ID: "soi_q", SalesOrderID: order.ID, ProductID: "p", Quantity: 0, UnitPrice: decimal.NewFromInt(10)}
	_ = orderItemRepo.Create(ctx, item)

	_, err := svc.ConfirmSalesOrder(ctx, order.ID)
	if !errors.Is(err, domain.ErrInvalidItemQuantity) {
		t.Errorf("err = %v, want ErrInvalidItemQuantity", err)
	}
}
