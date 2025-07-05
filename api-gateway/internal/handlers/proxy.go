// File: api-gateway/internal/handlers/proxy.go
package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
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
			c.Request.Header.Set("X-User-ID", string(rune(userID.(uint))))
		}
		if username, exists := c.Get("username"); exists {
			c.Request.Header.Set("X-Username", username.(string))
		}

		// Proxy the request
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}