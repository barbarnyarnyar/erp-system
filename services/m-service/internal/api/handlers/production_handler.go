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
		SalesOrderID  string    `json:"sales_order_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	po, err := h.svc.CreateProductionOrder(c.Request.Context(), req.BomID, req.Quantity, req.ScheduledDate, req.SalesOrderID)
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

func (h *ProductionHandler) CompleteProductionOrder(c *gin.Context) {
	id := c.Param("id")
	po, err := h.svc.CompleteProductionOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": po})
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
	id := c.Param("id") // work_order_id
	var req struct {
		EmployeeID  string `json:"employee_id"`
		HoursWorked string `json:"hours_worked"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hoursDec, err := decimal.NewFromString(req.HoursWorked)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hours_worked decimal"})
		return
	}

	lr, err := h.svc.ReportLabor(c.Request.Context(), id, req.EmployeeID, hoursDec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": lr})
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
