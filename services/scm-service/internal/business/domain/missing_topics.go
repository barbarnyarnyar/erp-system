package domain

const (
	TopicScmInventoryReceived           = "scm.inventory.received"
	TopicScmInventoryShipped            = "scm.inventory.shipped"
	TopicScmInventoryAdjusted           = "scm.inventory.adjusted"
	TopicScmInventoryOutOfStock         = "scm.inventory.out.of.stock"
	TopicScmInventoryValued             = "scm.inventory.valued"
	TopicScmProductCreated              = "scm.product.created"
	TopicScmProductUpdated              = "scm.product.updated"
	TopicScmProductDiscontinued         = "scm.product.discontinued"
	TopicScmPurchaseOrderReceived       = "scm.purchase.order.received"
	TopicScmPurchaseOrderCancelled      = "scm.purchase.order.cancelled"
	TopicScmPurchaseOrderSent           = "scm.purchase.order.sent"
	TopicScmVendorCreated               = "scm.vendor.created"
	TopicScmVendorUpdated               = "scm.vendor.updated"
	TopicScmVendorPerformanceEvaluated  = "scm.vendor.performance.evaluated"
	TopicScmMaterialDelivered           = "scm.material.delivered"
	TopicScmShipmentCreated             = "scm.shipment.created"
	TopicScmShipmentDelivered           = "scm.shipment.delivered"
	TopicScmShipmentDelayed             = "scm.shipment.delayed"
	TopicScmTrainingRequired            = "scm.training.required"

	// Missing Consumer topics
	TopicCrmSalesOrderCreated      = "crm.sales.order.created"
	TopicCrmCustomerDemandForecast = "crm.customer.demand.forecast"
	TopicMfgMaterialRequired       = "mfg.material.required"
	TopicMfgMaterialConsumed       = "mfg.material.consumed"
	TopicMfgProductionCompleted    = "mfg.production.completed"
	TopicPrjMaterialRequested      = "prj.material.requested"
)
