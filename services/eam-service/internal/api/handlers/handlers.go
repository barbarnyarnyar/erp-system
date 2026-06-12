package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/erp-system/eam-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type EamHandler struct {
	eqSvc  *service.EquipmentService
	maintSvc *service.MaintenanceService
	telSvc *service.TelemetryIngestionService
	resp   *utils.ResponseHelper
}

func NewEamHandler(eqSvc *service.EquipmentService, maintSvc *service.MaintenanceService, telSvc *service.TelemetryIngestionService, resp *utils.ResponseHelper) *EamHandler {
	return &EamHandler{
		eqSvc:    eqSvc,
		maintSvc: maintSvc,
		telSvc:   telSvc,
		resp:     resp,
	}
}

// Facility Handlers
func (h *EamHandler) CreateFacility(c *gin.Context) {
	var req struct {
		LegalEntityID   string `json:"legal_entity_id" binding:"required"`
		Name            string `json:"name" binding:"required"`
		PhysicalAddress string `json:"physical_address" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	f, err := h.eqSvc.CreateFacility(c.Request.Context(), req.LegalEntityID, req.Name, req.PhysicalAddress)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": f})
}

// Equipment Handlers
func (h *EamHandler) RegisterEquipment(c *gin.Context) {
	var req struct {
		LegalEntityID string `json:"legal_entity_id" binding:"required"`
		FacilityID    string `json:"facility_id" binding:"required"`
		AssetTag      string `json:"asset_tag" binding:"required"`
		Name          string `json:"name" binding:"required"`
		SerialNumber  string `json:"serial_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	eq, err := h.eqSvc.RegisterEquipment(c.Request.Context(), req.LegalEntityID, req.FacilityID, req.AssetTag, req.Name, req.SerialNumber)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": eq})
}

func (h *EamHandler) UpdateEquipmentStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status domain.EquipmentStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	eq, err := h.eqSvc.UpdateEquipmentStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": eq})
}

func (h *EamHandler) AssociateFinancialAsset(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		FinancialAssetID string `json:"financial_asset_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	eq, err := h.eqSvc.AssociateFinancialAsset(c.Request.Context(), id, req.FinancialAssetID)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": eq})
}

func (h *EamHandler) FetchTargetTenantAssets(c *gin.Context) {
	tenantID := c.Query("legal_entity_id")
	status := domain.EquipmentStatus(c.Query("status"))
	if tenantID == "" {
		h.resp.BadRequest(c, "legal_entity_id query param required")
		return
	}
	list, err := h.eqSvc.FetchTargetTenantAssets(c.Request.Context(), tenantID, status)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// Work Order Handlers
func (h *EamHandler) FileMachineIncident(c *gin.Context) {
	var req struct {
		LegalEntityID string                   `json:"legal_entity_id" binding:"required"`
		EquipmentID   string                   `json:"equipment_id" binding:"required"`
		ReportedBy    string                   `json:"reported_by" binding:"required"`
		Title         string                   `json:"title" binding:"required"`
		Priority      domain.WorkOrderPriority `json:"priority" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	wo, err := h.maintSvc.FileMachineIncident(c.Request.Context(), req.LegalEntityID, req.EquipmentID, req.ReportedBy, req.Title, req.Priority)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": wo})
}

func (h *EamHandler) RouteToTechnician(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		AssignedTech string `json:"assigned_tech" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	wo, err := h.maintSvc.RouteToTechnician(c.Request.Context(), id, req.AssignedTech)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wo})
}

func (h *EamHandler) TransitionToActiveState(c *gin.Context) {
	id := c.Param("id")
	wo, err := h.maintSvc.TransitionToActiveState(c.Request.Context(), id)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wo})
}

func (h *EamHandler) FinalizeResolution(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Notes string `json:"notes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	wo, err := h.maintSvc.FinalizeResolution(c.Request.Context(), id, req.Notes)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wo})
}

// Telemetry Handlers
func (h *EamHandler) QueueSensorMetrics(c *gin.Context) {
	var req struct {
		LegalEntityID string          `json:"legal_entity_id" binding:"required"`
		EquipmentID   string          `json:"equipment_id" binding:"required"`
		SensorKey     string          `json:"sensor_key" binding:"required"`
		Value         decimal.Decimal `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.BadRequest(c, err.Error())
		return
	}
	err := h.telSvc.QueueSensorMetrics(c.Request.Context(), req.LegalEntityID, req.EquipmentID, req.SensorKey, req.Value)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "queued"})
}

func (h *EamHandler) FlushStagedMetrics(c *gin.Context) {
	var req struct {
		Limit int `json:"limit" defaultValue:"100"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Limit <= 0 {
		req.Limit = 100
	}
	ids, err := h.telSvc.FlushStagedMetricsToTimeSeriesStore(c.Request.Context(), req.Limit)
	if err != nil {
		h.resp.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"flushed_count": len(ids), "ids": ids})
}
