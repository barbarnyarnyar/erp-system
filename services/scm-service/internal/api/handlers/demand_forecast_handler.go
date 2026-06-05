package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type DemandForecastHandler struct {
	svc *service.DemandPlanningService
}

func NewDemandForecastHandler(svc *service.DemandPlanningService) *DemandForecastHandler {
	return &DemandForecastHandler{svc: svc}
}

func (h *DemandForecastHandler) GetForecasts(c *gin.Context) {
	list, err := h.svc.ListForecasts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *DemandForecastHandler) CreateForecast(c *gin.Context) {
	var req struct {
		ProductID        string `json:"product_id"`
		ForecastDate     string `json:"forecast_date"`
		ForecastQuantity int    `json:"forecast_quantity"`
		ConfidenceLevel  string `json:"confidence_level"`
		Notes            string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fDate, err := time.Parse(time.RFC3339, req.ForecastDate)
	if err != nil {
		fDate = time.Now().AddDate(0, 1, 0) // default to 1 month out
	}

	confDec, err := decimal.NewFromString(req.ConfidenceLevel)
	if err != nil {
		confDec = decimal.NewFromFloat(0.8) // default to 80% confidence
	}

	df, err := h.svc.CreateForecast(c.Request.Context(), req.ProductID, fDate, req.ForecastQuantity, confDec, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": df})
}

func (h *DemandForecastHandler) GetForecast(c *gin.Context) {
	id := c.Param("id")
	df, err := h.svc.GetForecast(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "forecast not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": df})
}

func (h *DemandForecastHandler) UpdateForecast(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ForecastDate     string `json:"forecast_date"`
		ForecastQuantity int    `json:"forecast_quantity"`
		ConfidenceLevel  string `json:"confidence_level"`
		Notes            string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fDate, err := time.Parse(time.RFC3339, req.ForecastDate)
	if err != nil {
		fDate = time.Now().AddDate(0, 1, 0)
	}

	confDec, err := decimal.NewFromString(req.ConfidenceLevel)
	if err != nil {
		confDec = decimal.NewFromFloat(0.8)
	}

	df, err := h.svc.UpdateForecast(c.Request.Context(), id, fDate, req.ForecastQuantity, confDec, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": df})
}
