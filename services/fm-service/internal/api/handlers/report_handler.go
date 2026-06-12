package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	svc      *service.GeneralLedgerService
	response *utils.ResponseHelper
}

func NewReportHandler(svc *service.GeneralLedgerService, response *utils.ResponseHelper) *ReportHandler {
	return &ReportHandler{
		svc:      svc,
		response: response,
	}
}

func (h *ReportHandler) GetBalanceSheet(c *gin.Context) {
	report, err := h.svc.GetBalanceSheet(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": report})
}

func (h *ReportHandler) GetIncomeStatement(c *gin.Context) {
	report, err := h.svc.GetIncomeStatement(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": report})
}

func (h *ReportHandler) GetCashFlow(c *gin.Context) {
	report, err := h.svc.GetCashFlow(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": report})
}
