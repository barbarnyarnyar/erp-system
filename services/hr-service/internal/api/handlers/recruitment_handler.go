package handlers

import (
	"net/http"

	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/gin-gonic/gin"
)

type RecruitmentHandler struct {
	svc *service.RecruitmentService
}

func NewRecruitmentHandler(svc *service.RecruitmentService) *RecruitmentHandler {
	return &RecruitmentHandler{svc: svc}
}

func (h *RecruitmentHandler) GetJobPostings(c *gin.Context) {
	list, err := h.svc.ListJobPostings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *RecruitmentHandler) CreateJobPosting(c *gin.Context) {
	var req struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		DepartmentID string `json:"department_id"`
		Location     string `json:"location"`
		SalaryRange  string `json:"salary_range"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jp, err := h.svc.CreateJobPosting(c.Request.Context(), req.Title, req.Description, req.DepartmentID, req.Location, req.SalaryRange)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": jp})
}

func (h *RecruitmentHandler) GetJobPosting(c *gin.Context) {
	id := c.Param("id")
	jp, err := h.svc.GetJobPosting(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job posting not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": jp})
}

func (h *RecruitmentHandler) UpdateJobPosting(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		SalaryRange string `json:"salary_range"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jp, err := h.svc.UpdateJobPosting(c.Request.Context(), id, req.Title, req.Description, req.Location, req.SalaryRange, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": jp})
}

func (h *RecruitmentHandler) DeleteJobPosting(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteJobPosting(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "job posting deleted successfully"})
}

func (h *RecruitmentHandler) GetApplications(c *gin.Context) {
	list, err := h.svc.ListApplications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *RecruitmentHandler) CreateApplication(c *gin.Context) {
	var req struct {
		JobPostingID  string `json:"job_posting_id"`
		ApplicantName string `json:"applicant_name"`
		Email         string `json:"email"`
		Phone         string `json:"phone"`
		ResumeURL     string `json:"resume_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ja, err := h.svc.CreateApplication(c.Request.Context(), req.JobPostingID, req.ApplicantName, req.Email, req.Phone, req.ResumeURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ja})
}

func (h *RecruitmentHandler) GetApplication(c *gin.Context) {
	id := c.Param("id")
	ja, err := h.svc.GetApplication(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job application not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ja})
}

func (h *RecruitmentHandler) UpdateApplication(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ja, err := h.svc.UpdateApplication(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ja})
}
