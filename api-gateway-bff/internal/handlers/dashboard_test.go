package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-gateway-bff/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func TestGetSalesDashboard_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock server that returns mock responses for CRM, SCM, and FM
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/api/v1/orders/so-1/lines") {
			lines := []CRMOrderLineResponse{
				{
					ID:              "soi-1",
					SalesOrderID:    "so-1",
					MaterialID:      "mat-1",
					LineSequence:    10,
					QuantityOrdered: decimal.NewFromInt(2),
					UnitSellPrice:   decimal.NewFromFloat(150.0),
					NetLineAmount:   decimal.NewFromFloat(300.0),
				},
			}
			json.NewEncoder(w).Encode(lines)
			return
		}

		if strings.Contains(r.URL.Path, "/api/v1/orders/so-1") {
			order := CRMOrderResponse{
				ID:              "so-1",
				LegalEntityID:   "le-1",
				CustomerID:      "cust-1",
				OrderNumber:     "SO-00001",
				Status:          "DRAFT",
				TotalGrossValue: decimal.NewFromFloat(300.0),
			}
			json.NewEncoder(w).Encode(order)
			return
		}

		if strings.Contains(r.URL.Path, "/api/v1/customers/cust-1/credit") {
			credit := FMCustomerCreditEnvelope{
				Data: FMCustomerCreditResponse{
					ID:             "credit-1",
					CustomerID:     "cust-1",
					CreditLimit:    decimal.NewFromFloat(1000.0),
					CurrentBalance: decimal.NewFromFloat(100.0),
					IsOnHold:       false,
				},
			}
			json.NewEncoder(w).Encode(credit)
			return
		}

		if strings.Contains(r.URL.Path, "/api/v1/materials/mat-1") {
			material := SCMMaterialEnvelope{
				Data: SCMMaterialResponse{
					ID:          "mat-1",
					ProductCode: "P001",
					ProductName: "Base Material",
					IsActive:    true,
				},
			}
			json.NewEncoder(w).Encode(material)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		Port:          "8085",
		CRMServiceURL: mockServer.URL,
		SCMServiceURL: mockServer.URL,
		FMServiceURL:  mockServer.URL,
	}

	handler := NewDashboardHandler(cfg)
	router := gin.New()
	router.GET("/api/v1/ui/sales-dashboard/:order_id", handler.GetSalesDashboard)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ui/sales-dashboard/so-1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp SalesDashboardResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Order.ID != "so-1" {
		t.Errorf("Expected order ID 'so-1', got '%s'", resp.Order.ID)
	}
	if len(resp.Lines) != 1 || resp.Lines[0].MaterialID != "mat-1" {
		t.Errorf("Expected 1 line with material ID 'mat-1'")
	}
	if resp.CustomerCredit == nil || resp.CustomerCredit.CreditLimit.Cmp(decimal.NewFromFloat(1000.0)) != 0 {
		t.Errorf("Expected credit limit 1000")
	}
	if resp.Materials == nil || resp.Materials["mat-1"].ProductCode != "P001" {
		t.Errorf("Expected material code 'P001'")
	}
	if resp.InventoryStatus != "AVAILABLE" {
		t.Errorf("Expected inventory status AVAILABLE, got '%s'", resp.InventoryStatus)
	}
}

func TestGetSalesDashboard_SCM_Degradation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/api/v1/orders/so-1/lines") {
			lines := []CRMOrderLineResponse{
				{
					ID:           "soi-1",
					SalesOrderID: "so-1",
					MaterialID:   "mat-1",
				},
			}
			json.NewEncoder(w).Encode(lines)
			return
		}

		if strings.Contains(r.URL.Path, "/api/v1/orders/so-1") {
			order := CRMOrderResponse{
				ID:         "so-1",
				CustomerID: "cust-1",
			}
			json.NewEncoder(w).Encode(order)
			return
		}

		if strings.Contains(r.URL.Path, "/api/v1/customers/cust-1/credit") {
			credit := FMCustomerCreditEnvelope{
				Data: FMCustomerCreditResponse{
					ID:         "credit-1",
					CustomerID: "cust-1",
				},
			}
			json.NewEncoder(w).Encode(credit)
			return
		}

		if strings.Contains(r.URL.Path, "/api/v1/materials/mat-1") {
			// SCM service fails
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		Port:          "8085",
		CRMServiceURL: mockServer.URL,
		SCMServiceURL: mockServer.URL,
		FMServiceURL:  mockServer.URL,
	}

	handler := NewDashboardHandler(cfg)
	router := gin.New()
	router.GET("/api/v1/ui/sales-dashboard/:order_id", handler.GetSalesDashboard)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ui/sales-dashboard/so-1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp SalesDashboardResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.InventoryStatus != "UNAVAILABLE" {
		t.Errorf("Expected inventory status UNAVAILABLE, got '%s'", resp.InventoryStatus)
	}
}
