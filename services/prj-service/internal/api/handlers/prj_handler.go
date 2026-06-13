package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PrjHandler struct {
	projTrackingSvc service.ProjectTrackingService
	wbsSvc          service.WbsStructureService
	timeSvc         service.TimeTrackingService
	projRepo        domain.ProjectRepository
	wbsRepo         domain.WbsNodeRepository
	timeRepo        domain.TimeLogRepository
}

func NewPrjHandler(
	projTrackingSvc service.ProjectTrackingService,
	wbsSvc service.WbsStructureService,
	timeSvc service.TimeTrackingService,
	projRepo domain.ProjectRepository,
	wbsRepo domain.WbsNodeRepository,
	timeRepo domain.TimeLogRepository,
) *PrjHandler {
	return &PrjHandler{
		projTrackingSvc: projTrackingSvc,
		wbsSvc:          wbsSvc,
		timeSvc:         timeSvc,
		projRepo:        projRepo,
		wbsRepo:         wbsRepo,
		timeRepo:        timeRepo,
	}
}

func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

// ============================================================================
// Project Handlers
// ============================================================================

func (h *PrjHandler) InitializeProject(c *gin.Context) {
	var req struct {
		LegalEntityID string `json:"legal_entity_id" binding:"required"`
		CustomerID    string `json:"customer_id" binding:"required"`
		ProjectCode   string `json:"project_code" binding:"required"`
		Name          string `json:"name" binding:"required"`
		BillingMethod string `json:"billing_method" binding:"required"`
		StartDate     string `json:"start_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
		return
	}

	proj, err := h.projTrackingSvc.InitializeProject(
		c.Request.Context(),
		req.LegalEntityID,
		req.CustomerID,
		req.ProjectCode,
		req.Name,
		domain.BillingMethod(req.BillingMethod),
		startDate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, proj)
}

func (h *PrjHandler) TransitionProjectStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proj, err := h.projTrackingSvc.TransitionProjectStatus(
		c.Request.Context(),
		id,
		domain.ProjectStatus(req.Status),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proj)
}

func (h *PrjHandler) ListProjects(c *gin.Context) {
	list, err := h.projRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *PrjHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	proj, err := h.projRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proj)
}

// ============================================================================
// WbsNode Handlers
// ============================================================================

func (h *PrjHandler) AppendWbsNode(c *gin.Context) {
	projectID := c.Param("id")
	var req struct {
		ParentNodeID   *string         `json:"parent_node_id"`
		NodeCode       string          `json:"node_code" binding:"required"`
		Title          string          `json:"title" binding:"required"`
		NodeType       string          `json:"node_type" binding:"required"`
		EstimatedHours decimal.Decimal `json:"estimated_hours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node, err := h.wbsSvc.AppendWbsNode(
		c.Request.Context(),
		projectID,
		req.ParentNodeID,
		req.NodeCode,
		req.Title,
		domain.WbsNodeType(req.NodeType),
		req.EstimatedHours,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, node)
}

func (h *PrjHandler) DeclareNodeCompletion(c *gin.Context) {
	nodeID := c.Param("node_id")
	var req struct {
		CompletionHrID string `json:"completion_hr_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node, err := h.wbsSvc.DeclareNodeCompletion(
		c.Request.Context(),
		nodeID,
		req.CompletionHrID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, node)
}

func (h *PrjHandler) FetchProjectTree(c *gin.Context) {
	projectID := c.Param("id")
	tree, err := h.wbsSvc.FetchProjectTree(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tree)
}

// ============================================================================
// TimeLog Handlers
// ============================================================================

func (h *PrjHandler) LogOperationalHoursBulk(c *gin.Context) {
	var req struct {
		LegalEntityID string `json:"legal_entity_id" binding:"required"`
		EmployeeID    string `json:"employee_id" binding:"required"`
		Logs          []struct {
			WbsNodeID        string          `json:"wbs_node_id" binding:"required"`
			WorkDate         string          `json:"work_date" binding:"required"`
			HoursSpent       decimal.Decimal `json:"hours_spent" binding:"required"`
			InternalCostRate decimal.Decimal `json:"internal_cost_rate" binding:"required"`
			BillingRate      decimal.Decimal `json:"billing_rate" binding:"required"`
			IsBillable       bool            `json:"is_billable"`
		} `json:"logs" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var domainLogs []domain.TimeLogSubmissionInput
	for _, l := range req.Logs {
		wd, err := parseDate(l.WorkDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work_date format, use YYYY-MM-DD"})
			return
		}
		domainLogs = append(domainLogs, domain.TimeLogSubmissionInput{
			WbsNodeID:        l.WbsNodeID,
			WorkDate:         wd,
			HoursSpent:       l.HoursSpent,
			InternalCostRate: l.InternalCostRate,
			BillingRate:      l.BillingRate,
			IsBillable:       l.IsBillable,
		})
	}

	err := h.timeSvc.LogOperationalHoursBulk(
		c.Request.Context(),
		req.LegalEntityID,
		req.EmployeeID,
		domainLogs,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Time logs logged successfully"})
}

func (h *PrjHandler) ProcessTimesheetApproval(c *gin.Context) {
	var req struct {
		TimeLogIDs   []string `json:"time_log_ids" binding:"required"`
		ApproverHrID string   `json:"approver_hr_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.timeSvc.ProcessTimesheetApproval(
		c.Request.Context(),
		req.TimeLogIDs,
		req.ApproverHrID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Time logs approved successfully"})
}
