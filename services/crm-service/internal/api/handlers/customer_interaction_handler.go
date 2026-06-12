package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type CustomerInteractionHandler struct {
	interactionSvc *service.CustomerInteractionService
	response *utils.ResponseHelper
}

func NewCustomerInteractionHandler(interactionSvc *service.CustomerInteractionService, response *utils.ResponseHelper) *CustomerInteractionHandler {
	return &CustomerInteractionHandler{
		interactionSvc: interactionSvc,
		response: response,
	}
}

type CreateCustomerInteractionReq struct {
	CustomerID      string    `json:"customer_id" binding:"required"`
	Type            string    `json:"type" binding:"required"`
	Subject         string    `json:"subject"`
	Description     string    `json:"description"`
	InteractionDate time.Time `json:"interaction_date"`
	CreatedBy       string    `json:"created_by"`
}

func (h *CustomerInteractionHandler) CreateCustomerInteraction(c *gin.Context) {
	var req CreateCustomerInteractionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	if req.InteractionDate.IsZero() {
		req.InteractionDate = time.Now()
	}

	ci, err := h.interactionSvc.CreateCustomerInteraction(
		c.Request.Context(),
		req.CustomerID,
		req.Type,
		req.Subject,
		req.Description,
		req.InteractionDate,
		req.CreatedBy,
	)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, ci)
}

func (h *CustomerInteractionHandler) GetCustomerInteraction(c *gin.Context) {
	id := c.Param("id")
	ci, err := h.interactionSvc.GetCustomerInteraction(c.Request.Context(), id)
	if err != nil {
		h.response.NotFoundErr(c, err)
		return
	}

	c.JSON(http.StatusOK, ci)
}

func (h *CustomerInteractionHandler) ListCustomerInteractions(c *gin.Context) {
	customerID := c.Query("customer_id")
	if customerID == "" {
		h.response.BadRequest(c, "customer_id query parameter is required")
		return
	}

	list, err := h.interactionSvc.ListCustomerInteractions(c.Request.Context(), customerID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *CustomerInteractionHandler) DeleteCustomerInteraction(c *gin.Context) {
	id := c.Param("id")
	err := h.interactionSvc.DeleteCustomerInteraction(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
