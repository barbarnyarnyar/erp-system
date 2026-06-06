package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erp-system/auth-service/internal/api/handlers"
	"github.com/erp-system/auth-service/internal/api/routes"
	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/business/service"
	"github.com/erp-system/auth-service/internal/config"
	"github.com/erp-system/auth-service/internal/data/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting auth-service in %s environment...", cfg.Server.Env)

	// 2. Initialize in-memory repositories
	userRepo := memory.NewUserRepository()
	sessRepo := memory.NewSessionRepository()
	roleRepo := memory.NewRoleRepository()
	permRepo := memory.NewPermissionRepository()
	urRepo := memory.NewUserRoleRepository()
	usRepo := memory.NewUserStoreRepository()
	rpRepo := memory.NewRolePermissionRepository()

	// 3. Initialize business service
	idSvc := service.NewIdentityService(
		userRepo,
		sessRepo,
		roleRepo,
		permRepo,
		urRepo,
		usRepo,
		rpRepo,
		cfg,
	)

	// 4. Seed initial roles, permissions, and users
	seedAuthData(idSvc)

	// 5. Setup Gin routing
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "auth-service",
			"status":  "healthy",
			"port":    cfg.Server.Port,
		})
	})

	handler := handlers.NewIdentityHandler(idSvc)
	routes.SetupAuthRoutes(r, handler)

	// 6. Start HTTP server with graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Auth HTTP Server listening on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down auth-service...")

	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("auth-service stopped gracefully.")
}

func seedAuthData(s *service.IdentityService) {
	ctx := context.Background()
	log.Println("Seeding Auth / RBAC definitions...")

	// Create roles
	adminRole, _ := s.CreateRole(ctx, "Admin", "Super admin with all permissions")
	managerRole, _ := s.CreateRole(ctx, "Manager", "Store manager permissions")
	clerkRole, _ := s.CreateRole(ctx, "Clerk", "Basic clerk permissions")

	// Create permissions
	pCreateProduct, _ := s.CreatePermission(ctx, "scm:product:create", "Create products")
	pReadProduct, _ := s.CreatePermission(ctx, "scm:product:read", "View products")
	pCreateCustomer, _ := s.CreatePermission(ctx, "crm:customer:create", "Create customers")
	pReadCustomer, _ := s.CreatePermission(ctx, "crm:customer:read", "View customers")

	// Link permissions to Admin Role
	_ = s.LinkRolePermission(ctx, adminRole.ID, pCreateProduct.ID)
	_ = s.LinkRolePermission(ctx, adminRole.ID, pReadProduct.ID)
	_ = s.LinkRolePermission(ctx, adminRole.ID, pCreateCustomer.ID)
	_ = s.LinkRolePermission(ctx, adminRole.ID, pReadCustomer.ID)

	// Link permissions to Manager Role
	_ = s.LinkRolePermission(ctx, managerRole.ID, pReadProduct.ID)
	_ = s.LinkRolePermission(ctx, managerRole.ID, pReadCustomer.ID)

	// Link permissions to Clerk Role
	_ = s.LinkRolePermission(ctx, clerkRole.ID, pReadProduct.ID)

	// Create initial admin user
	adminUser := &domain.User{
		Username:     "admin",
		Email:        "admin@erp.com",
		PasswordHash: "admin123", // Simple plain password for mock purposes
		FirstName:    "System",
		LastName:     "Administrator",
	}

	_, err := s.CreateUser(ctx, adminUser, "store_default", []string{adminRole.ID})
	if err != nil {
		log.Printf("Failed to seed admin user: %v", err)
		return
	}

	log.Println("Auth / RBAC mock data seeded successfully.")
}