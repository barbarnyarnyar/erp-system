package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"testing"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
)

func TestCustomerService_All(t *testing.T) {
	repo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewCustomerService(repo, pub)

	ctx := context.Background()

	// 1. Create customer
	cust, err := svc.CreateCustomer(ctx, "Acme Inc", "John Doe", "john@acme.com", "12345", "Enterprise", "")
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}
	if cust.CompanyName != "Acme Inc" {
		t.Errorf("expected CompanyName 'Acme Inc', got %q", cust.CompanyName)
	}


	// Verify events published
	foundCreated := false
	foundActivated := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmCustomerCreated {
			foundCreated = true
		}
		if ev.Topic == domain.TopicCrmCustomerActivated {
			foundActivated = true
		}
	}
	if !foundCreated {
		t.Errorf("expected customer created event to be published")
	}
	if !foundActivated {
		t.Errorf("expected customer activated event to be published")
	}

	_, err = svc.CreateCustomer(ctx, "Beta Corp", "", "", "", "", "parent_id")
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}


	// 3. Get customer
	fetched, err := svc.GetCustomer(ctx, cust.ID)
	if err != nil {
		t.Fatalf("failed to get customer: %v", err)
	}
	if fetched.ID != cust.ID {
		t.Errorf("expected customer ID %q, got %q", cust.ID, fetched.ID)
	}

	// 4. List customers
	list, err := svc.ListCustomers(ctx)
	if err != nil {
		t.Fatalf("failed to list customers: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected list length 2, got %d", len(list))
	}

	// 5. Update customer success
	pub.Events = nil // clear events
	updated, err := svc.UpdateCustomer(ctx, cust.ID, "Acme Corp Updated", "Jane Doe", "jane@acme.com", "54321", "ACTIVE", "SME")
	if err != nil {
		t.Fatalf("failed to update customer: %v", err)
	}
	if updated.CompanyName != "Acme Corp Updated" {
		t.Errorf("expected CompanyName 'Acme Corp Updated', got %q", updated.CompanyName)
	}

	// Check updated event
	foundUpdated := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmCustomerUpdated {
			foundUpdated = true
		}
	}
	if !foundUpdated {
		t.Errorf("expected customer updated event to be published")
	}

	// 6. Update customer status to INACTIVE (trigger deactivation event)
	pub.Events = nil
	updated, err = svc.UpdateCustomer(ctx, cust.ID, "Acme Corp Updated", "Jane Doe", "jane@acme.com", "54321", "INACTIVE", "SME")
	if err != nil {
		t.Fatalf("failed to update customer: %v", err)
	}
	if updated.Status != domain.CustomerStatusINACTIVE {
		t.Errorf("expected status INACTIVE, got %q", updated.Status)
	}

	foundDeactivated := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmCustomerDeactivated {
			foundDeactivated = true
		}
	}
	if !foundDeactivated {
		t.Errorf("expected customer deactivated event to be published")
	}



	// 8. Delete customer
	err = svc.DeleteCustomer(ctx, cust.ID)
	if err != nil {
		t.Fatalf("failed to delete customer: %v", err)
	}

	// Verify deletion
	_, err = svc.GetCustomer(ctx, cust.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted customer, got nil")
	}
}

func TestCustomerService_UpdateInvalidStatus(t *testing.T) {
	repo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewCustomerService(repo, pub)

	ctx := context.Background()
	_, err := svc.UpdateCustomer(ctx, "some-id", "Acme", "John", "john@acme.com", "123", "INVALID_STATUS", "")
	if err == nil {
		t.Errorf("expected error with invalid customer status, got nil")
	}
}

func TestCustomerService_UpdateNotFound(t *testing.T) {
	repo := memory.NewCustomerRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewCustomerService(repo, pub)

	ctx := context.Background()
	_, err := svc.UpdateCustomer(ctx, "non-existent", "Acme", "John", "john@acme.com", "123", "ACTIVE", "")
	if err == nil {
		t.Errorf("expected error updating non-existent customer, got nil")
	}
}
