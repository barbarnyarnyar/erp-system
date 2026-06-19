package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/crm-service/internal/api/handlers"
	"github.com/erp-system/crm-service/internal/api/routes"
	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// -------------------------------------------------------------
// Failing Repositories to trigger 500 Internal Server Errors
// -------------------------------------------------------------

type failingCustomerRepo struct{}

func (r *failingCustomerRepo) Create(ctx context.Context, customer *domain.CustomerProfile) error {
	return errors.New("db error")
}
func (r *failingCustomerRepo) GetByID(ctx context.Context, id string) (*domain.CustomerProfile, error) {
	return nil, errors.New("db error")
}
func (r *failingCustomerRepo) List(ctx context.Context) ([]domain.CustomerProfile, error) {
	return nil, errors.New("db error")
}
func (r *failingCustomerRepo) Update(ctx context.Context, customer *domain.CustomerProfile) error {
	return errors.New("db error")
}
func (r *failingCustomerRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type failingLeadRepo struct{}

func (r *failingLeadRepo) Create(ctx context.Context, lead *domain.Lead) error {
	return errors.New("db error")
}
func (r *failingLeadRepo) GetByID(ctx context.Context, id string) (*domain.Lead, error) {
	return nil, errors.New("db error")
}
func (r *failingLeadRepo) List(ctx context.Context) ([]domain.Lead, error) {
	return nil, errors.New("db error")
}
func (r *failingLeadRepo) Update(ctx context.Context, lead *domain.Lead) error {
	return errors.New("db error")
}
func (r *failingLeadRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type failingOpportunityRepo struct{}

func (r *failingOpportunityRepo) Create(ctx context.Context, opp *domain.Opportunity) error {
	return errors.New("db error")
}
func (r *failingOpportunityRepo) GetByID(ctx context.Context, id string) (*domain.Opportunity, error) {
	return nil, errors.New("db error")
}
func (r *failingOpportunityRepo) List(ctx context.Context) ([]domain.Opportunity, error) {
	return nil, errors.New("db error")
}
func (r *failingOpportunityRepo) Update(ctx context.Context, opp *domain.Opportunity) error {
	return errors.New("db error")
}
func (r *failingOpportunityRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type failingOpportunityStageHistoryRepo struct{}

func (r *failingOpportunityStageHistoryRepo) Create(ctx context.Context, osh *domain.OpportunityStageHistory) error {
	return errors.New("db error")
}
func (r *failingOpportunityStageHistoryRepo) ListByOpportunityID(ctx context.Context, opportunityID string) ([]domain.OpportunityStageHistory, error) {
	return nil, errors.New("db error")
}

type failingSalesOrderRepo struct{}

func (r *failingSalesOrderRepo) Create(ctx context.Context, order *domain.SalesOrder) error {
	return errors.New("db error")
}
func (r *failingSalesOrderRepo) GetByID(ctx context.Context, id string) (*domain.SalesOrder, error) {
	return nil, errors.New("db error")
}
func (r *failingSalesOrderRepo) List(ctx context.Context) ([]domain.SalesOrder, error) {
	return nil, errors.New("db error")
}
func (r *failingSalesOrderRepo) Update(ctx context.Context, order *domain.SalesOrder) error {
	return errors.New("db error")
}
func (r *failingSalesOrderRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type failingSalesOrderLineRepo struct{}

func (r *failingSalesOrderLineRepo) Create(ctx context.Context, item *domain.SalesOrderLine) error {
	return errors.New("db error")
}
func (r *failingSalesOrderLineRepo) ListByOrderID(ctx context.Context, orderID string) ([]domain.SalesOrderLine, error) {
	return nil, errors.New("db error")
}

type failingQuoteRepo struct{}

func (r *failingQuoteRepo) Create(ctx context.Context, quote *domain.Quote) error {
	return errors.New("db error")
}
func (r *failingQuoteRepo) GetByID(ctx context.Context, id string) (*domain.Quote, error) {
	return nil, errors.New("db error")
}
func (r *failingQuoteRepo) List(ctx context.Context) ([]domain.Quote, error) {
	return nil, errors.New("db error")
}
func (r *failingQuoteRepo) Update(ctx context.Context, quote *domain.Quote) error {
	return errors.New("db error")
}
func (r *failingQuoteRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type failingQuoteLineItemRepo struct{}

func (r *failingQuoteLineItemRepo) Create(ctx context.Context, item *domain.QuoteLineItem) error {
	return errors.New("db error")
}
func (r *failingQuoteLineItemRepo) ListByQuoteID(ctx context.Context, quoteID string) ([]domain.QuoteLineItem, error) {
	return nil, errors.New("db error")
}

type failingPriceBookHeaderRepo struct{}

func (r *failingPriceBookHeaderRepo) Create(ctx context.Context, priceBook *domain.PriceBookHeader) error {
	return errors.New("db error")
}
func (r *failingPriceBookHeaderRepo) GetByID(ctx context.Context, id string) (*domain.PriceBookHeader, error) {
	return nil, errors.New("db error")
}
func (r *failingPriceBookHeaderRepo) List(ctx context.Context) ([]domain.PriceBookHeader, error) {
	return nil, errors.New("db error")
}
func (r *failingPriceBookHeaderRepo) Update(ctx context.Context, priceBook *domain.PriceBookHeader) error {
	return errors.New("db error")
}
func (r *failingPriceBookHeaderRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type failingPriceBookEntryRepo struct{}

func (r *failingPriceBookEntryRepo) Create(ctx context.Context, item *domain.PriceBookEntry) error {
	return errors.New("db error")
}
func (r *failingPriceBookEntryRepo) ListByPriceBookID(ctx context.Context, priceBookID string) ([]domain.PriceBookEntry, error) {
	return nil, errors.New("db error")
}

type failingServiceTicketRepo struct{}

func (r *failingServiceTicketRepo) Create(ctx context.Context, ticket *domain.ServiceTicket) error {
	return errors.New("db error")
}
func (r *failingServiceTicketRepo) GetByID(ctx context.Context, id string) (*domain.ServiceTicket, error) {
	return nil, errors.New("db error")
}
func (r *failingServiceTicketRepo) List(ctx context.Context) ([]domain.ServiceTicket, error) {
	return nil, errors.New("db error")
}
func (r *failingServiceTicketRepo) Update(ctx context.Context, ticket *domain.ServiceTicket) error {
	return errors.New("db error")
}
func (r *failingServiceTicketRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type failingCampaignRepo struct{}

func (r *failingCampaignRepo) Create(ctx context.Context, campaign *domain.Campaign) error {
	return errors.New("db error")
}
func (r *failingCampaignRepo) GetByID(ctx context.Context, id string) (*domain.Campaign, error) {
	return nil, errors.New("db error")
}
func (r *failingCampaignRepo) List(ctx context.Context) ([]domain.Campaign, error) {
	return nil, errors.New("db error")
}
func (r *failingCampaignRepo) Update(ctx context.Context, campaign *domain.Campaign) error {
	return errors.New("db error")
}
func (r *failingCampaignRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type failingCustomerInteractionRepo struct{}

func (r *failingCustomerInteractionRepo) Create(ctx context.Context, ci *domain.CustomerInteraction) error {
	return errors.New("db error")
}
func (r *failingCustomerInteractionRepo) GetByID(ctx context.Context, id string) (*domain.CustomerInteraction, error) {
	return nil, errors.New("db error")
}
func (r *failingCustomerInteractionRepo) ListByCustomerID(ctx context.Context, customerID string) ([]domain.CustomerInteraction, error) {
	return nil, errors.New("db error")
}
func (r *failingCustomerInteractionRepo) Delete(ctx context.Context, id string) error {
	return errors.New("db error")
}

type dummyPublisher struct{}

func (p *dummyPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

func setupFailingTestEnv() *gin.Engine {
	custRepo := &failingCustomerRepo{}
	leadRepo := &failingLeadRepo{}
	oppRepo := &failingOpportunityRepo{}
	orderRepo := &failingSalesOrderRepo{}
	orderLineRepo := &failingSalesOrderLineRepo{}
	quoteRepo := &failingQuoteRepo{}
	quoteLineRepo := &failingQuoteLineItemRepo{}
	pbHeaderRepo := &failingPriceBookHeaderRepo{}
	pbEntryRepo := &failingPriceBookEntryRepo{}
	ticketRepo := &failingServiceTicketRepo{}
	campRepo := &failingCampaignRepo{}
	historyRepo := &failingOpportunityStageHistoryRepo{}
	interactRepo := &failingCustomerInteractionRepo{}
	publisher := &dummyPublisher{}

	custSvc := service.NewCustomerService(custRepo, publisher)
	oppSvc := service.NewOpportunityService(oppRepo, historyRepo, publisher)
	leadSvc := service.NewLeadService(leadRepo, custSvc, oppSvc, publisher)
	orderSvc := service.NewSalesOrderService(orderRepo, orderLineRepo, custRepo, publisher)
	quoteSvc := service.NewQuoteService(quoteRepo, quoteLineRepo, publisher)
	ticketSvc := service.NewServiceTicketService(ticketRepo, publisher)
	campSvc := service.NewCampaignService(campRepo, publisher)
	plSvc := service.NewPriceListService(pbHeaderRepo, pbEntryRepo)
	ciSvc := service.NewCustomerInteractionService(interactRepo, publisher)

	response := utils.NewResponseHelper("crm-service")

	custLeadHandler := handlers.NewCustomerLeadHandler(custSvc, leadSvc, response)
	salesOppHandler := handlers.NewSalesOpportunityHandler(oppSvc, orderSvc, quoteSvc, ticketSvc, campSvc, plSvc, response)
	custInteractionHandler := handlers.NewCustomerInteractionHandler(ciSvc, response)

	router := gin.New()
	routes.SetupCRMRoutes(router, custLeadHandler, salesOppHandler, custInteractionHandler)

	return router
}

// -------------------------------------------------------------
// Test HTTP 500 Responses (Internal Server Error Branches)
// -------------------------------------------------------------

func TestHandlers_InternalErrors(t *testing.T) {
	router := setupFailingTestEnv()

	// 1. Create Customer -> 500
	custBody, _ := json.Marshal(map[string]interface{}{
		"company_name": "Failing Acme",
		"contact_name": "John Doe",
		"email":        "john@acme.com",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewBuffer(custBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 2. List Customers -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 3. Update Customer -> 500
	updateCustBody, _ := json.Marshal(map[string]interface{}{
		"company_name": "Failing Acme",
		"contact_name": "John Doe",
		"email":        "john@acme.com",
		"status":       "ACTIVE",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/customers/cust-1", bytes.NewBuffer(updateCustBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 4. Delete Customer -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/customers/cust-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 5. Create Lead -> 500
	leadBody, _ := json.Marshal(map[string]interface{}{
		"first_name": "Failing",
		"last_name":  "Lead",
		"company":    "Bad DB",
		"email":      "lead@db.com",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/leads", bytes.NewBuffer(leadBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 6. List Leads -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/leads", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 7. Update Lead -> 500
	updateLeadBody, _ := json.Marshal(map[string]interface{}{
		"first_name": "Failing",
		"last_name":  "Lead",
		"company":    "Bad DB",
		"status":     "NEW",
		"score":      50,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/leads/lead-1", bytes.NewBuffer(updateLeadBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 8. Delete Lead -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/leads/lead-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 9. Convert Lead -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/leads/lead-1/convert", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 10. Create Opportunity -> 500
	oppBody, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"title":       "Failing deal",
		"value":       decimal.NewFromInt(100),
		"stage":       "DISCOVERY",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/opportunities", bytes.NewBuffer(oppBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 11. List Opportunities -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/opportunities", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 12. Update Opportunity -> 500
	updateOppBody, _ := json.Marshal(map[string]interface{}{
		"title":       "Failing deal",
		"value":       decimal.NewFromInt(100),
		"status":      "OPEN",
		"stage":       "QUALIFIED",
		"probability": decimal.NewFromFloat(0.5),
		"changed_by":  "user-1",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/opportunities/opp-1", bytes.NewBuffer(updateOppBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 13. Delete Opportunity -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/opportunities/opp-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 14. Get Opportunity Stage History -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/opportunities/opp-1/stage-history", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 15. Create Sales Order -> 500
	soBody, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"items": []map[string]interface{}{
			{
				"product_id": "mat-1",
				"quantity":   5,
				"unit_price": decimal.NewFromInt(10),
			},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/sales-orders", bytes.NewBuffer(soBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 16. List Sales Orders -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/sales-orders", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 17. Update Sales Order -> 500
	updateSoBody, _ := json.Marshal(map[string]interface{}{
		"status": "APPROVED",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/sales-orders/so-1", bytes.NewBuffer(updateSoBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 18. Delete Sales Order -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/sales-orders/so-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 19. Create Quote -> 500
	quoteBody, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"title":       "Failing Quote",
		"valid_until": time.Now().Add(24 * time.Hour),
		"items": []map[string]interface{}{
			{
				"product_id": "mat-1",
				"quantity":   2,
				"unit_price": decimal.NewFromInt(50),
			},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/quotes", bytes.NewBuffer(quoteBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 20. List Quotes -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/quotes", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 21. Update Quote -> 500
	updateQuoteBody, _ := json.Marshal(map[string]interface{}{
		"status": "APPROVED",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/quotes/q-1", bytes.NewBuffer(updateQuoteBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 22. Delete Quote -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/quotes/q-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 23. Send Quote -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/quotes/q-1/send", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 24. Create Service Ticket -> 500
	ticketBody, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"title":       "Failing Ticket",
		"description": "Fail",
		"priority":    "HIGH",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/service-tickets", bytes.NewBuffer(ticketBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 25. List Service Tickets -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/service-tickets", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 26. Update Service Ticket -> 500
	updateTicketBody, _ := json.Marshal(map[string]interface{}{
		"status":   "CLOSED",
		"priority": "LOW",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/service-tickets/t-1", bytes.NewBuffer(updateTicketBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 27. Delete Service Ticket -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/service-tickets/t-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 28. Create Campaign -> 500
	campBody, _ := json.Marshal(map[string]interface{}{
		"name":   "Failing Camp",
		"type":   "EMAIL",
		"budget": decimal.NewFromInt(1000),
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/campaigns", bytes.NewBuffer(campBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 29. List Campaigns -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/campaigns", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 30. Update Campaign -> 500
	updateCampBody, _ := json.Marshal(map[string]interface{}{
		"status": "LAUNCHED",
		"budget": decimal.NewFromInt(1500),
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/campaigns/camp-1", bytes.NewBuffer(updateCampBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 31. Delete Campaign -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/campaigns/camp-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 32. Create Price List -> 500
	plBody, _ := json.Marshal(map[string]interface{}{
		"name": "Failing PL",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/price-lists", bytes.NewBuffer(plBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 33. List Price Lists -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/price-lists", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 34. Update Price List -> 500
	updatePlBody, _ := json.Marshal(map[string]interface{}{
		"name": "Failing PL updated",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/price-lists/pl-1", bytes.NewBuffer(updatePlBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 35. Delete Price List -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/price-lists/pl-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 36. Create Customer Interaction -> 500
	ciBody, _ := json.Marshal(map[string]interface{}{
		"customer_id": "cust-1",
		"type":        "CALL",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/customer-interactions", bytes.NewBuffer(ciBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 37. List Customer Interactions with query param but failing repo -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/customer-interactions?customer_id=cust-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 38. Delete Customer Interaction -> 500
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/customer-interactions/ci-1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// -------------------------------------------------------------
// Test HTTP 400 Responses (Update Input Validation Error Branches)
// -------------------------------------------------------------

func TestHandlers_UpdateValidationErrors(t *testing.T) {
	router := setupFailingTestEnv()
	badJson := []byte("{invalid-json}")

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodPut, "/api/v1/customers/cust-1"},
		{http.MethodPut, "/api/v1/leads/lead-1"},
		{http.MethodPut, "/api/v1/opportunities/opp-1"},
		{http.MethodPut, "/api/v1/sales-orders/so-1"},
		{http.MethodPut, "/api/v1/quotes/q-1"},
		{http.MethodPut, "/api/v1/service-tickets/t-1"},
		{http.MethodPut, "/api/v1/campaigns/camp-1"},
		{http.MethodPut, "/api/v1/price-lists/pl-1"},
	}

	for _, ep := range endpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(ep.method, ep.url, bytes.NewBuffer(badJson))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for %s %s, got %d", ep.method, ep.url, w.Code)
		}
	}
}
