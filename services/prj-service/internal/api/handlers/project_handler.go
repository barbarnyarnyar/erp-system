package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type ProjectHandler struct {
	planningSvc  *service.ProjectPlanningService
	taskSvc      *service.TaskManagementService
	resourceSvc  *service.ResourceManagementService
	timeSvc      *service.TimeExpenseService
	collabSvc    *service.CollaborationService
	analyticsSvc *service.PortfolioAnalyticsService
	response *utils.ResponseHelper
}

func NewProjectHandler(planningSvc *service.ProjectPlanningService,
	taskSvc *service.TaskManagementService,
	resourceSvc *service.ResourceManagementService,
	timeSvc *service.TimeExpenseService,
	collabSvc *service.CollaborationService,
	analyticsSvc *service.PortfolioAnalyticsService, response *utils.ResponseHelper) *ProjectHandler {
	return &ProjectHandler{
		planningSvc:  planningSvc,
		taskSvc:      taskSvc,
		resourceSvc:  resourceSvc,
		timeSvc:      timeSvc,
		collabSvc:    collabSvc,
		analyticsSvc: analyticsSvc,
		response: response,
	}
}

// ==========================================
// Portfolios
// ==========================================

type CreatePortfolioReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ManagerID   string `json:"manager_id"`
}

func (h *ProjectHandler) CreatePortfolio(c *gin.Context) {
	var req CreatePortfolioReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	p, err := h.planningSvc.CreatePortfolio(c.Request.Context(), req.Name, req.Description, req.ManagerID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *ProjectHandler) ListPortfolios(c *gin.Context) {
	list, err := h.planningSvc.ListPortfolios(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ProjectHandler) GetPortfolio(c *gin.Context) {
	id := c.Param("id")
	p, err := h.planningSvc.GetPortfolio(c.Request.Context(), id)
	if err != nil {
		h.response.NotFoundErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProjectHandler) GetPortfolioSummary(c *gin.Context) {
	id := c.Param("id")
	summary, err := h.analyticsSvc.GetPortfolioSummary(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

// ==========================================
// Projects
// ==========================================

type CreateProjectReq struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date"`
	PortfolioID string     `json:"portfolio_id"`
	BudgetID    string     `json:"budget_id"`
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	p, err := h.planningSvc.CreateProject(c.Request.Context(), req.Name, req.Description, req.StartDate, req.EndDate, req.PortfolioID, req.BudgetID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	list, err := h.planningSvc.ListProjects(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	p, err := h.planningSvc.GetProject(c.Request.Context(), id)
	if err != nil {
		h.response.NotFoundErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

type UpdateProjectStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func (h *ProjectHandler) UpdateProjectStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateProjectStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	p, err := h.planningSvc.UpdateProjectStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

// ==========================================
// Tasks
// ==========================================

type CreateTaskReq struct {
	ParentID       string          `json:"parent_id"`
	Title          string          `json:"title" binding:"required"`
	Description    string          `json:"description"`
	AssignedTo     string          `json:"assigned_to"`
	StartDate      *time.Time      `json:"start_date"`
	EndDate        *time.Time      `json:"end_date"`
	EstimatedHours decimal.Decimal `json:"estimated_hours"`
}

func (h *ProjectHandler) CreateTask(c *gin.Context) {
	projectID := c.Param("id")
	var req CreateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	t, err := h.taskSvc.CreateTask(c.Request.Context(), projectID, req.ParentID, req.Title, req.Description, req.AssignedTo, req.StartDate, req.EndDate, req.EstimatedHours)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *ProjectHandler) ListTasks(c *gin.Context) {
	projectID := c.Param("id")
	list, err := h.taskSvc.ListTasksByProject(c.Request.Context(), projectID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

type UpdateTaskProgressReq struct {
	Progress    int             `json:"progress" binding:"required"`
	ActualHours decimal.Decimal `json:"actual_hours" binding:"required"`
	Status      string          `json:"status" binding:"required"`
}

func (h *ProjectHandler) UpdateTaskProgress(c *gin.Context) {
	taskID := c.Param("task_id")
	var req UpdateTaskProgressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	t, err := h.taskSvc.UpdateTaskProgress(c.Request.Context(), taskID, req.Progress, req.ActualHours, req.Status)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

type AssignTaskReq struct {
	EmployeeID string `json:"employee_id" binding:"required"`
}

func (h *ProjectHandler) AssignTask(c *gin.Context) {
	taskID := c.Param("task_id")
	var req AssignTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	t, err := h.taskSvc.AssignTask(c.Request.Context(), taskID, req.EmployeeID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// ==========================================
// Dependencies
// ==========================================

type AddDependencyReq struct {
	DependsOnTaskID string `json:"depends_on_task_id" binding:"required"`
	DependencyType  string `json:"dependency_type" binding:"required"` // FS, SS, etc.
}

func (h *ProjectHandler) AddTaskDependency(c *gin.Context) {
	taskID := c.Param("task_id")
	var req AddDependencyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	dep, err := h.taskSvc.AddTaskDependency(c.Request.Context(), taskID, req.DependsOnTaskID, req.DependencyType)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, dep)
}

// ==========================================
// Resource Allocations
// ==========================================

type AllocateResourceReq struct {
	UserID               string     `json:"user_id" binding:"required"`
	Role                 string     `json:"role" binding:"required"`
	AllocationPercentage int        `json:"allocation_percentage" binding:"required"`
	StartDate            time.Time  `json:"start_date" binding:"required"`
	EndDate              *time.Time `json:"end_date"`
}

func (h *ProjectHandler) AllocateResource(c *gin.Context) {
	projectID := c.Param("id")
	var req AllocateResourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	alloc, err := h.resourceSvc.AllocateResource(c.Request.Context(), projectID, req.UserID, req.Role, req.AllocationPercentage, req.StartDate, req.EndDate)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, alloc)
}

func (h *ProjectHandler) ListAllocations(c *gin.Context) {
	projectID := c.Param("id")
	list, err := h.resourceSvc.ListAllocations(c.Request.Context(), projectID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// ==========================================
// Time Entries
// ==========================================

type LogTimeReq struct {
	TaskID      string          `json:"task_id" binding:"required"`
	UserID      string          `json:"user_id" binding:"required"`
	EntryDate   time.Time       `json:"entry_date" binding:"required"`
	Hours       decimal.Decimal `json:"hours" binding:"required"`
	Description string          `json:"description"`
}

func (h *ProjectHandler) LogTime(c *gin.Context) {
	projectID := c.Param("id")
	var req LogTimeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	entry, err := h.timeSvc.LogTime(c.Request.Context(), projectID, req.TaskID, req.UserID, req.EntryDate, req.Hours, req.Description)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, entry)
}

type ApproveTimeReq struct {
	ApprovedBy string `json:"approved_by" binding:"required"`
}

func (h *ProjectHandler) ApproveTime(c *gin.Context) {
	entryID := c.Param("time_id")
	var req ApproveTimeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	entry, err := h.timeSvc.ApproveTime(c.Request.Context(), entryID, req.ApprovedBy)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *ProjectHandler) ListTimeEntries(c *gin.Context) {
	projectID := c.Param("id")
	list, err := h.timeSvc.ListTimeEntries(c.Request.Context(), projectID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// ==========================================
// Expenses
// ==========================================

type LogExpenseReq struct {
	TaskID      string          `json:"task_id"`
	UserID      string          `json:"user_id" binding:"required"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
	Currency    string          `json:"currency" binding:"required"`
	ExpenseDate time.Time       `json:"expense_date" binding:"required"`
	Category    string          `json:"category" binding:"required"`
	Description string          `json:"description" binding:"required"`
}

func (h *ProjectHandler) LogExpense(c *gin.Context) {
	projectID := c.Param("id")
	var req LogExpenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	exp, err := h.timeSvc.LogExpense(c.Request.Context(), projectID, req.TaskID, req.UserID, req.Amount, req.Currency, req.ExpenseDate, req.Category, req.Description)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, exp)
}

type ApproveExpenseReq struct {
	ApprovedBy string `json:"approved_by" binding:"required"`
}

func (h *ProjectHandler) ApproveExpense(c *gin.Context) {
	expenseID := c.Param("expense_id")
	var req ApproveExpenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	exp, err := h.timeSvc.ApproveExpense(c.Request.Context(), expenseID, req.ApprovedBy)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, exp)
}

func (h *ProjectHandler) ListExpenses(c *gin.Context) {
	projectID := c.Param("id")
	list, err := h.timeSvc.ListExpenses(c.Request.Context(), projectID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// ==========================================
// Documents
// ==========================================

type UploadDocReq struct {
	Name     string `json:"name" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
	FileSize int    `json:"file_size" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
}

func (h *ProjectHandler) UploadDocument(c *gin.Context) {
	projectID := c.Param("id")
	var req UploadDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	doc, err := h.collabSvc.UploadDocument(c.Request.Context(), projectID, req.Name, req.FilePath, req.FileSize, req.UserID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, doc)
}

func (h *ProjectHandler) ListDocuments(c *gin.Context) {
	projectID := c.Param("id")
	list, err := h.collabSvc.ListDocuments(c.Request.Context(), projectID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// ==========================================
// Issues
// ==========================================

type LogIssueReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Severity    string `json:"severity" binding:"required"` // LOW, MEDIUM, HIGH, CRITICAL
	UserID      string `json:"user_id" binding:"required"`
}

func (h *ProjectHandler) LogIssue(c *gin.Context) {
	projectID := c.Param("id")
	var req LogIssueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	issue, err := h.collabSvc.LogIssue(c.Request.Context(), projectID, req.Title, req.Description, req.Severity, req.UserID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, issue)
}

type ResolveIssueReq struct {
	AssignedTo string `json:"assigned_to"`
}

func (h *ProjectHandler) ResolveIssue(c *gin.Context) {
	issueID := c.Param("issue_id")
	var req ResolveIssueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	issue, err := h.collabSvc.ResolveIssue(c.Request.Context(), issueID, req.AssignedTo)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, issue)
}

func (h *ProjectHandler) ListIssues(c *gin.Context) {
	projectID := c.Param("id")
	list, err := h.collabSvc.ListIssues(c.Request.Context(), projectID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// ==========================================
// Change Requests
// ==========================================

type CreateChangeReq struct {
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
	ImpactAnalysis string `json:"impact_analysis" binding:"required"`
	UserID         string `json:"user_id" binding:"required"`
}

func (h *ProjectHandler) CreateChangeRequest(c *gin.Context) {
	projectID := c.Param("id")
	var req CreateChangeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	cr, err := h.collabSvc.CreateChangeRequest(c.Request.Context(), projectID, req.Title, req.Description, req.Reason, req.ImpactAnalysis, req.UserID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, cr)
}

type ApproveChangeReq struct {
	ApprovedBy string `json:"approved_by" binding:"required"`
}

func (h *ProjectHandler) ApproveChangeRequest(c *gin.Context) {
	requestID := c.Param("request_id")
	var req ApproveChangeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	cr, err := h.collabSvc.ApproveChangeRequest(c.Request.Context(), requestID, req.ApprovedBy)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, cr)
}

func (h *ProjectHandler) ListChangeRequests(c *gin.Context) {
	projectID := c.Param("id")
	list, err := h.collabSvc.ListChangeRequests(c.Request.Context(), projectID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// ==========================================
// Cross-Service Integration Triggers
// ==========================================

type RequestMaterialReq struct {
	TaskID    string `json:"task_id" binding:"required"`
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

func (h *ProjectHandler) RequestMaterial(c *gin.Context) {
	projectID := c.Param("id")
	var req RequestMaterialReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	err := h.taskSvc.RequestMaterial(c.Request.Context(), projectID, req.TaskID, req.ProductID, req.Quantity)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Material request event successfully published to SCM"})
}

type RequestCustomOrderReq struct {
	CustomItemID string    `json:"custom_item_id" binding:"required"`
	Quantity     int       `json:"quantity" binding:"required"`
	RequiredBy   time.Time `json:"required_by" binding:"required"`
}

func (h *ProjectHandler) RequestCustomOrder(c *gin.Context) {
	projectID := c.Param("id")
	var req RequestCustomOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	err := h.planningSvc.RequestCustomOrder(c.Request.Context(), projectID, req.CustomItemID, req.Quantity, req.RequiredBy)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Custom order request event successfully published to Manufacturing"})
}
