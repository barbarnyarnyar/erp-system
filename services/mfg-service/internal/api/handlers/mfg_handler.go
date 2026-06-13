package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/erp-system/m-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type MfgHandler struct {
	floorSvc service.FloorConfigurationService
	execSvc  service.WorkOrderExecutionService
	teleSvc  service.ShopFloorTelemetryService
}

func NewMfgHandler(
	floorSvc service.FloorConfigurationService,
	execSvc service.WorkOrderExecutionService,
	teleSvc service.ShopFloorTelemetryService,
) *MfgHandler {
	return &MfgHandler{
		floorSvc: floorSvc,
		execSvc:  execSvc,
		teleSvc:  teleSvc,
	}
}

// 1. EstablishWorkCenter
type EstablishWorkCenterInput struct {
	LegalEntityID string `json:"legal_entity_id" binding:"required"`
	Code          string `json:"code" binding:"required"`
	Name          string `json:"name" binding:"required"`
}

func (h *MfgHandler) EstablishWorkCenter(c *gin.Context) {
	var input EstablishWorkCenterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wc, err := h.floorSvc.EstablishWorkCenter(c.Request.Context(), input.LegalEntityID, input.Code, input.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wc)
}

// 2. AppendStationToCenter
type AppendStationInput struct {
	RoutingCode           string             `json:"routing_code" binding:"required"`
	StationType           domain.StationType `json:"station_type" binding:"required"`
	EquipmentID           *string            `json:"equipment_id"`
	StandardSetupTimeMins int                `json:"standard_setup_time_mins"`
	StandardRunTimeMins   int                `json:"standard_run_time_mins"`
}

func (h *MfgHandler) AppendStationToCenter(c *gin.Context) {
	wcID := c.Param("id")
	var input AppendStationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	station, err := h.floorSvc.AppendStationToCenter(
		c.Request.Context(),
		wcID,
		input.RoutingCode,
		input.StationType,
		input.EquipmentID,
		input.StandardSetupTimeMins,
		input.StandardRunTimeMins,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, station)
}

// 3. InstantiateWorkOrder
type InstantiateWorkOrderInput struct {
	LegalEntityID  string          `json:"legal_entity_id" binding:"required"`
	MaterialID     string          `json:"material_id" binding:"required"`
	BomHeaderID    string          `json:"bom_header_id" binding:"required"`
	QuantityTarget decimal.Decimal `json:"quantity_target" binding:"required"`
	ScheduledStart time.Time       `json:"scheduled_start" binding:"required"`
	ScheduledEnd   time.Time       `json:"scheduled_end" binding:"required"`
}

func (h *MfgHandler) InstantiateWorkOrder(c *gin.Context) {
	var input InstantiateWorkOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wo, err := h.execSvc.InstantiateWorkOrder(
		c.Request.Context(),
		input.LegalEntityID,
		input.MaterialID,
		input.BomHeaderID,
		input.QuantityTarget,
		input.ScheduledStart,
		input.ScheduledEnd,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wo)
}

// 4. TransitionWorkOrderState
type TransitionWorkOrderStateInput struct {
	CurrentState domain.WorkOrderState `json:"current_state" binding:"required"`
	TargetState  domain.WorkOrderState `json:"target_state" binding:"required"`
}

func (h *MfgHandler) TransitionWorkOrderState(c *gin.Context) {
	woID := c.Param("id")
	var input TransitionWorkOrderStateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wo, err := h.execSvc.TransitionWorkOrderState(c.Request.Context(), woID, input.CurrentState, input.TargetState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wo)
}

// 5. RerouteWorkOrderStation
type RerouteWorkOrderStationInput struct {
	CurrentStationID string `json:"current_station_id" binding:"required"`
	TargetStationID  string `json:"target_station_id" binding:"required"`
	IsRework         bool   `json:"is_rework"`
}

func (h *MfgHandler) RerouteWorkOrderStation(c *gin.Context) {
	woID := c.Param("id")
	var input RerouteWorkOrderStationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.execSvc.RerouteWorkOrderStation(c.Request.Context(), woID, input.CurrentStationID, input.TargetStationID, input.IsRework)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "rerouted successfully"})
}

// 6. RecordBulkMaterialConsumption
type RecordBulkMaterialConsumptionInput struct {
	LegalEntityID string                              `json:"legal_entity_id" binding:"required"`
	Lines         []domain.ConsumptionSubmissionInput `json:"lines" binding:"required,dive"`
}

func (h *MfgHandler) RecordBulkMaterialConsumption(c *gin.Context) {
	woID := c.Param("id")
	var input RecordBulkMaterialConsumptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.teleSvc.RecordBulkMaterialConsumption(c.Request.Context(), input.LegalEntityID, woID, input.Lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "consumption recorded successfully"})
}

// 7. CommitProductionYield
type CommitProductionYieldInput struct {
	LegalEntityID string          `json:"legal_entity_id" binding:"required"`
	StationID     string          `json:"station_id" binding:"required"`
	QuantityGood  decimal.Decimal `json:"quantity_good" binding:"required"`
	QuantityScrap decimal.Decimal `json:"quantity_scrap"`
	OperatorHrID  string          `json:"operator_hr_id" binding:"required"`
}

func (h *MfgHandler) CommitProductionYield(c *gin.Context) {
	woID := c.Param("id")
	var input CommitProductionYieldInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.teleSvc.CommitProductionYield(c.Request.Context(), input.LegalEntityID, woID, input.StationID, input.QuantityGood, input.QuantityScrap, input.OperatorHrID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "yield committed successfully"})
}
