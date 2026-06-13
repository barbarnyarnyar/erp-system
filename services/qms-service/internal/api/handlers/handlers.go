package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/erp-system/qms-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type QmsHandler struct {
	planSvc  *service.InspectionPlanService
	execSvc  *service.InspectionExecutionService
	ncSvc    *service.NonConformanceService
	analySvc *service.QualityAnalyticsService
	resp     *utils.ResponseHelper
}

func NewQmsHandler(
	planSvc *service.InspectionPlanService,
	execSvc *service.InspectionExecutionService,
	ncSvc *service.NonConformanceService,
	analySvc *service.QualityAnalyticsService,
	resp *utils.ResponseHelper,
) *QmsHandler {
	return &QmsHandler{
		planSvc:  planSvc,
		execSvc:  execSvc,
		ncSvc:    ncSvc,
		analySvc: analySvc,
		resp:     resp,
	}
}

// Inspection Plan Handlers
func (h *QmsHandler) ConfigurePlan(c *gin.Context) {
	var req struct {
		LegalEntityID string `json:"legal_entity_id" binding:"required"`
		MaterialID    string `json:"material_id" binding:"required"`
		PlanName      string `json:"plan_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	p, err := h.planSvc.ConfigurePlan(c.Request.Context(), req.LegalEntityID, req.MaterialID, req.PlanName)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p})
}

func (h *QmsHandler) RegisterPlanMetric(c *gin.Context) {
	var req struct {
		PlanID            string              `json:"plan_id" binding:"required"`
		MetricKey         string              `json:"metric_key" binding:"required"`
		DisplayName       string              `json:"display_name" binding:"required"`
		DataType          domain.MetricDataType `json:"data_type" binding:"required"`
		MinToleranceLimit *decimal.Decimal    `json:"min_tolerance_limit"`
		MaxToleranceLimit *decimal.Decimal    `json:"max_tolerance_limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	m, err := h.planSvc.RegisterPlanMetric(c.Request.Context(), req.PlanID, req.MetricKey, req.DisplayName, req.DataType, req.MinToleranceLimit, req.MaxToleranceLimit)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": m})
}

// Inspection Execution Handlers
func (h *QmsHandler) StageInspection(c *gin.Context) {
	var req struct {
		LegalEntityID string                       `json:"legal_entity_id" binding:"required"`
		PlanID        string                       `json:"plan_id" binding:"required"`
		TriggerSource domain.InspectionTriggerType `json:"trigger_source" binding:"required"`
		SourceDocID   string                       `json:"source_doc_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	qi, err := h.execSvc.StageInspection(c.Request.Context(), req.LegalEntityID, req.PlanID, req.TriggerSource, req.SourceDocID)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": qi})
}

func (h *QmsHandler) AssignInspector(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		InspectorHrID string `json:"inspector_hr_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	qi, err := h.execSvc.AssignInspector(c.Request.Context(), id, req.InspectorHrID)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": qi})
}

func (h *QmsHandler) RecordBulkMeasurements(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Samples []domain.MetricSubmissionInput `json:"samples" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	err := h.execSvc.RecordBulkMeasurements(c.Request.Context(), id, req.Samples)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

// Non Conformance Handlers
func (h *QmsHandler) LogFailureIncident(c *gin.Context) {
	var req struct {
		LegalEntityID  string          `json:"legal_entity_id" binding:"required"`
		InspectionID   string          `json:"inspection_id" binding:"required"`
		Description    string          `json:"description" binding:"required"`
		QtyDefective   decimal.Decimal `json:"qty_defective" binding:"required"`
		AutoQuarantine bool            `json:"auto_quarantine"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	nc, err := h.ncSvc.LogFailureIncident(c.Request.Context(), req.LegalEntityID, req.InspectionID, req.Description, req.QtyDefective, req.AutoQuarantine)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": nc})
}

func (h *QmsHandler) ExecuteDisposition(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Action       domain.DispositionAction `json:"action" binding:"required"`
		Notes        string                   `json:"notes" binding:"required"`
		ResolverHrID string                   `json:"resolver_hr_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	nc, err := h.ncSvc.ExecuteDisposition(c.Request.Context(), id, req.Action, req.Notes, req.ResolverHrID)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": nc})
}

// Analytics Handlers
func (h *QmsHandler) ComputeSpcDistribution(c *gin.Context) {
	planId := c.Query("plan_id")
	metricDefId := c.Query("metric_def_id")
	if planId == "" || metricDefId == "" {
		h.resp.BadRequest(c, "plan_id and metric_def_id are required query params")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	start := time.Now().AddDate(0, 0, -30) // Default last 30 days
	end := time.Now()

	if startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			start = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			end = t
		}
	}

	window := domain.TimeRange{
		StartDate: start,
		EndDate:   end,
	}

	summary, err := h.analySvc.ComputeSpcDistribution(c.Request.Context(), planId, metricDefId, window)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}
