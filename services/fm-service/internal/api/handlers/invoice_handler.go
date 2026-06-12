package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type InvoiceHandler struct {
	svc      *service.AccountsReceivableService
	response *utils.ResponseHelper
}

func NewInvoiceHandler(svc *service.AccountsReceivableService, response *utils.ResponseHelper) *InvoiceHandler {
	return &InvoiceHandler{
		svc:      svc,
		response: response,
	}
}

func (h *InvoiceHandler) GetInvoices(c *gin.Context) {
	invoices, err := h.svc.ListInvoices(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": invoices})
}

func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req struct {
		LegalEntityID string    `json:"legal_entity_id"`
		CustomerID    string    `json:"customer_id"`
		SalesOrderID  string    `json:"sales_order_id"`
		TotalAmount   string    `json:"total_amount"`
		TaxAmount     string    `json:"tax_amount"`
		DueDate       time.Time `json:"due_date"`
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

	invoice, err := h.svc.CreateInvoice(c.Request.Context(), req.LegalEntityID, req.CustomerID, req.SalesOrderID, totalDec, taxDec, req.DueDate)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": invoice})
}

func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	invoice, err := h.svc.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "invoice not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": invoice,
	})
}

func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	invoice, err := h.svc.UpdateInvoice(c.Request.Context(), id, req)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoice})
}

func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteInvoice(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "invoice deleted successfully"})
}

func (h *InvoiceHandler) SendInvoice(c *gin.Context) {
	id := c.Param("id")
	ok, err := h.svc.SendInvoice(c.Request.Context(), id)
	if err != nil || !ok {
		h.response.InternalServerError(c, "failed to send invoice", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "invoice sent successfully"})
}

func (h *InvoiceHandler) GetInvoiceLines(c *gin.Context) {
	id := c.Param("id")
	_, err := h.svc.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "invoice not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": []string{},
	})
}
