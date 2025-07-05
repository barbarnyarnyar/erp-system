# Makefile
.PHONY: help build run stop clean logs test

# Default target
help: ## Show this help message
	@echo "ERP Microservices - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build all services
	@echo "Building ERP Microservices..."
	docker-compose build --parallel

run: ## Start all services
	@echo "Starting ERP Microservices..."
	docker-compose up -d
	@echo "Services are starting up..."
	@echo "API Gateway: http://localhost:8080"
	@echo "Finance Service: http://localhost:8001"
	@echo "HR Service: http://localhost:8002"
	@echo "SCM Service: http://localhost:8003"
	@echo "Manufacturing Service: http://localhost:8004"
	@echo "CRM Service: http://localhost:8005"
	@echo "Projects Service: http://localhost:8006"
	@echo ""
	@echo "RabbitMQ Management: http://localhost:15672 (user: erp_user, pass: erp_password)"

run-dev: ## Start services in development mode with logs
	@echo "Starting ERP Microservices in development mode..."
	docker-compose up

stop: ## Stop all services
	@echo "Stopping ERP Microservices..."
	docker-compose down

restart: ## Restart all services
	@echo "Restarting ERP Microservices..."
	docker-compose restart

clean: ## Clean up containers, volumes, and images
	@echo "Cleaning up ERP Microservices..."
	docker-compose down -v --rmi all --remove-orphans
	docker system prune -f

logs: ## Show logs for all services
	docker-compose logs -f

logs-service: ## Show logs for specific service (usage: make logs-service SERVICE=finance)
	docker-compose logs -f $(SERVICE)

status: ## Show status of all services
	docker-compose ps

health: ## Check health of all services
	@echo "Checking service health..."
	@echo ""
	@echo "API Gateway:"
	@curl -s http://localhost:8080/health | jq '.' 2>/dev/null || echo "Service not responding"
	@echo ""
	@echo "Finance Service:"
	@curl -s http://localhost:8001/health | jq '.' 2>/dev/null || echo "Service not responding"
	@echo ""
	@echo "HR Service:"
	@curl -s http://localhost:8002/health | jq '.' 2>/dev/null || echo "Service not responding"
	@echo ""
	@echo "SCM Service:"
	@curl -s http://localhost:8003/health | jq '.' 2>/dev/null || echo "Service not responding"
	@echo ""
	@echo "Manufacturing Service:"
	@curl -s http://localhost:8004/health | jq '.' 2>/dev/null || echo "Service not responding"
	@echo ""
	@echo "CRM Service:"
	@curl -s http://localhost:8005/health | jq '.' 2>/dev/null || echo "Service not responding"
	@echo ""
	@echo "Projects Service:"
	@curl -s http://localhost:8006/health | jq '.' 2>/dev/null || echo "Service not responding"

test-apis: ## Test all API endpoints
	@echo "Testing ERP Microservices APIs..."
	@echo ""
	@echo "=== Finance Service ==="
	@curl -s http://localhost:8080/api/v1/finance/gl | jq '.'
	@echo ""
	@echo "=== HR Service ==="
	@curl -s http://localhost:8080/api/v1/hr/employees | jq '.'
	@echo ""
	@echo "=== SCM Service ==="
	@curl -s http://localhost:8080/api/v1/scm/products | jq '.'
	@echo ""
	@echo "=== Manufacturing Service ==="
	@curl -s http://localhost:8080/api/v1/manufacturing/boms | jq '.'
	@echo ""
	@echo "=== CRM Service ==="
	@curl -s http://localhost:8080/api/v1/crm/customers | jq '.'
	@echo ""
	@echo "=== Projects Service ==="
	@curl -s http://localhost:8080/api/v1/projects/projects | jq '.'

setup-dev: ## Setup development environment
	@echo "Setting up development environment..."
	@./scripts/setup-dev.sh

# Individual service commands
build-finance: ## Build finance service
	docker-compose build finance-service

build-hr: ## Build HR service
	docker-compose build hr-service

build-scm: ## Build SCM service
	docker-compose build scm-service

build-manufacturing: ## Build manufacturing service
	docker-compose build manufacturing-service

build-crm: ## Build CRM service
	docker-compose build crm-service

build-projects: ## Build projects service
	docker-compose build projects-service

# Infrastructure commands
infra-up: ## Start only infrastructure services
	docker-compose up -d postgres redis rabbitmq

infra-down: ## Stop infrastructure services
	docker-compose stop postgres redis rabbitmq

# Development helpers
go-mod-tidy: ## Run go mod tidy for all services
	@echo "Running go mod tidy for all services..."
	cd shared && go mod tidy
	cd services/finance && go mod tidy
	cd services/hr && go mod tidy
	cd services/scm && go mod tidy
	cd services/manufacturing && go mod tidy
	cd services/crm && go mod tidy
	cd services/projects && go mod tidy
	cd api-gateway && go mod tidy

init-modules: ## Initialize Go modules for all services
	@echo "Initializing Go modules..."
	cd shared && go mod init github.com/your-org/erp-microservices/shared
	cd services/finance && go mod init github.com/your-org/erp-microservices/services/finance
	cd services/hr && go mod init github.com/your-org/erp-microservices/services/hr
	cd services/scm && go mod init github.com/your-org/erp-microservices/services/scm
	cd services/manufacturing && go mod init github.com/your-org/erp-microservices/services/manufacturing
	cd services/crm && go mod init github.com/your-org/erp-microservices/services/crm
	cd services/projects && go mod init github.com/your-org/erp-microservices/services/projects
	cd api-gateway && go mod init github.com/your-org/erp-microservices/api-gateway
