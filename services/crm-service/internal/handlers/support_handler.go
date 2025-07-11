package handlers

import (
	"net/http"
	"strconv"
	"time"

	"crm-service/internal/models"
	"crm-service/internal/repositories"
	"crm-service/internal/utils"

	"github.com/gin-gonic/gin"
)

type SupportHandler struct {
	ticketRepo   *repositories.SupportTicketRepository
	responseRepo *repositories.SupportResponseRepository
	response     *utils.ResponseHelper
}

func NewSupportHandler(ticketRepo *repositories.SupportTicketRepository, responseRepo *repositories.SupportResponseRepository) *SupportHandler {
	return &SupportHandler{
		ticketRepo:   ticketRepo,
		responseRepo: responseRepo,
		response:     utils.NewResponseHelper("crm-service"),
	}
}

// generateTicketNumber generates a unique ticket number
func generateTicketNumber() string {
	return "TKT-" + strconv.FormatInt(time.Now().Unix(), 10)
}

// CreateTicket creates a new support ticket
func (h *SupportHandler) CreateTicket(c *gin.Context) {
	var req models.SupportTicketCreateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	ticket := &models.SupportTicket{
		ContactID:    req.ContactID,
		TicketNumber: generateTicketNumber(),
		Subject:      req.Subject,
		Description:  req.Description,
		Category:     req.Category,
		Priority:     req.Priority,
		Status:       "open",
	}

	if err := h.ticketRepo.Create(ticket); err != nil {
		h.response.InternalServerError(c, "Failed to create support ticket", err)
		return
	}

	h.response.Success(c, "Support ticket created successfully", ticket)
}

// GetTicket retrieves a support ticket by ID
func (h *SupportHandler) GetTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid ticket ID")
		return
	}

	ticket, err := h.ticketRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Support ticket not found")
		return
	}

	h.response.Success(c, "Support ticket retrieved successfully", ticket)
}

// GetTicketByNumber retrieves a support ticket by ticket number
func (h *SupportHandler) GetTicketByNumber(c *gin.Context) {
	ticketNumber := c.Param("ticketNumber")
	if ticketNumber == "" {
		h.response.BadRequest(c, "Ticket number is required")
		return
	}

	ticket, err := h.ticketRepo.GetByTicketNumber(ticketNumber)
	if err != nil {
		h.response.NotFound(c, "Support ticket not found")
		return
	}

	h.response.Success(c, "Support ticket retrieved successfully", ticket)
}

// GetAllTickets retrieves all support tickets with pagination
func (h *SupportHandler) GetAllTickets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	tickets, total, err := h.ticketRepo.GetAll(page, limit)
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve support tickets", err)
		return
	}

	h.response.Success(c, "Support tickets retrieved successfully", gin.H{
		"tickets": tickets,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetTicketsByStatus retrieves support tickets by status
func (h *SupportHandler) GetTicketsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		h.response.BadRequest(c, "Status is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	tickets, total, err := h.ticketRepo.GetByStatus(status, page, limit)
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve support tickets", err)
		return
	}

	h.response.Success(c, "Support tickets retrieved successfully", gin.H{
		"tickets": tickets,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"status":  status,
	})
}

// UpdateTicket updates a support ticket
func (h *SupportHandler) UpdateTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid ticket ID")
		return
	}

	var req models.SupportTicketCreateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	ticket, err := h.ticketRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Support ticket not found")
		return
	}

	// Update fields
	ticket.ContactID = req.ContactID
	ticket.Subject = req.Subject
	ticket.Description = req.Description
	ticket.Category = req.Category
	ticket.Priority = req.Priority

	if err := h.ticketRepo.Update(ticket); err != nil {
		h.response.InternalServerError(c, "Failed to update support ticket", err)
		return
	}

	h.response.Success(c, "Support ticket updated successfully", ticket)
}

// CloseTicket closes a support ticket
func (h *SupportHandler) CloseTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid ticket ID")
		return
	}

	ticket, err := h.ticketRepo.GetByID(uint(id))
	if err != nil {
		h.response.NotFound(c, "Support ticket not found")
		return
	}

	now := time.Now()
	ticket.Status = "closed"
	ticket.ClosedAt = &now

	if err := h.ticketRepo.Update(ticket); err != nil {
		h.response.InternalServerError(c, "Failed to close support ticket", err)
		return
	}

	h.response.Success(c, "Support ticket closed successfully", ticket)
}

// GetTicketsByContact retrieves support tickets for a specific contact
func (h *SupportHandler) GetTicketsByContact(c *gin.Context) {
	contactID, err := strconv.ParseUint(c.Param("contactId"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid contact ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	tickets, total, err := h.ticketRepo.GetByContactID(uint(contactID), page, limit)
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve support tickets", err)
		return
	}

	h.response.Success(c, "Support tickets retrieved successfully", gin.H{
		"tickets":    tickets,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"contact_id": contactID,
	})
}

// CreateResponse creates a new response to a support ticket
func (h *SupportHandler) CreateResponse(c *gin.Context) {
	var req models.SupportResponseCreateRequest
	if !h.response.ValidateJSON(c, &req) {
		return
	}

	response := &models.SupportResponse{
		TicketID:   req.TicketID,
		Response:   req.Response,
		IsInternal: req.IsInternal,
	}

	if err := h.responseRepo.Create(response); err != nil {
		h.response.InternalServerError(c, "Failed to create support response", err)
		return
	}

	h.response.Success(c, "Support response created successfully", response)
}

// GetResponsesByTicket retrieves all responses for a specific ticket
func (h *SupportHandler) GetResponsesByTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("ticketId"), 10, 32)
	if err != nil {
		h.response.BadRequest(c, "Invalid ticket ID")
		return
	}

	responses, err := h.responseRepo.GetByTicketID(uint(ticketID))
	if err != nil {
		h.response.InternalServerError(c, "Failed to retrieve support responses", err)
		return
	}

	h.response.Success(c, "Support responses retrieved successfully", responses)
}