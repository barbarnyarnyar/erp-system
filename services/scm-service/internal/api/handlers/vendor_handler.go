package handlers

import (
	"erp-system/shared/utils"
	"net/http"
	"time"

	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type VendorHandler struct {
	svc *service.SupplierManagementService
	response *utils.ResponseHelper
}

func NewVendorHandler(svc *service.SupplierManagementService, response *utils.ResponseHelper) *VendorHandler {
	return &VendorHandler{
		svc: svc,
		response: response,
	}
}

func (h *VendorHandler) GetVendors(c *gin.Context) {
	list, err := h.svc.ListSuppliers(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *VendorHandler) CreateVendor(c *gin.Context) {
	var req struct {
		SupplierCode string `json:"supplier_code"`
		SupplierName string `json:"supplier_name"`
		ContactName  string `json:"contact_name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	v, err := h.svc.CreateSupplier(c.Request.Context(), req.SupplierCode, req.SupplierName, req.ContactName, req.Email, req.Phone)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": v})
}

func (h *VendorHandler) GetVendor(c *gin.Context) {
	id := c.Param("id")
	v, err := h.svc.GetSupplier(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "vendor not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": v})
}

func (h *VendorHandler) UpdateVendor(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		SupplierCode string `json:"supplier_code"`
		SupplierName string `json:"supplier_name"`
		ContactName  string `json:"contact_name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		IsActive     bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	v, err := h.svc.UpdateSupplier(c.Request.Context(), id, req.SupplierCode, req.SupplierName, req.ContactName, req.Email, req.Phone, req.IsActive)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": v})
}

func (h *VendorHandler) DeleteVendor(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteSupplier(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vendor deleted successfully"})
}

// Vendor Contracts CRUD

func (h *VendorHandler) GetContracts(c *gin.Context) {
	list, err := h.svc.ListContracts(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *VendorHandler) CreateContract(c *gin.Context) {
	var req struct {
		ContractNumber string `json:"contract_number"`
		SupplierID     string `json:"supplier_id"`
		StartDate      string `json:"start_date"` // YYYY-MM-DD
		EndDate        string `json:"end_date"`   // YYYY-MM-DD
		Terms          string `json:"terms"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		h.response.BadRequest(c, "invalid start_date format, use YYYY-MM-DD")
		return
	}

	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		h.response.BadRequest(c, "invalid end_date format, use YYYY-MM-DD")
		return
	}

	vc, err := h.svc.CreateContract(c.Request.Context(), req.ContractNumber, req.SupplierID, start, end, req.Terms)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": vc})
}

func (h *VendorHandler) GetContract(c *gin.Context) {
	id := c.Param("id")
	vc, err := h.svc.GetContract(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "contract not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vc})
}

func (h *VendorHandler) UpdateContract(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ContractNumber string `json:"contract_number"`
		SupplierID     string `json:"supplier_id"`
		StartDate      string `json:"start_date"` // YYYY-MM-DD
		EndDate        string `json:"end_date"`   // YYYY-MM-DD
		Terms          string `json:"terms"`
		Status         string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		h.response.BadRequest(c, "invalid start_date format, use YYYY-MM-DD")
		return
	}

	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		h.response.BadRequest(c, "invalid end_date format, use YYYY-MM-DD")
		return
	}

	vc, err := h.svc.UpdateContract(c.Request.Context(), id, req.ContractNumber, req.SupplierID, start, end, req.Terms, req.Status)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vc})
}

func (h *VendorHandler) DeleteContract(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteContract(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "contract deleted successfully"})
}
