package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type TrainingHandler struct {
	svc *service.TrainingService
	response *utils.ResponseHelper
}

func NewTrainingHandler(svc *service.TrainingService, response *utils.ResponseHelper) *TrainingHandler {
	return &TrainingHandler{
		svc: svc,
		response: response,
	}
}

func (h *TrainingHandler) GetTrainingPrograms(c *gin.Context) {
	list, err := h.svc.ListTrainingPrograms(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *TrainingHandler) CreateTrainingProgram(c *gin.Context) {
	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Trainer     string    `json:"trainer"`
		StartDate   time.Time `json:"start_date"`
		EndDate     time.Time `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	tp, err := h.svc.CreateTrainingProgram(c.Request.Context(), req.Title, req.Description, req.Trainer, req.StartDate, req.EndDate)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": tp})
}

func (h *TrainingHandler) GetTrainingProgram(c *gin.Context) {
	id := c.Param("id")
	tp, err := h.svc.GetTrainingProgram(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "training program not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tp})
}

func (h *TrainingHandler) UpdateTrainingProgram(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Trainer     string    `json:"trainer"`
		StartDate   time.Time `json:"start_date"`
		EndDate     time.Time `json:"end_date"`
		Status      string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	tp, err := h.svc.UpdateTrainingProgram(c.Request.Context(), id, req.Title, req.Description, req.Trainer, req.StartDate, req.EndDate, req.Status)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tp})
}

func (h *TrainingHandler) EnrollEmployee(c *gin.Context) {
	trainingID := c.Param("id")
	var req struct {
		EmployeeID string `json:"employee_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	enrollment, err := h.svc.EnrollEmployee(c.Request.Context(), trainingID, req.EmployeeID)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": enrollment})
}

func (h *TrainingHandler) CompleteTraining(c *gin.Context) {
	enrollmentID := c.Param("enrollmentId")

	enrollment, err := h.svc.CompleteTraining(c.Request.Context(), enrollmentID)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": enrollment})
}
