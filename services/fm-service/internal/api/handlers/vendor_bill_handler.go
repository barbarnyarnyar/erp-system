package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type VendorBillHandler struct {
	svc      *service.AccountsPayableService
	response *utils.ResponseHelper
}

func NewVendorBillHandler(svc *service.AccountsPayableService, response *utils.ResponseHelper) *VendorBillHandler {
	return &VendorBillHandler{
		svc:      svc,
		response: response,
	}
}

func (h *VendorBillHandler) GetVendorBills(c *gin.Context) {
	bills, err := h.svc.ListVendorBills(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": bills})
}

func (h *VendorBillHandler) CreateVendorBill(c *gin.Context) {
	var req struct {
		LegalEntityID   string    `json:"legal_entity_id"`
		VendorID        string    `json:"vendor_id"`
		BillNumber      string    `json:"bill_number"`
		PurchaseOrderID string    `json:"purchase_order_id"`
		DueDate         time.Time `json:"due_date"`
		TotalAmount     string    `json:"total_amount"`
		TaxAmount       string    `json:"tax_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	totalDec, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		totalDec = decimal.Zero
	}
	taxDec, err := decimal.NewFromString(req.TaxAmount)
	if err != nil {
		taxDec = decimal.Zero
	}

	bill, err := h.svc.CreateVendorBill(
		c.Request.Context(),
		req.LegalEntityID,
		req.VendorID,
		req.BillNumber,
		req.PurchaseOrderID,
		req.DueDate,
		totalDec,
		taxDec,
	)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": bill})
}

func (h *VendorBillHandler) GetVendorBillLines(c *gin.Context) {
	id := c.Param("id")
	_, err := h.svc.GetVendorBill(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "vendor bill not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": []string{},
	})
}

