package handlers

import (
	"net/http"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	svc *service.GeneralLedgerService
}

func NewAccountHandler(svc *service.GeneralLedgerService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	accs, err := h.svc.ListAccounts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": accs})
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req struct {
		AccountNumber string `json:"account_number"`
		Name          string `json:"name"`
		Type          string `json:"type"`
		ParentID      string `json:"parent_id"`
		Currency      string `json:"currency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.svc.CreateAccount(c.Request.Context(), req.AccountNumber, req.Name, req.Type, req.ParentID, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": acc})
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	acc, err := h.svc.GetAccount(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": acc})
}

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		ParentID string `json:"parent_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.svc.UpdateAccount(c.Request.Context(), id, req.Name, req.Type, req.ParentID, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": acc})
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteAccount(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}

func (h *AccountHandler) GetAccountBalance(c *gin.Context) {
	id := c.Param("id")
	balance, err := h.svc.GetAccountBalance(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": balance})
}
