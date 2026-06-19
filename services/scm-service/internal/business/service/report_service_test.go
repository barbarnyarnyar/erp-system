package service

import (
	"context"
	"errors"
	"testing"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type MockReportServiceRepos struct {
	prodListErr     error
	invListErr      error
	supListErr      error
	poListErr       error
	forecastListErr error
}

type MockReportProductRepo struct {
	domain.ProductRepository
	err error
}

func (m *MockReportProductRepo) List(ctx context.Context) ([]domain.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ProductRepository.List(ctx)
}

type MockReportInventoryRepo struct {
	domain.InventoryItemRepository
	err error
}

func (m *MockReportInventoryRepo) List(ctx context.Context) ([]domain.InventoryItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.InventoryItemRepository.List(ctx)
}

type MockReportSupplierRepo struct {
	domain.SupplierRepository
	err error
}

func (m *MockReportSupplierRepo) List(ctx context.Context) ([]domain.Supplier, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.SupplierRepository.List(ctx)
}

type MockReportPORepo struct {
	domain.PurchaseOrderRepository
	err error
}

func (m *MockReportPORepo) List(ctx context.Context) ([]domain.PurchaseOrder, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.PurchaseOrderRepository.List(ctx)
}

type MockReportForecastRepo struct {
	domain.DemandForecastRepository
	listByProductErr error
}

func (m *MockReportForecastRepo) ListByProductID(ctx context.Context, productID string) ([]domain.DemandForecast, error) {
	if m.listByProductErr != nil {
		return nil, m.listByProductErr
	}
	return m.DemandForecastRepository.ListByProductID(ctx, productID)
}

func TestReportService_GetInventoryLevelsReport(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		prodRepo := memory.NewMemoryProductRepo()
		invRepo := memory.NewMemoryInventoryItemRepo()
		svc := NewReportService(prodRepo, invRepo, nil, nil, nil, nil)

		_ = prodRepo.Create(ctx, &domain.Product{ID: "p-1", ProductCode: "P1", ProductName: "Prod 1"})
		_ = invRepo.Create(ctx, &domain.InventoryItem{ID: "inv-1", ProductID: "p-1", LocationID: "loc-1", QuantityOnHand: 10, ReorderPoint: 5, UnitCost: decimal.NewFromFloat(12.5)})
		_ = invRepo.Create(ctx, &domain.InventoryItem{ID: "inv-2", ProductID: "p-nonexistent", LocationID: "loc-1"})

		report, err := svc.GetInventoryLevelsReport(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report) != 1 {
			t.Errorf("expected 1 item, got %d", len(report))
		}
		if report[0].Valuation.InexactFloat64() != 125.0 {
			t.Errorf("expected valuation 125, got %s", report[0].Valuation)
		}
		if report[0].IsCritical {
			t.Error("expected stock level to not be critical")
		}
	})

	t.Run("prod repo error", func(t *testing.T) {
		prodRepo := &MockReportProductRepo{err: errors.New("db error")}
		svc := NewReportService(prodRepo, nil, nil, nil, nil, nil)
		_, err := svc.GetInventoryLevelsReport(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("inv repo error", func(t *testing.T) {
		prodRepo := memory.NewMemoryProductRepo()
		invRepo := &MockReportInventoryRepo{err: errors.New("db error")}
		svc := NewReportService(prodRepo, invRepo, nil, nil, nil, nil)
		_, err := svc.GetInventoryLevelsReport(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestReportService_GetVendorPerformanceReport(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		supRepo := memory.NewMemorySupplierRepo()
		poRepo := memory.NewMemoryPurchaseOrderRepo()
		svc := NewReportService(nil, nil, supRepo, poRepo, nil, nil)

		_ = supRepo.Create(ctx, &domain.Supplier{ID: "sup-1", SupplierCode: "S1", SupplierName: "Supplier 1"})
		_ = supRepo.Create(ctx, &domain.Supplier{ID: "sup-2", SupplierCode: "S2", SupplierName: "Supplier 2"})

		_ = poRepo.Create(ctx, &domain.PurchaseOrder{ID: "po-1", SupplierID: "sup-1", TotalAmount: decimal.NewFromFloat(100.0), Status: "DELIVERED"})
		_ = poRepo.Create(ctx, &domain.PurchaseOrder{ID: "po-2", SupplierID: "sup-1", TotalAmount: decimal.NewFromFloat(150.0), Status: "CANCELLED"})
		_ = poRepo.Create(ctx, &domain.PurchaseOrder{ID: "po-3", SupplierID: "sup-nonexistent", TotalAmount: decimal.NewFromFloat(200.0)})

		report, err := svc.GetVendorPerformanceReport(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(report) != 2 {
			t.Errorf("expected 2 items, got %d", len(report))
		}

		var s1Perf *SupplierPerformance
		for i := range report {
			if report[i].SupplierID == "sup-1" {
				s1Perf = &report[i]
			}
		}

		if s1Perf == nil {
			t.Fatal("sup-1 performance not found")
		}
		if s1Perf.TotalOrders != 2 {
			t.Errorf("expected 2 orders, got %d", s1Perf.TotalOrders)
		}
		if s1Perf.CompletedOrders != 1 {
			t.Errorf("expected 1 completed, got %d", s1Perf.CompletedOrders)
		}
		if !s1Perf.CompletionRate.Equal(decimal.NewFromFloat(0.5)) {
			t.Errorf("expected 0.5 completion rate, got %s", s1Perf.CompletionRate)
		}
		if !s1Perf.TotalSpend.Equal(decimal.NewFromFloat(250.0)) {
			t.Errorf("expected 250.0 total spend, got %s", s1Perf.TotalSpend)
		}
	})

	t.Run("sup repo error", func(t *testing.T) {
		supRepo := &MockReportSupplierRepo{err: errors.New("db error")}
		svc := NewReportService(nil, nil, supRepo, nil, nil, nil)
		_, err := svc.GetVendorPerformanceReport(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("po repo error", func(t *testing.T) {
		supRepo := memory.NewMemorySupplierRepo()
		poRepo := &MockReportPORepo{err: errors.New("db error")}
		svc := NewReportService(nil, nil, supRepo, poRepo, nil, nil)
		_, err := svc.GetVendorPerformanceReport(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestReportService_GetProcurementMetricsReport(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		poRepo := memory.NewMemoryPurchaseOrderRepo()
		svc := NewReportService(nil, nil, nil, poRepo, nil, nil)

		_ = poRepo.Create(ctx, &domain.PurchaseOrder{ID: "po-1", TotalAmount: decimal.NewFromFloat(100.0), Status: "APPROVED"})
		_ = poRepo.Create(ctx, &domain.PurchaseOrder{ID: "po-2", TotalAmount: decimal.NewFromFloat(200.0), Status: "DELIVERED"})

		metrics, err := svc.GetProcurementMetricsReport(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !metrics.TotalProcurementSpend.Equal(decimal.NewFromFloat(300.0)) {
			t.Errorf("expected 300 total spend, got %s", metrics.TotalProcurementSpend)
		}
		if !metrics.AverageOrderAmount.Equal(decimal.NewFromFloat(150.0)) {
			t.Errorf("expected 150 average, got %s", metrics.AverageOrderAmount)
		}
		if metrics.OrdersCountByStatus["APPROVED"] != 1 || metrics.OrdersCountByStatus["DELIVERED"] != 1 {
			t.Errorf("unexpected orders count: %v", metrics.OrdersCountByStatus)
		}
	})

	t.Run("empty orders", func(t *testing.T) {
		poRepo := memory.NewMemoryPurchaseOrderRepo()
		svc := NewReportService(nil, nil, nil, poRepo, nil, nil)
		metrics, err := svc.GetProcurementMetricsReport(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !metrics.AverageOrderAmount.IsZero() {
			t.Errorf("expected average 0, got %s", metrics.AverageOrderAmount)
		}
	})

	t.Run("po repo error", func(t *testing.T) {
		poRepo := &MockReportPORepo{err: errors.New("db error")}
		svc := NewReportService(nil, nil, nil, poRepo, nil, nil)
		_, err := svc.GetProcurementMetricsReport(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestReportService_GetSafetyStockReport(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		prodRepo := memory.NewMemoryProductRepo()
		invRepo := memory.NewMemoryInventoryItemRepo()
		forecastRepo := memory.NewMemoryDemandForecastRepo()
		svc := NewReportService(prodRepo, invRepo, nil, nil, nil, forecastRepo)

		_ = prodRepo.Create(ctx, &domain.Product{ID: "p-1", ProductCode: "P1", ProductName: "Product 1"})
		_ = prodRepo.Create(ctx, &domain.Product{ID: "p-2", ProductCode: "P2", ProductName: "Product 2"})
		_ = prodRepo.Create(ctx, &domain.Product{ID: "p-3", ProductCode: "P3", ProductName: "Product 3"})

		// P1: StockLevel = OK
		_ = invRepo.Create(ctx, &domain.InventoryItem{ID: "inv-1", ProductID: "p-1", QuantityOnHand: 100, ReorderPoint: 10})
		// P2: StockLevel = RESTOCK_SOON
		_ = invRepo.Create(ctx, &domain.InventoryItem{ID: "inv-2", ProductID: "p-2", QuantityOnHand: 25, ReorderPoint: 10})
		// P3: StockLevel = REORDER_IMMEDIATELY (no inventory seeded, qtyOnHand=0)

		// Seed Forecasts
		_ = forecastRepo.Create(ctx, &domain.DemandForecast{ID: "f-1", ProductID: "p-1", ForecastQuantity: 10})
		_ = forecastRepo.Create(ctx, &domain.DemandForecast{ID: "f-2", ProductID: "p-1", ForecastQuantity: 20})
		_ = forecastRepo.Create(ctx, &domain.DemandForecast{ID: "f-3", ProductID: "p-2", ForecastQuantity: 10})

		report, err := svc.GetSafetyStockReport(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report) != 3 {
			t.Errorf("expected 3 items, got %d", len(report))
		}

		var p1Rec, p2Rec, p3Rec *SafetyStockRecommendation
		for i := range report {
			if report[i].ProductID == "p-1" {
				p1Rec = &report[i]
			} else if report[i].ProductID == "p-2" {
				p2Rec = &report[i]
			} else if report[i].ProductID == "p-3" {
				p3Rec = &report[i]
			}
		}

		if p1Rec.Recommendation != "STOCK_LEVEL_OK" {
			t.Errorf("expected p1 STOCK_LEVEL_OK, got %s", p1Rec.Recommendation)
		}
		if p2Rec.Recommendation != "RESTOCK_SOON" {
			t.Errorf("expected p2 RESTOCK_SOON, got %s", p2Rec.Recommendation)
		}
		if p3Rec.Recommendation != "REORDER_IMMEDIATELY" {
			t.Errorf("expected p3 REORDER_IMMEDIATELY, got %s", p3Rec.Recommendation)
		}
	})

	t.Run("forecast list error (falls back to nil forecasts)", func(t *testing.T) {
		prodRepo := memory.NewMemoryProductRepo()
		invRepo := memory.NewMemoryInventoryItemRepo()
		forecastRepo := &MockReportForecastRepo{
			listByProductErr: errors.New("forecast fetch error"),
		}
		svc := NewReportService(prodRepo, invRepo, nil, nil, nil, forecastRepo)

		_ = prodRepo.Create(ctx, &domain.Product{ID: "p-1", ProductCode: "P1", ProductName: "Product 1"})

		report, err := svc.GetSafetyStockReport(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report) != 1 {
			t.Errorf("expected 1 item, got %d", len(report))
		}
		// avgForecast should be 0, calculated safety stock should be 10, qty=0, so REORDER_IMMEDIATELY
		if report[0].Recommendation != "REORDER_IMMEDIATELY" {
			t.Errorf("expected REORDER_IMMEDIATELY, got %s", report[0].Recommendation)
		}
	})

	t.Run("prod repo error", func(t *testing.T) {
		prodRepo := &MockReportProductRepo{err: errors.New("db error")}
		svc := NewReportService(prodRepo, nil, nil, nil, nil, nil)
		_, err := svc.GetSafetyStockReport(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("inv repo error", func(t *testing.T) {
		prodRepo := memory.NewMemoryProductRepo()
		invRepo := &MockReportInventoryRepo{err: errors.New("db error")}
		svc := NewReportService(prodRepo, invRepo, nil, nil, nil, nil)
		_, err := svc.GetSafetyStockReport(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
