package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/erp-system/fm-service/internal/business/service"
)

type ReportHandler struct {
	svc *service.FinanceService
}

func NewReportHandler(svc *service.FinanceService) *ReportHandler {
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
	c.JSON(http.StatusOK, gin.H{
		"message": "Income statement report generated successfully",
		"report":  "income_statement",
	})
}

func (h *ReportHandler) GetCashFlow(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Cash flow report generated successfully",
		"report":  "cash_flow",
	})
}