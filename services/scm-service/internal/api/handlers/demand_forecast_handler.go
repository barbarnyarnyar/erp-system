package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type DemandForecastHandler struct {
	svc      *service.DemandPlanningService
	response *utils.ResponseHelper
}

func NewDemandForecastHandler(svc *service.DemandPlanningService, response *utils.ResponseHelper) *DemandForecastHandler {
	return &DemandForecastHandler{
		svc:      svc,
		response: response,
	}
}

func (h *DemandForecastHandler) GetForecasts(c *gin.Context) {
	list, err := h.svc.ListForecasts(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
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
		h.response.BadRequest(c, err.Error())
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

	df, err := h.svc.CreateForecast(c.Request.Context(), req.ProductID, fDate, decimal.NewFromInt(int64(req.ForecastQuantity)), confDec, req.Notes)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": df})
}

func (h *DemandForecastHandler) GetForecast(c *gin.Context) {
	id := c.Param("id")
	df, err := h.svc.GetForecast(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "forecast not found")
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
		h.response.BadRequest(c, err.Error())
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

	df, err := h.svc.UpdateForecast(c.Request.Context(), id, fDate, decimal.NewFromInt(int64(req.ForecastQuantity)), confDec, req.Notes)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": df})
}
