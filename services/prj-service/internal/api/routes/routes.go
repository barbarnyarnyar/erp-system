package routes

import (
	"github.com/erp-system/pm-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupPMRoutes(r *gin.Engine, handler *handlers.PrjHandler) {
	v1 := r.Group("/api/v1")
	{
		// Projects
		v1.POST("/projects", handler.InitializeProject)
		v1.PUT("/projects/:id/status", handler.TransitionProjectStatus)
		v1.GET("/projects", handler.ListProjects)
		v1.GET("/projects/:id", handler.GetProject)

		// WBS Structure
		v1.POST("/projects/:id/wbs", handler.AppendWbsNode)
		v1.PUT("/wbs/:node_id/complete", handler.DeclareNodeCompletion)
		v1.GET("/projects/:id/wbs", handler.FetchProjectTree)

		// Time Tracking
		v1.POST("/time-logs/bulk", handler.LogOperationalHoursBulk)
		v1.POST("/time-logs/approve", handler.ProcessTimesheetApproval)
	}
}
