package handlers

import (
	"net/http"

	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type CustomerLeadHandler struct {
	custSvc *service.CustomerService
	leadSvc *service.LeadService
}

func NewCustomerLeadHandler(custSvc *service.CustomerService, leadSvc *service.LeadService) *CustomerLeadHandler {
	return &CustomerLeadHandler{
		custSvc: custSvc,
		leadSvc: leadSvc,
	}
}

// Customers

type CreateCustomerReq struct {
	CompanyName      string `json:"company_name" binding:"required"`
	ContactName      string `json:"contact_name" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Phone            string `json:"phone"`
	Category         string `json:"category"`
	ParentCustomerID string `json:"parent_customer_id"`
}

func (h *CustomerLeadHandler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cust, err := h.custSvc.CreateCustomer(c.Request.Context(), req.CompanyName, req.ContactName, req.Email, req.Phone, req.Category, req.ParentCustomerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cust)
}

func (h *CustomerLeadHandler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	cust, err := h.custSvc.GetCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cust)
}

func (h *CustomerLeadHandler) ListCustomers(c *gin.Context) {
	list, err := h.custSvc.ListCustomers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

type UpdateCustomerReq struct {
	CompanyName string `json:"company_name" binding:"required"`
	ContactName string `json:"contact_name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Phone       string `json:"phone"`
	Status      string `json:"status" binding:"required"`
	Category    string `json:"category"`
}

func (h *CustomerLeadHandler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cust, err := h.custSvc.UpdateCustomer(c.Request.Context(), id, req.CompanyName, req.ContactName, req.Email, req.Phone, req.Status, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cust)
}

func (h *CustomerLeadHandler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	err := h.custSvc.DeleteCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// Leads

type CreateLeadReq struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Company   string `json:"company" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Phone     string `json:"phone"`
	Source    string `json:"source"`
}

func (h *CustomerLeadHandler) CreateLead(c *gin.Context) {
	var req CreateLeadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lead, err := h.leadSvc.CreateLead(c.Request.Context(), req.FirstName, req.LastName, req.Company, req.Email, req.Phone, req.Source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lead)
}

func (h *CustomerLeadHandler) GetLead(c *gin.Context) {
	id := c.Param("id")
	lead, err := h.leadSvc.GetLead(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lead)
}

func (h *CustomerLeadHandler) ListLeads(c *gin.Context) {
	list, err := h.leadSvc.ListLeads(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

type UpdateLeadReq struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Company   string `json:"company" binding:"required"`
	Status    string `json:"status" binding:"required"`
	Score     int    `json:"score" binding:"required"`
}

func (h *CustomerLeadHandler) UpdateLead(c *gin.Context) {
	id := c.Param("id")
	var req UpdateLeadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lead, err := h.leadSvc.UpdateLead(c.Request.Context(), id, req.FirstName, req.LastName, req.Company, req.Status, req.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lead)
}

func (h *CustomerLeadHandler) DeleteLead(c *gin.Context) {
	id := c.Param("id")
	err := h.leadSvc.DeleteLead(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead deleted successfully"})
}

func (h *CustomerLeadHandler) ConvertLead(c *gin.Context) {
	id := c.Param("id")
	opp, err := h.leadSvc.ConvertLead(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, opp)
}
