package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type ProductHandler struct {
	svc *service.ProductManagementService
	response *utils.ResponseHelper
}

func NewProductHandler(svc *service.ProductManagementService, response *utils.ResponseHelper) *ProductHandler {
	return &ProductHandler{
		svc: svc,
		response: response,
	}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	list, err := h.svc.ListProducts(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req struct {
		ProductCode   string  `json:"product_code"`
		ProductName   string  `json:"product_name"`
		Description   string  `json:"description"`
		ProductType   string  `json:"product_type"`
		CategoryID    *string `json:"category_id"`
		UnitOfMeasure string  `json:"unit_of_measure"`
		StandardCost  string  `json:"standard_cost"`
		ListPrice     string  `json:"list_price"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	costDec, err := decimal.NewFromString(req.StandardCost)
	if err != nil {
		costDec = decimal.Zero
	}
	priceDec, err := decimal.NewFromString(req.ListPrice)
	if err != nil {
		priceDec = decimal.Zero
	}

	p, err := h.svc.CreateProduct(c.Request.Context(), req.ProductCode, req.ProductName, req.Description, req.ProductType, req.UnitOfMeasure, costDec, priceDec, req.CategoryID)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": p})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.GetProduct(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "product not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ProductCode   string  `json:"product_code"`
		ProductName   string  `json:"product_name"`
		Description   string  `json:"description"`
		ProductType   string  `json:"product_type"`
		CategoryID    *string `json:"category_id"`
		UnitOfMeasure string  `json:"unit_of_measure"`
		StandardCost  string  `json:"standard_cost"`
		ListPrice     string  `json:"list_price"`
		IsActive      bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	costDec, err := decimal.NewFromString(req.StandardCost)
	if err != nil {
		costDec = decimal.Zero
	}
	priceDec, err := decimal.NewFromString(req.ListPrice)
	if err != nil {
		priceDec = decimal.Zero
	}

	p, err := h.svc.UpdateProduct(c.Request.Context(), id, req.ProductCode, req.ProductName, req.Description, req.ProductType, req.UnitOfMeasure, costDec, priceDec, req.IsActive, req.CategoryID)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}

// Product Categories CRUD

func (h *ProductHandler) GetCategories(c *gin.Context) {
	list, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	pc, err := h.svc.CreateCategory(c.Request.Context(), req.Code, req.Name, req.Description)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": pc})
}

func (h *ProductHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")
	pc, err := h.svc.GetCategory(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "category not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pc})
}

func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	pc, err := h.svc.UpdateCategory(c.Request.Context(), id, req.Code, req.Name, req.Description)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pc})
}

func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}

func (h *ProductHandler) GetLocations(c *gin.Context) {
	list, err := h.svc.ListLocations(c.Request.Context())
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *ProductHandler) CreateLocation(c *gin.Context) {
	var req struct {
		LocationCode string `json:"location_code"`
		LocationName string `json:"location_name"`
		LocationType string `json:"location_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	loc, err := h.svc.CreateLocation(c.Request.Context(), req.LocationCode, req.LocationName, req.LocationType)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": loc})
}

func (h *ProductHandler) GetLocation(c *gin.Context) {
	id := c.Param("id")
	loc, err := h.svc.GetLocation(c.Request.Context(), id)
	if err != nil {
		h.response.NotFound(c, "location not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": loc})
}

func (h *ProductHandler) UpdateLocation(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		LocationCode string `json:"location_code"`
		LocationName string `json:"location_name"`
		LocationType string `json:"location_type"`
		IsActive     bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	loc, err := h.svc.UpdateLocation(c.Request.Context(), id, req.LocationCode, req.LocationName, req.LocationType, req.IsActive)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": loc})
}

func (h *ProductHandler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteLocation(c.Request.Context(), id)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "location deleted successfully"})
}
