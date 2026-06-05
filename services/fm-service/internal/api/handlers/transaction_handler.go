package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/shopspring/decimal"
)

type TransactionHandler struct {
	svc *service.GeneralLedgerService
}

func NewTransactionHandler(svc *service.GeneralLedgerService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	entries, err := h.svc.ListJournalEntries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entries})
}

func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req struct {
		Reference   string `json:"reference"`
		Description string `json:"description"`
		Lines       []struct {
			AccountID    string `json:"account_id"`
			DebitAmount  string `json:"debit_amount"`
			CreditAmount string `json:"credit_amount"`
			Description  string `json:"description"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Helper function inside handler to parse decimal inputs safely
	domainLines := make([]domain.JournalEntryLine, len(req.Lines))
	for i, l := range req.Lines {
		debitDec, err := decimal.NewFromString(l.DebitAmount)
		if err != nil {
			debitDec = decimal.Zero
		}
		creditDec, err := decimal.NewFromString(l.CreditAmount)
		if err != nil {
			creditDec = decimal.Zero
		}

		domainLines[i] = domain.JournalEntryLine{
			AccountID:    l.AccountID,
			DebitAmount:  debitDec,
			CreditAmount: creditDec,
			Description:  l.Description,
		}
	}

	entry, err := h.svc.CreateJournalEntry(c.Request.Context(), req.Reference, req.Description, domainLines)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": entry})
}

func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	id := c.Param("id")
	entry, lines, err := h.svc.GetJournalEntry(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "journal entry not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  entry,
		"lines": lines,
	})
}

func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reference   string `json:"reference"`
		Description string `json:"description"`
		Lines       []struct {
			AccountID    string `json:"account_id"`
			DebitAmount  string `json:"debit_amount"`
			CreditAmount string `json:"credit_amount"`
			Description  string `json:"description"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domainLines := make([]domain.JournalEntryLine, len(req.Lines))
	for i, l := range req.Lines {
		debitDec, err := decimal.NewFromString(l.DebitAmount)
		if err != nil {
			debitDec = decimal.Zero
		}
		creditDec, err := decimal.NewFromString(l.CreditAmount)
		if err != nil {
			creditDec = decimal.Zero
		}

		domainLines[i] = domain.JournalEntryLine{
			AccountID:    l.AccountID,
			DebitAmount:  debitDec,
			CreditAmount: creditDec,
			Description:  l.Description,
		}
	}

	entry, err := h.svc.UpdateJournalEntry(c.Request.Context(), id, req.Reference, req.Description, domainLines)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entry})
}

func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteJournalEntry(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "journal entry deleted successfully"})
}