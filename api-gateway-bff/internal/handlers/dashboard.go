package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"api-gateway-bff/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type DashboardHandler struct {
	cfg *config.Config
}

func NewDashboardHandler(cfg *config.Config) *DashboardHandler {
	return &DashboardHandler{cfg: cfg}
}

// CRM responses
type CRMOrderResponse struct {
	ID              string          `json:"id"`
	LegalEntityID   string          `json:"legal_entity_id"`
	CustomerID      string          `json:"customer_id"`
	PriceBookID     string          `json:"price_book_id"`
	OrderNumber     string          `json:"order_number"`
	Status          string          `json:"status"`
	TotalGrossValue decimal.Decimal `json:"total_gross_value"`
	TotalTaxValue   decimal.Decimal `json:"total_tax_value"`
	Version         int             `json:"version"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type CRMOrderLineResponse struct {
	ID              string          `json:"id"`
	SalesOrderID    string          `json:"sales_order_id"`
	MaterialID      string          `json:"material_id"`
	LineSequence    int             `json:"line_sequence"`
	QuantityOrdered decimal.Decimal `json:"quantity_ordered"`
	QuantityShipped decimal.Decimal `json:"quantity_shipped"`
	UnitSellPrice   decimal.Decimal `json:"unit_sell_price"`
	DiscountApplied decimal.Decimal `json:"discount_applied"`
	NetLineAmount   decimal.Decimal `json:"net_line_amount"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// FM responses
type FMCustomerCreditEnvelope struct {
	Data FMCustomerCreditResponse `json:"data"`
}

type FMCustomerCreditResponse struct {
	ID             string          `json:"id"`
	CustomerID     string          `json:"customer_id"`
	CreditLimit    decimal.Decimal `json:"credit_limit"`
	CurrentBalance decimal.Decimal `json:"current_balance"`
	IsOnHold       bool            `json:"is_on_hold"`
}

// SCM responses
type SCMMaterialEnvelope struct {
	Data SCMMaterialResponse `json:"data"`
}

type SCMMaterialResponse struct {
	ID            string          `json:"id"`
	ProductCode   string          `json:"product_code"`
	ProductName   string          `json:"product_name"`
	Description   string          `json:"description"`
	ProductType   string          `json:"product_type"`
	UnitOfMeasure string          `json:"unit_of_measure"`
	StandardCost  decimal.Decimal `json:"standard_cost"`
	ListPrice     decimal.Decimal `json:"list_price"`
	IsActive      bool            `json:"is_active"`
}

type SalesDashboardResponse struct {
	Order           CRMOrderResponse               `json:"order"`
	Lines           []CRMOrderLineResponse         `json:"lines"`
	CustomerCredit  *FMCustomerCreditResponse      `json:"customer_credit,omitempty"`
	Materials       map[string]SCMMaterialResponse `json:"materials,omitempty"`
	InventoryStatus string                         `json:"inventory_status"`
}

func (h *DashboardHandler) GetSalesDashboard(c *gin.Context) {
	orderID := c.Param("order_id")
	authHeader := c.GetHeader("Authorization")

	// Use a timed context to enforce the 3-second strict timeout requirement
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// 1. Fetch CRM Order (Primary Context)
	var order CRMOrderResponse
	orderURL := fmt.Sprintf("%s/api/v1/orders/%s", h.cfg.CRMServiceURL, orderID)
	if err := fetchJSON(ctx, "GET", orderURL, authHeader, &order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch order: %v", err)})
		return
	}

	// 2. Fetch CRM Order Lines (Primary Context)
	var lines []CRMOrderLineResponse
	linesURL := fmt.Sprintf("%s/api/v1/orders/%s/lines", h.cfg.CRMServiceURL, orderID)
	if err := fetchJSON(ctx, "GET", linesURL, authHeader, &lines); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch order lines: %v", err)})
		return
	}

	customerID := order.CustomerID
	uniqueMaterialIDs := make(map[string]bool)
	for _, line := range lines {
		if line.MaterialID != "" {
			uniqueMaterialIDs[line.MaterialID] = true
		}
	}

	// 3. Concurrently fetch FM customer credit and SCM material details
	var credit FMCustomerCreditEnvelope
	materialsMap := make(map[string]SCMMaterialResponse)
	var materialsMu sync.Mutex

	g, gCtx := errgroup.WithContext(ctx)

	// Fetch customer credit (if FM fails, we treat it as a hard error/failure of primary credit info)
	g.Go(func() error {
		creditURL := fmt.Sprintf("%s/api/v1/customers/%s/credit", h.cfg.FMServiceURL, customerID)
		if err := fetchJSON(gCtx, "GET", creditURL, authHeader, &credit); err != nil {
			return fmt.Errorf("failed to fetch customer credit: %w", err)
		}
		return nil
	})

	// Fetch materials from SCM
	var scmFailed bool
	var scmErr error
	var scmMu sync.Mutex

	for matID := range uniqueMaterialIDs {
		id := matID
		g.Go(func() error {
			matURL := fmt.Sprintf("%s/api/v1/materials/%s", h.cfg.SCMServiceURL, id)
			var mat SCMMaterialEnvelope
			if err := fetchJSON(gCtx, "GET", matURL, authHeader, &mat); err != nil {
				scmMu.Lock()
				scmFailed = true
				scmErr = err
				scmMu.Unlock()
				// Return nil so we don't abort other sibling calls (like FM credit fetch)
				return nil
			}
			materialsMu.Lock()
			materialsMap[id] = mat.Data
			materialsMu.Unlock()
			return nil
		})
	}

	// Wait for concurrent calls
	if err := g.Wait(); err != nil {
		// Hard error if FM call fails
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Determine inventory status based on SCM failure or timeouts
	inventoryStatus := "AVAILABLE"
	scmMu.Lock()
	if scmFailed || scmErr != nil {
		inventoryStatus = "UNAVAILABLE"
	}
	scmMu.Unlock()

	resp := SalesDashboardResponse{
		Order:           order,
		Lines:           lines,
		CustomerCredit:  &credit.Data,
		Materials:       materialsMap,
		InventoryStatus: inventoryStatus,
	}

	c.JSON(http.StatusOK, resp)
}

func fetchJSON(ctx context.Context, method, url, authHeader string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
