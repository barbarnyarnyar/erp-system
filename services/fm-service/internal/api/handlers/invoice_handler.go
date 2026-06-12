package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type InvoiceHandler struct {
	svc *service.AccountsReceivableService
	response *utils.ResponseHelper
}

func NewInvoiceHandler(svc *service.AccountsReceivableService, response *utils.ResponseHelper) *InvoiceHandler {
	return &InvoiceHandler{
		svc: svc,
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
		CustomerID string    `json:"customer_id"`
		IssueDate  time.Time `json:"issue_date"`
		DueDate    time.Time `json:"due_date"`
		Lines      []struct {
			Description string `json:"description"`
			Quantity    int    `json:"quantity"`
			UnitPrice   string `json:"unit_price"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	domainLines := make([]domain.InvoiceLine, len(req.Lines))
	for i, l := range req.Lines {
		priceDec, err := decimal.NewFromString(l.UnitPrice)
		if err != nil {
			priceDec = decimal.Zero
		}
		domainLines[i] = domain.InvoiceLine{
			Description: l.Description,
			Quantity:    l.Quantity,
			UnitPrice:   priceDec,
			LineTotal:   priceDec.Mul(decimal.NewFromInt(int64(l.Quantity))),
		}
	}

	invoice, err := h.svc.CreateInvoice(c.Request.Context(), req.CustomerID, req.IssueDate, req.DueDate, domainLines)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": invoice})
}

func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	invoice, lines, err := h.svc.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "invoice not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  invoice,
		"lines": lines,
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
	_, lines, err := h.svc.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "invoice not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": lines,
	})
}
