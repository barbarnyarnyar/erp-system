package handlers

import (
	"net/http"
	"time"

	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type SalesOpportunityHandler struct {
	oppSvc    *service.OpportunityService
	orderSvc  *service.SalesOrderService
	quoteSvc  *service.QuoteService
	ticketSvc *service.ServiceTicketService
	campSvc   *service.CampaignService
	plSvc     *service.PriceListService
}

func NewSalesOpportunityHandler(
	oppSvc *service.OpportunityService,
	orderSvc *service.SalesOrderService,
	quoteSvc *service.QuoteService,
	ticketSvc *service.ServiceTicketService,
	campSvc *service.CampaignService,
	plSvc *service.PriceListService,
) *SalesOpportunityHandler {
	return &SalesOpportunityHandler{
		oppSvc:    oppSvc,
		orderSvc:  orderSvc,
		quoteSvc:  quoteSvc,
		ticketSvc: ticketSvc,
		campSvc:   campSvc,
		plSvc:     plSvc,
	}
}

// Opportunities

type CreateOpportunityReq struct {
	CustomerID string          `json:"customer_id" binding:"required"`
	Title      string          `json:"title" binding:"required"`
	Value      decimal.Decimal `json:"value" binding:"required"`
	Stage      string          `json:"stage" binding:"required"`
}

func (h *SalesOpportunityHandler) CreateOpportunity(c *gin.Context) {
	var req CreateOpportunityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opp, err := h.oppSvc.CreateOpportunity(c.Request.Context(), req.CustomerID, req.Title, req.Value, req.Stage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, opp)
}

func (h *SalesOpportunityHandler) GetOpportunity(c *gin.Context) {
	id := c.Param("id")
	opp, err := h.oppSvc.GetOpportunity(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, opp)
}

func (h *SalesOpportunityHandler) ListOpportunities(c *gin.Context) {
	list, err := h.oppSvc.ListOpportunities(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

type UpdateOpportunityReq struct {
	Title       string          `json:"title" binding:"required"`
	Value       decimal.Decimal `json:"value" binding:"required"`
	Status      string          `json:"status" binding:"required"`
	Stage       string          `json:"stage" binding:"required"`
	Probability decimal.Decimal `json:"probability" binding:"required"`
}

func (h *SalesOpportunityHandler) UpdateOpportunity(c *gin.Context) {
	id := c.Param("id")
	var req UpdateOpportunityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opp, err := h.oppSvc.UpdateOpportunity(c.Request.Context(), id, req.Title, req.Value, req.Status, req.Stage, req.Probability)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, opp)
}

func (h *SalesOpportunityHandler) DeleteOpportunity(c *gin.Context) {
	id := c.Param("id")
	err := h.oppSvc.DeleteOpportunity(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Opportunity deleted successfully"})
}

// Sales Orders

type CreateSalesOrderReq struct {
	CustomerID string                        `json:"customer_id" binding:"required"`
	Items      []service.SalesOrderItemInput `json:"items" binding:"required"`
}

func (h *SalesOpportunityHandler) CreateSalesOrder(c *gin.Context) {
	var req CreateSalesOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderSvc.CreateSalesOrder(c.Request.Context(), req.CustomerID, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *SalesOpportunityHandler) GetSalesOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderSvc.GetSalesOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *SalesOpportunityHandler) ListSalesOrders(c *gin.Context) {
	list, err := h.orderSvc.ListSalesOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

type UpdateSalesOrderReq struct {
	Status string `json:"status" binding:"required"`
}

func (h *SalesOpportunityHandler) UpdateSalesOrder(c *gin.Context) {
	id := c.Param("id")
	var req UpdateSalesOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderSvc.UpdateSalesOrder(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *SalesOpportunityHandler) DeleteSalesOrder(c *gin.Context) {
	id := c.Param("id")
	err := h.orderSvc.DeleteSalesOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sales order deleted/cancelled successfully"})
}

// Quotes

type CreateQuoteReq struct {
	CustomerID string                       `json:"customer_id" binding:"required"`
	Title      string                       `json:"title" binding:"required"`
	ValidUntil time.Time                    `json:"valid_until" binding:"required"`
	Items      []service.QuoteLineItemInput `json:"items" binding:"required"`
}

func (h *SalesOpportunityHandler) CreateQuote(c *gin.Context) {
	var req CreateQuoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quote, err := h.quoteSvc.CreateQuote(c.Request.Context(), req.CustomerID, req.Title, req.ValidUntil, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, quote)
}

func (h *SalesOpportunityHandler) GetQuote(c *gin.Context) {
	id := c.Param("id")
	quote, err := h.quoteSvc.GetQuote(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quote)
}

func (h *SalesOpportunityHandler) ListQuotes(c *gin.Context) {
	list, err := h.quoteSvc.ListQuotes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

type UpdateQuoteReq struct {
	Status string `json:"status" binding:"required"`
}

func (h *SalesOpportunityHandler) UpdateQuote(c *gin.Context) {
	id := c.Param("id")
	var req UpdateQuoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quote, err := h.quoteSvc.UpdateQuote(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quote)
}

func (h *SalesOpportunityHandler) DeleteQuote(c *gin.Context) {
	id := c.Param("id")
	err := h.quoteSvc.DeleteQuote(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quote deleted successfully"})
}

func (h *SalesOpportunityHandler) SendQuote(c *gin.Context) {
	id := c.Param("id")
	quote, err := h.quoteSvc.SendQuote(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quote)
}

// Service Tickets

type CreateServiceTicketReq struct {
	CustomerID  string `json:"customer_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Priority    string `json:"priority" binding:"required"`
}

func (h *SalesOpportunityHandler) CreateServiceTicket(c *gin.Context) {
	var req CreateServiceTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.ticketSvc.CreateServiceTicket(c.Request.Context(), req.CustomerID, req.Title, req.Description, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (h *SalesOpportunityHandler) GetServiceTicket(c *gin.Context) {
	id := c.Param("id")
	ticket, err := h.ticketSvc.GetServiceTicket(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *SalesOpportunityHandler) ListServiceTickets(c *gin.Context) {
	list, err := h.ticketSvc.ListServiceTickets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

type UpdateServiceTicketReq struct {
	Status   string `json:"status" binding:"required"`
	Priority string `json:"priority" binding:"required"`
}

func (h *SalesOpportunityHandler) UpdateServiceTicket(c *gin.Context) {
	id := c.Param("id")
	var req UpdateServiceTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.ticketSvc.UpdateServiceTicket(c.Request.Context(), id, req.Status, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *SalesOpportunityHandler) DeleteServiceTicket(c *gin.Context) {
	id := c.Param("id")
	err := h.ticketSvc.DeleteServiceTicket(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service ticket deleted successfully"})
}

// Campaigns

type CreateCampaignReq struct {
	Name   string          `json:"name" binding:"required"`
	Type   string          `json:"type" binding:"required"`
	Budget decimal.Decimal `json:"budget" binding:"required"`
}

func (h *SalesOpportunityHandler) CreateCampaign(c *gin.Context) {
	var req CreateCampaignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	camp, err := h.campSvc.CreateCampaign(c.Request.Context(), req.Name, req.Type, req.Budget)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, camp)
}

func (h *SalesOpportunityHandler) GetCampaign(c *gin.Context) {
	id := c.Param("id")
	camp, err := h.campSvc.GetCampaign(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, camp)
}

func (h *SalesOpportunityHandler) ListCampaigns(c *gin.Context) {
	list, err := h.campSvc.ListCampaigns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

type UpdateCampaignReq struct {
	Status string          `json:"status" binding:"required"`
	Budget decimal.Decimal `json:"budget" binding:"required"`
}

func (h *SalesOpportunityHandler) UpdateCampaign(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCampaignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	camp, err := h.campSvc.UpdateCampaign(c.Request.Context(), id, req.Status, req.Budget)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, camp)
}

func (h *SalesOpportunityHandler) DeleteCampaign(c *gin.Context) {
	id := c.Param("id")
	err := h.campSvc.DeleteCampaign(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campaign deleted successfully"})
}

// Price Lists

type CreatePriceListReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *SalesOpportunityHandler) CreatePriceList(c *gin.Context) {
	var req CreatePriceListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pl, err := h.plSvc.CreatePriceList(c.Request.Context(), req.Name, req.Description, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pl)
}

func (h *SalesOpportunityHandler) GetPriceList(c *gin.Context) {
	id := c.Param("id")
	pl, err := h.plSvc.GetPriceList(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pl)
}

func (h *SalesOpportunityHandler) ListPriceLists(c *gin.Context) {
	list, err := h.plSvc.ListPriceLists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

type UpdatePriceListReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *SalesOpportunityHandler) UpdatePriceList(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePriceListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pl, err := h.plSvc.UpdatePriceList(c.Request.Context(), id, req.Name, req.Description, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pl)
}

func (h *SalesOpportunityHandler) DeletePriceList(c *gin.Context) {
	id := c.Param("id")
	err := h.plSvc.DeletePriceList(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Price list deleted successfully"})
}
