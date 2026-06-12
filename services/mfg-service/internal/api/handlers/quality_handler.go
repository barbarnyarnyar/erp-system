package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/m-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type QualityHandler struct {
	svc *service.QualityService
	response *utils.ResponseHelper
}

func NewQualityHandler(svc *service.QualityService, response *utils.ResponseHelper) *QualityHandler {
	return &QualityHandler{
		svc: svc,
		response: response,
	}
}

func (h *QualityHandler) RecordQualityInspection(c *gin.Context) {
	workOrderID := c.Param("id")
	var req struct {
		InspectorID string `json:"inspector_id"`
		Result      string `json:"result"`
		Remarks     string `json:"remarks"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	qi, err := h.svc.RecordQualityInspection(c.Request.Context(), workOrderID, req.InspectorID, req.Result, req.Remarks)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": qi})
}

func (h *QualityHandler) ListQualityInspections(c *gin.Context) {
	list, err := h.svc.ListQualityInspections(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *QualityHandler) GetQualityInspectionDetails(c *gin.Context) {
	id := c.Param("id")
	qi, err := h.svc.GetQualityInspection(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "quality inspection not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": qi})
}

func (h *QualityHandler) UpdateQualityInspection(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Result  string `json:"result"`
		Remarks string `json:"remarks"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	qi, err := h.svc.UpdateQualityInspection(c.Request.Context(), id, req.Result, req.Remarks)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": qi})
}

func (h *QualityHandler) GetNonConformances(c *gin.Context) {
	inspectionID := c.Param("id")
	list, err := h.svc.ListNonConformancesByInspectionID(c.Request.Context(), inspectionID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
