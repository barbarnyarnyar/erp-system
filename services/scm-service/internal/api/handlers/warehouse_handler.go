package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type WarehouseHandler struct {
	svc      *service.WarehouseService
	response *utils.ResponseHelper
}

func NewWarehouseHandler(svc *service.WarehouseService, response *utils.ResponseHelper) *WarehouseHandler {
	return &WarehouseHandler{
		svc:      svc,
		response: response,
	}
}

// ============================================================================
// RECEIPTS ENDPOINTS
// ============================================================================

func (h *WarehouseHandler) GetReceipts(c *gin.Context) {
	list, err := h.svc.ListReceipts(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *WarehouseHandler) CreateReceipt(c *gin.Context) {
	var req struct {
		PurchaseOrderID string `json:"purchase_order_id"`
		Notes           string `json:"notes"`
		Lines           []struct {
			ProductID        string `json:"product_id"`
			QuantityReceived int    `json:"quantity_received"`
			LocationID       string `json:"location_id"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	linesInput := make([]service.ReceiptLineInput, 0, len(req.Lines))
	for _, l := range req.Lines {
		linesInput = append(linesInput, service.ReceiptLineInput{
			ProductID:        l.ProductID,
			QuantityReceived: l.QuantityReceived,
			LocationID:       l.LocationID,
		})
	}

	rec, err := h.svc.CreateReceipt(c.Request.Context(), req.PurchaseOrderID, req.Notes, linesInput)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": rec})
}

func (h *WarehouseHandler) GetReceipt(c *gin.Context) {
	id := c.Param("id")
	rec, err := h.svc.GetReceipt(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "receipt not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rec})
}

func (h *WarehouseHandler) UpdateReceipt(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	rec, err := h.svc.UpdateReceipt(c.Request.Context(), id, req.Status, req.Notes)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rec})
}

// ============================================================================
// SHIPMENTS ENDPOINTS
// ============================================================================

func (h *WarehouseHandler) GetShipments(c *gin.Context) {
	list, err := h.svc.ListShipments(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *WarehouseHandler) CreateShipment(c *gin.Context) {
	var req struct {
		Carrier           string `json:"carrier"`
		TrackingNumber    string `json:"tracking_number"`
		EstimatedDelivery string `json:"estimated_delivery"`
		Notes             string `json:"notes"`
		Lines             []struct {
			ProductID       string `json:"product_id"`
			QuantityShipped int    `json:"quantity_shipped"`
			LocationID      string `json:"location_id"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	estDeliveryTime, err := time.Parse(time.RFC3339, req.EstimatedDelivery)
	if err != nil {
		estDeliveryTime = time.Now().AddDate(0, 0, 3) // default to 3 days out
	}

	linesInput := make([]service.ShipmentLineInput, 0, len(req.Lines))
	for _, l := range req.Lines {
		linesInput = append(linesInput, service.ShipmentLineInput{
			ProductID:       l.ProductID,
			QuantityShipped: l.QuantityShipped,
			LocationID:      l.LocationID,
		})
	}

	shp, err := h.svc.CreateShipment(c.Request.Context(), req.Carrier, req.TrackingNumber, estDeliveryTime, req.Notes, linesInput)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": shp})
}

func (h *WarehouseHandler) GetShipment(c *gin.Context) {
	id := c.Param("id")
	shp, err := h.svc.GetShipment(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "shipment not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shp})
}

func (h *WarehouseHandler) UpdateShipment(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	shp, err := h.svc.UpdateShipment(c.Request.Context(), id, req.Status, req.Notes)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": shp})
}

func (h *WarehouseHandler) GetReceiptLines(c *gin.Context) {
	id := c.Param("id")
	lines, err := h.svc.ListReceiptLines(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lines})
}

func (h *WarehouseHandler) GetShipmentLines(c *gin.Context) {
	id := c.Param("id")
	lines, err := h.svc.ListShipmentLines(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lines})
}
