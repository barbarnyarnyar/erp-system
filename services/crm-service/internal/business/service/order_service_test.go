package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"testing"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func TestSalesOrderService_All(t *testing.T) {
	orderRepo := memory.NewSalesOrderRepository()
	orderLineRepo := memory.NewSalesOrderLineRepository()
	custRepo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewSalesOrderService(orderRepo, orderLineRepo, custRepo, pub)

	ctx := context.Background()

	// 1. Create Sales Order
	items := []service.SalesOrderItemInput{
		{
			ProductID: "prod_1",
			Quantity:  2,
			UnitPrice: decimal.NewFromInt(50),
			Discount:  decimal.NewFromInt(10),
		},
	}
	order, err := svc.CreateSalesOrder(ctx, "cust_1", items)
	if err != nil {
		t.Fatalf("failed to create sales order: %v", err)
	}
	if order.CustomerID != "cust_1" {
		t.Errorf("expected customer ID 'cust_1', got %q", order.CustomerID)
	}
	expectedTotal := decimal.NewFromInt(90) // 2 * 50 - 10
	if !order.TotalGrossValue.Equal(expectedTotal) {
		t.Errorf("expected total gross value %s, got %s", expectedTotal, order.TotalGrossValue)
	}

	// Verify event
	foundCreated := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmSalesOrderCreated {
			foundCreated = true
		}
	}
	if !foundCreated {
		t.Errorf("expected sales order created event to be published")
	}

	// Verify order lines
	lines, err := orderLineRepo.ListByOrderID(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to list order lines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 order line, got %d", len(lines))
	}
	if lines[0].MaterialID != "prod_1" {
		t.Errorf("expected material ID 'prod_1', got %q", lines[0].MaterialID)
	}

	// 2. Get Sales Order
	fetched, err := svc.GetSalesOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to get order: %v", err)
	}
	if fetched.ID != order.ID {
		t.Errorf("expected order ID %q, got %q", order.ID, fetched.ID)
	}

	// 3. List Sales Orders
	list, err := svc.ListSalesOrders(ctx)
	if err != nil {
		t.Fatalf("failed to list orders: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list length 1, got %d", len(list))
	}

	// 4. Update Sales Order status to CONFIRMED
	pub.Events = nil
	updated, err := svc.UpdateSalesOrder(ctx, order.ID, "CONFIRMED")
	if err != nil {
		t.Fatalf("failed to update order: %v", err)
	}
	if updated.Status != "CONFIRMED" {
		t.Errorf("expected status CONFIRMED, got %q", updated.Status)
	}
	foundConfirmed := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmSalesOrderConfirmed {
			foundConfirmed = true
		}
	}
	if !foundConfirmed {
		t.Errorf("expected sales order confirmed event to be published")
	}

	// 5. Update Sales Order status to SHIPPED
	pub.Events = nil
	updated, err = svc.UpdateSalesOrder(ctx, order.ID, "SHIPPED")
	if err != nil {
		t.Fatalf("failed to update order to SHIPPED: %v", err)
	}
	foundShipped := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmSalesOrderShipped {
			foundShipped = true
		}
	}
	if !foundShipped {
		t.Errorf("expected sales order shipped event to be published")
	}

	// 6. Update Sales Order status to DELIVERED
	pub.Events = nil
	updated, err = svc.UpdateSalesOrder(ctx, order.ID, "DELIVERED")
	if err != nil {
		t.Fatalf("failed to update order to DELIVERED: %v", err)
	}
	foundDelivered := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmSalesOrderDelivered {
			foundDelivered = true
		}
	}
	if !foundDelivered {
		t.Errorf("expected sales order delivered event to be published")
	}

	// 7. Update Sales Order status to CANCELLED
	pub.Events = nil
	updated, err = svc.UpdateSalesOrder(ctx, order.ID, "CANCELLED")
	if err != nil {
		t.Fatalf("failed to update order to CANCELLED: %v", err)
	}
	foundCancelled := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmSalesOrderCancelled {
			foundCancelled = true
		}
	}
	if !foundCancelled {
		t.Errorf("expected sales order cancelled event to be published")
	}

	// 8. Delete Sales Order
	pub.Events = nil
	err = svc.DeleteSalesOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to delete sales order: %v", err)
	}
	foundDeletedCancel := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmSalesOrderCancelled {
			foundDeletedCancel = true
		}
	}
	if !foundDeletedCancel {
		t.Errorf("expected sales order cancelled event during deletion to be published")
	}

	// 9. Receive Sales Order
	pub.Events = nil
	err = svc.ReceiveSalesOrder(ctx, "so_received", "cust_rec", decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("failed to receive sales order: %v", err)
	}
	foundReceived := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmSalesOrderReceived {
			foundReceived = true
		}
	}
	if !foundReceived {
		t.Errorf("expected sales order received event to be published")
	}
}

func TestSalesOrderService_Errors(t *testing.T) {
	orderRepo := memory.NewSalesOrderRepository()
	orderLineRepo := memory.NewSalesOrderLineRepository()
	custRepo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewSalesOrderService(orderRepo, orderLineRepo, custRepo, pub)

	ctx := context.Background()

	// Update non-existent
	_, err := svc.UpdateSalesOrder(ctx, "non-existent", "CONFIRMED")
	if err == nil {
		t.Errorf("expected error updating non-existent order, got nil")
	}
}
