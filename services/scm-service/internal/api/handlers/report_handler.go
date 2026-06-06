package handlers

import (
	"net/http"

	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) GetInventoryLevelsReport(c *gin.Context) {
	rep, err := h.svc.GetInventoryLevelsReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}

func (h *ReportHandler) GetVendorPerformanceReport(c *gin.Context) {
	rep, err := h.svc.GetVendorPerformanceReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}

func (h *ReportHandler) GetProcurementMetricsReport(c *gin.Context) {
	rep, err := h.svc.GetProcurementMetricsReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}

func (h *ReportHandler) GetSafetyStockReport(c *gin.Context) {
	rep, err := h.svc.GetSafetyStockReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}
