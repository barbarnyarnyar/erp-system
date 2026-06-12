package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PaymentHandler struct {
	svc *service.CashManagementService
	response *utils.ResponseHelper
}

func NewPaymentHandler(svc *service.CashManagementService, response *utils.ResponseHelper) *PaymentHandler {
	return &PaymentHandler{
		svc: svc,
		response: response,
	}
}

func (h *PaymentHandler) GetPayments(c *gin.Context) {
	payments, err := h.svc.ListPayments(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payments})
}

func (h *PaymentHandler) RecordPayment(c *gin.Context) {
	var req struct {
		InvoiceID     string `json:"invoice_id"`
		BillID        string `json:"bill_id"`
		BankAccountID string `json:"bank_account_id"`
		Amount        string `json:"amount"`
		PaymentMethod string `json:"payment_method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	amountDec, err := decimal.NewFromString(req.Amount)
	if err != nil {
		h.response.BadRequest(c, "invalid payment amount")
		return
	}

	payment, err := h.svc.RecordPayment(c.Request.Context(), req.InvoiceID, req.BillID, req.BankAccountID, amountDec, req.PaymentMethod)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": payment})
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")
	payment, err := h.svc.GetPayment(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "payment not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payment})
}

func (h *PaymentHandler) GetBankStatementLines(c *gin.Context) {
	id := c.Param("id")
	_, lines, err := h.svc.GetBankStatement(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "bank statement not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": lines,
	})
}
