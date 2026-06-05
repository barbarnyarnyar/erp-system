package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/erp-system/hr-service/internal/business/service"
)

type LeaveHandler struct {
	svc *service.LeaveManagementService
}

func NewLeaveHandler(svc *service.LeaveManagementService) *LeaveHandler {
	return &LeaveHandler{svc: svc}
}

func (h *LeaveHandler) GetLeaveRequests(c *gin.Context) {
	list, err := h.svc.ListLeaveRequests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *LeaveHandler) CreateLeaveRequest(c *gin.Context) {
	var req struct {
		EmployeeID string    `json:"employee_id"`
		LeaveType  string    `json:"leave_type"`
		StartDate  time.Time `json:"start_date"`
		EndDate    time.Time `json:"end_date"`
		Reason     string    `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lr, err := h.svc.CreateLeaveRequest(c.Request.Context(), req.EmployeeID, req.LeaveType, req.StartDate, req.EndDate, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": lr})
}

func (h *LeaveHandler) GetLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	lr, err := h.svc.GetLeaveRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave request not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lr})
}

func (h *LeaveHandler) UpdateLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		LeaveType string    `json:"leave_type"`
		StartDate time.Time `json:"start_date"`
		EndDate   time.Time `json:"end_date"`
		Reason    string    `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lr, err := h.svc.UpdateLeaveRequest(c.Request.Context(), id, req.LeaveType, req.StartDate, req.EndDate, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": lr})
}

func (h *LeaveHandler) ApproveLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ApprovedBy string `json:"approved_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lr, err := h.svc.ApproveLeaveRequest(c.Request.Context(), id, req.ApprovedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": lr})
}

func (h *LeaveHandler) RejectLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		RejectedBy string `json:"rejected_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lr, err := h.svc.RejectLeaveRequest(c.Request.Context(), id, req.RejectedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": lr})
}

