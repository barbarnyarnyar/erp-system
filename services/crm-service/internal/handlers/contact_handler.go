package handlers

import (
	"strconv"

	"crm-service/internal/models"
	"crm-service/internal/repositories"
	"crm-service/internal/utils"

	"github.com/gin-gonic/gin"
)

type ContactHandler struct {
	contactRepo *repositories.ContactRepository
	response    *utils.ResponseHelper
}

func NewContactHandler(contactRepo *repositories.ContactRepository) *ContactHandler {
	return &ContactHandler{
		contactRepo: contactRepo,
		response:    utils.NewResponseHelper("crm-service"),
	}
}

// CreateContact creates a new contact
func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req models.ContactCreateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	contact := &models.Contact{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Phone:      req.Phone,
		Company:    req.Company,
		Position:   req.Position,
		Department: req.Department,
		Address:    req.Address,
		City:       req.City,
		State:      req.State,
		Country:    req.Country,
		PostalCode: req.PostalCode,
		LeadSource: req.LeadSource,
		Notes:      req.Notes,
		Tags:       req.Tags,
	}

	if err := h.contactRepo.Create(contact); err != nil {
		h.response.InternalServerError(c, "Failed to create contact", err)
		return
	}

	h.response.Success(c, "Contact created successfully", contact)
}

// GetContact retrieves a contact by ID
func (h *ContactHandler) GetContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid contact ID")
		return
	}

	contact, err := h.contactRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Contact not found")
		return
	}

	h.response.Success(c, "Contact retrieved successfully", contact)
}

// GetAllContacts retrieves all contacts with pagination
func (h *ContactHandler) GetAllContacts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	contacts, total, err := h.contactRepo.GetAll(page, limit)
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve contacts", err)
		return
	}

	h.response.Success(c, "Contacts retrieved successfully", gin.H{
		"contacts": contacts,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// UpdateContact updates a contact
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid contact ID")
		return
	}

	var req models.ContactUpdateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	contact, err := h.contactRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Contact not found")
		return
	}

	// Update fields if provided
	if req.FirstName != "" {
		contact.FirstName = req.FirstName
	}
	if req.LastName != "" {
		contact.LastName = req.LastName
	}
	if req.Email != "" {
		contact.Email = req.Email
	}
	if req.Phone != "" {
		contact.Phone = req.Phone
	}
	if req.Company != "" {
		contact.Company = req.Company
	}
	if req.Position != "" {
		contact.Position = req.Position
	}
	if req.Department != "" {
		contact.Department = req.Department
	}
	if req.Address != "" {
		contact.Address = req.Address
	}
	if req.City != "" {
		contact.City = req.City
	}
	if req.State != "" {
		contact.State = req.State
	}
	if req.Country != "" {
		contact.Country = req.Country
	}
	if req.PostalCode != "" {
		contact.PostalCode = req.PostalCode
	}
	if req.LeadSource != "" {
		contact.LeadSource = req.LeadSource
	}
	if req.Status != "" {
		contact.Status = req.Status
	}
	if req.Notes != "" {
		contact.Notes = req.Notes
	}
	if req.Tags != "" {
		contact.Tags = req.Tags
	}

	if err := h.contactRepo.Update(contact); err != nil {
		h.response.InternalServerError(c, "Failed to update contact", err)
		return
	}

	h.response.Success(c, "Contact updated successfully", contact)
}

// DeleteContact deletes a contact
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid contact ID")
		return
	}

	if err := h.contactRepo.Delete(uint(id)); err != nil {
		h.response.InternalServerError(c, "Failed to delete contact", err)
		return
	}

	h.response.Success(c, "Contact deleted successfully", nil)
}

// SearchContacts searches contacts by query
func (h *ContactHandler) SearchContacts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		h.response.BadRequest(c, "Search query is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	contacts, total, err := h.contactRepo.Search(query, page, limit)
	if err != nil {
		h.response.InternalServerError(c, "Failed to search contacts", err)
		return
	}

	h.response.Success(c, "Contacts search completed", gin.H{
		"contacts": contacts,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"query":    query,
	})
}
