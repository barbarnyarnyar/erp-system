package handlers

import (
	"net/http"

	"github.com/erp-system/m-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type QualityHandler struct {
	svc *service.QualityService
}

func NewQualityHandler(svc *service.QualityService) *QualityHandler {
	return &QualityHandler{svc: svc}
}

func (h *QualityHandler) RecordQualityInspection(c *gin.Context) {
	workOrderID := c.Param("id")
	var req struct {
		InspectorID string `json:"inspector_id"`
		Result      string `json:"result"`
		Remarks     string `json:"remarks"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qi, err := h.svc.RecordQualityInspection(c.Request.Context(), workOrderID, req.InspectorID, req.Result, req.Remarks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": qi})
}

func (h *QualityHandler) ListQualityInspections(c *gin.Context) {
	list, err := h.svc.ListQualityInspections(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *QualityHandler) GetQualityInspectionDetails(c *gin.Context) {
	id := c.Param("id")
	qi, err := h.svc.GetQualityInspection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quality inspection not found"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qi, err := h.svc.UpdateQualityInspection(c.Request.Context(), id, req.Result, req.Remarks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": qi})
}
