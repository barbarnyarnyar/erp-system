package handlers

import (
	"context"
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
	// Run MRP asynchronously in a background goroutine to avoid HTTP thread saturation
	go func() {
		_ = h.svc.RunMRP(context.Background())
	}()
	c.JSON(http.StatusAccepted, gin.H{"message": "MRP run initiated in the background"})
}
