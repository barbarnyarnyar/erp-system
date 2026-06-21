package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const gatewayURL = "http://localhost:8080"

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type productResponse struct {
	Data struct {
		ID          string `json:"id"`
		ProductCode string `json:"product_code"`
	} `json:"data"`
}

type locationResponse struct {
	Data struct {
		ID           string `json:"id"`
		LocationCode string `json:"location_code"`
	} `json:"data"`
}

type inventoryResponse struct {
	Data struct {
		ID             string `json:"id"`
		QuantityOnHand string `json:"quantity_on_hand"`
	} `json:"data"`
}

type leadResponse struct {
	ID        string `json:"id"`
	Company   string `json:"company"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type convertResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
}

type salesOrderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Status     string `json:"status"`
}

type workCenterResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type stationResponse struct {
	ID          string `json:"id"`
	RoutingCode string `json:"routing_code"`
}

type workOrderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type glAccountResponse struct {
	Data struct {
		ID          string `json:"id"`
		AccountCode string `json:"account_code"`
	} `json:"data"`
}

type invoiceResponse struct {
	Data struct {
		ID            string `json:"id"`
		InvoiceNumber string `json:"invoice_number"`
	} `json:"data"`
}

type paymentResponse struct {
	Data struct {
		ID            string `json:"id"`
		PaymentNumber string `json:"payment_number"`
	} `json:"data"`
}

type balanceResponse struct {
	Balance string `json:"balance"`
}

func TestE2E_SalesToCashLifecycle(t *testing.T) {
	// 1. Login to Gateway to obtain JWT
	token := login(t)
	t.Logf("Step 1 Passed: JWT Token obtained")

	// 2. Create SCM Product
	productID := createProduct(t, token)
	t.Logf("Step 2 Passed: SCM Product created with ID: %s", productID)

	// 3. Create SCM Location
	locationID := createLocation(t, token)
	t.Logf("Step 3 Passed: SCM Location created with ID: %s", locationID)

	// 4. Create SCM Inventory Item (Initial Stock = 2)
	invItemID := createInventoryItem(t, token, productID, locationID, 2)
	t.Logf("Step 4 Passed: SCM Inventory Item created with ID: %s", invItemID)

	// 5. Create Lead (CRM)
	leadID := createLead(t, token)
	t.Logf("Step 5 Passed: CRM Lead created with ID: %s", leadID)

	// 6. Convert Lead to Customer and Opportunity
	customerID, opportunityID := convertLead(t, token, leadID)
	t.Logf("Step 6 Passed: CRM Lead converted. Customer ID: %s, Opportunity ID: %s", customerID, opportunityID)

	// 7. Create Sales Order (CRM) of Quantity 10
	salesOrderID := createSalesOrder(t, token, customerID, productID, 10)
	t.Logf("Step 7 Passed: CRM Sales Order created with ID: %s", salesOrderID)

	// 8. Assert Stock Reservation fails because qty on hand is 2 (requested 10)
	tryReserveStockFails(t, token, productID, locationID, salesOrderID, 10)
	t.Logf("Step 8 Passed: SCM Stock Reservation failed as expected due to low stock")

	// 9. Establish Work Center (MFG)
	wcID := establishWorkCenter(t, token)
	t.Logf("Step 9 Passed: MFG Work Center established with ID: %s", wcID)

	// 10. Append Routing Station to Center (MFG)
	stationID := appendStation(t, token, wcID)
	t.Logf("Step 10 Passed: MFG Routing Station appended with ID: %s", stationID)

	// 11. Instantiate Work Order (MFG) for target quantity 50
	woID := instantiateWorkOrder(t, token, productID)
	t.Logf("Step 11 Passed: MFG Work Order instantiated with ID: %s", woID)

	// 12. Transition Work Order: STAGED -> RELEASED
	transitionWorkOrder(t, token, woID, "STAGED", "RELEASED")
	t.Logf("Step 12 Passed: Work Order transitioned STAGED -> RELEASED")

	// 13. Transition Work Order: RELEASED -> IN_PROGRESS
	transitionWorkOrder(t, token, woID, "RELEASED", "IN_PROGRESS")
	t.Logf("Step 13 Passed: Work Order transitioned RELEASED -> IN_PROGRESS")

	// 14. Commit Production Yield (MFG: 50 units good)
	commitYield(t, token, woID, stationID, "50.0")
	t.Logf("Step 14 Passed: MFG Production Yield committed successfully")

	// 15. Transition Work Order: IN_PROGRESS -> COMPLETED
	transitionWorkOrder(t, token, woID, "IN_PROGRESS", "COMPLETED")
	t.Logf("Step 15 Passed: Work Order transitioned IN_PROGRESS -> COMPLETED")

	// 16. Update SCM Inventory to reflect release of finished goods (New Qty = 52)
	updateInventoryOnHand(t, token, invItemID, 52)
	t.Logf("Step 16 Passed: SCM Inventory Item updated to reflect new stock (qty=52)")

	// 17. Confirm Sales Order (CRM)
	confirmSalesOrder(t, token, salesOrderID)
	t.Logf("Step 17 Passed: CRM Sales Order confirmed")

	// 18. Reserve Stock (SCM: 10 units). This should now succeed!
	reserveStockSucceeds(t, token, productID, locationID, salesOrderID, 10)
	t.Logf("Step 18 Passed: SCM Stock Reservation succeeded for Sales Order")

	// 19. Create General Ledger Accounts (FM: Bank Account and AR Account)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	// Keep suffix length reasonable so total length is < 50 chars (max length for account_code)
	if len(suffix) > 15 {
		suffix = suffix[len(suffix)-15:]
	}
	bankAccountCode := "1010-e2e-" + suffix
	arAccountCode := "1100-e2e-" + suffix

	bankAccountID := createGLAccount(t, token, bankAccountCode, "E2E Bank Account")
	arAccountID := createGLAccount(t, token, arAccountCode, "E2E Accounts Receivable")
	t.Logf("Step 19 Passed: FM GL Accounts created. Bank: %s (%s), AR: %s (%s)", bankAccountID, bankAccountCode, arAccountID, arAccountCode)

	// 20. Generate Invoice (FM)
	invoiceID := createInvoice(t, token, customerID, salesOrderID, "2500.00", "200.00")
	t.Logf("Step 20 Passed: FM Invoice generated with ID: %s", invoiceID)

	// 21. Record Payment (FM)
	paymentID := recordPayment(t, token, invoiceID, bankAccountID, "2500.00")
	t.Logf("Step 21 Passed: FM Payment recorded with ID: %s", paymentID)

	// 22. Post balanced GL Journal Entry representing payment collection
	createJournalEntry(t, token, bankAccountID, arAccountID, invoiceID, "2500.00")
	t.Logf("Step 22 Passed: GL Journal Entry posted successfully")

	// 23. Assert GL Bank Account balance reflects the updated balance (+2500.00)
	assertGLBalance(t, token, bankAccountID, "2500")
	t.Logf("Step 23 Passed: GL Balance verified successfully! Balance is exactly 2500.00")
}

func login(t *testing.T) string {
	url := fmt.Sprintf("%s/api/v1/auth/login", gatewayURL)
	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "admin123",
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login returned status code %d", resp.StatusCode)
	}

	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	return lr.AccessToken
}

func newRequest(t *testing.T, method, path string, body interface{}, token string) *http.Request {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, gatewayURL+path, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func doRequest(t *testing.T, req *http.Request) *http.Response {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	return resp
}

func createProduct(t *testing.T, token string) string {
	req := newRequest(t, "POST", "/api/v1/scm/products", map[string]interface{}{
		"product_code":    fmt.Sprintf("PROD-E2E-%d", time.Now().Unix()),
		"product_name":    "E2E Finished Product",
		"description":     "Product generated for Sales-to-Cash E2E flow testing",
		"product_type":    "FINISHED_GOOD",
		"unit_of_measure": "EA",
		"standard_cost":   "100.00",
		"list_price":      "250.00",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CreateProduct failed with status %d: %s", resp.StatusCode, string(body))
	}

	var pr productResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		t.Fatalf("Failed to decode product response: %v", err)
	}

	return pr.Data.ID
}

func createLocation(t *testing.T, token string) string {
	req := newRequest(t, "POST", "/api/v1/scm/locations", map[string]interface{}{
		"location_code": fmt.Sprintf("LOC-E2E-%d", time.Now().Unix()),
		"location_name": "E2E Storage Warehouse Location",
		"location_type": "WAREHOUSE",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CreateLocation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var lr locationResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		t.Fatalf("Failed to decode location response: %v", err)
	}

	return lr.Data.ID
}

func createInventoryItem(t *testing.T, token string, productID, locationID string, qty int) string {
	req := newRequest(t, "POST", "/api/v1/scm/inventory", map[string]interface{}{
		"product_id":       productID,
		"location_id":      locationID,
		"quantity_on_hand": fmt.Sprintf("%d", qty),
		"reorder_point":    5,
		"maximum_stock":    100,
		"unit_cost":        "100.00",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CreateInventoryItem failed with status %d: %s", resp.StatusCode, string(body))
	}

	var ir inventoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		t.Fatalf("Failed to decode inventory response: %v", err)
	}

	return ir.Data.ID
}

func createLead(t *testing.T, token string) string {
	req := newRequest(t, "POST", "/api/v1/crm/leads", map[string]interface{}{
		"first_name": "E2E",
		"last_name":  "Buyer",
		"company":    "E2E Testing Corporation",
		"email":      "e2ebuyer@e2etesting.corp",
		"phone":      "+1-555-9876",
		"source":     "REFERRAL",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CreateLead failed with status %d: %s", resp.StatusCode, string(body))
	}

	var lr leadResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		t.Fatalf("Failed to decode lead response: %v", err)
	}

	return lr.ID
}

func convertLead(t *testing.T, token string, leadID string) (string, string) {
	req := newRequest(t, "POST", fmt.Sprintf("/api/v1/crm/leads/%s/convert", leadID), nil, token)
	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("ConvertLead failed with status %d: %s", resp.StatusCode, string(body))
	}

	var cr convertResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		t.Fatalf("Failed to decode conversion response: %v", err)
	}

	if cr.CustomerID == "" {
		t.Fatalf("Conversion response did not contain customer ID")
	}

	return cr.CustomerID, cr.ID
}

func createSalesOrder(t *testing.T, token string, customerID, productID string, qty int) string {
	req := newRequest(t, "POST", "/api/v1/crm/sales-orders", map[string]interface{}{
		"customer_id": customerID,
		"items": []map[string]interface{}{
			{
				"product_id": productID,
				"quantity":   qty,
				"unit_price": "250.00",
				"discount":   "0.0",
			},
		},
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CreateSalesOrder failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sor salesOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&sor); err != nil {
		t.Fatalf("Failed to decode sales order response: %v", err)
	}

	return sor.ID
}

func tryReserveStockFails(t *testing.T, token string, productID, locationID, salesOrderID string, qty int) {
	req := newRequest(t, "POST", "/api/v1/scm/inventory/reserve", map[string]interface{}{
		"product_id":   productID,
		"location_id":  locationID,
		"quantity":     fmt.Sprintf("%d", qty),
		"reference_id": salesOrderID,
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("Expected stock reservation to FAIL due to low inventory, but it succeeded (HTTP 200)")
	}
}

func establishWorkCenter(t *testing.T, token string) string {
	req := newRequest(t, "POST", "/api/v1/manufacturing/mfg/work-centers", map[string]interface{}{
		"legal_entity_id": "le_e2e",
		"code":            fmt.Sprintf("WC-%d", time.Now().Unix()),
		"name":            "E2E Production Center",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("EstablishWorkCenter failed with status %d: %s", resp.StatusCode, string(body))
	}

	var wcr workCenterResponse
	if err := json.NewDecoder(resp.Body).Decode(&wcr); err != nil {
		t.Fatalf("Failed to decode work center response: %v", err)
	}

	return wcr.ID
}

func appendStation(t *testing.T, token string, wcID string) string {
	req := newRequest(t, "POST", fmt.Sprintf("/api/v1/manufacturing/mfg/work-centers/%s/stations", wcID), map[string]interface{}{
		"routing_code":             "ROUT-E2E",
		"station_type":             "ASSEMBLY",
		"standard_setup_time_mins": 10,
		"standard_run_time_mins":   20,
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("AppendStation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sr stationResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		t.Fatalf("Failed to decode station response: %v", err)
	}

	return sr.ID
}

func instantiateWorkOrder(t *testing.T, token string, productID string) string {
	req := newRequest(t, "POST", "/api/v1/manufacturing/mfg/work-orders", map[string]interface{}{
		"legal_entity_id": "le_e2e",
		"material_id":     productID,
		"bom_header_id":   "bom_default",
		"quantity_target": "50.0",
		"scheduled_start": time.Now().Add(-1 * time.Hour),
		"scheduled_end":   time.Now().Add(4 * time.Hour),
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("InstantiateWorkOrder failed with status %d: %s", resp.StatusCode, string(body))
	}

	var wor workOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&wor); err != nil {
		t.Fatalf("Failed to decode work order response: %v", err)
	}

	return wor.ID
}

func transitionWorkOrder(t *testing.T, token string, woID string, current, target string) {
	req := newRequest(t, "POST", fmt.Sprintf("/api/v1/manufacturing/mfg/work-orders/%s/transition", woID), map[string]interface{}{
		"current_state": current,
		"target_state":  target,
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("TransitionWorkOrder from %s to %s failed with status %d: %s", current, target, resp.StatusCode, string(body))
	}
}

func commitYield(t *testing.T, token string, woID string, stationID string, qtyGood string) {
	req := newRequest(t, "POST", fmt.Sprintf("/api/v1/manufacturing/mfg/work-orders/%s/yield", woID), map[string]interface{}{
		"legal_entity_id": "le_e2e",
		"station_id":      stationID,
		"quantity_good":   qtyGood,
		"quantity_scrap":  "0.0",
		"operator_hr_id":  "op_e2e",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CommitYield failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func updateInventoryOnHand(t *testing.T, token string, invItemID string, qty int) {
	req := newRequest(t, "PUT", fmt.Sprintf("/api/v1/scm/inventory/%s", invItemID), map[string]interface{}{
		"quantity_on_hand":  fmt.Sprintf("%d", qty),
		"quantity_reserved": "0",
		"reorder_point":     5,
		"maximum_stock":     100,
		"unit_cost":         "100.00",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("UpdateInventoryOnHand failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func confirmSalesOrder(t *testing.T, token string, salesOrderID string) {
	req := newRequest(t, "PUT", fmt.Sprintf("/api/v1/crm/sales-orders/%s", salesOrderID), map[string]interface{}{
		"status": "CONFIRMED",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("ConfirmSalesOrder failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func reserveStockSucceeds(t *testing.T, token string, productID, locationID, salesOrderID string, qty int) {
	req := newRequest(t, "POST", "/api/v1/scm/inventory/reserve", map[string]interface{}{
		"product_id":   productID,
		"location_id":  locationID,
		"quantity":     fmt.Sprintf("%d", qty),
		"reference_id": salesOrderID,
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("ReserveStock failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func createGLAccount(t *testing.T, token string, code, name string) string {
	req := newRequest(t, "POST", "/api/v1/finance/accounts", map[string]interface{}{
		"legal_entity_id": "le_e2e",
		"account_code":    code,
		"account_name":    name,
		"type":            "ASSET",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CreateGLAccount (%s) failed with status %d: %s", code, resp.StatusCode, string(body))
	}

	var gar glAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&gar); err != nil {
		t.Fatalf("Failed to decode GL Account response: %v", err)
	}

	return gar.Data.ID
}

func createInvoice(t *testing.T, token string, customerID, salesOrderID string, total, tax string) string {
	req := newRequest(t, "POST", "/api/v1/finance/invoices", map[string]interface{}{
		"legal_entity_id": "le_e2e",
		"customer_id":     customerID,
		"sales_order_id":  salesOrderID,
		"total_amount":    total,
		"tax_amount":      tax,
		"due_date":        time.Now().AddDate(0, 0, 30),
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CreateInvoice failed with status %d: %s", resp.StatusCode, string(body))
	}

	var ir invoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		t.Fatalf("Failed to decode invoice response: %v", err)
	}

	return ir.Data.ID
}

func recordPayment(t *testing.T, token string, invoiceID, bankAccountID string, amount string) string {
	req := newRequest(t, "POST", "/api/v1/finance/payments", map[string]interface{}{
		"invoice_id":      invoiceID,
		"bank_account_id": bankAccountID,
		"amount":          amount,
		"payment_method":  "CASH",
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("RecordPayment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var pr paymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		t.Fatalf("Failed to decode payment response: %v", err)
	}

	return pr.Data.ID
}

func createJournalEntry(t *testing.T, token string, bankAccountID, arAccountID, invoiceID string, amount string) {
	req := newRequest(t, "POST", "/api/v1/finance/journal-entries", map[string]interface{}{
		"legal_entity_id":    "le_e2e",
		"source_module":      "AR",
		"source_document_id": invoiceID,
		"posting_date":       time.Now(),
		"lines": []map[string]interface{}{
			{
				"account_id":             bankAccountID,
				"amount_functional":      amount,
				"amount_transactional":   amount,
				"currency_transactional": "USD",
			},
			{
				"account_id":             arAccountID,
				"amount_functional":      "-" + amount,
				"amount_transactional":   "-" + amount,
				"currency_transactional": "USD",
			},
		},
	}, token)

	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("CreateJournalEntry failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func assertGLBalance(t *testing.T, token string, accountID string, expected string) {
	req := newRequest(t, "GET", fmt.Sprintf("/api/v1/finance/accounts/%s/balance", accountID), nil, token)
	resp := doRequest(t, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GetAccountBalance failed with status %d: %s", resp.StatusCode, string(body))
	}

	var br balanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil {
		t.Fatalf("Failed to decode balance response: %v", err)
	}

	if br.Balance != expected {
		t.Fatalf("GL balance mismatch: expected %s, got %s", expected, br.Balance)
	}
}
