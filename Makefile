# Makefile
.PHONY: help build run stop clean logs health test

help: ## Show help
	@echo "ERP Microservices - Hello World"
	@echo ""
	@echo "Available commands:"
	@echo "  build    - Build all services"
	@echo "  run      - Start all services"
	@echo "  stop     - Stop all services"
	@echo "  clean    - Clean up containers"
	@echo "  logs     - Show logs"
	@echo "  health   - Check service health"
	@echo "  test     - Test Hello World APIs"

build: ## Build all services
	@echo "🔨 Building ERP Microservices..."
	docker compose build

run: ## Start all services
	@echo "🚀 Starting ERP Microservices..."
	docker compose up -d
	@echo ""
	@echo "✅ Services started!"
	@echo "API Gateway: http://localhost:8080"
	@echo "Service Discovery: http://localhost:8080/services"
	@echo ""
	@echo "Individual Services:"
	@echo "  Finance:      http://localhost:8001"
	@echo "  HR:           http://localhost:8002"
	@echo "  SCM:          http://localhost:8003"
	@echo "  Manufacturing: http://localhost:8004"
	@echo "  CRM:          http://localhost:8005"
	@echo "  Projects:     http://localhost:8006"

stop: ## Stop all services
	@echo "🛑 Stopping ERP Microservices..."
	docker compose down

clean: ## Clean up containers and images
	@echo "🧹 Cleaning up..."
	docker compose down --rmi all --volumes --remove-orphans

logs: ## Show logs for all services
	docker compose logs -f

health: ## Check health of all services
	@echo "🏥 Checking service health..."
	@echo ""
	@echo "API Gateway:"
	@curl -s http://localhost:8080/health || echo "❌ Not responding"
	@echo ""
	@echo "Finance Service:"
	@curl -s http://localhost:8001/health || echo "❌ Not responding"
	@echo ""
	@echo "HR Service:"
	@curl -s http://localhost:8002/health || echo "❌ Not responding"
	@echo ""
	@echo "SCM Service:"
	@curl -s http://localhost:8003/health || echo "❌ Not responding"
	@echo ""
	@echo "Manufacturing Service:"
	@curl -s http://localhost:8004/health || echo "❌ Not responding"
	@echo ""
	@echo "CRM Service:"
	@curl -s http://localhost:8005/health || echo "❌ Not responding"
	@echo ""
	@echo "Projects Service:"
	@curl -s http://localhost:8006/health || echo "❌ Not responding"

test: ## Test Hello World APIs
	@echo "🧪 Testing Hello World APIs..."
	@echo ""
	@echo "=== API Gateway ==="
	@curl -s http://localhost:8080/ | jq '.message' || echo "❌ Failed"
	@echo ""
	@echo "=== Finance Service ==="
	@curl -s http://localhost:8080/api/v1/finance/hello | jq '.message' || echo "❌ Failed"
	@echo ""
	@echo "=== HR Service ==="
	@curl -s http://localhost:8080/api/v1/hr/hello | jq '.message' || echo "❌ Failed"
	@echo ""
	@echo "=== SCM Service ==="
	@curl -s http://localhost:8080/api/v1/scm/hello | jq '.message' || echo "❌ Failed"
	@echo ""
	@echo "=== Manufacturing Service ==="
	@curl -s http://localhost:8080/api/v1/manufacturing/hello | jq '.message' || echo "❌ Failed"
	@echo ""
	@echo "=== CRM Service ==="
	@curl -s http://localhost:8080/api/v1/crm/hello | jq '.message' || echo "❌ Failed"
	@echo ""
	@echo "=== Projects Service ==="
	@curl -s http://localhost:8080/api/v1/projects/hello | jq '.message' || echo "❌ Failed"

test-direct: ## Test services directly (bypass gateway)
	@echo "🧪 Testing services directly..."
	@echo ""
	@echo "Finance: " && curl -s http://localhost:8001/ | jq '.message'
	@echo "HR: " && curl -s http://localhost:8002/ | jq '.message'
	@echo "SCM: " && curl -s http://localhost:8003/ | jq '.message'
	@echo "Manufacturing: " && curl -s http://localhost:8004/ | jq '.message'
	@echo "CRM: " && curl -s http://localhost:8005/ | jq '.message'
	@echo "Projects: " && curl -s http://localhost:8006/ | jq '.message'