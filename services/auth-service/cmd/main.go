package main

import (
	"context"
	sharedkafka "erp-system/shared/kafka"
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
	"github.com/erp-system/auth-service/internal/data/kafka"
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

	// 2. Initialize Event Publisher (Kafka)
	publisher := sharedkafka.NewPublisher(cfg.Kafka.Brokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Failed to close Kafka publisher: %v", err)
		}
	}()

	// 3. Initialize in-memory repositories
	userRepo := memory.NewUserRepository()
	sessRepo := memory.NewSessionRepository()
	roleRepo := memory.NewRoleRepository()
	permRepo := memory.NewPermissionRepository()
	urRepo := memory.NewUserRoleRepository()
	usRepo := memory.NewUserStoreRepository()
	rpRepo := memory.NewRolePermissionRepository()

	// 4. Initialize business services (split components)
	rbacSvc := service.NewRBACService(
		roleRepo,
		permRepo,
		urRepo,
		rpRepo,
		publisher,
	)

	userSvc := service.NewUserService(
		userRepo,
		usRepo,
		urRepo,
		publisher,
	)

	authSvc := service.NewAuthService(
		userRepo,
		sessRepo,
		rbacSvc,
		publisher,
		cfg,
	)

	// 5. Seed initial roles, permissions, and users
	seedAuthData(userSvc, rbacSvc)

	// 5b. Start Kafka consumer for HR offboarding events (Phase S4.8)
	// Closes the offboarding loop: when HR publishes hr.employee.terminated,
	// Auth automatically deactivates the corresponding user.
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()
	consumer := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, "auth-service", publisher, userSvc)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}()
	go consumer.Start(consumerCtx)

	// 6. Setup Gin routing
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

	handler := handlers.NewIdentityHandler(authSvc, userSvc, rbacSvc)
	rbacHandler := handlers.NewRBACHandler(rbacSvc)
	routes.SetupAuthRoutes(r, handler, rbacHandler)

	// 7. Start HTTP server with graceful shutdown
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

func seedAuthData(userSvc *service.UserService, rbacSvc *service.RBACService) {
	ctx := context.Background()
	log.Println("Seeding Auth / RBAC definitions...")

	// Create roles
	adminRole, _ := rbacSvc.CreateRole(ctx, "Admin", "Super admin with all permissions")
	managerRole, _ := rbacSvc.CreateRole(ctx, "Manager", "Store manager permissions")
	clerkRole, _ := rbacSvc.CreateRole(ctx, "Clerk", "Basic clerk permissions")

	// Create permissions
	pCreateProduct, _ := rbacSvc.CreatePermission(ctx, "scm:product:create", "Create products")
	pReadProduct, _ := rbacSvc.CreatePermission(ctx, "scm:product:read", "View products")
	pCreateCustomer, _ := rbacSvc.CreatePermission(ctx, "crm:customer:create", "Create customers")
	pReadCustomer, _ := rbacSvc.CreatePermission(ctx, "crm:customer:read", "View customers")

	pReadHR, _ := rbacSvc.CreatePermission(ctx, "hr:*:read", "Read HR")
	pReadSCM, _ := rbacSvc.CreatePermission(ctx, "scm:*:read", "Read SCM")
	pReadM, _ := rbacSvc.CreatePermission(ctx, "m:*:read", "Read Manufacturing")
	pReadCRM, _ := rbacSvc.CreatePermission(ctx, "crm:*:read", "Read CRM")
	pReadPM, _ := rbacSvc.CreatePermission(ctx, "pm:*:read", "Read Projects")
	pReadFM, _ := rbacSvc.CreatePermission(ctx, "fm:*:read", "Read Finance")

	// Link permissions to Admin Role
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pCreateProduct.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pReadProduct.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pCreateCustomer.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pReadCustomer.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pReadHR.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pReadSCM.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pReadM.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pReadCRM.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pReadPM.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, adminRole.ID, pReadFM.ID)

	// Link permissions to Manager Role
	_ = rbacSvc.AssignPermissionToRole(ctx, managerRole.ID, pReadProduct.ID)
	_ = rbacSvc.AssignPermissionToRole(ctx, managerRole.ID, pReadCustomer.ID)

	// Link permissions to Clerk Role
	_ = rbacSvc.AssignPermissionToRole(ctx, clerkRole.ID, pReadProduct.ID)

	// Create initial admin user (CreateUser hashes the password)
	adminUser := &domain.User{
		Username:     "admin",
		Email:        "admin@erp.com",
		PasswordHash: "admin123",
		FirstName:    "System",
		LastName:     "Administrator",
	}

	_, err := userSvc.CreateUser(ctx, adminUser, "store_default", []string{adminRole.ID})
	if err != nil {
		log.Printf("Failed to seed admin user: %v", err)
		return
	}

	log.Println("Auth / RBAC mock data seeded successfully.")
}
