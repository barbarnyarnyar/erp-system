// File: api-gateway/internal/handlers/proxy.go
package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	services map[string]*httputil.ReverseProxy
}

func NewProxyHandler(serviceURLs map[string]string) *ProxyHandler {
	proxies := make(map[string]*httputil.ReverseProxy)
	
	for service, serviceURL := range serviceURLs {
		target, _ := url.Parse(serviceURL)
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Custom director to modify requests
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			// Add custom headers if needed
			req.Header.Set("X-Forwarded-By", "api-gateway")
		}
		
		proxies[service] = proxy
	}
	
	return &ProxyHandler{services: proxies}
}

func (p *ProxyHandler) ProxyToService(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy, exists := p.services[serviceName]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Add user context to headers
		if userID, exists := c.Get("user_id"); exists {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}
		if username, exists := c.Get("username"); exists {
			c.Request.Header.Set("X-Username", username.(string))
		}

		// Rewrite path for backend services that do not expect the service name prefix
		originalPath := c.Request.URL.Path
		switch serviceName {
		case "fm":
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, "/api/v1/finance", "/api/v1", 1)
		case "hr":
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, "/api/v1/hr", "/api/v1", 1)
		case "scm":
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, "/api/v1/scm", "/api/v1", 1)
		case "m":
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, "/api/v1/manufacturing", "/api/v1", 1)
		case "crm":
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, "/api/v1/crm", "/api/v1", 1)
		}

		// Proxy the request
		proxy.ServeHTTP(c.Writer, c.Request)

		// Restore original path
		c.Request.URL.Path = originalPath
	}
}