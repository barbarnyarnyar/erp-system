// api-gateway/cmd/main.go
package main

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/sithuhlaing/erp-system/shared/utils"
)

func main() {
    port := getEnv("PORT", "8080")
    
    utils.InitLogger("api-gateway")
    utils.Logger.Info("Starting API Gateway")

    r := gin.Default()

    // Health check
    r.GET("/health", func(c *gin.Context) {
        utils.SuccessResponse(c, "API Gateway is healthy", gin.H{
            "service": "api-gateway",
            "version": "1.0.0",
        })
    })

    // Service discovery
    r.GET("/", func(c *gin.Context) {
        utils.SuccessResponse(c, "Hello from ERP API Gateway!", gin.H{
            "message": "Welcome to ERP Microservices System",
            "services": []gin.H{
                {"name": "finance", "url": "/api/v1/finance", "port": "8001"},
                {"name": "hr", "url": "/api/v1/hr", "port": "8002"},
                {"name": "scm", "url": "/api/v1/scm", "port": "8003"},
                {"name": "manufacturing", "url": "/api/v1/manufacturing", "port": "8004"},
                {"name": "crm", "url": "/api/v1/crm", "port": "8005"},
                {"name": "projects", "url": "/api/v1/projects", "port": "8006"},
            },
        })
    })

    r.GET("/services", func(c *gin.Context) {
        utils.SuccessResponse(c, "Available ERP Services", gin.H{
            "finance": gin.H{
                "name": "Financial Management",
                "endpoint": "/api/v1/finance/hello",
                "description": "GL, AP, AR, Financial Reports",
            },
            "hr": gin.H{
                "name": "Human Resources",
                "endpoint": "/api/v1/hr/hello",
                "description": "Employees, Payroll, Time Tracking",
            },
            "scm": gin.H{
                "name": "Supply Chain Management",
                "endpoint": "/api/v1/scm/hello",
                "description": "Products, Vendors, Inventory",
            },
            "manufacturing": gin.H{
                "name": "Manufacturing",
                "endpoint": "/api/v1/manufacturing/hello",
                "description": "BOMs, Production, Quality",
            },
            "crm": gin.H{
                "name": "Customer Relationship Management",
                "endpoint": "/api/v1/crm/hello",
                "description": "Customers, Leads, Sales",
            },
            "projects": gin.H{
                "name": "Project Management",
                "endpoint": "/api/v1/projects/hello",
                "description": "Projects, Tasks, Resources",
            },
        })
    })

    // Setup service proxies
    services := map[string]string{
        "finance":      getEnv("FINANCE_SERVICE_URL", "http://finance-service:8001"),
        "hr":          getEnv("HR_SERVICE_URL", "http://hr-service:8002"),
        "scm":         getEnv("SCM_SERVICE_URL", "http://scm-service:8003"),
        "manufacturing": getEnv("MANUFACTURING_SERVICE_URL", "http://manufacturing-service:8004"),
        "crm":         getEnv("CRM_SERVICE_URL", "http://crm-service:8005"),
        "projects":    getEnv("PROJECTS_SERVICE_URL", "http://projects-service:8006"),
    }

    for serviceName, serviceURL := range services {
        setupProxy(r, serviceName, serviceURL)
    }

    utils.Logger.WithField("port", port).Info("API Gateway starting")
    if err := r.Run(":" + port); err != nil {
        log.Fatal("Failed to start API Gateway:", err)
    }
}

func setupProxy(r *gin.Engine, serviceName, serviceURL string) {
    target, err := url.Parse(serviceURL)
    if err != nil {
        utils.Logger.WithField("service", serviceName).WithError(err).Error("Failed to parse service URL")
        return
    }

    proxy := httputil.NewSingleHostReverseProxy(target)
    
    r.Any("/api/v1/"+serviceName+"/*path", func(c *gin.Context) {
        proxy.ServeHTTP(c.Writer, c.Request)
    })

    utils.Logger.WithField("service", serviceName).WithField("url", serviceURL).Info("Proxy configured")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}