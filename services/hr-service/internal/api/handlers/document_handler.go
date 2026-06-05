package handlers

import (
	"net/http"

	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	svc *service.EmployeeDocumentService
}

func NewDocumentHandler(svc *service.EmployeeDocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func (h *DocumentHandler) GetEmployeeDocuments(c *gin.Context) {
	empID := c.Param("id")
	list, err := h.svc.ListDocuments(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := h.svc.UploadDocument(c.Request.Context(), empID, req.DocType, req.FileName, req.FileUrl)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": doc})
}

func (h *DocumentHandler) DeleteEmployeeDocument(c *gin.Context) {
	docID := c.Param("docId")
	err := h.svc.DeleteDocument(c.Request.Context(), docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
}
