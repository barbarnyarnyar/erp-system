package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/scm-service/internal/api/handlers"
	"github.com/erp-system/scm-service/internal/api/routes"
	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/erp-system/scm-service/internal/data/sql"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
	utils.InitLogger("scm-service-test")
}

type mockPublisher struct{}

func (m *mockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

type testEnv struct {
	router *gin.Engine
	db     *gorm.DB
}

func setupTestEnv(t *testing.T) *testEnv {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&sql.ProductCategory{},
		&sql.Product{},
		&sql.Location{},
		&sql.Supplier{},
		&sql.VendorContract{},
		&sql.InventoryItem{},
		&sql.InventoryMovement{},
		&sql.StockTransfer{},
		&sql.PurchaseRequisition{},
		&sql.PurchaseRequisitionLine{},
		&sql.PurchaseOrder{},
		&sql.PurchaseOrderLine{},
		&sql.Receipt{},
		&sql.ReceiptLine{},
		&sql.Shipment{},
		&sql.ShipmentLine{},
		&sql.DemandForecast{},
		&sql.KafkaEventInbox{},
		&sql.TransactionalOutbox{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// SQL Repositories
	prodRepo := sql.NewSQLProductRepo(db)
	catRepo := sql.NewSQLProductCategoryRepo(db)
	locRepo := sql.NewSQLLocationRepo(db)
	supRepo := sql.NewSQLSupplierRepo(db)
	contRepo := sql.NewSQLVendorContractRepo(db)
	invRepo := sql.NewSQLInventoryItemRepo(db)
	moveRepo := sql.NewSQLInventoryMovementRepo(db)
	poRepo := sql.NewSQLPurchaseOrderRepo(db)
	lineRepo := sql.NewSQLPurchaseOrderLineRepo(db)
	reqRepo := sql.NewSQLPurchaseRequisitionRepo(db)
	reqLineRepo := sql.NewSQLPurchaseRequisitionLineRepo(db)
	recRepo := sql.NewSQLReceiptRepo(db)
	recLRepo := sql.NewSQLReceiptLineRepo(db)
	shipRepo := sql.NewSQLShipmentRepo(db)
	shipLRepo := sql.NewSQLShipmentLineRepo(db)
	forecastRepo := sql.NewSQLDemandForecastRepo(db)
	transferRepo := sql.NewSQLStockTransferRepo(db)

	publisher := &mockPublisher{}
	tm := sql.NewGORMTransactionManager(db)

	// Seed default warehouse location
	_ = locRepo.Create(context.Background(), &domain.Location{
		ID:           "loc_default",
		LocationCode: "WH-MAIN",
		LocationName: "Main Distribution Center",
		LocationType: "WAREHOUSE",
		IsActive:     true,
	})

	prodSvc := service.NewProductManagementService(prodRepo, catRepo, locRepo, publisher)
	supSvc := service.NewSupplierManagementService(supRepo, contRepo, publisher)
	poSvc := service.NewPurchaseOrderService(poRepo, lineRepo, reqRepo, reqLineRepo, publisher, tm)
	invSvc := service.NewInventoryService(invRepo, moveRepo, transferRepo, publisher, tm)
	whSvc := service.NewWarehouseService(recRepo, recLRepo, shipRepo, shipLRepo, poRepo, lineRepo, invSvc, publisher, tm)
	demandSvc := service.NewDemandPlanningService(forecastRepo)
	reportSvc := service.NewReportService(prodRepo, invRepo, supRepo, poRepo, moveRepo, forecastRepo)

	responseHelper := utils.NewResponseHelper("scm-service")

	prodHandler := handlers.NewProductHandler(prodSvc, responseHelper)
	vendorHandler := handlers.NewVendorHandler(supSvc, responseHelper)
	poHandler := handlers.NewPurchaseOrderHandler(poSvc, responseHelper)
	invHandler := handlers.NewInventoryHandler(invSvc, responseHelper)
	whHandler := handlers.NewWarehouseHandler(whSvc, responseHelper)
	demandHandler := handlers.NewDemandForecastHandler(demandSvc, responseHelper)
	reportHandler := handlers.NewReportHandler(reportSvc, responseHelper)

	router := gin.New()
	routes.RegisterRoutes(router, prodHandler, vendorHandler, poHandler, invHandler, whHandler, demandHandler, reportHandler)

	return &testEnv{
		router: router,
		db:     db,
	}
}

func TestProductCategoryEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Create Category
	body, _ := json.Marshal(map[string]interface{}{
		"code":        "CAT001",
		"name":        "Raw Materials",
		"description": "Base manufacturing items",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/product-categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res struct {
		Data domain.ProductCategory `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	catID := res.Data.ID

	// Get Categories List
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/product-categories", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Category
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/product-categories/"+catID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Category
	body, _ = json.Marshal(map[string]interface{}{
		"name":        "Raw Materials Updated",
		"description": "Updated description",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/product-categories/"+catID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete Category
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/product-categories/"+catID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Category (404)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/product-categories/"+catID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestProductEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Create Product
	body, _ := json.Marshal(map[string]interface{}{
		"product_code":    "PROD001",
		"product_name":    "Iron Rod",
		"description":     "Structural iron rod",
		"product_type":    "RAW_MATERIAL",
		"unit_of_measure": "PCS",
		"standard_cost":   "10.50",
		"list_price":      "15.00",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res struct {
		Data domain.Product `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	prodID := res.Data.ID

	// List Products
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/products", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Product
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/products/"+prodID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Product
	body, _ = json.Marshal(map[string]interface{}{
		"product_code":    "PROD001",
		"product_name":    "Iron Rod Premium",
		"description":     "Premium structural iron rod",
		"product_type":    "RAW_MATERIAL",
		"unit_of_measure": "PCS",
		"standard_cost":   "12.00",
		"list_price":      "18.00",
		"is_active":       true,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/products/"+prodID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete Product
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/products/"+prodID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Product (404)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/products/"+prodID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestLocationEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Create Location
	body, _ := json.Marshal(map[string]interface{}{
		"location_code": "LOC001",
		"location_name": "Aisle A",
		"location_type": "AISLE",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/locations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res struct {
		Data domain.Location `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	locID := res.Data.ID

	// List Locations
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/locations", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Location
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/locations/"+locID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Location
	body, _ = json.Marshal(map[string]interface{}{
		"location_code": "LOC001",
		"location_name": "Aisle A - Updated",
		"location_type": "AISLE",
		"is_active":     true,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/locations/"+locID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete Location
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/locations/"+locID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Location (404)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/locations/"+locID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestVendorEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Create Vendor
	body, _ := json.Marshal(map[string]interface{}{
		"supplier_code": "VEND001",
		"supplier_name": "Acme Metal Corp",
		"contact_name":  "Wile E. Coyote",
		"email":         "wile@acme.com",
		"phone":         "555-0199",
		"payment_terms": "NET30",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vendors", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res struct {
		Data domain.Supplier `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	vendorID := res.Data.ID

	// List Vendors
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vendors", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Vendor
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vendors/"+vendorID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Vendor
	body, _ = json.Marshal(map[string]interface{}{
		"supplier_code": "VEND001",
		"supplier_name": "Acme Metal Corp Ltd",
		"payment_terms": "NET60",
		"is_active":     true,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/vendors/"+vendorID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Create Vendor Contract
	contractBody, _ := json.Marshal(map[string]interface{}{
		"supplier_id":     vendorID,
		"contract_number": "CON-2026-001",
		"start_date":      "2026-06-01",
		"end_date":        "2027-06-01",
		"terms":           "Standard terms",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/vendor-contracts", bytes.NewBuffer(contractBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var contractRes struct {
		Data domain.VendorContract `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &contractRes)
	contractID := contractRes.Data.ID

	// Get Contracts
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vendor-contracts", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Contract
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vendor-contracts/"+contractID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Contract
	contractBody, _ = json.Marshal(map[string]interface{}{
		"contract_number": "CON-2026-001-REV1",
		"supplier_id":     vendorID,
		"start_date":      "2026-06-01",
		"end_date":        "2027-06-01",
		"terms":           "Standard terms revised",
		"status":          "ACTIVE",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/vendor-contracts/"+contractID, bytes.NewBuffer(contractBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Delete Contract
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/vendor-contracts/"+contractID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete Vendor
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/vendors/"+vendorID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestPurchaseOrderEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Create Supplier
	supplier := &sql.Supplier{
		ID:           "sup-123",
		SupplierCode: "SUPP-01",
		SupplierName: "Acme Sup",
		IsActive:     true,
	}
	_ = env.db.Create(supplier).Error

	// Create Product
	product := &sql.Product{
		ID:          "prod-123",
		ProductCode: "ROD-01",
		ProductName: "Steel Rod",
		IsActive:    true,
	}
	_ = env.db.Create(product).Error

	// Create Requisition
	body, _ := json.Marshal(map[string]interface{}{
		"requisition_number": "REQ-100",
		"requested_by":       "emp-001",
		"department_id":      "dept-01",
		"lines": []map[string]interface{}{
			{
				"product_id":  "prod-123",
				"quantity":    10,
				"target_date": time.Now().AddDate(0, 0, 7).Format(time.RFC3339),
			},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/purchase-requisitions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var reqRes struct {
		Data domain.PurchaseRequisition `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &reqRes)
	reqID := reqRes.Data.ID

	// Approve Requisition
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/purchase-requisitions/"+reqID+"/approve", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Create Purchase Order
	poBody, _ := json.Marshal(map[string]interface{}{
		"supplier_id": "sup-123",
		"order_date":  time.Now().Format(time.RFC3339),
		"lines": []map[string]interface{}{
			{
				"product_id": "prod-123",
				"quantity":   5,
				"unit_price": "12.50",
			},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/purchase-orders", bytes.NewBuffer(poBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var poRes struct {
		Data domain.PurchaseOrder `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &poRes)
	poID := poRes.Data.ID

	// Send Purchase Order
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/purchase-orders/"+poID+"/send", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get PO list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/purchase-orders", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get PO lines
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/purchase-orders/"+poID+"/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestInventoryAndTransferEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Create Item
	body, _ := json.Marshal(map[string]interface{}{
		"product_id":     "prod-123",
		"location_id":    "loc_default",
		"quantity_on_hand": 100,
		"reorder_point":  10,
		"maximum_stock":  500,
		"unit_cost":      "5.00",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/inventory", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var invRes struct {
		Data domain.InventoryItem `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &invRes)
	invID := invRes.Data.ID

	// Get Inventory item
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/inventory/"+invID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Reserve Stock
	resBody, _ := json.Marshal(map[string]interface{}{
		"product_id":  "prod-123",
		"location_id": "loc_default",
		"quantity":    15,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/inventory/reserve", bytes.NewBuffer(resBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Release Reservation
	relBody, _ := json.Marshal(map[string]interface{}{
		"product_id":  "prod-123",
		"location_id": "loc_default",
		"quantity":    10,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/inventory/release", bytes.NewBuffer(relBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Inventory Movements
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/inventory/movements", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Create Stock Transfer
	stBody, _ := json.Marshal(map[string]interface{}{
		"product_id":       "prod-123",
		"from_location_id": "loc_default",
		"to_location_id":   "loc_default", // same or diff
		"quantity":         5,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/stock-transfers", bytes.NewBuffer(stBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestWarehouseAndForecastEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Create Receipt
	recBody, _ := json.Marshal(map[string]interface{}{
		"purchase_order_id": "po-123",
		"received_by":       "emp-001",
		"received_date":     time.Now().Format(time.RFC3339),
		"lines": []map[string]interface{}{
			{
				"purchase_order_line_id": "pol-123",
				"quantity_received":      10,
			},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/receipts", bytes.NewBuffer(recBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var recRes struct {
		Data domain.Receipt `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &recRes)
	recID := recRes.Data.ID

	// Get Receipt
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/receipts/"+recID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Create Shipment
	shipBody, _ := json.Marshal(map[string]interface{}{
		"sales_order_id": "so-123",
		"shipped_by":      "emp-001",
		"shipped_date":    time.Now().Format(time.RFC3339),
		"lines": []map[string]interface{}{
			{
				"sales_order_line_id": "sol-123",
				"quantity_shipped":    10,
			},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/shipments", bytes.NewBuffer(shipBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var shipRes struct {
		Data domain.Shipment `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &shipRes)
	shipID := shipRes.Data.ID

	// Get Shipment
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/shipments/"+shipID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Create Forecast
	fcBody, _ := json.Marshal(map[string]interface{}{
		"product_id":       "prod-123",
		"forecast_period":  "2026-06",
		"forecast_quantity": 250,
		"confidence_level": "0.85",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/demand-forecasts", bytes.NewBuffer(fcBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var fcRes struct {
		Data domain.DemandForecast `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &fcRes)
	fcID := fcRes.Data.ID

	// Get Forecast
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/demand-forecasts/"+fcID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// List Forecasts
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/demand-forecasts", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestReportsEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	reports := []string{
		"/api/v1/reports/inventory-levels",
		"/api/v1/reports/vendor-performance",
		"/api/v1/reports/procurement-metrics",
		"/api/v1/reports/safety-stock",
	}

	for _, url := range reports {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		env.router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 for %s, got %d", url, w.Code)
		}
	}
}

func TestScmErrorPaths(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Product Category 404
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/product-categories/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 2. Create Category validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/product-categories", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 3. Product 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/products/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 4. Location 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/locations/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/locations", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 5. Vendor contract 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vendor-contracts/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/vendor-contracts", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 6. Purchase Requisition 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/purchase-requisitions/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/purchase-requisitions", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 7. Purchase Order 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/purchase-orders/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/purchase-orders", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 8. Inventory Item 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/inventory/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/inventory", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 9. Receipt 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/receipts/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/receipts", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 10. Shipment 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/shipments/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/shipments", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 11. Demand Forecast 404 & validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/demand-forecasts/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/demand-forecasts", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 12. Context cancellation database errors
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	listUrls := []string{
		"/api/v1/products",
		"/api/v1/product-categories",
		"/api/v1/locations",
		"/api/v1/vendors",
		"/api/v1/vendor-contracts",
		"/api/v1/purchase-orders",
		"/api/v1/purchase-requisitions",
		"/api/v1/inventory",
		"/api/v1/stock-transfers",
		"/api/v1/inventory/movements",
		"/api/v1/receipts",
		"/api/v1/shipments",
		"/api/v1/demand-forecasts",
	}

	for _, url := range listUrls {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, url, nil)
		req = req.WithContext(canceledCtx)
		env.router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500 for canceled context on %s, got %d", url, w.Code)
		}
	}

	// 13. PUT operations on non-existent resources
	putErrors := []struct {
		url  string
		body interface{}
	}{
		{"/api/v1/product-categories/non-existent", map[string]interface{}{"name": "Non-existent"}},
		{"/api/v1/products/non-existent", map[string]interface{}{"product_name": "Non-existent"}},
		{"/api/v1/locations/non-existent", map[string]interface{}{"location_name": "Non-existent"}},
		{"/api/v1/vendors/non-existent", map[string]interface{}{"supplier_name": "Non-existent"}},
		{"/api/v1/vendor-contracts/non-existent", map[string]interface{}{"contract_number": "CON-ERR", "start_date": "2026-06-01", "end_date": "2027-06-01"}},
		{"/api/v1/purchase-requisitions/non-existent", map[string]interface{}{"requisition_number": "REQ-ERR"}},
		{"/api/v1/purchase-orders/non-existent", map[string]interface{}{"supplier_id": "sup-123", "order_date": "2026-06-14T00:00:00Z"}},
		{"/api/v1/inventory/non-existent", map[string]interface{}{"quantity_on_hand": 10}},
		{"/api/v1/receipts/non-existent", map[string]interface{}{"status": "RECEIVED"}},
		{"/api/v1/shipments/non-existent", map[string]interface{}{"status": "SHIPPED"}},
		{"/api/v1/demand-forecasts/non-existent", map[string]interface{}{"forecast_period": "2026-07", "confidence_level": "0.90"}},
	}

	for _, item := range putErrors {
		bodyBytes, _ := json.Marshal(item.body)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodPut, item.url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		env.router.ServeHTTP(w, req)
		if w.Code == http.StatusOK || w.Code == http.StatusCreated {
			t.Errorf("expected error code for PUT %s, got %d", item.url, w.Code)
		}
	}
}

func TestRemainingScmEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Seed Category
	cat := &sql.ProductCategory{ID: "cat-123", Code: "RAW", Name: "Raw Materials"}
	_ = env.db.Create(cat).Error

	// Seed Product
	product := &sql.Product{
		ID:          "prod-123",
		ProductCode: "ROD-01",
		ProductName: "Steel Rod",
		IsActive:    true,
		CategoryID:  &cat.ID,
	}
	_ = env.db.Create(product).Error

	// Seed InventoryItem
	invItem := &sql.InventoryItem{
		ID:                "inv-123",
		ProductID:         "prod-123",
		LocationID:        "loc_default",
		QuantityOnHand:    100,
		QuantityReserved:  0,
		QuantityAvailable: 100,
		MaximumStock:      1000,
	}
	_ = env.db.Create(invItem).Error

	// Get Inventory Items List
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/inventory", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Inventory Item
	updateInvBody, _ := json.Marshal(map[string]interface{}{
		"quantity_on_hand": 120,
		"reorder_point":    20,
		"maximum_stock":    1100,
		"unit_cost":        "5.50",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/inventory/inv-123", bytes.NewBuffer(updateInvBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Seed Demand Forecast
	fc := &sql.DemandForecast{
		ID:               "fc-123",
		ProductID:        "prod-123",
		ForecastDate:     time.Now(),
		ForecastQuantity: 300,
		ConfidenceLevel:  decimal.NewFromFloat(0.90),
	}
	_ = env.db.Create(fc).Error

	// Update Forecast
	fcBody, _ := json.Marshal(map[string]interface{}{
		"product_id":        "prod-123",
		"forecast_period":   "2026-07",
		"forecast_quantity": 350,
		"confidence_level":  "0.95",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/demand-forecasts/fc-123", bytes.NewBuffer(fcBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Stock Transfers
	st := &sql.StockTransfer{
		ID:             "st-123",
		ProductID:      "prod-123",
		FromLocationID: "loc_default",
		ToLocationID:   "loc_default",
		Quantity:       10,
		Status:         "PENDING",
	}
	_ = env.db.Create(st).Error

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/stock-transfers", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/stock-transfers/st-123", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Reserve Stock for the transfer
	resStBody, _ := json.Marshal(map[string]interface{}{
		"product_id":   "prod-123",
		"location_id":  "loc_default",
		"quantity":     10,
		"reference_id": "st-123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/inventory/reserve", bytes.NewBuffer(resStBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/stock-transfers/st-123/execute", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Purchase Requisition Operations
	reqObj := &sql.PurchaseRequisition{
		ID:        "pr-123",
		ReqNumber: "REQ-200",
		Status:    "PENDING",
	}
	_ = env.db.Create(reqObj).Error

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/purchase-requisitions", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/purchase-requisitions/pr-123", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Purchase Requisition
	prUpdateBody, _ := json.Marshal(map[string]interface{}{
		"requisition_number": "REQ-200-REV1",
		"requested_by":       "emp-002",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/purchase-requisitions/pr-123", bytes.NewBuffer(prUpdateBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Reject PR
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/purchase-requisitions/pr-123/reject", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Get PR Lines
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/purchase-requisitions/pr-123/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete PR
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/purchase-requisitions/pr-123", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Purchase Orders operations
	poObj := &sql.PurchaseOrder{
		ID:       "po-123",
		PoNumber: "PO-200",
		Status:   "PENDING",
	}
	_ = env.db.Create(poObj).Error

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/purchase-orders/po-123", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Purchase Order
	poUpdateBody, _ := json.Marshal(map[string]interface{}{
		"supplier_id": "sup-123",
		"order_date":  time.Now().Format(time.RFC3339),
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/purchase-orders/po-123", bytes.NewBuffer(poUpdateBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Delete Purchase Order
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/purchase-orders/po-123", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Receipts and Shipments operations
	recObj := &sql.Receipt{
		ID:            "rec-123",
		ReceiptNumber: "REC-100",
		Status:        "PENDING",
	}
	_ = env.db.Create(recObj).Error

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/receipts", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/receipts/rec-123/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	recUpdateBody, _ := json.Marshal(map[string]interface{}{
		"status": "RECEIVED",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/receipts/rec-123", bytes.NewBuffer(recUpdateBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	shipObj := &sql.Shipment{
		ID:             "ship-123",
		ShipmentNumber: "SHP-100",
		Status:         "PENDING",
	}
	_ = env.db.Create(shipObj).Error

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/shipments", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/shipments/ship-123/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	shipUpdateBody, _ := json.Marshal(map[string]interface{}{
		"status": "SHIPPED",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/shipments/ship-123", bytes.NewBuffer(shipUpdateBody))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Delete Inventory Item
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/inventory/inv-123", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

