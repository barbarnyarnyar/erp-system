package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type InventoryHandler struct {
	svc      *service.InventoryService
	response *utils.ResponseHelper
}

func NewInventoryHandler(svc *service.InventoryService, response *utils.ResponseHelper) *InventoryHandler {
	return &InventoryHandler{
		svc:      svc,
		response: response,
	}
}

func (h *InventoryHandler) GetInventoryItems(c *gin.Context) {
	list, err := h.svc.ListInventory(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *InventoryHandler) CreateInventoryItem(c *gin.Context) {
	var req struct {
		MaterialID     string `json:"material_id"`
		ProductID      string `json:"product_id"`
		LocationID     string `json:"location_id"`
		QuantityOnHand string `json:"quantity_on_hand"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	matID := req.MaterialID
	if matID == "" {
		matID = req.ProductID
	}

	qtyDec, err := decimal.NewFromString(req.QuantityOnHand)
	if err != nil {
		qtyDec = decimal.Zero
	}

	sb, err := h.svc.CreateStockBalance(c.Request.Context(), matID, req.LocationID, qtyDec)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": sb})
}

func (h *InventoryHandler) GetInventoryItem(c *gin.Context) {
	id := c.Param("id")
	sb, err := h.svc.GetStockBalance(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "stock balance not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sb})
}

func (h *InventoryHandler) UpdateInventoryItem(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		QuantityOnHand   string `json:"quantity_on_hand"`
		QuantityReserved string `json:"quantity_reserved"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	qtyOnHandDec, err := decimal.NewFromString(req.QuantityOnHand)
	if err != nil {
		qtyOnHandDec = decimal.Zero
	}
	qtyReservedDec, err := decimal.NewFromString(req.QuantityReserved)
	if err != nil {
		qtyReservedDec = decimal.Zero
	}

	var sb *domain.StockBalance
	for i := 0; i < 5; i++ {
		sb, err = h.svc.UpdateStockBalance(c.Request.Context(), id, qtyOnHandDec, qtyReservedDec)
		if err != domain.ErrOptimisticLock {
			break
		}
	}
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sb})
}

func (h *InventoryHandler) DeleteInventoryItem(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "inventory tracker deleted successfully. ID: " + id})
}

func (h *InventoryHandler) ReserveStock(c *gin.Context) {
	var req struct {
		MaterialID  string `json:"material_id"`
		ProductID   string `json:"product_id"`
		LocationID  string `json:"location_id"`
		Quantity    string `json:"quantity"`
		ReferenceID string `json:"reference_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	matID := req.MaterialID
	if matID == "" {
		matID = req.ProductID
	}

	qtyDec, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		qtyDec = decimal.Zero
	}

	for i := 0; i < 5; i++ {
		err = h.svc.ReserveStock(c.Request.Context(), matID, req.LocationID, qtyDec, req.ReferenceID)
		if err != domain.ErrOptimisticLock {
			break
		}
	}
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock reserved successfully"})
}

func (h *InventoryHandler) ReleaseReservation(c *gin.Context) {
	var req struct {
		ReferenceID string `json:"reference_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	var err error
	for i := 0; i < 5; i++ {
		err = h.svc.ReleaseReservation(c.Request.Context(), req.ReferenceID)
		if err != domain.ErrOptimisticLock {
			break
		}
	}
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock reservation released successfully"})
}

func (h *InventoryHandler) CreateStockTransfer(c *gin.Context) {
	var req struct {
		FromLocationID string `json:"from_location_id"`
		ToLocationID   string `json:"to_location_id"`
		ProductID      string `json:"product_id"`
		Quantity       int    `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	var st *domain.StockTransfer
	var err error
	for i := 0; i < 5; i++ {
		st, err = h.svc.CreateStockTransfer(c.Request.Context(), req.FromLocationID, req.ToLocationID, req.ProductID, req.Quantity)
		if err != domain.ErrOptimisticLock {
			break
		}
	}
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": st})
}

func (h *InventoryHandler) GetStockTransfer(c *gin.Context) {
	id := c.Param("id")
	st, err := h.svc.GetStockTransfer(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "stock transfer not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": st})
}

func (h *InventoryHandler) GetStockTransfers(c *gin.Context) {
	list, err := h.svc.ListStockTransfers(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *InventoryHandler) ExecuteStockTransfer(c *gin.Context) {
	id := c.Param("id")
	var st *domain.StockTransfer
	var err error
	for i := 0; i < 5; i++ {
		st, err = h.svc.ExecuteStockTransfer(c.Request.Context(), id)
		if err != domain.ErrOptimisticLock {
			break
		}
	}
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": st})
}

func (h *InventoryHandler) GetInventoryMovements(c *gin.Context) {
	list, err := h.svc.ListMovements(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
