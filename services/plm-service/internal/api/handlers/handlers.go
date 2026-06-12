package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"strconv"

	"github.com/erp-system/plm-service/internal/business/domain"
	"github.com/erp-system/plm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type PlmHandler struct {
	matSvc  *service.MaterialService
	bomSvc  *service.BomService
	changeSvc *service.EngineeringChangeService
	resp    *utils.ResponseHelper
}

func NewPlmHandler(matSvc *service.MaterialService, bomSvc *service.BomService, changeSvc *service.EngineeringChangeService, resp *utils.ResponseHelper) *PlmHandler {
	return &PlmHandler{
		matSvc:    matSvc,
		bomSvc:    bomSvc,
		changeSvc: changeSvc,
		resp:      resp,
	}
}

// Material Master Handlers
func (h *PlmHandler) CreateMaterial(c *gin.Context) {
	var req struct {
		LegalEntityID   string                 `json:"legal_entity_id" binding:"required"`
		Sku             string                 `json:"sku" binding:"required"`
		Description     string                 `json:"description" binding:"required"`
		Uom             string                 `json:"uom" binding:"required"`
		ProcurementType domain.ProcurementType `json:"procurement_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	m, err := h.matSvc.CreateMaterial(c.Request.Context(), req.LegalEntityID, req.Sku, req.Description, req.Uom, req.ProcurementType)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": m})
}

func (h *PlmHandler) UpdateTechnicalSpecs(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Specs string `json:"specs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	m, err := h.matSvc.UpdateTechnicalSpecs(c.Request.Context(), id, req.Specs)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": m})
}

func (h *PlmHandler) TransitionStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status domain.MaterialStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	m, err := h.matSvc.TransitionStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": m})
}

// Bill of Materials Handlers
func (h *PlmHandler) EstablishBomHeader(c *gin.Context) {
	var req struct {
		LegalEntityID string                 `json:"legal_entity_id" binding:"required"`
		MaterialID    string                 `json:"material_id" binding:"required"`
		VersionString string                 `json:"version_string" binding:"required"`
		Lines         []service.BomLineInput `json:"lines" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	bh, err := h.bomSvc.EstablishBomHeader(c.Request.Context(), req.LegalEntityID, req.MaterialID, req.VersionString, req.Lines)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": bh})
}

func (h *PlmHandler) ReleaseBom(c *gin.Context) {
	id := c.Param("id")
	bh, err := h.bomSvc.ReleaseBom(c.Request.Context(), id)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": bh})
}

func (h *PlmHandler) ExplodeBillOfMaterials(c *gin.Context) {
	id := c.Param("id")
	depthStr := c.DefaultQuery("depth", "5")
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		h.resp.BadRequest(c, "invalid depth param")
		return
	}
	graph, err := h.bomSvc.ExplodeBillOfMaterials(c.Request.Context(), id, depth)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": graph})
}

// Engineering Change Request Handlers
func (h *PlmHandler) InitiateChangeRequest(c *gin.Context) {
	var req struct {
		LegalEntityID string `json:"legal_entity_id" binding:"required"`
		MaterialID    string `json:"material_id" binding:"required"`
		RequesterID   string `json:"requester_id" binding:"required"`
		Title         string `json:"title" binding:"required"`
		Description   string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	eco, err := h.changeSvc.InitiateChangeRequest(c.Request.Context(), req.LegalEntityID, req.MaterialID, req.RequesterID, req.Title, req.Description)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": eco})
}

func (h *PlmHandler) ProcessApprovalAction(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ApproverHrID string           `json:"approver_hr_id" binding:"required"`
		Action       domain.EcoStatus `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	eco, err := h.changeSvc.ProcessApprovalAction(c.Request.Context(), id, req.ApproverHrID, req.Action)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": eco})
}
