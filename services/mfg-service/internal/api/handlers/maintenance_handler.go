package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/m-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type MaintenanceHandler struct {
	svc *service.MaintenanceService
}

func NewMaintenanceHandler(svc *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc}
}

func (h *MaintenanceHandler) LogMachineStatus(c *gin.Context) {
	id := c.Param("id") // work_center_id
	var req struct {
		StatusCode string `json:"status_code"`
		Message    string `json:"message"`
		Severity   string `json:"severity"` // INFO, WARNING, ERROR, CRITICAL
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	severity := req.Severity
	if severity == "" {
		severity = "INFO"
	}

	logEntry, err := h.svc.LogMachineStatus(c.Request.Context(), id, req.StatusCode, req.Message, severity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": logEntry})
}

func (h *MaintenanceHandler) CreateEquipment(c *gin.Context) {
	var req struct {
		WorkCenterID string `json:"work_center_id"`
		Name         string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eq, err := h.svc.CreateEquipment(c.Request.Context(), req.WorkCenterID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": eq})
}

func (h *MaintenanceHandler) ScheduleMaintenance(c *gin.Context) {
	equipmentID := c.Param("id")
	var req struct {
		Description     string `json:"description"`
		MaintenanceType string `json:"maintenance_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mo, err := h.svc.ScheduleMaintenance(c.Request.Context(), equipmentID, req.Description, req.MaintenanceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": mo})
}

func (h *MaintenanceHandler) CompleteMaintenance(c *gin.Context) {
	id := c.Param("id")
	mo, err := h.svc.CompleteMaintenance(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mo})
}

func (h *MaintenanceHandler) ListMaintenanceSchedules(c *gin.Context) {
	list, err := h.svc.ListMaintenanceSchedules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *MaintenanceHandler) GetMaintenanceScheduleDetails(c *gin.Context) {
	id := c.Param("id")
	mo, err := h.svc.GetMaintenanceSchedule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "maintenance schedule not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mo})
}

func (h *MaintenanceHandler) UpdateMaintenanceSchedule(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status      string     `json:"status"`
		CompletedAt *time.Time `json:"completed_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mo, err := h.svc.UpdateMaintenanceSchedule(c.Request.Context(), id, req.Status, req.CompletedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mo})
}
