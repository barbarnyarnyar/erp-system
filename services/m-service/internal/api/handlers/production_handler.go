package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/m-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type ProductionHandler struct {
	svc *service.ProductionService
}

func NewProductionHandler(svc *service.ProductionService) *ProductionHandler {
	return &ProductionHandler{svc: svc}
}

func (h *ProductionHandler) CreateProductionPlan(c *gin.Context) {
	var req struct {
		BomID         string    `json:"bom_id"`
		Quantity      int       `json:"quantity"`
		ScheduledDate time.Time `json:"scheduled_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	po, err := h.svc.CreateProductionOrder(c.Request.Context(), req.BomID, req.Quantity, req.ScheduledDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": po})
}

func (h *ProductionHandler) ListProductionPlans(c *gin.Context) {
	list, err := h.svc.ListProductionPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *ProductionHandler) GetProductionPlanDetails(c *gin.Context) {
	id := c.Param("id")
	po, err := h.svc.GetProductionPlan(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "production plan not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": po})
}

func (h *ProductionHandler) UpdateProductionPlan(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Quantity      int       `json:"quantity"`
		ScheduledDate time.Time `json:"scheduled_date"`
		Status        string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	po, err := h.svc.UpdateProductionPlan(c.Request.Context(), id, req.Quantity, req.ScheduledDate, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": po})
}

func (h *ProductionHandler) RunMRP(c *gin.Context) {
	err := h.svc.RunMRP(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MRP run completed successfully"})
}

func (h *ProductionHandler) ListWorkOrders(c *gin.Context) {
	list, err := h.svc.ListWorkOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *ProductionHandler) CreateWorkOrder(c *gin.Context) {
	var req struct {
		ProductionOrderID string    `json:"production_order_id"`
		SequenceNumber    int       `json:"sequence_number"`
		WorkCenterID      string    `json:"work_center_id"`
		ScheduledStart    time.Time `json:"scheduled_start"`
		ScheduledEnd      time.Time `json:"scheduled_end"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wo, err := h.svc.CreateWorkOrder(c.Request.Context(), req.ProductionOrderID, req.SequenceNumber, req.WorkCenterID, req.ScheduledStart, req.ScheduledEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": wo})
}

func (h *ProductionHandler) GetWorkOrderDetails(c *gin.Context) {
	id := c.Param("id")
	wo, err := h.svc.GetWorkOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work order not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wo})
}

func (h *ProductionHandler) UpdateWorkOrder(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status         string     `json:"status"`
		ScheduledStart time.Time  `json:"scheduled_start"`
		ScheduledEnd   time.Time  `json:"scheduled_end"`
		ActualStart    *time.Time `json:"actual_start"`
		ActualEnd      *time.Time `json:"actual_end"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wo, err := h.svc.UpdateWorkOrder(c.Request.Context(), id, req.Status, req.ScheduledStart, req.ScheduledEnd, req.ActualStart, req.ActualEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wo})
}

func (h *ProductionHandler) DeleteWorkOrder(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteWorkOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "work order deleted successfully"})
}

func (h *ProductionHandler) StartWorkOrder(c *gin.Context) {
	id := c.Param("id")
	wo, err := h.svc.StartWorkOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wo})
}

func (h *ProductionHandler) ReportLabor(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		EmployeeID  string `json:"employee_id"`
		HoursWorked string `json:"hours_worked"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hoursDec, _ := decimal.NewFromString(req.HoursWorked)

	lr, err := h.svc.ReportLabor(c.Request.Context(), id, req.EmployeeID, hoursDec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": lr})
}

func (h *ProductionHandler) LogMachineStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		StatusCode string `json:"status_code"`
		Message    string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logEntry, err := h.svc.LogMachineStatus(c.Request.Context(), id, req.StatusCode, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": logEntry})
}

func (h *ProductionHandler) CompleteWorkOrder(c *gin.Context) {
	id := c.Param("id")
	wo, err := h.svc.CompleteWorkOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wo})
}

func (h *ProductionHandler) RecordQualityInspection(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		InspectorID string `json:"inspector_id"`
		Result      string `json:"result"`
		Remarks     string `json:"remarks"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qi, err := h.svc.RecordQualityInspection(c.Request.Context(), id, req.InspectorID, req.Result, req.Remarks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": qi})
}

func (h *ProductionHandler) ListQualityInspections(c *gin.Context) {
	list, err := h.svc.ListQualityInspections(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *ProductionHandler) GetQualityInspectionDetails(c *gin.Context) {
	id := c.Param("id")
	qi, err := h.svc.GetQualityInspection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quality inspection not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": qi})
}

func (h *ProductionHandler) UpdateQualityInspection(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Result  string `json:"result"`
		Remarks string `json:"remarks"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qi, err := h.svc.UpdateQualityInspection(c.Request.Context(), id, req.Result, req.Remarks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": qi})
}

func (h *ProductionHandler) GetCosting(c *gin.Context) {
	poID := c.Param("id")
	cost, err := h.svc.GetCostingRecord(c.Request.Context(), poID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Costing record not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cost})
}

func (h *ProductionHandler) CreateEquipment(c *gin.Context) {
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

func (h *ProductionHandler) ListMaintenanceSchedules(c *gin.Context) {
	list, err := h.svc.ListMaintenanceSchedules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *ProductionHandler) ScheduleMaintenance(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Description     string `json:"description"`
		MaintenanceType string `json:"maintenance_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mo, err := h.svc.ScheduleMaintenance(c.Request.Context(), id, req.Description, req.MaintenanceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": mo})
}

func (h *ProductionHandler) GetMaintenanceScheduleDetails(c *gin.Context) {
	id := c.Param("id")
	mo, err := h.svc.GetMaintenanceSchedule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "maintenance schedule not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mo})
}

func (h *ProductionHandler) UpdateMaintenanceSchedule(c *gin.Context) {
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
