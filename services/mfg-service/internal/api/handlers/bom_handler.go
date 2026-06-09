package handlers

import (
	"net/http"

	"github.com/erp-system/m-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type BOMHandler struct {
	svc *service.BOMService
}

func NewBOMHandler(svc *service.BOMService) *BOMHandler {
	return &BOMHandler{svc: svc}
}

func (h *BOMHandler) CreateBillOfMaterials(c *gin.Context) {
	var req struct {
		ProductID   string `json:"product_id"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bom, err := h.svc.CreateBillOfMaterials(c.Request.Context(), req.ProductID, req.Version, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": bom})
}

func (h *BOMHandler) GetBillOfMaterials(c *gin.Context) {
	id := c.Param("id")
	bom, err := h.svc.GetBillOfMaterials(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "BOM not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": bom})
}

func (h *BOMHandler) ListBOMs(c *gin.Context) {
	list, err := h.svc.ListBOMs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BOMHandler) UpdateBillOfMaterials(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ProductID   string `json:"product_id"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bom, err := h.svc.UpdateBillOfMaterials(c.Request.Context(), id, req.ProductID, req.Version, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bom})
}

func (h *BOMHandler) DeleteBillOfMaterials(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteBillOfMaterials(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "BOM deleted successfully"})
}

func (h *BOMHandler) GetBOMComponents(c *gin.Context) {
	bomID := c.Param("id")
	components, err := h.svc.GetBOMComponents(c.Request.Context(), bomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": components})
}

func (h *BOMHandler) AddBOMComponent(c *gin.Context) {
	bomID := c.Param("id")
	var req struct {
		ComponentProductID string `json:"component_product_id"`
		Quantity           string `json:"quantity"`
		WasteFactor        string `json:"waste_factor"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qtyDec, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quantity decimal"})
		return
	}
	wasteDec, err := decimal.NewFromString(req.WasteFactor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid waste_factor decimal"})
		return
	}

	comp, err := h.svc.AddBOMComponent(c.Request.Context(), bomID, req.ComponentProductID, qtyDec, wasteDec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": comp})
}

func (h *BOMHandler) RemoveBOMComponent(c *gin.Context) {
	bomID := c.Param("id")
	componentID := c.Param("componentId")
	err := h.svc.RemoveBOMComponent(c.Request.Context(), bomID, componentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "BOM component removed successfully"})
}

func (h *BOMHandler) CreateWorkCenter(c *gin.Context) {
	var req struct {
		Code          string `json:"code"`
		Name          string `json:"name"`
		CapacityHours string `json:"capacity_hours"`
		HourlyRate    string `json:"hourly_rate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	capacityDec, _ := decimal.NewFromString(req.CapacityHours)
	rateDec, _ := decimal.NewFromString(req.HourlyRate)

	wc, err := h.svc.CreateWorkCenter(c.Request.Context(), req.Code, req.Name, capacityDec, rateDec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": wc})
}

func (h *BOMHandler) GetWorkCenterDetails(c *gin.Context) {
	id := c.Param("id")
	wc, err := h.svc.GetWorkCenter(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work center not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wc})
}

func (h *BOMHandler) ListWorkCenters(c *gin.Context) {
	list, err := h.svc.ListWorkCenters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BOMHandler) UpdateWorkCenter(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Code          string `json:"code"`
		Name          string `json:"name"`
		Status        string `json:"status"`
		CapacityHours string `json:"capacity_hours"`
		HourlyRate    string `json:"hourly_rate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	capacityDec, _ := decimal.NewFromString(req.CapacityHours)
	rateDec, _ := decimal.NewFromString(req.HourlyRate)

	wc, err := h.svc.UpdateWorkCenter(c.Request.Context(), id, req.Code, req.Name, req.Status, capacityDec, rateDec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wc})
}

func (h *BOMHandler) DeleteWorkCenter(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteWorkCenter(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "work center deleted successfully"})
}

func (h *BOMHandler) ListRoutings(c *gin.Context) {
	list, err := h.svc.ListRoutings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BOMHandler) CreateRouting(c *gin.Context) {
	var req struct {
		BomID          string `json:"bom_id"`
		SequenceNumber int    `json:"sequence_number"`
		WorkCenterID   string `json:"work_center_id"`
		OperationName  string `json:"operation_name"`
		SetupTime      string `json:"setup_time"`
		RunTime        string `json:"run_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setupDec, _ := decimal.NewFromString(req.SetupTime)
	runDec, _ := decimal.NewFromString(req.RunTime)

	op, err := h.svc.CreateRoutingOperation(c.Request.Context(), req.BomID, req.SequenceNumber, req.WorkCenterID, req.OperationName, setupDec, runDec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": op})
}

func (h *BOMHandler) GetRoutingDetails(c *gin.Context) {
	id := c.Param("id")
	op, err := h.svc.GetRoutingByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "routing not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": op})
}

func (h *BOMHandler) UpdateRouting(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		BomID          string `json:"bom_id"`
		SequenceNumber int    `json:"sequence_number"`
		WorkCenterID   string `json:"work_center_id"`
		OperationName  string `json:"operation_name"`
		SetupTime      string `json:"setup_time"`
		RunTime        string `json:"run_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setupDec, _ := decimal.NewFromString(req.SetupTime)
	runDec, _ := decimal.NewFromString(req.RunTime)

	op, err := h.svc.UpdateRouting(c.Request.Context(), id, req.BomID, req.SequenceNumber, req.WorkCenterID, req.OperationName, setupDec, runDec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": op})
}

func (h *BOMHandler) DeleteRouting(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteRouting(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "routing deleted successfully"})
}
