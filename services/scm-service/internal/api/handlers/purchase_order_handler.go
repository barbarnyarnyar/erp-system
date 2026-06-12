package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PurchaseOrderHandler struct {
	svc *service.PurchaseOrderService
	response *utils.ResponseHelper
}

func NewPurchaseOrderHandler(svc *service.PurchaseOrderService, response *utils.ResponseHelper) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		svc: svc,
		response: response,
	}
}

func (h *PurchaseOrderHandler) GetPurchaseOrders(c *gin.Context) {
	list, err := h.svc.ListPurchaseOrders(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *PurchaseOrderHandler) CreatePurchaseOrder(c *gin.Context) {
	var req struct {
		SupplierID       string `json:"supplier_id"`
		ExpectedDelivery string `json:"expected_delivery"`
		Notes            string `json:"notes"`
		Lines            []struct {
			ProductID       string `json:"product_id"`
			QuantityOrdered int    `json:"quantity_ordered"`
			UnitPrice       string `json:"unit_price"`
			Description     string `json:"description"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	deliveryTime, err := time.Parse(time.RFC3339, req.ExpectedDelivery)
	if err != nil {
		deliveryTime = time.Now().AddDate(0, 0, 7) // default to 7 days out
	}

	linesInput := make([]service.POLineInput, 0, len(req.Lines))
	for _, l := range req.Lines {
		priceDec, err := decimal.NewFromString(l.UnitPrice)
		if err != nil {
			priceDec = decimal.Zero
		}
		linesInput = append(linesInput, service.POLineInput{
			ProductID:       l.ProductID,
			QuantityOrdered: l.QuantityOrdered,
			UnitPrice:       priceDec,
			Description:     l.Description,
		})
	}

	po, err := h.svc.CreatePurchaseOrder(c.Request.Context(), req.SupplierID, deliveryTime, req.Notes, linesInput)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": po})
}

func (h *PurchaseOrderHandler) GetPurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	po, err := h.svc.GetPurchaseOrder(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "purchase order not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": po})
}

func (h *PurchaseOrderHandler) UpdatePurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ExpectedDelivery string `json:"expected_delivery"`
		Status           string `json:"status"`
		Notes            string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	deliveryTime, err := time.Parse(time.RFC3339, req.ExpectedDelivery)
	if err != nil {
		deliveryTime = time.Now().AddDate(0, 0, 7)
	}

	po, err := h.svc.UpdatePurchaseOrder(c.Request.Context(), id, deliveryTime, req.Status, req.Notes)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": po})
}

func (h *PurchaseOrderHandler) DeletePurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeletePurchaseOrder(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "purchase order deleted successfully"})
}

func (h *PurchaseOrderHandler) SendPurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	po, err := h.svc.SendPurchaseOrder(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": po})
}

// Purchase Requisitions CRUD & Approval

func (h *PurchaseOrderHandler) GetPurchaseRequisitions(c *gin.Context) {
	list, err := h.svc.ListPurchaseRequisitions(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *PurchaseOrderHandler) CreatePurchaseRequisition(c *gin.Context) {
	var req struct {
		RequesterID string `json:"requester_id"`
		RequestDate string `json:"request_date"` // YYYY-MM-DD
		Notes       string `json:"notes"`
		Lines       []struct {
			ProductID          string `json:"product_id"`
			QuantityRequested  int    `json:"quantity_requested"`
			EstimatedUnitPrice string `json:"estimated_unit_price"`
		} `json:"lines"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	reqDate, err := time.Parse("2006-01-02", req.RequestDate)
	if err != nil {
		reqDate = time.Now()
	}

	linesInput := make([]service.RequisitionLineInput, 0, len(req.Lines))
	for _, l := range req.Lines {
		priceDec, err := decimal.NewFromString(l.EstimatedUnitPrice)
		if err != nil {
			priceDec = decimal.Zero
		}
		linesInput = append(linesInput, service.RequisitionLineInput{
			ProductID:          l.ProductID,
			QuantityRequested:  l.QuantityRequested,
			EstimatedUnitPrice: priceDec,
		})
	}

	pr, err := h.svc.CreatePurchaseRequisition(c.Request.Context(), req.RequesterID, reqDate, req.Notes, linesInput)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": pr})
}

func (h *PurchaseOrderHandler) GetPurchaseRequisition(c *gin.Context) {
	id := c.Param("id")
	pr, err := h.svc.GetPurchaseRequisition(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "purchase requisition not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *PurchaseOrderHandler) UpdatePurchaseRequisition(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		RequestDate string `json:"request_date"`
		Status      string `json:"status"`
		Notes       string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	reqDate, err := time.Parse("2006-01-02", req.RequestDate)
	if err != nil {
		reqDate = time.Now()
	}

	pr, err := h.svc.UpdatePurchaseRequisition(c.Request.Context(), id, reqDate, req.Status, req.Notes)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *PurchaseOrderHandler) DeletePurchaseRequisition(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeletePurchaseRequisition(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "purchase requisition deleted successfully"})
}

func (h *PurchaseOrderHandler) ApprovePurchaseRequisition(c *gin.Context) {
	id := c.Param("id")
	pr, err := h.svc.ApprovePurchaseRequisition(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *PurchaseOrderHandler) RejectPurchaseRequisition(c *gin.Context) {
	id := c.Param("id")
	pr, err := h.svc.RejectPurchaseRequisition(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pr})
}

func (h *PurchaseOrderHandler) GetPurchaseOrderLines(c *gin.Context) {
	poID := c.Param("id")
	lines, err := h.svc.ListPurchaseOrderLines(c.Request.Context(), poID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lines})
}

func (h *PurchaseOrderHandler) GetPurchaseRequisitionLines(c *gin.Context) {
	reqID := c.Param("id")
	lines, err := h.svc.ListPurchaseRequisitionLines(c.Request.Context(), reqID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lines})
}
