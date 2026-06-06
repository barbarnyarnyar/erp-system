package handlers

import (
	"net/http"

	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	svc *service.GeneralLedgerService
}

func NewReportHandler(svc *service.GeneralLedgerService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) GetBalanceSheet(c *gin.Context) {
	report, err := h.svc.GetBalanceSheet(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": report})
}

func (h *ReportHandler) GetIncomeStatement(c *gin.Context) {
	report, err := h.svc.GetIncomeStatement(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": report})
}

func (h *ReportHandler) GetCashFlow(c *gin.Context) {
	report, err := h.svc.GetCashFlow(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": report})
}
