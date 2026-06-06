package routes

import (
	"github.com/erp-system/pm-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupPMRoutes(r *gin.Engine, handler *handlers.ProjectHandler) {
	v1 := r.Group("/api/v1")
	{
		// Portfolios
		v1.GET("/projects/portfolios", handler.ListPortfolios)
		v1.POST("/projects/portfolios", handler.CreatePortfolio)
		v1.GET("/projects/portfolios/:id", handler.GetPortfolio)
		v1.GET("/projects/portfolios/:id/summary", handler.GetPortfolioSummary)

		// Projects
		v1.GET("/projects", handler.ListProjects)
		v1.POST("/projects", handler.CreateProject)
		v1.GET("/projects/:id", handler.GetProject)
		v1.PUT("/projects/:id/status", handler.UpdateProjectStatus)

		// Tasks
		v1.GET("/projects/:id/tasks", handler.ListTasks)
		v1.POST("/projects/:id/tasks", handler.CreateTask)
		v1.PUT("/projects/tasks/:task_id/progress", handler.UpdateTaskProgress)
		v1.PUT("/projects/tasks/:task_id/assign", handler.AssignTask)
		v1.POST("/projects/tasks/:task_id/dependencies", handler.AddTaskDependency)

		// Resource Allocations
		v1.GET("/projects/:id/allocations", handler.ListAllocations)
		v1.POST("/projects/:id/allocations", handler.AllocateResource)

		// Time Tracking
		v1.GET("/projects/:id/time", handler.ListTimeEntries)
		v1.POST("/projects/:id/time", handler.LogTime)
		v1.PUT("/projects/time/:time_id/approve", handler.ApproveTime)

		// Expense Management
		v1.GET("/projects/:id/expenses", handler.ListExpenses)
		v1.POST("/projects/:id/expenses", handler.LogExpense)
		v1.PUT("/projects/expenses/:expense_id/approve", handler.ApproveExpense)

		// Documents
		v1.GET("/projects/:id/documents", handler.ListDocuments)
		v1.POST("/projects/:id/documents", handler.UploadDocument)

		// Issues
		v1.GET("/projects/:id/issues", handler.ListIssues)
		v1.POST("/projects/:id/issues", handler.LogIssue)
		v1.PUT("/projects/issues/:issue_id/resolve", handler.ResolveIssue)

		// Change Requests
		v1.GET("/projects/:id/change-requests", handler.ListChangeRequests)
		v1.POST("/projects/:id/change-requests", handler.CreateChangeRequest)
		v1.PUT("/projects/change-requests/:request_id/approve", handler.ApproveChangeRequest)

		// Cross-Service Integration Triggers
		v1.POST("/projects/:id/request-material", handler.RequestMaterial)
		v1.POST("/projects/:id/request-custom-order", handler.RequestCustomOrder)
	}
}
