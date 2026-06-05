package routes

import (
	"github.com/erp-system/m-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	bomHandler *handlers.BOMHandler,
	prodHandler *handlers.ProductionHandler,
) {
	v1 := r.Group("/api/v1")
	{
		// BOM Management
		v1.GET("/boms", bomHandler.ListBOMs)
		v1.POST("/boms", bomHandler.CreateBillOfMaterials)
		v1.GET("/boms/:id", bomHandler.GetBillOfMaterials)
		v1.PUT("/boms/:id", bomHandler.UpdateBillOfMaterials)
		v1.DELETE("/boms/:id", bomHandler.DeleteBillOfMaterials)

		// Routing Management
		v1.GET("/routings", bomHandler.ListRoutings)
		v1.POST("/routings", bomHandler.CreateRouting)
		v1.GET("/routings/:id", bomHandler.GetRoutingDetails)
		v1.PUT("/routings/:id", bomHandler.UpdateRouting)
		v1.DELETE("/routings/:id", bomHandler.DeleteRouting)

		// Work Orders
		v1.GET("/work-orders", prodHandler.ListWorkOrders)
		v1.POST("/work-orders", prodHandler.CreateWorkOrder)
		v1.GET("/work-orders/:id", prodHandler.GetWorkOrderDetails)
		v1.PUT("/work-orders/:id", prodHandler.UpdateWorkOrder)
		v1.DELETE("/work-orders/:id", prodHandler.DeleteWorkOrder)
		v1.POST("/work-orders/:id/start", prodHandler.StartWorkOrder)
		v1.POST("/work-orders/:id/complete", prodHandler.CompleteWorkOrder)
		v1.POST("/work-orders/:id/labor", prodHandler.ReportLabor)
		v1.POST("/work-orders/:id/inspect", prodHandler.RecordQualityInspection)

		// Production Planning
		v1.GET("/production-plans", prodHandler.ListProductionPlans)
		v1.POST("/production-plans", prodHandler.CreateProductionPlan)
		v1.GET("/production-plans/:id", prodHandler.GetProductionPlanDetails)
		v1.PUT("/production-plans/:id", prodHandler.UpdateProductionPlan)
		v1.POST("/mrp/run", prodHandler.RunMRP)

		// Quality Control
		v1.GET("/quality-inspections", prodHandler.ListQualityInspections)
		v1.POST("/quality-inspections", prodHandler.RecordQualityInspection)
		v1.GET("/quality-inspections/:id", prodHandler.GetQualityInspectionDetails)
		v1.PUT("/quality-inspections/:id", prodHandler.UpdateQualityInspection)

		// Work Centers
		v1.GET("/work-centers", bomHandler.ListWorkCenters)
		v1.POST("/work-centers", bomHandler.CreateWorkCenter)
		v1.GET("/work-centers/:id", bomHandler.GetWorkCenterDetails)
		v1.PUT("/work-centers/:id", bomHandler.UpdateWorkCenter)
		v1.DELETE("/work-centers/:id", bomHandler.DeleteWorkCenter)
		v1.POST("/work-centers/:id/machine-log", prodHandler.LogMachineStatus)

		// Maintenance
		v1.GET("/maintenance-schedules", prodHandler.ListMaintenanceSchedules)
		v1.POST("/maintenance-schedules", prodHandler.ScheduleMaintenance)
		v1.GET("/maintenance-schedules/:id", prodHandler.GetMaintenanceScheduleDetails)
		v1.PUT("/maintenance-schedules/:id", prodHandler.UpdateMaintenanceSchedule)
	}
}
