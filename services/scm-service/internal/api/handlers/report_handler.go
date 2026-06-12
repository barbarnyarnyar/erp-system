package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	svc *service.ReportService
	response *utils.ResponseHelper
}

func NewReportHandler(svc *service.ReportService, response *utils.ResponseHelper) *ReportHandler {
	return &ReportHandler{
		svc: svc,
		response: response,
	}
}

func (h *ReportHandler) GetInventoryLevelsReport(c *gin.Context) {
	rep, err := h.svc.GetInventoryLevelsReport(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}

func (h *ReportHandler) GetVendorPerformanceReport(c *gin.Context) {
	rep, err := h.svc.GetVendorPerformanceReport(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}

func (h *ReportHandler) GetProcurementMetricsReport(c *gin.Context) {
	rep, err := h.svc.GetProcurementMetricsReport(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}

func (h *ReportHandler) GetSafetyStockReport(c *gin.Context) {
	rep, err := h.svc.GetSafetyStockReport(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}
