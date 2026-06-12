package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type AssetHandler struct {
	svc      *service.CapitalAssetService
	response *utils.ResponseHelper
}

func NewAssetHandler(svc *service.CapitalAssetService, response *utils.ResponseHelper) *AssetHandler {
	return &AssetHandler{
		svc:      svc,
		response: response,
	}
}

func (h *AssetHandler) CapitalizeAsset(c *gin.Context) {
	var req struct {
		LegalEntityID    string  `json:"legal_entity_id"`
		AssetTag         string  `json:"asset_tag"`
		AcquisitionCost  string  `json:"acquisition_cost"`
		UsefulLifeMonths int     `json:"useful_life_months"`
		EamEquipmentID   *string `json:"eam_equipment_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	costDec, err := decimal.NewFromString(req.AcquisitionCost)
	if err != nil {
		h.response.BadRequest(c, "invalid acquisition cost format")
		return
	}

	asset, err := h.svc.CapitalizeAsset(c.Request.Context(), req.LegalEntityID, req.AssetTag, costDec, req.UsefulLifeMonths, req.EamEquipmentID)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": asset})
}

func (h *AssetHandler) GenerateDepreciationSchedule(c *gin.Context) {
	id := c.Param("id")
	schedule, err := h.svc.GenerateDepreciationSchedule(c.Request.Context(), id)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": schedule})
}

func (h *AssetHandler) PostMonthlyDepreciation(c *gin.Context) {
	var req struct {
		LegalEntityID string `json:"legal_entity_id"`
		FiscalYear    int    `json:"fiscal_year"`
		PeriodNumber  int    `json:"period_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	err := h.svc.PostMonthlyStraightLineDepreciation(c.Request.Context(), req.LegalEntityID, req.FiscalYear, req.PeriodNumber)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "depreciation posted successfully"})
}

func (h *AssetHandler) GetAsset(c *gin.Context) {
	id := c.Param("id")
	asset, err := h.svc.GetAsset(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "asset not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": asset})
}

func (h *AssetHandler) GetAssets(c *gin.Context) {
	list, err := h.svc.ListAssets(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
