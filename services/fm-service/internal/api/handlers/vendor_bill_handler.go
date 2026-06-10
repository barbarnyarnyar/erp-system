package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type VendorBillHandler struct {
	svc *service.AccountsPayableService
}

func NewVendorBillHandler(svc *service.AccountsPayableService) *VendorBillHandler {
	return &VendorBillHandler{svc: svc}
}

func (h *VendorBillHandler) GetVendorBills(c *gin.Context) {
	bills, err := h.svc.ListVendorBills(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": bills})
}

func (h *VendorBillHandler) CreateVendorBill(c *gin.Context) {
	var req struct {
		SupplierID      string    `json:"supplier_id"`
		BillNumber      string    `json:"bill_number"`
		PurchaseOrderID string    `json:"purchase_order_id"`
		IssueDate       time.Time `json:"issue_date"`
		DueDate         time.Time `json:"due_date"`
		TotalAmount     string    `json:"total_amount"`
		Lines           []struct {
			Description string `json:"description"`
			Quantity    int    `json:"quantity"`
			UnitPrice   string `json:"unit_price"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalDec, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		totalDec = decimal.Zero
	}

	domainLines := make([]domain.VendorBillLine, len(req.Lines))
	for i, l := range req.Lines {
		priceDec, err := decimal.NewFromString(l.UnitPrice)
		if err != nil {
			priceDec = decimal.Zero
		}
		domainLines[i] = domain.VendorBillLine{
			Description: l.Description,
			Quantity:    l.Quantity,
			UnitPrice:   priceDec,
			LineTotal:   priceDec.Mul(decimal.NewFromInt(int64(l.Quantity))),
		}
	}

	bill, err := h.svc.CreateVendorBill(
		c.Request.Context(),
		req.SupplierID,
		req.BillNumber,
		req.PurchaseOrderID,
		req.IssueDate,
		req.DueDate,
		totalDec,
		domainLines,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": bill})
}

func (h *VendorBillHandler) GetVendorBillLines(c *gin.Context) {
	id := c.Param("id")
	_, lines, err := h.svc.GetVendorBill(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vendor bill not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": lines,
	})
}
