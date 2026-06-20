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
	"github.com/erp-system/crm-service/internal/api/handlers"
	"github.com/erp-system/crm-service/internal/api/routes"
	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func init() {
	gin.SetMode(gin.TestMode)
	utils.InitLogger("crm-service-test")
}

type testEnv struct {
	router         *gin.Engine
	custRepo       *memory.CustomerRepository
	leadRepo       *memory.LeadRepository
	oppRepo        *memory.OpportunityRepository
	orderRepo      *memory.SalesOrderRepository
	orderLineRepo  *memory.SalesOrderLineRepository
	quoteRepo      *memory.QuoteRepository
	quoteLineRepo  *memory.QuoteLineItemRepository
	pbHeaderRepo   *memory.PriceBookHeaderRepository
	pbEntryRepo    *memory.PriceBookEntryRepository
	ticketRepo     *memory.ServiceTicketRepository
	campRepo       *memory.CampaignRepository
	historyRepo    *memory.OpportunityStageHistoryRepository
	interactRepo   *memory.CustomerInteractionRepository
	publisher      *mockPublisher
}

type mockPublisher struct {
	Published []publishedEvent
}

type publishedEvent struct {
	Topic   string
	Key     string
	Payload interface{}
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	m.Published = append(m.Published, publishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

func setupTestEnv() *testEnv {
	custRepo := memory.NewCustomerRepository()
	leadRepo := memory.NewLeadRepository()
	oppRepo := memory.NewOpportunityRepository()
	orderRepo := memory.NewSalesOrderRepository()
	orderLineRepo := memory.NewSalesOrderLineRepository()
	quoteRepo := memory.NewQuoteRepository()
	quoteLineRepo := memory.NewQuoteLineItemRepository()
	pbHeaderRepo := memory.NewPriceBookHeaderRepository()
	pbEntryRepo := memory.NewPriceBookEntryRepository()
	ticketRepo := memory.NewServiceTicketRepository()
	campRepo := memory.NewCampaignRepository()
	historyRepo := memory.NewOpportunityStageHistoryRepository()
	interactRepo := memory.NewCustomerInteractionRepository()
	publisher := &mockPublisher{}

	custSvc := service.NewCustomerService(custRepo, publisher)
	oppSvc := service.NewOpportunityService(oppRepo, historyRepo, publisher)
	leadSvc := service.NewLeadService(leadRepo, custSvc, oppSvc, publisher)
	orderSvc := service.NewSalesOrderService(orderRepo, orderLineRepo, custRepo, publisher)
	quoteSvc := service.NewQuoteService(quoteRepo, quoteLineRepo, publisher)
	ticketSvc := service.NewServiceTicketService(ticketRepo, publisher)
	campSvc := service.NewCampaignService(campRepo, publisher)
	plSvc := service.NewPriceListService(pbHeaderRepo, pbEntryRepo)

	response := utils.NewResponseHelper("crm-service")

	custLeadHandler := handlers.NewCustomerLeadHandler(custSvc, leadSvc, response)
	salesOppHandler := handlers.NewSalesOpportunityHandler(oppSvc, orderSvc, quoteSvc, ticketSvc, campSvc, plSvc, response)
	custInteractionHandler := handlers.NewCustomerInteractionHandler(service.NewCustomerInteractionService(interactRepo, publisher), response)

	router := gin.New()
	routes.SetupCRMRoutes(router, custLeadHandler, salesOppHandler, custInteractionHandler)

	return &testEnv{
		router:         router,
		custRepo:       custRepo,
		leadRepo:       leadRepo,
		oppRepo:        oppRepo,
		orderRepo:      orderRepo,
		orderLineRepo:  orderLineRepo,
		quoteRepo:      quoteRepo,
		quoteLineRepo:  quoteLineRepo,
		pbHeaderRepo:   pbHeaderRepo,
		pbEntryRepo:    pbEntryRepo,
		ticketRepo:     ticketRepo,
		campRepo:       campRepo,
		historyRepo:    historyRepo,
		interactRepo:   interactRepo,
		publisher:      publisher,
	}
}

func TestCustomerEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Create Customer invalid input
	body, _ := json.Marshal(map[string]interface{}{
		"email": "invalid-customer",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad customer req, got %d", w.Code)
	}

	// 2. Create Customer success
	body, _ = json.Marshal(map[string]interface{}{
		"company_name": "ACME Corp",
		"contact_name": "John Doe",
		"email":        "john@acme.com",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var customer domain.CustomerProfile
	_ = json.Unmarshal(w.Body.Bytes(), &customer)

	// 3. Get Customer success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/customers/"+customer.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 4. Update Customer success
	body, _ = json.Marshal(map[string]interface{}{
		"company_name": "ACME Corp updated",
		"contact_name": "John Doe updated",
		"email":        "john.updated@acme.com",
		"status":       "ACTIVE",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/customers/"+customer.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for update, got %d", w.Code)
	}

	// 5. List Customers
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 6. Delete Customer
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/customers/"+customer.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLeadEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Create Lead success
	body, _ := json.Marshal(map[string]interface{}{
		"first_name":   "Jane",
		"last_name":    "Doe",
		"company":      "Globex",
		"email":        "jane@globex.com",
		"phone":        "555-1234",
		"source":       "WEB",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/leads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var lead domain.Lead
	_ = json.Unmarshal(w.Body.Bytes(), &lead)

	// 2. Convert Lead
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/leads/"+lead.ID+"/convert", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestOpportunityEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Create Opportunity success
	body, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"title":       "Big Deal",
		"value":       decimal.NewFromInt(10000),
		"stage":       "DISCOVERY",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/opportunities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var opp domain.Opportunity
	_ = json.Unmarshal(w.Body.Bytes(), &opp)

	// 2. Update stage
	body, _ = json.Marshal(map[string]interface{}{
		"title":       "Big Deal",
		"value":       decimal.NewFromInt(10000),
		"status":      "OPEN",
		"stage":       "CLOSED_WON",
		"probability": decimal.NewFromFloat(0.8),
		"changed_by":  "emp-1",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/opportunities/"+opp.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCustomerInteractions(t *testing.T) {
	env := setupTestEnv()

	body, _ := json.Marshal(map[string]interface{}{
		"customer_id":      "cust-1",
		"type":             "EMAIL",
		"subject":          "Intro",
		"description":      "First meeting description",
		"interaction_date": time.Now(),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/customer-interactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestSalesOrderEndpoints(t *testing.T) {
	env := setupTestEnv()

	// Seed customer
	cust := &domain.CustomerProfile{ID: "cust-1", CompanyName: "ACME", Status: domain.CustomerStatusACTIVE}
	_ = env.custRepo.Create(context.Background(), cust)

	body, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"items": []map[string]interface{}{
			{
				"material_id": "mat-1",
				"quantity":    2,
				"unit_price":  decimal.NewFromInt(100),
			},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sales-orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestAdditionalOpportunityEndpoints(t *testing.T) {
	env := setupTestEnv()
	// Seed opp
	opp := &domain.Opportunity{ID: "opp-1", CustomerID: "cust-1", Title: "Opp 1", Value: decimal.NewFromInt(500), Status: "OPEN", Stage: domain.OpportunityStageDISCOVERY}
	_ = env.oppRepo.Create(context.Background(), opp)

	// Get opportunity
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/opportunities/opp-1", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// List opportunities
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/opportunities", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Stage history
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/opportunities/opp-1/stage-history", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete opportunity
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/opportunities/opp-1", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestQuoteEndpoints(t *testing.T) {
	env := setupTestEnv()

	body, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"title":       "Quote 1",
		"valid_until": time.Now().Add(48 * time.Hour),
		"items": []map[string]interface{}{
			{
				"material_id": "mat-1",
				"quantity":    5,
				"unit_price":  decimal.NewFromInt(120),
			},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/quotes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var quote domain.Quote
	_ = json.Unmarshal(w.Body.Bytes(), &quote)

	// List quotes
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/quotes", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get quote
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/quotes/"+quote.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Send quote
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/quotes/"+quote.ID+"/send", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update quote
	body, _ = json.Marshal(map[string]interface{}{
		"title":       "Quote 1 updated",
		"valid_until": time.Now().Add(96 * time.Hour),
		"status":      "APPROVED",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/quotes/"+quote.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete quote
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/quotes/"+quote.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestServiceTicketEndpoints(t *testing.T) {
	env := setupTestEnv()

	body, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"title":       "Problem with login",
		"description": "Cannot login",
		"priority":    "HIGH",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/service-tickets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var ticket domain.ServiceTicket
	_ = json.Unmarshal(w.Body.Bytes(), &ticket)

	// List
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/service-tickets", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/service-tickets/"+ticket.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update
	body, _ = json.Marshal(map[string]interface{}{
		"subject":     "Support Issue Updated",
		"description": "Resolved",
		"priority":    "MEDIUM",
		"status":      "CLOSED",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/service-tickets/"+ticket.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/service-tickets/"+ticket.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCampaignEndpoints(t *testing.T) {
	env := setupTestEnv()

	body, _ := json.Marshal(map[string]interface{}{
		"name":       "Summer Sale",
		"type":       "EMAIL",
		"status":     "ACTIVE",
		"budget":     decimal.NewFromInt(5000),
		"start_date": time.Now(),
		"end_date":   time.Now().Add(720 * time.Hour),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/campaigns", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var camp domain.Campaign
	_ = json.Unmarshal(w.Body.Bytes(), &camp)

	// List
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/campaigns", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/campaigns/"+camp.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update
	body, _ = json.Marshal(map[string]interface{}{
		"name":   "Summer Sale Updated",
		"type":   "EMAIL",
		"status": "COMPLETED",
		"budget": decimal.NewFromInt(6000),
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/campaigns/"+camp.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/campaigns/"+camp.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestPriceListEndpoints(t *testing.T) {
	env := setupTestEnv()

	body, _ := json.Marshal(map[string]interface{}{
		"name":        "Wholesale Price Book",
		"type":        "STANDARD",
		"description": "Standard wholesale price list",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/price-lists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var pb domain.PriceBookHeader
	_ = json.Unmarshal(w.Body.Bytes(), &pb)

	// List
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/price-lists", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/price-lists/"+pb.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update
	body, _ = json.Marshal(map[string]interface{}{
		"name":        "Wholesale Price Book Updated",
		"description": "Updated standard wholesale price list",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/price-lists/"+pb.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/price-lists/"+pb.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCustomerInteractionEndpoints(t *testing.T) {
	env := setupTestEnv()

	// Seed interaction
	ci := &domain.CustomerInteraction{ID: "ci-1", CustomerID: "cust-1", Type: "CALL", Subject: "Intro Call"}
	_ = env.interactRepo.Create(context.Background(), ci)

	// List
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/customer-interactions?customer_id=cust-1", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/customer-interactions/ci-1", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/customer-interactions/ci-1", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestAdditionalSalesOrderEndpoints(t *testing.T) {
	env := setupTestEnv()

	// Seed order
	so := &domain.SalesOrder{ID: "so-1", CustomerID: "cust-1", TotalGrossValue: decimal.NewFromInt(200), Status: domain.SalesOrderStateDRAFT}
	_ = env.orderRepo.Create(context.Background(), so)

	// List
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/sales-orders", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/sales-orders/so-1", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update
	body, _ := json.Marshal(map[string]interface{}{
		"status": "APPROVED",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/sales-orders/so-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Delete
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/sales-orders/so-1", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestErrorPathsAndValidation(t *testing.T) {
	env := setupTestEnv()

	// 1. Lead validation error
	body, _ := json.Marshal(map[string]interface{}{
		"email": "invalid-email",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/leads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 2. Get Lead not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/leads/lead-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 3. Update Lead validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/leads/lead-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 4. Update Opportunity validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/opportunities/opp-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 5. Quote validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/quotes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 6. Service Ticket validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/service-tickets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 7. Campaign validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/campaigns", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 8. Price list validation error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/price-lists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 9. Get Customer not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/customers/cust-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 10. Get Opportunity not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/opportunities/opp-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 11. Get Quote not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/quotes/q-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 12. Get Service Ticket not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/service-tickets/t-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 13. Get Campaign not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/campaigns/camp-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 14. Get Price List not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/price-lists/pb-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 15. Get Customer Interaction not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/customer-interactions/ci-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 16. Get Sales Order not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/sales-orders/so-999", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

