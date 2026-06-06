package routes

import (
	"github.com/erp-system/m-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	bomHandler *handlers.BOMHandler,
	prodHandler *handlers.ProductionHandler,
	qualityHandler *handlers.QualityHandler,
	maintHandler *handlers.MaintenanceHandler,
	costingHandler *handlers.CostingHandler,
) {
	v1 := r.Group("/api/v1")
	{
		// BOM Management
		v1.GET("/boms", bomHandler.ListBOMs)
		v1.POST("/boms", bomHandler.CreateBillOfMaterials)
		v1.GET("/boms/:id", bomHandler.GetBillOfMaterials)
		v1.PUT("/boms/:id", bomHandler.UpdateBillOfMaterials)
		v1.DELETE("/boms/:id", bomHandler.DeleteBillOfMaterials)
		
		// BOM Components
		v1.GET("/boms/:id/components", bomHandler.GetBOMComponents)
		v1.POST("/boms/:id/components", bomHandler.AddBOMComponent)
		v1.DELETE("/boms/:id/components/:componentId", bomHandler.RemoveBOMComponent)

		// Routing Operations
		v1.GET("/routings", bomHandler.ListRoutings)
		v1.POST("/routings", bomHandler.CreateRouting)
		v1.GET("/routings/:id", bomHandler.GetRoutingDetails)
		v1.PUT("/routings/:id", bomHandler.UpdateRouting)
		v1.DELETE("/routings/:id", bomHandler.DeleteRouting)

		// Work Centers
		v1.GET("/work-centers", bomHandler.ListWorkCenters)
		v1.POST("/work-centers", bomHandler.CreateWorkCenter)
		v1.GET("/work-centers/:id", bomHandler.GetWorkCenterDetails)
		v1.PUT("/work-centers/:id", bomHandler.UpdateWorkCenter)
		v1.DELETE("/work-centers/:id", bomHandler.DeleteWorkCenter)
		v1.POST("/work-centers/:id/machine-log", maintHandler.LogMachineStatus)

		// Equipment
		v1.POST("/equipment", maintHandler.CreateEquipment)

		// Maintenance
		v1.GET("/maintenance-schedules", maintHandler.ListMaintenanceSchedules)
		v1.POST("/maintenance-schedules/:id", maintHandler.ScheduleMaintenance)
		v1.GET("/maintenance-schedules/:id/details", maintHandler.GetMaintenanceScheduleDetails)
		v1.PUT("/maintenance-schedules/:id", maintHandler.UpdateMaintenanceSchedule)
		v1.POST("/maintenance-schedules/:id/complete", maintHandler.CompleteMaintenance)

		// Production Planning
		v1.GET("/production-plans", prodHandler.ListProductionPlans)
		v1.POST("/production-plans", prodHandler.CreateProductionPlan)
		v1.GET("/production-plans/:id", prodHandler.GetProductionPlanDetails)
		v1.PUT("/production-plans/:id", prodHandler.UpdateProductionPlan)
		v1.POST("/production-plans/:id/complete", prodHandler.CompleteProductionOrder)

		// Work Orders
		v1.GET("/work-orders", prodHandler.ListWorkOrders)
		v1.POST("/work-orders", prodHandler.CreateWorkOrder)
		v1.GET("/work-orders/:id", prodHandler.GetWorkOrderDetails)
		v1.PUT("/work-orders/:id", prodHandler.UpdateWorkOrder)
		v1.DELETE("/work-orders/:id", prodHandler.DeleteWorkOrder)
		v1.POST("/work-orders/:id/start", prodHandler.StartWorkOrder)
		v1.POST("/work-orders/:id/complete", prodHandler.CompleteWorkOrder)
		v1.POST("/work-orders/:id/labor", prodHandler.ReportLabor)
		v1.POST("/work-orders/:id/inspect", qualityHandler.RecordQualityInspection)

		// Quality Control
		v1.GET("/quality-inspections", qualityHandler.ListQualityInspections)
		v1.POST("/quality-inspections", qualityHandler.RecordQualityInspection)
		v1.GET("/quality-inspections/:id", qualityHandler.GetQualityInspectionDetails)
		v1.PUT("/quality-inspections/:id", qualityHandler.UpdateQualityInspection)

		// Costing & MRP
		v1.GET("/production-plans/:id/costing", costingHandler.GetCosting)
		v1.POST("/mrp/run", costingHandler.RunMRP)
	}
}
