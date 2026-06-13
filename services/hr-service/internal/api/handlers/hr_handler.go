package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type HrHandler struct {
	empService   service.EmployeeService
	payrollSvc   service.PayrollService
	expenseSvc   service.ExpenseService
	deptRepo     domain.DepartmentRepository
	empRepo      domain.EmployeeMasterRepository
	payrollRepo  domain.PayrollRunRepository
	expenseRepo  domain.ExpenseClaimRepository
	lineRepo     domain.ExpenseClaimLineRepository
}

func NewHrHandler(
	empService service.EmployeeService,
	payrollSvc service.PayrollService,
	expenseSvc service.ExpenseService,
	deptRepo domain.DepartmentRepository,
	empRepo domain.EmployeeMasterRepository,
	payrollRepo domain.PayrollRunRepository,
	expenseRepo domain.ExpenseClaimRepository,
	lineRepo domain.ExpenseClaimLineRepository,
) *HrHandler {
	return &HrHandler{
		empService:   empService,
		payrollSvc:   payrollSvc,
		expenseSvc:   expenseSvc,
		deptRepo:     deptRepo,
		empRepo:      empRepo,
		payrollRepo:  payrollRepo,
		expenseRepo:  expenseRepo,
		lineRepo:     lineRepo,
	}
}

// ==========================================
// Department Handlers
// ==========================================

func (h *HrHandler) CreateDepartment(c *gin.Context) {
	var req struct {
		ID             string `json:"id"`
		LegalEntityID  string `json:"legal_entity_id" binding:"required"`
		DepartmentCode string `json:"department_code" binding:"required"`
		Name           string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	dept := &domain.Department{
		ID:             req.ID,
		LegalEntityID:  req.LegalEntityID,
		DepartmentCode: req.DepartmentCode,
		Name:           req.Name,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.deptRepo.Create(c.Request.Context(), dept); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dept)
}

func (h *HrHandler) GetDepartments(c *gin.Context) {
	depts, err := h.deptRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, depts)
}

func (h *HrHandler) GetDepartment(c *gin.Context) {
	id := c.Param("id")
	dept, err := h.deptRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dept)
}

func (h *HrHandler) UpdateDepartment(c *gin.Context) {
	id := c.Param("id")
	dept, err := h.deptRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept.Name = req.Name
	dept.IsActive = req.IsActive
	dept.UpdatedAt = time.Now()

	if err := h.deptRepo.Update(c.Request.Context(), dept); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dept)
}

// ==========================================
// Employee Handlers
// ==========================================

func (h *HrHandler) HireEmployee(c *gin.Context) {
	var req struct {
		LegalEntityID  string          `json:"legal_entity_id" binding:"required"`
		DepartmentID   string          `json:"department_id" binding:"required"`
		ManagerHrID    *string         `json:"manager_hr_id"`
		EmployeeNumber string          `json:"employee_number" binding:"required"`
		FirstName      string          `json:"first_name" binding:"required"`
		LastName       string          `json:"last_name" binding:"required"`
		Email          string          `json:"email" binding:"required"`
		BaseSalary     decimal.Decimal `json:"base_salary" binding:"required"`
		Type           string          `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	empType := domain.EmploymentType(req.Type)
	emp, err := h.empService.HireEmployee(
		c.Request.Context(),
		req.LegalEntityID,
		req.DepartmentID,
		req.ManagerHrID,
		req.EmployeeNumber,
		req.FirstName,
		req.LastName,
		req.Email,
		req.BaseSalary,
		empType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, emp)
}

func (h *HrHandler) GetEmployees(c *gin.Context) {
	emps, err := h.empRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emps)
}

func (h *HrHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	emp, err := h.empRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emp)
}

func (h *HrHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	emp, err := h.empRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		Status    string `json:"status" binding:"required"`
		Type      string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emp.FirstName = req.FirstName
	emp.LastName = req.LastName
	emp.Status = domain.EmployeeStatus(req.Status)
	emp.Type = domain.EmploymentType(req.Type)
	emp.UpdatedAt = time.Now()

	if err := h.empRepo.Update(c.Request.Context(), emp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emp)
}

func (h *HrHandler) TerminateEmployee(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		TerminationDate time.Time `json:"termination_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.TerminationDate = time.Now()
	}

	emp, err := h.empService.TerminateEmployee(c.Request.Context(), id, req.TerminationDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emp)
}

func (h *HrHandler) UpdateCompensation(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		TargetSalary decimal.Decimal `json:"target_salary" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.empService.AdjustCompensation(c.Request.Context(), id, req.TargetSalary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emp)
}

func (h *HrHandler) GetManagementChain(c *gin.Context) {
	id := c.Param("id")
	chain, err := h.empService.FetchManagementChain(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chain)
}

// ==========================================
// Payroll Handlers
// ==========================================

func (h *HrHandler) InitiatePayrollRun(c *gin.Context) {
	var req struct {
		LegalEntityID string `json:"legal_entity_id" binding:"required"`
		FiscalYear    int    `json:"fiscal_year" binding:"required"`
		PeriodNumber  int    `json:"period_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	run, err := h.payrollSvc.InitiatePeriodRun(c.Request.Context(), req.LegalEntityID, req.FiscalYear, req.PeriodNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, run)
}

func (h *HrHandler) ExecutePayrollCalculations(c *gin.Context) {
	id := c.Param("id")
	run, err := h.payrollSvc.ExecuteCalculations(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *HrHandler) CloseAndApprovePayroll(c *gin.Context) {
	id := c.Param("id")
	run, err := h.payrollSvc.CloseAndApprovePayroll(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *HrHandler) GetPayrollRuns(c *gin.Context) {
	runs, err := h.payrollRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, runs)
}

func (h *HrHandler) GetPayrollRun(c *gin.Context) {
	id := c.Param("id")
	run, err := h.payrollRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, run)
}

// ==========================================
// Expense Handlers
// ==========================================

func (h *HrHandler) SubmitExpenseClaim(c *gin.Context) {
	var req struct {
		LegalEntityID string                         `json:"legal_entity_id" binding:"required"`
		EmployeeID    string                         `json:"employee_id" binding:"required"`
		ClaimNumber   string                         `json:"claim_number" binding:"required"`
		Purpose       string                         `json:"purpose" binding:"required"`
		CostCenterTag string                         `json:"cost_center_tag" binding:"required"`
		Lines         []domain.ExpenseClaimLineInput `json:"lines" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claim, err := h.expenseSvc.SubmitClaim(
		c.Request.Context(),
		req.LegalEntityID,
		req.EmployeeID,
		req.ClaimNumber,
		req.Purpose,
		req.CostCenterTag,
		req.Lines,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, claim)
}

func (h *HrHandler) VerifyAndApproveClaim(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ReviewerHrID string `json:"reviewer_hr_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claim, err := h.expenseSvc.VerifyAndApproveClaim(c.Request.Context(), id, req.ReviewerHrID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, claim)
}

func (h *HrHandler) PayClaim(c *gin.Context) {
	id := c.Param("id")
	err := h.expenseSvc.ClearClaimForPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "paid"})
}

func (h *HrHandler) GetExpenseClaims(c *gin.Context) {
	claims, err := h.expenseRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, claims)
}

func (h *HrHandler) GetExpenseClaim(c *gin.Context) {
	id := c.Param("id")
	claim, err := h.expenseRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, claim)
}

func (h *HrHandler) GetExpenseClaimLines(c *gin.Context) {
	id := c.Param("id")
	lines, err := h.lineRepo.ListByClaimID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lines)
}
