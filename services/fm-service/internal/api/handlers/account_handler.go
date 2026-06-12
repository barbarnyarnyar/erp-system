package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	svc      *service.GeneralLedgerService
	response *utils.ResponseHelper
}

func NewAccountHandler(svc *service.GeneralLedgerService, response *utils.ResponseHelper) *AccountHandler {
	return &AccountHandler{
		svc:      svc,
		response: response,
	}
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	accs, err := h.svc.ListAccounts(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": accs})
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req struct {
		LegalEntityID string `json:"legal_entity_id"`
		AccountCode   string `json:"account_code"`
		AccountName   string `json:"account_name"`
		Type          string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	acc, err := h.svc.CreateAccount(c.Request.Context(), req.LegalEntityID, req.AccountCode, req.AccountName, req.Type)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": acc})
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	acc, err := h.svc.GetAccount(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "account not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": acc})
}

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		AccountName string `json:"account_name"`
		Type        string `json:"type"`
		IsActive    bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	acc, err := h.svc.UpdateAccount(c.Request.Context(), id, req.AccountName, req.Type, req.IsActive)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": acc})
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteAccount(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}

func (h *AccountHandler) GetAccountBalance(c *gin.Context) {
	id := c.Param("id")
	balance, err := h.svc.GetAccountBalance(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "account not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": balance})
}
