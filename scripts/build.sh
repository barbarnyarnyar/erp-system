# scripts/build.sh
#!/bin/bash

echo "🔨 Building ERP Microservices..."

# Build shared module first
echo "📦 Building shared module..."
cd shared && go mod tidy && cd ..

# Build each service
services=(
	"auth-service"
	"crm-service"
	"eam-service"
	"fm-service"
	"hr-service"
	"mfg-service"
	"plm-service"
	"prj-service"
	"qms-service"
	"scm-service"
)

for service in "${services[@]}"; do
	echo "Resuming building $service service..."
	cd "services/$service"
	go mod tidy
	go build -o bin/main cmd/main.go
	cd ../..
done

# Build API Gateway
echo "🔨 Building API Gateway..."
cd api-gateway
go mod tidy
go build -o bin/main cmd/main.go
cd ..

echo "✅ All services built successfully!"