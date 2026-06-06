package routes

import (
	"github.com/erp-system/crm-service/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupCRMRoutes(
	r *gin.Engine,
	custLeadHandler *handlers.CustomerLeadHandler,
	salesOppHandler *handlers.SalesOpportunityHandler,
	custInteractionHandler *handlers.CustomerInteractionHandler,
) {
	v1 := r.Group("/api/v1")
	{
		// Customers
		v1.GET("/customers", custLeadHandler.ListCustomers)
		v1.POST("/customers", custLeadHandler.CreateCustomer)
		v1.GET("/customers/:id", custLeadHandler.GetCustomer)
		v1.PUT("/customers/:id", custLeadHandler.UpdateCustomer)
		v1.DELETE("/customers/:id", custLeadHandler.DeleteCustomer)

		// Customer Interactions
		v1.GET("/customer-interactions", custInteractionHandler.ListCustomerInteractions)
		v1.POST("/customer-interactions", custInteractionHandler.CreateCustomerInteraction)
		v1.GET("/customer-interactions/:id", custInteractionHandler.GetCustomerInteraction)
		v1.DELETE("/customer-interactions/:id", custInteractionHandler.DeleteCustomerInteraction)

		// Leads
		v1.GET("/leads", custLeadHandler.ListLeads)
		v1.POST("/leads", custLeadHandler.CreateLead)
		v1.GET("/leads/:id", custLeadHandler.GetLead)
		v1.PUT("/leads/:id", custLeadHandler.UpdateLead)
		v1.DELETE("/leads/:id", custLeadHandler.DeleteLead)
		v1.POST("/leads/:id/convert", custLeadHandler.ConvertLead)

		// Opportunities
		v1.GET("/opportunities", salesOppHandler.ListOpportunities)
		v1.POST("/opportunities", salesOppHandler.CreateOpportunity)
		v1.GET("/opportunities/:id", salesOppHandler.GetOpportunity)
		v1.PUT("/opportunities/:id", salesOppHandler.UpdateOpportunity)
		v1.DELETE("/opportunities/:id", salesOppHandler.DeleteOpportunity)
		v1.GET("/opportunities/:id/stage-history", salesOppHandler.GetOpportunityStageHistory)

		// Sales Orders
		v1.GET("/sales-orders", salesOppHandler.ListSalesOrders)
		v1.POST("/sales-orders", salesOppHandler.CreateSalesOrder)
		v1.GET("/sales-orders/:id", salesOppHandler.GetSalesOrder)
		v1.PUT("/sales-orders/:id", salesOppHandler.UpdateSalesOrder)
		v1.DELETE("/sales-orders/:id", salesOppHandler.DeleteSalesOrder)

		// Quotes
		v1.GET("/quotes", salesOppHandler.ListQuotes)
		v1.POST("/quotes", salesOppHandler.CreateQuote)
		v1.GET("/quotes/:id", salesOppHandler.GetQuote)
		v1.PUT("/quotes/:id", salesOppHandler.UpdateQuote)
		v1.DELETE("/quotes/:id", salesOppHandler.DeleteQuote)
		v1.POST("/quotes/:id/send", salesOppHandler.SendQuote)

		// Service Tickets
		v1.GET("/service-tickets", salesOppHandler.ListServiceTickets)
		v1.POST("/service-tickets", salesOppHandler.CreateServiceTicket)
		v1.GET("/service-tickets/:id", salesOppHandler.GetServiceTicket)
		v1.PUT("/service-tickets/:id", salesOppHandler.UpdateServiceTicket)
		v1.DELETE("/service-tickets/:id", salesOppHandler.DeleteServiceTicket)

		// Campaigns
		v1.GET("/campaigns", salesOppHandler.ListCampaigns)
		v1.POST("/campaigns", salesOppHandler.CreateCampaign)
		v1.GET("/campaigns/:id", salesOppHandler.GetCampaign)
		v1.PUT("/campaigns/:id", salesOppHandler.UpdateCampaign)
		v1.DELETE("/campaigns/:id", salesOppHandler.DeleteCampaign)

		// Price Lists
		v1.GET("/price-lists", salesOppHandler.ListPriceLists)
		v1.POST("/price-lists", salesOppHandler.CreatePriceList)
		v1.GET("/price-lists/:id", salesOppHandler.GetPriceList)
		v1.PUT("/price-lists/:id", salesOppHandler.UpdatePriceList)
		v1.DELETE("/price-lists/:id", salesOppHandler.DeletePriceList)
	}
}
