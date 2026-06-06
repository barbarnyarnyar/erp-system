package handlers

import (
	"net/http"

	"github.com/erp-system/m-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type CostingHandler struct {
	svc *service.CostingService
}

func NewCostingHandler(svc *service.CostingService) *CostingHandler {
	return &CostingHandler{svc: svc}
}

func (h *CostingHandler) GetCosting(c *gin.Context) {
	poID := c.Param("id")
	cost, err := h.svc.GetCostingRecord(c.Request.Context(), poID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "costing record not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cost})
}

func (h *CostingHandler) RunMRP(c *gin.Context) {
	err := h.svc.RunMRP(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MRP run completed successfully"})
}
