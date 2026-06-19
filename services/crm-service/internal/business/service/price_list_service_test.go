package service_test

import (
	"context"
	"testing"

	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
)

func TestPriceListService_All(t *testing.T) {
	headerRepo := memory.NewPriceBookHeaderRepository()
	entryRepo := memory.NewPriceBookEntryRepository()
	svc := service.NewPriceListService(headerRepo, entryRepo)

	ctx := context.Background()

	// 1. Create Price List
	pl, err := svc.CreatePriceList(ctx, "Wholesale Price Book", "Standard wholesale price list", true)
	if err != nil {
		t.Fatalf("failed to create price list: %v", err)
	}
	if pl.Name != "Wholesale Price Book" {
		t.Errorf("expected name 'Wholesale Price Book', got %q", pl.Name)
	}
	if pl.IsActive != true {
		t.Errorf("expected IsActive true, got %t", pl.IsActive)
	}

	// 2. Get Price List
	fetched, err := svc.GetPriceList(ctx, pl.ID)
	if err != nil {
		t.Fatalf("failed to get price list: %v", err)
	}
	if fetched.ID != pl.ID {
		t.Errorf("expected price list ID %q, got %q", pl.ID, fetched.ID)
	}

	// 3. List Price Lists
	list, err := svc.ListPriceLists(ctx)
	if err != nil {
		t.Fatalf("failed to list price lists: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list length 1, got %d", len(list))
	}

	// 4. Update Price List
	updated, err := svc.UpdatePriceList(ctx, pl.ID, "Wholesale Price Book Updated", "Updated description", false)
	if err != nil {
		t.Fatalf("failed to update price list: %v", err)
	}
	if updated.Name != "Wholesale Price Book Updated" {
		t.Errorf("expected updated name, got %q", updated.Name)
	}
	if updated.IsActive != false {
		t.Errorf("expected IsActive false, got %t", updated.IsActive)
	}

	// 5. Delete Price List
	err = svc.DeletePriceList(ctx, pl.ID)
	if err != nil {
		t.Fatalf("failed to delete price list: %v", err)
	}

	// Verify deletion
	_, err = svc.GetPriceList(ctx, pl.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted price list, got nil")
	}
}

func TestPriceListService_UpdateNotFound(t *testing.T) {
	headerRepo := memory.NewPriceBookHeaderRepository()
	entryRepo := memory.NewPriceBookEntryRepository()
	svc := service.NewPriceListService(headerRepo, entryRepo)

	ctx := context.Background()
	_, err := svc.UpdatePriceList(ctx, "non-existent", "Test", "Desc", true)
	if err == nil {
		t.Errorf("expected error updating non-existent price list, got nil")
	}
}
