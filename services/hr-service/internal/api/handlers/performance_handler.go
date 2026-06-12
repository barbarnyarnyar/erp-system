package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type PerformanceHandler struct {
	svc *service.PerformanceService
	response *utils.ResponseHelper
}

func NewPerformanceHandler(svc *service.PerformanceService, response *utils.ResponseHelper) *PerformanceHandler {
	return &PerformanceHandler{
		svc: svc,
		response: response,
	}
}

func (h *PerformanceHandler) GetPerformanceReviews(c *gin.Context) {
	list, err := h.svc.ListPerformanceReviews(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *PerformanceHandler) CreatePerformanceReview(c *gin.Context) {
	var req struct {
		EmployeeID  string    `json:"employee_id"`
		ReviewerID  string    `json:"reviewer_id"`
		ReviewDate  time.Time `json:"review_date"`
		PeriodStart time.Time `json:"period_start"`
		PeriodEnd   time.Time `json:"period_end"`
		Rating      int       `json:"rating"`
		Feedback    string    `json:"feedback"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	pr, err := h.svc.CreatePerformanceReview(c.Request.Context(), req.EmployeeID, req.ReviewerID, req.ReviewDate, req.PeriodStart, req.PeriodEnd, req.Rating, req.Feedback)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": pr})
}

func (h *PerformanceHandler) GetPerformanceReview(c *gin.Context) {
	id := c.Param("id")
	pr, err := h.svc.GetPerformanceReview(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "performance review not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *PerformanceHandler) UpdatePerformanceReview(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Rating   int    `json:"rating"`
		Feedback string `json:"feedback"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	pr, err := h.svc.UpdatePerformanceReview(c.Request.Context(), id, req.Rating, req.Feedback)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pr})
}
