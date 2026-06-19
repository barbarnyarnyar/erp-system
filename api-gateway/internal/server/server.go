// File: api-gateway/internal/server/server.go
package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
)

type Server struct {
	config *config.Config
	router *gin.Engine
}

func New(cfg *config.Config) *Server {
	router := gin.Default()
	
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	return &Server{
		config: cfg,
		router: router,
	}
}

func (s *Server) Start() error {
	s.setupRoutes()
	return s.router.Run(":" + s.config.Port)
}

func (s *Server) setupRoutes() {
	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(s.config.JWTSecret, s.config.Services.AuthService)
	
	// Proxy handler
	proxyHandler := handlers.NewProxyHandler(map[string]string{
		"auth": s.config.Services.AuthService,
		"fm":   s.config.Services.FMService,
		"hr":   s.config.Services.HRService,
		"scm":  s.config.Services.SCMService,
		"m":    s.config.Services.MService,
		"crm":  s.config.Services.CRMService,
		"pm":   s.config.Services.PMService,
		"eam":  s.config.Services.EAMService,
		"plm":  s.config.Services.PLMService,
		"qms":  s.config.Services.QMSService,
	})

	// API Gateway health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// Auth service health passthrough
	s.router.GET("/health/auth", proxyHandler.ProxyToService("auth"))

	// Public routes (no authentication required)
	public := s.router.Group("/api/v1")
	{
		// Auth routes
		public.POST("/auth/login", proxyHandler.ProxyToService("auth"))
		public.POST("/auth/register", proxyHandler.ProxyToService("auth"))
		public.POST("/auth/refresh", proxyHandler.ProxyToService("auth"))
	}

	// Protected routes (authentication required)
	protected := s.router.Group("/api/v1")
	protected.Use(authMiddleware.ValidateToken())
	{
		// Auth routes
		authGroup := protected.Group("/auth")
		{
			authGroup.POST("/logout", proxyHandler.ProxyToService("auth"))
			authGroup.GET("/profile", proxyHandler.ProxyToService("auth"))
			authGroup.GET("/validate", proxyHandler.ProxyToService("auth"))
		}

		// Financial Management routes
		fmGroup := protected.Group("/finance")
		fmGroup.Use(authMiddleware.RequirePermission("fm", "*", "read"))
		{
			// Accounts
			fmGroup.GET("/accounts", proxyHandler.ProxyToService("fm"))
			fmGroup.POST("/accounts", 
				authMiddleware.RequirePermission("fm", "accounts", "write"),
				proxyHandler.ProxyToService("fm"))
			fmGroup.PUT("/accounts/:id", 
				authMiddleware.RequirePermission("fm", "accounts", "write"),
				proxyHandler.ProxyToService("fm"))
			fmGroup.DELETE("/accounts/:id", 
				authMiddleware.RequirePermission("fm", "accounts", "delete"),
				proxyHandler.ProxyToService("fm"))
			fmGroup.GET("/accounts/:id/balance", proxyHandler.ProxyToService("fm"))

			// Parties (Customers/Vendors)
			fmGroup.GET("/parties", proxyHandler.ProxyToService("fm"))
			fmGroup.POST("/parties", 
				authMiddleware.RequirePermission("fm", "parties", "write"),
				proxyHandler.ProxyToService("fm"))
			fmGroup.PUT("/parties/:id", 
				authMiddleware.RequirePermission("fm", "parties", "write"),
				proxyHandler.ProxyToService("fm"))

			// Invoices
			fmGroup.GET("/invoices", proxyHandler.ProxyToService("fm"))
			fmGroup.POST("/invoices", 
				authMiddleware.RequirePermission("fm", "invoices", "write"),
				proxyHandler.ProxyToService("fm"))
			fmGroup.PUT("/invoices/:id", 
				authMiddleware.RequirePermission("fm", "invoices", "write"),
				proxyHandler.ProxyToService("fm"))
			fmGroup.GET("/invoices/:id/lines", proxyHandler.ProxyToService("fm"))

			// Vendor Bills
			fmGroup.GET("/vendor-bills", proxyHandler.ProxyToService("fm"))
			fmGroup.POST("/vendor-bills",
				authMiddleware.RequirePermission("fm", "invoices", "write"),
				proxyHandler.ProxyToService("fm"))
			fmGroup.GET("/vendor-bills/:id/lines", proxyHandler.ProxyToService("fm"))

			// Bank Statements
			fmGroup.GET("/bank-statements/:id/lines", proxyHandler.ProxyToService("fm"))

			// Payments
			fmGroup.GET("/payments", proxyHandler.ProxyToService("fm"))
			fmGroup.POST("/payments", 
				authMiddleware.RequirePermission("fm", "payments", "write"),
				proxyHandler.ProxyToService("fm"))

			// Journal Entries
			fmGroup.GET("/journal-entries", proxyHandler.ProxyToService("fm"))
			fmGroup.POST("/journal-entries", 
				authMiddleware.RequirePermission("fm", "journal", "write"),
				proxyHandler.ProxyToService("fm"))
			fmGroup.POST("/journal-entries/:id/post", 
				authMiddleware.RequirePermission("fm", "journal", "post"),
				proxyHandler.ProxyToService("fm"))

			// Reports
			fmGroup.GET("/reports/*path", 
				authMiddleware.RequirePermission("fm", "reports", "read"),
				proxyHandler.ProxyToService("fm"))
		}

		// HR routes
		hrGroup := protected.Group("/hr")
		hrGroup.Use(authMiddleware.RequirePermission("hr", "*", "read"))
		{
			hrGroup.Any("", proxyHandler.ProxyToService("hr"))
			hrGroup.Any("/*path", proxyHandler.ProxyToService("hr"))
		}

		// SCM routes
		scmGroup := protected.Group("/scm")
		scmGroup.Use(authMiddleware.RequirePermission("scm", "*", "read"))
		{
			scmGroup.Any("", proxyHandler.ProxyToService("scm"))
			scmGroup.Any("/*path", proxyHandler.ProxyToService("scm"))
		}

		// Enterprise Asset Management (EAM) routes
		eamGroup := protected.Group("/eam")
		eamGroup.Use(authMiddleware.RequirePermission("eam", "*", "read"))
		{
			eamGroup.Any("", proxyHandler.ProxyToService("eam"))
			eamGroup.Any("/*path", proxyHandler.ProxyToService("eam"))
		}

		// Product Lifecycle Management (PLM) routes
		plmGroup := protected.Group("/plm")
		plmGroup.Use(authMiddleware.RequirePermission("plm", "*", "read"))
		{
			plmGroup.Any("", proxyHandler.ProxyToService("plm"))
			plmGroup.Any("/*path", proxyHandler.ProxyToService("plm"))
		}

		// Quality Management System (QMS) routes
		qmsGroup := protected.Group("/qms")
		qmsGroup.Use(authMiddleware.RequirePermission("qms", "*", "read"))
		{
			qmsGroup.Any("", proxyHandler.ProxyToService("qms"))
			qmsGroup.Any("/*path", proxyHandler.ProxyToService("qms"))
		}

		// Manufacturing routes
		mGroup := protected.Group("/manufacturing")
		mGroup.Use(authMiddleware.RequirePermission("m", "*", "read"))
		{
			mGroup.Any("", proxyHandler.ProxyToService("m"))
			mGroup.Any("/*path", proxyHandler.ProxyToService("m"))
		}

		// CRM routes
		crmGroup := protected.Group("/crm")
		crmGroup.Use(authMiddleware.RequirePermission("crm", "*", "read"))
		{
			crmGroup.Any("", proxyHandler.ProxyToService("crm"))
			crmGroup.Any("/*path", proxyHandler.ProxyToService("crm"))
		}

		// Project Management routes
		pmGroup := protected.Group("/projects")
		pmGroup.Use(authMiddleware.RequirePermission("pm", "*", "read"))
		{
			pmGroup.Any("", proxyHandler.ProxyToService("pm"))
			pmGroup.Any("/*path", proxyHandler.ProxyToService("pm"))
		}

		// Admin routes (require admin role)
		adminGroup := protected.Group("/admin")
		adminGroup.Use(authMiddleware.RequireRole("admin"))
		{
			adminGroup.Any("/auth/*path", proxyHandler.ProxyToService("auth"))
			adminGroup.GET("/services/status", s.getServicesStatus)
		}
	}
}

func (s *Server) getServicesStatus(c *gin.Context) {
	services := map[string]string{
		"auth": s.config.Services.AuthService,
		"fm":   s.config.Services.FMService,
		"hr":   s.config.Services.HRService,
		"scm":  s.config.Services.SCMService,
		"m":    s.config.Services.MService,
		"crm":  s.config.Services.CRMService,
		"pm":   s.config.Services.PMService,
		"eam":  s.config.Services.EAMService,
		"plm":  s.config.Services.PLMService,
		"qms":  s.config.Services.QMSService,
	}

	status := make(map[string]interface{})
	for name, url := range services {
		// Simple health check
		resp, err := http.Get(url + "/health")
		if err != nil {
			status[name] = map[string]interface{}{
				"status": "down",
				"error":  err.Error(),
			}
		} else {
			resp.Body.Close()
			status[name] = map[string]interface{}{
				"status": "up",
				"url":    url,
			}
		}
	}

	c.JSON(200, gin.H{"services": status})
}