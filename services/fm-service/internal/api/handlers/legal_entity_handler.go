package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type LegalEntityHandler struct {
	svc      *service.LegalEntityService
	response *utils.ResponseHelper
}

func NewLegalEntityHandler(svc *service.LegalEntityService, response *utils.ResponseHelper) *LegalEntityHandler {
	return &LegalEntityHandler{
		svc:      svc,
		response: response,
	}
}

func (h *LegalEntityHandler) CreateLegalEntity(c *gin.Context) {
	var req struct {
		CompanyCode           string `json:"company_code"`
		CompanyName           string `json:"company_name"`
		FunctionalCurrency    string `json:"functional_currency"`
		TaxRegistrationNumber string `json:"tax_registration_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	le, err := h.svc.CreateLegalEntity(c.Request.Context(), req.CompanyCode, req.CompanyName, req.FunctionalCurrency, req.TaxRegistrationNumber)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": le})
}

func (h *LegalEntityHandler) GetLegalEntity(c *gin.Context) {
	id := c.Param("id")
	le, err := h.svc.GetLegalEntity(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "legal entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": le})
}

func (h *LegalEntityHandler) GetLegalEntities(c *gin.Context) {
	list, err := h.svc.ListLegalEntities(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
