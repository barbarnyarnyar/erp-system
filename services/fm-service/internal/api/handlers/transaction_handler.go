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

type TransactionHandler struct {
	svc      *service.GeneralLedgerService
	response *utils.ResponseHelper
}

func NewTransactionHandler(svc *service.GeneralLedgerService, response *utils.ResponseHelper) *TransactionHandler {
	return &TransactionHandler{
		svc:      svc,
		response: response,
	}
}

func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	entries, err := h.svc.ListJournalEntries(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entries})
}

func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req struct {
		LegalEntityID    string    `json:"legal_entity_id"`
		SourceModule     string    `json:"source_module"`
		SourceDocumentID string    `json:"source_document_id"`
		PostingDate      time.Time `json:"posting_date"`
		Lines            []struct {
			AccountID             string `json:"account_id"`
			AmountFunctional      string `json:"amount_functional"`
			AmountTransactional   string `json:"amount_transactional"`
			CurrencyTransactional string `json:"currency_transactional"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	domainLines := make([]domain.UniversalJournalLine, len(req.Lines))
	for i, l := range req.Lines {
		amtFunc, err := decimal.NewFromString(l.AmountFunctional)
		if err != nil {
			amtFunc = decimal.Zero
		}
		amtTrans, err := decimal.NewFromString(l.AmountTransactional)
		if err != nil {
			amtTrans = decimal.Zero
		}

		domainLines[i] = domain.UniversalJournalLine{
			AccountID:             l.AccountID,
			AmountFunctional:      amtFunc,
			AmountTransactional:   amtTrans,
			CurrencyTransactional: l.CurrencyTransactional,
		}
	}

	entry, err := h.svc.CreateJournalEntry(c.Request.Context(), req.LegalEntityID, req.SourceModule, req.SourceDocumentID, req.PostingDate, domainLines)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": entry})
}

func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	id := c.Param("id")
	entry, lines, err := h.svc.GetJournalEntry(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "journal entry not found")
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
		LegalEntityID    string    `json:"legal_entity_id"`
		SourceModule     string    `json:"source_module"`
		SourceDocumentID string    `json:"source_document_id"`
		PostingDate      time.Time `json:"posting_date"`
		Lines            []struct {
			AccountID             string `json:"account_id"`
			AmountFunctional      string `json:"amount_functional"`
			AmountTransactional   string `json:"amount_transactional"`
			CurrencyTransactional string `json:"currency_transactional"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	domainLines := make([]domain.UniversalJournalLine, len(req.Lines))
	for i, l := range req.Lines {
		amtFunc, err := decimal.NewFromString(l.AmountFunctional)
		if err != nil {
			amtFunc = decimal.Zero
		}
		amtTrans, err := decimal.NewFromString(l.AmountTransactional)
		if err != nil {
			amtTrans = decimal.Zero
		}

		domainLines[i] = domain.UniversalJournalLine{
			AccountID:             l.AccountID,
			AmountFunctional:      amtFunc,
			AmountTransactional:   amtTrans,
			CurrencyTransactional: l.CurrencyTransactional,
		}
	}

	entry, err := h.svc.UpdateJournalEntry(c.Request.Context(), id, req.LegalEntityID, req.SourceModule, req.SourceDocumentID, req.PostingDate, domainLines)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entry})
}

func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteJournalEntry(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "journal entry deleted successfully"})
}

