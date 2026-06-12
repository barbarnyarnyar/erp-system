package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PayrollHandler struct {
	svc *service.PayrollService
	response *utils.ResponseHelper
}

func NewPayrollHandler(svc *service.PayrollService, response *utils.ResponseHelper) *PayrollHandler {
	return &PayrollHandler{
		svc: svc,
		response: response,
	}
}

func (h *PayrollHandler) GetPayrollRecords(c *gin.Context) {
	list, err := h.svc.ListPayrollRecords(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *PayrollHandler) ProcessPayroll(c *gin.Context) {
	var req struct {
		EmployeeID     string    `json:"employee_id"`
		PayPeriodStart time.Time `json:"pay_period_start"`
		PayPeriodEnd   time.Time `json:"pay_period_end"`
		RegularHours   string    `json:"regular_hours"`
		OvertimeHours  string    `json:"overtime_hours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	regHours, err := decimal.NewFromString(req.RegularHours)
	if err != nil {
		regHours = decimal.Zero
	}
	otHours, err := decimal.NewFromString(req.OvertimeHours)
	if err != nil {
		otHours = decimal.Zero
	}

	pr, err := h.svc.ProcessPayroll(c.Request.Context(), req.EmployeeID, req.PayPeriodStart, req.PayPeriodEnd, regHours, otHours)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": pr})
}

func (h *PayrollHandler) GetPayrollRecord(c *gin.Context) {
	id := c.Param("id")
	pr, err := h.svc.GetPayrollRecord(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "payroll record not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *PayrollHandler) UpdatePayrollRecord(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	pr, err := h.svc.UpdatePayrollRecord(c.Request.Context(), id, req.Status)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *PayrollHandler) GetEmployeePayroll(c *gin.Context) {
	empID := c.Param("id")
	list, err := h.svc.GetEmployeePayroll(c.Request.Context(), empID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
