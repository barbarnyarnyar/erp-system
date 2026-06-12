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

func (f *failingCustomerRepo) GetByID(ctx context.Context, id string) (*domain.CustomerProfile, error) {
	return nil, f.err
}

func setupConfirmFixtures(t *testing.T) (svc *service.SalesOrderService, orderRepo domain.SalesOrderRepository, orderItemRepo domain.SalesOrderLineRepository, custRepo domain.CustomerRepository, pub *sharedtesting.MockPublisher) {
	t.Helper()
	orderRepo = memory.NewSalesOrderRepository()
	orderItemRepo = memory.NewSalesOrderLineRepository()
	custRepo = memory.NewCustomerRepository()
	pub = &sharedtesting.MockPublisher{}
	svc = service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, pub)
	return
}

func seedDraftOrderWithCustomer(t *testing.T, orderRepo domain.SalesOrderRepository, orderItemRepo domain.SalesOrderLineRepository, custRepo domain.CustomerRepository) (orderID, customerID string) {
	t.Helper()
	ctx := context.Background()
	customer := &domain.CustomerProfile{
		ID:                 "cust_1",
		LegalEntityID:      "default_entity_id",
		CustomerCode:       "CODE-cust_1",
		CompanyName:        "Acme",
		AccountManagerHrID: "default_manager_id",
		Status:             domain.CustomerStatusACTIVE,
		CreditLimit:        decimal.NewFromInt(50000),
		Currency:           "USD",
		Version:            1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if err := custRepo.Create(ctx, customer); err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	order := &domain.SalesOrder{
		ID:              "so_1",
		LegalEntityID:   "default_entity_id",
		CustomerID:      customer.ID,
		PriceBookID:     "default_price_book_id",
		OrderNumber:     "SO-so_1",
		Status:          domain.SalesOrderStateDRAFT,
		TotalGrossValue: decimal.NewFromInt(100),
		TotalTaxValue:   decimal.Zero,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	item := &domain.SalesOrderLine{
		ID:              "soi_1",
		SalesOrderID:    order.ID,
		MaterialID:      "prod_1",
		LineSequence:    10,
		QuantityOrdered: decimal.NewFromInt(2),
		QuantityShipped: decimal.Zero,
		UnitSellPrice:   decimal.NewFromInt(50),
		DiscountApplied: decimal.Zero,
		NetLineAmount:   decimal.NewFromInt(100),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
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
	if got.Status != domain.SalesOrderStatusConfirmed {
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
	order.Status = domain.SalesOrderStatusConfirmed
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
	cust.Status = domain.CustomerStatusINACTIVE
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
	orderItemRepo := memory.NewSalesOrderLineRepository()
	custRepo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, pub)

	ctx := context.Background()
	order := &domain.SalesOrder{
		ID:              "so_x",
		LegalEntityID:   "default_entity_id",
		CustomerID:      "ghost",
		PriceBookID:     "default_price_book_id",
		OrderNumber:     "SO-so_x",
		Status:          domain.SalesOrderStateDRAFT,
		TotalGrossValue: decimal.NewFromInt(10),
		TotalTaxValue:   decimal.Zero,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = orderRepo.Create(ctx, order)
	item := &domain.SalesOrderLine{
		ID:              "soi_x",
		SalesOrderID:    order.ID,
		MaterialID:      "p",
		LineSequence:    10,
		QuantityOrdered: decimal.NewFromInt(1),
		QuantityShipped: decimal.Zero,
		UnitSellPrice:   decimal.NewFromInt(10),
		DiscountApplied: decimal.Zero,
		NetLineAmount:   decimal.NewFromInt(10),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = orderItemRepo.Create(ctx, item)

	_, err := svc.ConfirmSalesOrder(ctx, order.ID)
	if !errors.Is(err, domain.ErrCustomerNotFound) {
		t.Errorf("err = %v, want ErrCustomerNotFound", err)
	}
}

func TestConfirmSalesOrder_EmptyItems(t *testing.T) {
	orderRepo := memory.NewSalesOrderRepository()
	orderItemRepo := memory.NewSalesOrderLineRepository()
	custRepo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, pub)

	ctx := context.Background()
	cust := &domain.CustomerProfile{
		ID:                 "c1",
		LegalEntityID:      "default_entity_id",
		CustomerCode:       "CODE-c1",
		CompanyName:        "Acme",
		AccountManagerHrID: "default_manager_id",
		Status:             domain.CustomerStatusACTIVE,
		CreditLimit:        decimal.NewFromInt(50000),
		Currency:           "USD",
		Version:            1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	_ = custRepo.Create(ctx, cust)
	order := &domain.SalesOrder{
		ID:              "so_e",
		LegalEntityID:   "default_entity_id",
		CustomerID:      "c1",
		PriceBookID:     "default_price_book_id",
		OrderNumber:     "SO-so_e",
		Status:          domain.SalesOrderStateDRAFT,
		TotalGrossValue: decimal.Zero,
		TotalTaxValue:   decimal.Zero,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = orderRepo.Create(ctx, order)

	_, err := svc.ConfirmSalesOrder(ctx, order.ID)
	if !errors.Is(err, domain.ErrOrderHasNoItems) {
		t.Errorf("err = %v, want ErrOrderHasNoItems", err)
	}
}

func TestConfirmSalesOrder_InvalidItemQuantity(t *testing.T) {
	orderRepo := memory.NewSalesOrderRepository()
	orderItemRepo := memory.NewSalesOrderLineRepository()
	custRepo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewSalesOrderService(orderRepo, orderItemRepo, custRepo, pub)

	ctx := context.Background()
	cust := &domain.CustomerProfile{
		ID:                 "c1",
		LegalEntityID:      "default_entity_id",
		CustomerCode:       "CODE-c1",
		CompanyName:        "Acme",
		AccountManagerHrID: "default_manager_id",
		Status:             domain.CustomerStatusACTIVE,
		CreditLimit:        decimal.NewFromInt(50000),
		Currency:           "USD",
		Version:            1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	_ = custRepo.Create(ctx, cust)
	order := &domain.SalesOrder{
		ID:              "so_q",
		LegalEntityID:   "default_entity_id",
		CustomerID:      "c1",
		PriceBookID:     "default_price_book_id",
		OrderNumber:     "SO-so_q",
		Status:          domain.SalesOrderStateDRAFT,
		TotalGrossValue: decimal.Zero,
		TotalTaxValue:   decimal.Zero,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = orderRepo.Create(ctx, order)
	item := &domain.SalesOrderLine{
		ID:              "soi_q",
		SalesOrderID:    order.ID,
		MaterialID:      "p",
		LineSequence:    10,
		QuantityOrdered: decimal.Zero,
		QuantityShipped: decimal.Zero,
		UnitSellPrice:   decimal.NewFromInt(10),
		DiscountApplied: decimal.Zero,
		NetLineAmount:   decimal.Zero,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = orderItemRepo.Create(ctx, item)

	_, err := svc.ConfirmSalesOrder(ctx, order.ID)
	if !errors.Is(err, domain.ErrInvalidItemQuantity) {
		t.Errorf("err = %v, want ErrInvalidItemQuantity", err)
	}
}
