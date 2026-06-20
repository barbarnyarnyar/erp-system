package service

import (
	"context"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type ReportService struct {
	prodRepo     domain.ProductRepository
	invRepo      domain.StockBalanceRepository
	supRepo      domain.SupplierRepository
	poRepo       domain.PurchaseOrderRepository
	moveRepo     domain.InventoryMovementRepository
	forecastRepo domain.DemandForecastRepository
}

func NewReportService(
	prodRepo domain.ProductRepository,
	invRepo domain.StockBalanceRepository,
	supRepo domain.SupplierRepository,
	poRepo domain.PurchaseOrderRepository,
	moveRepo domain.InventoryMovementRepository,
	forecastRepo domain.DemandForecastRepository,
) *ReportService {
	return &ReportService{
		prodRepo:     prodRepo,
		invRepo:      invRepo,
		supRepo:      supRepo,
		poRepo:       poRepo,
		moveRepo:     moveRepo,
		forecastRepo: forecastRepo,
	}
}

type StockLevel struct {
	ProductID      string          `json:"product_id"`
	ProductCode    string          `json:"product_code"`
	ProductName    string          `json:"product_name"`
	LocationID     string          `json:"location_id"`
	QuantityOnHand decimal.Decimal `json:"quantity_on_hand"`
	ReorderPoint   decimal.Decimal `json:"reorder_point"`
	IsCritical     bool            `json:"is_critical"`
	Valuation      decimal.Decimal `json:"valuation"`
}

func (s *ReportService) GetInventoryLevelsReport(ctx context.Context) ([]StockLevel, error) {
	products, err := s.prodRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.invRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	prodMap := make(map[string]domain.Product)
	for _, p := range products {
		prodMap[p.ID] = p
	}

	report := make([]StockLevel, 0, len(items))

	for _, item := range items {
		p, ok := prodMap[item.MaterialID]
		if !ok {
			continue
		}

		isCritical := item.QuantityOnHand.IsZero()
		val := p.StandardCost.Mul(item.QuantityOnHand)

		report = append(report, StockLevel{
			ProductID:      item.MaterialID,
			ProductCode:    p.ProductCode,
			ProductName:    p.ProductName,
			LocationID:     item.LocationID,
			QuantityOnHand: item.QuantityOnHand,
			ReorderPoint:   decimal.Zero,
			IsCritical:     isCritical,
			Valuation:      val,
		})
	}

	return report, nil
}

type SupplierPerformance struct {
	SupplierID      string          `json:"supplier_id"`
	SupplierCode    string          `json:"supplier_code"`
	SupplierName    string          `json:"supplier_name"`
	TotalOrders     int             `json:"total_orders"`
	CompletedOrders int             `json:"completed_orders"`
	CompletionRate  decimal.Decimal `json:"completion_rate"`
	TotalSpend      decimal.Decimal `json:"total_spend"`
}

func (s *ReportService) GetVendorPerformanceReport(ctx context.Context) ([]SupplierPerformance, error) {
	suppliers, err := s.supRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	orders, err := s.poRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	performanceMap := make(map[string]*SupplierPerformance)
	for _, sup := range suppliers {
		performanceMap[sup.ID] = &SupplierPerformance{
			SupplierID:     sup.ID,
			SupplierCode:   sup.SupplierCode,
			SupplierName:   sup.SupplierName,
			CompletionRate: decimal.Zero,
			TotalSpend:     decimal.Zero,
		}
	}

	for _, po := range orders {
		perf, ok := performanceMap[po.SupplierID]
		if !ok {
			continue
		}

		perf.TotalOrders++
		perf.TotalSpend = perf.TotalSpend.Add(po.TotalAmount)

		if po.Status == "DELIVERED" {
			perf.CompletedOrders++
		}
	}

	report := make([]SupplierPerformance, 0, len(suppliers))
	for _, perf := range performanceMap {
		if perf.TotalOrders > 0 {
			perf.CompletionRate = decimal.NewFromInt(int64(perf.CompletedOrders)).Div(decimal.NewFromInt(int64(perf.TotalOrders)))
		}
		report = append(report, *perf)
	}

	return report, nil
}

type ProcurementMetrics struct {
	TotalProcurementSpend decimal.Decimal `json:"total_procurement_spend"`
	AverageOrderAmount    decimal.Decimal `json:"average_order_amount"`
	OrdersCountByStatus   map[string]int  `json:"orders_count_by_status"`
}

func (s *ReportService) GetProcurementMetricsReport(ctx context.Context) (*ProcurementMetrics, error) {
	orders, err := s.poRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	metrics := &ProcurementMetrics{
		TotalProcurementSpend: decimal.Zero,
		AverageOrderAmount:    decimal.Zero,
		OrdersCountByStatus:   make(map[string]int),
	}

	for _, po := range orders {
		metrics.TotalProcurementSpend = metrics.TotalProcurementSpend.Add(po.TotalAmount)
		metrics.OrdersCountByStatus[string(po.Status)]++
	}

	if len(orders) > 0 {
		metrics.AverageOrderAmount = metrics.TotalProcurementSpend.Div(decimal.NewFromInt(int64(len(orders))))
	}

	return metrics, nil
}

type SafetyStockRecommendation struct {
	ProductID             string          `json:"product_id"`
	ProductCode           string          `json:"product_code"`
	ProductName           string          `json:"product_name"`
	QuantityOnHand        decimal.Decimal `json:"quantity_on_hand"`
	AverageForecastDemand decimal.Decimal `json:"average_forecast_demand"`
	CalculatedSafetyStock decimal.Decimal `json:"calculated_safety_stock"`
	CurrentReorderPoint   decimal.Decimal `json:"current_reorder_point"`
	Recommendation        string          `json:"recommendation"`
}

func (s *ReportService) GetSafetyStockReport(ctx context.Context) ([]SafetyStockRecommendation, error) {
	products, err := s.prodRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.invRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	itemMap := make(map[string]domain.StockBalance)
	for _, item := range items {
		itemMap[item.MaterialID] = item
	}

	report := make([]SafetyStockRecommendation, 0, len(products))

	for _, p := range products {
		item, hasItem := itemMap[p.ID]
		qtyOnHand := decimal.Zero
		if hasItem {
			qtyOnHand = item.QuantityOnHand
		}

		forecasts, err := s.forecastRepo.ListByMaterialID(ctx, p.ID)
		if err != nil {
			forecasts = nil
		}

		avgForecast := decimal.Zero
		if len(forecasts) > 0 {
			totalForecast := decimal.Zero
			for _, f := range forecasts {
				totalForecast = totalForecast.Add(f.ForecastQuantity)
			}
			avgForecast = totalForecast.Div(decimal.NewFromInt(int64(len(forecasts))))
		}

		// Simple calculated safety stock formula
		calcSafetyStock := avgForecast.Mul(decimal.NewFromFloat(1.25)).Add(decimal.NewFromInt(10))

		rec := "STOCK_LEVEL_OK"
		if qtyOnHand.LessThan(calcSafetyStock) {
			rec = "REORDER_IMMEDIATELY"
		} else if qtyOnHand.LessThan(calcSafetyStock.Add(decimal.NewFromInt(15))) {
			rec = "RESTOCK_SOON"
		}

		report = append(report, SafetyStockRecommendation{
			ProductID:             p.ID,
			ProductCode:           p.ProductCode,
			ProductName:           p.ProductName,
			QuantityOnHand:        qtyOnHand,
			AverageForecastDemand: avgForecast,
			CalculatedSafetyStock: calcSafetyStock,
			CurrentReorderPoint:   decimal.Zero,
			Recommendation:        rec,
		})
	}

	return report, nil
}
