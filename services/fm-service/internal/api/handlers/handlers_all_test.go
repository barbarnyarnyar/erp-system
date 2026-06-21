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
	"github.com/erp-system/fm-service/internal/api/handlers"
	"github.com/erp-system/fm-service/internal/api/routes"
	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/config"
	"github.com/erp-system/fm-service/internal/data/memory"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func init() {
	gin.SetMode(gin.TestMode)
	utils.InitLogger("fm-service-test")
}

type testEnv struct {
	router        *gin.Engine
	accounts      *memory.MemoryChartOfAccountsRepo
	entries       *memory.MemoryUniversalJournalEntryRepo
	invoices      *memory.MemoryArInvoiceRepo
	payments      *memory.MemoryPaymentRepo
	statements    *memory.MemoryBankStatementRepo
	bills         *memory.MemoryApVendorBillRepo
	outbox        *memory.MemoryTransactionalOutboxRepo
	legalEntities *memory.MemoryLegalEntityRepo
	assets        *memory.MemoryCapitalAssetRepo
	scheduleLines *memory.MemoryDepreciationScheduleLineRepo
	inbox         *memory.MemoryKafkaEventInboxRepo
}

func setupTestEnv() *testEnv {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	invoices := memory.NewMemoryArInvoiceRepo()
	payments := memory.NewMemoryPaymentRepo()
	statements := memory.NewMemoryBankStatementRepo()
	bills := memory.NewMemoryApVendorBillRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	legalEntities := memory.NewMemoryLegalEntityRepo()
	assets := memory.NewMemoryCapitalAssetRepo()
	scheduleLines := memory.NewMemoryDepreciationScheduleLineRepo()
	inbox := memory.NewMemoryKafkaEventInboxRepo()
	credits := memory.NewMemoryCustomerCreditRepo()

	tmGL := memory.NewMemoryTransactionManager(accounts, entries, outbox)
	glSvc := service.NewGeneralLedgerService(accounts, entries, outbox, tmGL)

	tmAR := memory.NewMemoryTransactionManager(invoices, outbox)
	arSvc := service.NewAccountsReceivableService(invoices, credits, outbox, tmAR)

	tmAP := memory.NewMemoryTransactionManager(bills, outbox)
	apSvc := service.NewAccountsPayableService(bills, outbox, tmAP)

	tmCM := memory.NewMemoryTransactionManager(payments, invoices, outbox)
	cmSvc := service.NewCashManagementService(payments, invoices, statements, outbox, tmCM)

	tmLE := memory.NewMemoryTransactionManager(legalEntities)
	leSvc := service.NewLegalEntityService(legalEntities, tmLE)

	tmAsset := memory.NewMemoryTransactionManager(assets, scheduleLines, accounts, entries, outbox)
	assetSvc := service.NewCapitalAssetService(assets, scheduleLines, accounts, entries, outbox, tmAsset)

	response := utils.NewResponseHelper("fm-service")

	accHandler := handlers.NewAccountHandler(glSvc, response)
	txHandler := handlers.NewTransactionHandler(glSvc, response)
	repHandler := handlers.NewReportHandler(glSvc, response)
	invHandler := handlers.NewInvoiceHandler(arSvc, response)
	payHandler := handlers.NewPaymentHandler(cmSvc, response)
	billHandler := handlers.NewVendorBillHandler(apSvc, response)
	leHandler := handlers.NewLegalEntityHandler(leSvc, response)
	assetHandler := handlers.NewAssetHandler(assetSvc, response)

	router := gin.New()
	routes.SetupRoutes(router, &config.Config{}, accHandler, txHandler, repHandler, invHandler, payHandler, billHandler, leHandler, assetHandler)

	return &testEnv{
		router:        router,
		accounts:      accounts,
		entries:       entries,
		invoices:      invoices,
		payments:      payments,
		statements:    statements,
		bills:         bills,
		outbox:        outbox,
		legalEntities: legalEntities,
		assets:        assets,
		scheduleLines: scheduleLines,
		inbox:         inbox,
	}
}

func TestHealthCheck(t *testing.T) {
	env := setupTestEnv()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAccountEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Create Account validation error (missing account code)
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "legal_123",
		"account_name":    "Cash",
		"type":            "ASSET",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError { // due to business logic returning validation error
		t.Errorf("expected 500 for missing account_code, got %d", w.Code)
	}

	// 2. Create Account success
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "legal_123",
		"account_code":    "1000",
		"account_name":    "Cash",
		"type":            "ASSET",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Parse ID
	var resp struct {
		Data domain.ChartOfAccounts `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	accID := resp.Data.ID

	// 3. Get Account success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts/"+accID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 4. Get Account non-existent
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 5. Get Account Balance success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts/"+accID+"/balance", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 6. Get Account Balance 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts/non-existent/balance", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 7. Update Account success
	body, _ = json.Marshal(map[string]interface{}{
		"account_name": "Main Cash",
		"type":         "ASSET",
		"is_active":    true,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/accounts/"+accID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 8. Update Account bad request (invalid JSON)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/accounts/"+accID, bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 9. Get all accounts
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 10. Delete Account
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/accounts/"+accID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestTransactionEndpoints(t *testing.T) {
	env := setupTestEnv()

	// Seed accounts
	acc1 := &domain.ChartOfAccounts{ID: "acc_1", LegalEntityID: "legal_123", AccountCode: "1000", AccountName: "Cash", Type: domain.AccountTypeASSET, IsActive: true}
	acc2 := &domain.ChartOfAccounts{ID: "acc_2", LegalEntityID: "legal_123", AccountCode: "4000", AccountName: "Revenue", Type: domain.AccountTypeREVENUE, IsActive: true}
	_ = env.accounts.Create(context.Background(), acc1)
	_ = env.accounts.Create(context.Background(), acc2)

	// 1. Create Transaction (success)
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id":    "legal_123",
		"source_module":      "FM",
		"source_document_id": "doc_123",
		"posting_date":       time.Now(),
		"lines": []map[string]interface{}{
			{"account_id": "acc_1", "amount_functional": "150.00", "amount_transactional": "150.00", "currency_transactional": "USD"},
			{"account_id": "acc_2", "amount_functional": "-150.00", "amount_transactional": "-150.00", "currency_transactional": "USD"},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/journal-entries", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data domain.UniversalJournalEntry `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	jeID := resp.Data.ID

	// 2. Create Transaction unbalanced
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id":    "legal_123",
		"source_module":      "FM",
		"source_document_id": "doc_123",
		"posting_date":       time.Now(),
		"lines": []map[string]interface{}{
			{"account_id": "acc_1", "amount_functional": "100.00", "amount_transactional": "100.00", "currency_transactional": "USD"},
			{"account_id": "acc_2", "amount_functional": "-50.00", "amount_transactional": "-50.00", "currency_transactional": "USD"},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/journal-entries", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unbalanced transaction, got %d", w.Code)
	}

	// 3. Create Transaction invalid JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/journal-entries", bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 4. Get Transaction
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/journal-entries/"+jeID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 5. Get Transaction 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/journal-entries/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 6. Update Transaction error (cannot update posted entry)
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id":    "legal_123",
		"source_module":      "FM",
		"source_document_id": "doc_123",
		"posting_date":       time.Now(),
		"lines": []map[string]interface{}{
			{"account_id": "acc_1", "amount_functional": "200.00", "amount_transactional": "200.00", "currency_transactional": "USD"},
			{"account_id": "acc_2", "amount_functional": "-200.00", "amount_transactional": "-200.00", "currency_transactional": "USD"},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/journal-entries/"+jeID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (not mutable), got %d", w.Code)
	}

	// Update Transaction bad request (invalid JSON)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/journal-entries/"+jeID, bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 7. Get all Transactions
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/journal-entries", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 8. Delete Transaction
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/journal-entries/"+jeID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestInvoiceEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Create Invoice success
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "legal_123",
		"customer_id":     "cust_999",
		"sales_order_id":  "so_123",
		"total_amount":    "150.00",
		"tax_amount":      "0.00",
		"due_date":        time.Now().AddDate(0, 0, 30),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/invoices", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data domain.ArInvoice `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	invID := resp.Data.ID

	// Create Invoice validation error (bad JSON)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/invoices", bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 2. Get Invoice success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/invoices/"+invID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 3. Get Invoice 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/invoices/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 4. Update Invoice success
	body, _ = json.Marshal(map[string]interface{}{
		"status": "OPEN",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/invoices/"+invID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Update Invoice bad request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/invoices/"+invID, bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Update Invoice missing/non-existent
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/invoices/non-existent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 5. Send Invoice success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/invoices/"+invID+"/send", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Send Invoice non-existent
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/invoices/non-existent/send", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 6. Get Invoice Lines success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/invoices/"+invID+"/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Invoice Lines 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/invoices/non-existent/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 7. Get all Invoices
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/invoices", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 8. Delete Invoice
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/invoices/"+invID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestPaymentEndpoints(t *testing.T) {
	env := setupTestEnv()

	// Seed Invoice
	inv := &domain.ArInvoice{ID: "inv_123", CustomerID: "cust_1", TotalAmount: decimal.NewFromInt(100), Status: domain.PaymentStatusOPEN}
	_ = env.invoices.Create(context.Background(), inv)

	// 1. Record Payment success
	body, _ := json.Marshal(map[string]interface{}{
		"invoice_id":      "inv_123",
		"bank_account_id": "bank_777",
		"amount":          "100.00",
		"payment_method":  "WIRE",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data domain.Payment `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	payID := resp.Data.ID

	// 2. Record Payment bad json
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 3. Record Payment bad amount
	body, _ = json.Marshal(map[string]interface{}{
		"amount": "not-a-decimal",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 4. Record Payment error (non-existent invoice ID inside business logic)
	body, _ = json.Marshal(map[string]interface{}{
		"invoice_id":     "non-existent-inv",
		"amount":         "100.00",
		"payment_method": "WIRE",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 5. Get Payment success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/payments/"+payID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 6. Get Payment 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/payments/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 7. Get Bank Statement Lines 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/bank-statements/non-existent/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// Seed statement and get lines (success)
	stmt := &domain.BankStatement{ID: "stmt_1"}
	stmtLines := []domain.BankStatementLine{{ID: "line_1", StatementID: "stmt_1"}}
	_ = env.statements.Create(context.Background(), stmt, stmtLines)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/bank-statements/stmt_1/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 8. Get all Payments
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/payments", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVendorBillEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Create Vendor Bill success
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id":   "legal_123",
		"vendor_id":         "supp_123",
		"bill_number":       "BILL-999",
		"purchase_order_id": "po_456",
		"due_date":          time.Now().AddDate(0, 0, 30),
		"total_amount":      "300.00",
		"tax_amount":        "0.00",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vendor-bills", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data domain.ApVendorBill `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	billID := resp.Data.ID

	// 2. Create Vendor Bill bad json
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/vendor-bills", bytes.NewBuffer([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 3. Get Vendor Bill Lines success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vendor-bills/"+billID+"/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Vendor Bill Lines 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vendor-bills/non-existent/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 4. Get all Vendor Bills
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vendor-bills", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestReportEndpoints(t *testing.T) {
	env := setupTestEnv()

	// Seed some active accounts
	_ = env.accounts.Create(context.Background(), &domain.ChartOfAccounts{
		ID:            "acc_asset",
		LegalEntityID: "legal_123",
		AccountCode:   "1000",
		AccountName:   "Cash at Bank",
		Type:          domain.AccountTypeASSET,
		IsActive:      true,
	})
	_ = env.accounts.Create(context.Background(), &domain.ChartOfAccounts{
		ID:            "acc_liability",
		LegalEntityID: "legal_123",
		AccountCode:   "2000",
		AccountName:   "Accounts Payable",
		Type:          domain.AccountTypeLIABILITY,
		IsActive:      true,
	})
	_ = env.accounts.Create(context.Background(), &domain.ChartOfAccounts{
		ID:            "acc_revenue",
		LegalEntityID: "legal_123",
		AccountCode:   "4000",
		AccountName:   "Sales Revenue",
		Type:          domain.AccountTypeREVENUE,
		IsActive:      true,
	})
	_ = env.accounts.Create(context.Background(), &domain.ChartOfAccounts{
		ID:            "acc_expense",
		LegalEntityID: "legal_123",
		AccountCode:   "5000",
		AccountName:   "Operating Expenses",
		Type:          domain.AccountTypeEXPENSE,
		IsActive:      true,
	})

	// 1. Balance Sheet
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/reports/balance-sheet", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 2. Income Statement
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/reports/income-statement", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 3. Cash Flow
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/reports/cash-flow", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLegalEntityEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Create Legal Entity success
	body, _ := json.Marshal(map[string]interface{}{
		"company_code":            "CORP_DE",
		"company_name":            "Corp DE",
		"functional_currency":     "EUR",
		"tax_registration_number": "DE123456789",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/legal-entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	// Parse response ID
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	leID := resp.Data.ID

	// 2. Get Legal Entity by ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/legal-entities/"+leID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Legal Entity 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/legal-entities/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 3. Get all Legal Entities
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/legal-entities", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 4. Create Legal Entity validation failure (missing company code)
	bodyBad, _ := json.Marshal(map[string]interface{}{
		"company_name": "Corp DE",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/legal-entities", bytes.NewBuffer(bodyBad))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected creation failure for missing code")
	}

	// Create Legal Entity bad JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/legal-entities", bytes.NewBuffer([]byte("{bad-json")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected creation failure for bad JSON")
	}
}

func TestAssetEndpoints(t *testing.T) {
	env := setupTestEnv()

	// 1. Capitalize Asset success
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id":    "legal_123",
		"asset_tag":          "EQ-001",
		"acquisition_cost":   "1200",
		"useful_life_months": 12,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/assets/capitalize", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assetID := resp.Data.ID

	// 2. Generate schedule
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/assets/"+assetID+"/depreciation-schedule", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 3. Post monthly depreciation
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "legal_123",
		"fiscal_year":     time.Now().Year(),
		"period_number":   int(time.Now().Month()),
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/assets/depreciate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	// 4. Get Asset by ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/assets/"+assetID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Asset 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/assets/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 5. Get all Assets
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/assets", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 6. Capitalize Asset bad JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/assets/capitalize", bytes.NewBuffer([]byte("{bad-json")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected capitalization failure for bad JSON")
	}

	// 7. Capitalize Asset invalid cost format
	bodyInvalidCost, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id":    "legal_123",
		"asset_tag":          "EQ-002",
		"acquisition_cost":   "invalid-decimal",
		"useful_life_months": 12,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/assets/capitalize", bytes.NewBuffer(bodyInvalidCost))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected capitalization failure for invalid cost format")
	}

	// 8. Capitalize Asset service failure
	bodyServiceFail, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id":    "", // Empty legal entity ID triggers service validation error
		"asset_tag":          "EQ-003",
		"acquisition_cost":   "1200",
		"useful_life_months": 12,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/assets/capitalize", bytes.NewBuffer(bodyServiceFail))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected capitalization failure for empty legal entity ID")
	}

	// 9. Generate schedule service failure
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/assets/non-existent/depreciation-schedule", nil)
	env.router.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected schedule generation failure for non-existent asset")
	}

	// 10. Post monthly depreciation bad JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/assets/depreciate", bytes.NewBuffer([]byte("{bad-json")))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected depreciation posting failure for bad JSON")
	}
}
