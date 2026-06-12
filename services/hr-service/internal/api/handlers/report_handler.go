package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/hr-service/internal/business/service"
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

func (h *ReportHandler) GetHeadcountReport(c *gin.Context) {
	rep, err := h.svc.GetHeadcountReport(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}

func (h *ReportHandler) GetPayrollReport(c *gin.Context) {
	rep, err := h.svc.GetPayrollReport(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}

func (h *ReportHandler) GetAttendanceReport(c *gin.Context) {
	rep, err := h.svc.GetAttendanceReport(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rep})
}
