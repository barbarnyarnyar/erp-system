package routes

import (
	"github.com/erp-system/m-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	mfgHandler *handlers.MfgHandler,
) {
	v1 := r.Group("/api/v1")
	{
		// Floor / Work Center Configuration
		v1.POST("/mfg/work-centers", mfgHandler.EstablishWorkCenter)
		v1.POST("/mfg/work-centers/:id/stations", mfgHandler.AppendStationToCenter)

		// Work Order Execution
		v1.POST("/mfg/work-orders", mfgHandler.InstantiateWorkOrder)
		v1.POST("/mfg/work-orders/:id/transition", mfgHandler.TransitionWorkOrderState)
		v1.POST("/mfg/work-orders/:id/reroute", mfgHandler.RerouteWorkOrderStation)

		// Shop Floor Telemetry
		v1.POST("/mfg/work-orders/:id/consumption", mfgHandler.RecordBulkMaterialConsumption)
		v1.POST("/mfg/work-orders/:id/yield", mfgHandler.CommitProductionYield)
	}
}
