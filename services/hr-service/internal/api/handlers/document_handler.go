package handlers

import (
	"erp-system/shared/utils"
	"net/http"

	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	svc *service.EmployeeDocumentService
	response *utils.ResponseHelper
}

func NewDocumentHandler(svc *service.EmployeeDocumentService, response *utils.ResponseHelper) *DocumentHandler {
	return &DocumentHandler{
		svc: svc,
		response: response,
	}
}

func (h *DocumentHandler) GetEmployeeDocuments(c *gin.Context) {
	empID := c.Param("id")
	list, err := h.svc.ListDocuments(c.Request.Context(), empID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *DocumentHandler) UploadEmployeeDocument(c *gin.Context) {
	empID := c.Param("id")
	var req struct {
		DocType  string `json:"doc_type"`
		FileName string `json:"file_name"`
		FileUrl  string `json:"file_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	doc, err := h.svc.UploadDocument(c.Request.Context(), empID, req.DocType, req.FileName, req.FileUrl)
	if err != nil {
		h.response.BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": doc})
}

func (h *DocumentHandler) DeleteEmployeeDocument(c *gin.Context) {
	docID := c.Param("docId")
	err := h.svc.DeleteDocument(c.Request.Context(), docID)
	if err != nil {
		h.response.InternalErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
}
