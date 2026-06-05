package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/shopspring/decimal"
)

type PaymentHandler struct {
	svc *service.CashManagementService
}

func NewPaymentHandler(svc *service.CashManagementService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) GetPayments(c *gin.Context) {
	payments, err := h.svc.ListPayments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payments})
}

func (h *PaymentHandler) RecordPayment(c *gin.Context) {
	var req struct {
		InvoiceID     string `json:"invoice_id"`
		BillID        string `json:"bill_id"`
		Amount        string `json:"amount"`
		PaymentMethod string `json:"payment_method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amountDec, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment amount"})
		return
	}

	payment, err := h.svc.RecordPayment(c.Request.Context(), req.InvoiceID, req.BillID, amountDec, req.PaymentMethod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": payment})
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")
	payment, err := h.svc.GetPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payment})
}
