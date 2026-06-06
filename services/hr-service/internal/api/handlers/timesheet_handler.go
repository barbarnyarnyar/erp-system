package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type TimesheetHandler struct {
	svc *service.TimeAttendanceService
}

func NewTimesheetHandler(svc *service.TimeAttendanceService) *TimesheetHandler {
	return &TimesheetHandler{svc: svc}
}

func (h *TimesheetHandler) GetTimesheets(c *gin.Context) {
	list, err := h.svc.ListTimesheets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *TimesheetHandler) CreateTimesheet(c *gin.Context) {
	var req struct {
		EmployeeID string    `json:"employee_id"`
		EntryDate  time.Time `json:"entry_date"`
		ClockIn    time.Time `json:"clock_in"`
		ClockOut   time.Time `json:"clock_out"`
		Notes      string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	te, err := h.svc.CreateTimesheet(c.Request.Context(), req.EmployeeID, req.EntryDate, req.ClockIn, req.ClockOut, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": te})
}

func (h *TimesheetHandler) GetTimesheet(c *gin.Context) {
	id := c.Param("id")
	te, err := h.svc.GetTimesheet(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "timesheet not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": te})
}

func (h *TimesheetHandler) UpdateTimesheet(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ClockIn  time.Time `json:"clock_in"`
		ClockOut time.Time `json:"clock_out"`
		Notes    string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	te, err := h.svc.UpdateTimesheet(c.Request.Context(), id, req.ClockIn, req.ClockOut, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": te})
}

func (h *TimesheetHandler) SubmitTimesheet(c *gin.Context) {
	id := c.Param("id")
	te, err := h.svc.SubmitTimesheet(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": te})
}

func (h *TimesheetHandler) ApproveTimesheet(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ApprovedBy string `json:"approved_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	te, err := h.svc.ApproveTimesheet(c.Request.Context(), id, req.ApprovedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": te})
}
