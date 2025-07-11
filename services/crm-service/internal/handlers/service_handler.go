package handlers

import (
	"strconv"
	"time"

	"crm-service/internal/models"
	"crm-service/internal/repositories"
	"crm-service/internal/utils"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	serviceRepo        *repositories.ServiceRepository
	serviceRequestRepo *repositories.ServiceRequestRepository
	response           *utils.ResponseHelper
}

func NewServiceHandler(serviceRepo *repositories.ServiceRepository, serviceRequestRepo *repositories.ServiceRequestRepository) *ServiceHandler {
	return &ServiceHandler{
		serviceRepo:        serviceRepo,
		serviceRequestRepo: serviceRequestRepo,
		response:           utils.NewResponseHelper("crm-service"),
	}
}

// CreateService creates a new service
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req models.ServiceCreateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	service := &models.Service{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       req.Price,
		Currency:    req.Currency,
		Duration:    req.Duration,
	}

	if err := h.serviceRepo.Create(service); err != nil {
		h.response.InternalServerError(c, "Failed to create service", err)
		return
	}

	h.response.Success(c, "Service created successfully", service)
}

// GetService retrieves a service by ID
func (h *ServiceHandler) GetService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid service ID")
		return
	}

	service, err := h.serviceRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Service not found")
		return
	}

	h.response.Success(c, "Service retrieved successfully", service)
}

// GetAllServices retrieves all services with pagination
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	services, total, err := h.serviceRepo.GetAll(page, limit)
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve services", err)
		return
	}

	h.response.Success(c, "Services retrieved successfully", gin.H{
		"services": services,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetActiveServices retrieves all active services
func (h *ServiceHandler) GetActiveServices(c *gin.Context) {
	services, err := h.serviceRepo.GetActiveServices()
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve active services", err)
		return
	}

	h.response.Success(c, "Active services retrieved successfully", services)
}

// UpdateService updates a service
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid service ID")
		return
	}

	var req models.ServiceCreateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	service, err := h.serviceRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Service not found")
		return
	}

	// Update fields
	service.Name = req.Name
	service.Description = req.Description
	service.Category = req.Category
	service.Price = req.Price
	service.Currency = req.Currency
	service.Duration = req.Duration

	if err := h.serviceRepo.Update(service); err != nil {
		h.response.InternalServerError(c, "Failed to update service", err)
		return
	}

	h.response.Success(c, "Service updated successfully", service)
}

// DeleteService deletes a service
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid service ID")
		return
	}

	if err := h.serviceRepo.Delete(uint(id)); err != nil {
		h.response.InternalServerError(c, "Failed to delete service", err)
		return
	}

	h.response.Success(c, "Service deleted successfully", nil)
}

// CreateServiceRequest creates a new service request
func (h *ServiceHandler) CreateServiceRequest(c *gin.Context) {
	var req models.ServiceRequestCreateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	serviceRequest := &models.ServiceRequest{
		ContactID:   req.ContactID,
		ServiceID:   req.ServiceID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		RequestedAt: time.Now(),
		ScheduledAt: req.ScheduledAt,
		Notes:       req.Notes,
	}

	if err := h.serviceRequestRepo.Create(serviceRequest); err != nil {
		h.response.InternalServerError(c, "Failed to create service request", err)
		return
	}

	h.response.Success(c, "Service request created successfully", serviceRequest)
}

// GetServiceRequest retrieves a service request by ID
func (h *ServiceHandler) GetServiceRequest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid service request ID")
		return
	}

	serviceRequest, err := h.serviceRequestRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Service request not found")
		return
	}

	h.response.Success(c, "Service request retrieved successfully", serviceRequest)
}

// GetAllServiceRequests retrieves all service requests with pagination
func (h *ServiceHandler) GetAllServiceRequests(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	serviceRequests, total, err := h.serviceRequestRepo.GetAll(page, limit)
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve service requests", err)
		return
	}

	h.response.Success(c, "Service requests retrieved successfully", gin.H{
		"service_requests": serviceRequests,
		"total":            total,
		"page":             page,
		"limit":            limit,
	})
}

// UpdateServiceRequest updates a service request
func (h *ServiceHandler) UpdateServiceRequest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid service request ID")
		return
	}

	var req models.ServiceRequestCreateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	serviceRequest, err := h.serviceRequestRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Service request not found")
		return
	}

	// Update fields
	serviceRequest.ContactID = req.ContactID
	serviceRequest.ServiceID = req.ServiceID
	serviceRequest.Title = req.Title
	serviceRequest.Description = req.Description
	serviceRequest.Priority = req.Priority
	serviceRequest.ScheduledAt = req.ScheduledAt
	serviceRequest.Notes = req.Notes

	if err := h.serviceRequestRepo.Update(serviceRequest); err != nil {
		h.response.InternalServerError(c, "Failed to update service request", err)
		return
	}

	h.response.Success(c, "Service request updated successfully", serviceRequest)
}

// GetServiceRequestsByContact retrieves service requests for a specific contact
func (h *ServiceHandler) GetServiceRequestsByContact(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("contactId"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid contact ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	serviceRequests, total, err := h.serviceRequestRepo.GetByContactID(uint(contactID), page, limit)
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve service requests", err)
		return
	}

	h.response.Success(c, "Service requests retrieved successfully", gin.H{
		"service_requests": serviceRequests,
		"total":            total,
		"page":             page,
		"limit":            limit,
		"contact_id":       contactID,
	})
}
