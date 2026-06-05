package routes

import (
	"github.com/erp-system/scm-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	prodHandler *handlers.ProductHandler,
	vendorHandler *handlers.VendorHandler,
	poHandler *handlers.PurchaseOrderHandler,
	invHandler *handlers.InventoryHandler,
	whHandler *handlers.WarehouseHandler,
	demandHandler *handlers.DemandForecastHandler,
	reportHandler *handlers.ReportHandler,
) {
	v1 := r.Group("/api/v1")
	{
		// Product Categories
		v1.GET("/product-categories", prodHandler.GetCategories)
		v1.POST("/product-categories", prodHandler.CreateCategory)
		v1.GET("/product-categories/:id", prodHandler.GetCategory)
		v1.PUT("/product-categories/:id", prodHandler.UpdateCategory)
		v1.DELETE("/product-categories/:id", prodHandler.DeleteCategory)

		// Product Management
		v1.GET("/products", prodHandler.GetProducts)
		v1.POST("/products", prodHandler.CreateProduct)
		v1.GET("/products/:id", prodHandler.GetProduct)
		v1.PUT("/products/:id", prodHandler.UpdateProduct)
		v1.DELETE("/products/:id", prodHandler.DeleteProduct)

		// Vendor Management
		v1.GET("/vendors", vendorHandler.GetVendors)
		v1.POST("/vendors", vendorHandler.CreateVendor)
		v1.GET("/vendors/:id", vendorHandler.GetVendor)
		v1.PUT("/vendors/:id", vendorHandler.UpdateVendor)
		v1.DELETE("/vendors/:id", vendorHandler.DeleteVendor)

		// Vendor Contracts
		v1.GET("/vendor-contracts", vendorHandler.GetContracts)
		v1.POST("/vendor-contracts", vendorHandler.CreateContract)
		v1.GET("/vendor-contracts/:id", vendorHandler.GetContract)
		v1.PUT("/vendor-contracts/:id", vendorHandler.UpdateContract)
		v1.DELETE("/vendor-contracts/:id", vendorHandler.DeleteContract)

		// Purchase Requisitions
		v1.GET("/purchase-requisitions", poHandler.GetPurchaseRequisitions)
		v1.POST("/purchase-requisitions", poHandler.CreatePurchaseRequisition)
		v1.GET("/purchase-requisitions/:id", poHandler.GetPurchaseRequisition)
		v1.PUT("/purchase-requisitions/:id", poHandler.UpdatePurchaseRequisition)
		v1.DELETE("/purchase-requisitions/:id", poHandler.DeletePurchaseRequisition)
		v1.POST("/purchase-requisitions/:id/approve", poHandler.ApprovePurchaseRequisition)
		v1.POST("/purchase-requisitions/:id/reject", poHandler.RejectPurchaseRequisition)

		// Purchase Orders
		v1.GET("/purchase-orders", poHandler.GetPurchaseOrders)
		v1.POST("/purchase-orders", poHandler.CreatePurchaseOrder)
		v1.GET("/purchase-orders/:id", poHandler.GetPurchaseOrder)
		v1.PUT("/purchase-orders/:id", poHandler.UpdatePurchaseOrder)
		v1.DELETE("/purchase-orders/:id", poHandler.DeletePurchaseOrder)
		v1.POST("/purchase-orders/:id/send", poHandler.SendPurchaseOrder)

		// Inventory
		v1.GET("/inventory", invHandler.GetInventoryItems)
		v1.POST("/inventory", invHandler.CreateInventoryItem)
		v1.GET("/inventory/:id", invHandler.GetInventoryItem)
		v1.PUT("/inventory/:id", invHandler.UpdateInventoryItem)
		v1.DELETE("/inventory/:id", invHandler.DeleteInventoryItem)

		// Warehouse Operations - Receipts
		v1.GET("/receipts", whHandler.GetReceipts)
		v1.POST("/receipts", whHandler.CreateReceipt)
		v1.GET("/receipts/:id", whHandler.GetReceipt)
		v1.PUT("/receipts/:id", whHandler.UpdateReceipt)

		// Warehouse Operations - Shipments
		v1.GET("/shipments", whHandler.GetShipments)
		v1.POST("/shipments", whHandler.CreateShipment)
		v1.GET("/shipments/:id", whHandler.GetShipment)
		v1.PUT("/shipments/:id", whHandler.UpdateShipment)

		// Demand Planning
		v1.GET("/demand-forecasts", demandHandler.GetForecasts)
		v1.POST("/demand-forecasts", demandHandler.CreateForecast)
		v1.GET("/demand-forecasts/:id", demandHandler.GetForecast)
		v1.PUT("/demand-forecasts/:id", demandHandler.UpdateForecast)

		// Reporting
		v1.GET("/reports/inventory-levels", reportHandler.GetInventoryLevelsReport)
		v1.GET("/reports/vendor-performance", reportHandler.GetVendorPerformanceReport)
		v1.GET("/reports/procurement-metrics", reportHandler.GetProcurementMetricsReport)
		v1.GET("/reports/safety-stock", reportHandler.GetSafetyStockReport)
	}
}
