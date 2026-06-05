package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type EmployeeHandler struct {
	svc *service.EmployeeManagementService
}

func NewEmployeeHandler(svc *service.EmployeeManagementService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

func (h *EmployeeHandler) GetEmployees(c *gin.Context) {
	list, err := h.svc.ListEmployees(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		DepartmentID string `json:"department_id"`
		PositionID   string `json:"position_id"`
		Salary       string `json:"salary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	salaryDec, err := decimal.NewFromString(req.Salary)
	if err != nil {
		salaryDec = decimal.Zero
	}

	emp, err := h.svc.CreateEmployee(c.Request.Context(), req.FirstName, req.LastName, req.Email, req.DepartmentID, req.PositionID, salaryDec)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": emp})
}

func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	emp, err := h.svc.GetEmployee(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": emp})
}

func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		DepartmentID string `json:"department_id"`
		PositionID   string `json:"position_id"`
		Salary       string `json:"salary"`
		Status       string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	salaryDec, err := decimal.NewFromString(req.Salary)
	if err != nil {
		salaryDec = decimal.Zero
	}

	emp, err := h.svc.UpdateEmployee(c.Request.Context(), id, req.FirstName, req.LastName, req.Email, req.DepartmentID, req.PositionID, salaryDec, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": emp})
}

func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteEmployee(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "employee deleted successfully"})
}

func (h *EmployeeHandler) SubmitExpenseClaim(c *gin.Context) {
	employeeID := c.Param("id")
	var req struct {
		ClaimDate time.Time `json:"claim_date"`
		Lines     []struct {
			Description string `json:"description"`
			Amount      string `json:"amount"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var claimLines []domain.ExpenseClaimLine
	for _, l := range req.Lines {
		amt, err := decimal.NewFromString(l.Amount)
		if err != nil {
			amt = decimal.Zero
		}
		claimLines = append(claimLines, domain.ExpenseClaimLine{
			Description: l.Description,
			Amount:      amt,
		})
	}

	claim, err := h.svc.SubmitExpenseClaim(c.Request.Context(), employeeID, req.ClaimDate, claimLines)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": claim})
}
