#!/bin/bash

# Setup Swagger for a specific service
# Usage: ./shared/setup-swagger.sh SERVICE_NAME

if [ $# -eq 0 ]; then
    echo "❌ Please provide a service name"
    echo "Usage: ./shared/setup-swagger.sh SERVICE_NAME"
    echo "Available services: auth-service, fm-service, hr-service, scm-service, m-service, crm-service, pm-service"
    exit 1
fi

SERVICE_NAME=$1

# Define service configurations
declare -A SERVICE_CONFIG
SERVICE_CONFIG["auth-service"]="Auth Service,8000,auth"
SERVICE_CONFIG["fm-service"]="Financial Management Service,8001,fm"
SERVICE_CONFIG["hr-service"]="HR Service,8002,hr"
SERVICE_CONFIG["scm-service"]="Supply Chain Management Service,8003,scm"
SERVICE_CONFIG["m-service"]="Manufacturing Service,8004,manufacturing"
SERVICE_CONFIG["crm-service"]="CRM Service,8005,crm"
SERVICE_CONFIG["pm-service"]="Project Management Service,8006,pm"

# Check if service exists
if [[ ! -v SERVICE_CONFIG[$SERVICE_NAME] ]]; then
    echo "❌ Unknown service: $SERVICE_NAME"
    echo "Available services: ${!SERVICE_CONFIG[@]}"
    exit 1
fi

# Parse service configuration
IFS=',' read -r SERVICE_TITLE PORT API_PREFIX <<< "${SERVICE_CONFIG[$SERVICE_NAME]}"

echo "🔄 Setting up Swagger for $SERVICE_NAME..."
echo "   Title: $SERVICE_TITLE"
echo "   Port: $PORT"
echo "   API Prefix: $API_PREFIX"

SERVICE_DIR="services/$SERVICE_NAME"

# Check if service directory exists
if [ ! -d "$SERVICE_DIR" ]; then
    echo "❌ Service directory $SERVICE_DIR not found!"
    exit 1
fi

# 1. Update go.mod to include Swagger dependencies
echo "📦 Updating go.mod..."
cat > "$SERVICE_DIR/go.mod" << EOF
module $SERVICE_NAME

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.1
)
EOF

# 2. Generate main.go from template
echo "📝 Generating main.go with Swagger..."
TEMPLATE_FILE="shared/templates/swagger-main.go.template"

if [ ! -f "$TEMPLATE_FILE" ]; then
    echo "❌ Template file $TEMPLATE_FILE not found!"
    exit 1
fi

# Replace template variables
sed "s/{{SERVICE_NAME}}/$SERVICE_NAME/g; s/{{SERVICE_TITLE}}/$SERVICE_TITLE/g; s/{{PORT}}/$PORT/g; s/{{API_PREFIX}}/$API_PREFIX/g" \
    "$TEMPLATE_FILE" > "$SERVICE_DIR/cmd/main.go"

# 3. Create generate-docs script
echo "📚 Creating documentation generation script..."
cat > "$SERVICE_DIR/generate-docs.sh" << EOF
#!/bin/bash

# Generate Swagger documentation for $SERVICE_TITLE

echo "🔄 Generating Swagger documentation for $SERVICE_TITLE..."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "📦 Installing swag CLI..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate docs
echo "📚 Generating docs..."
swag init -g cmd/main.go -o docs/

if [ \$? -eq 0 ]; then
    echo "✅ Swagger documentation generated successfully!"
    echo "📖 Documentation will be available at: http://localhost:$PORT/swagger/index.html"
    echo ""
    echo "📝 Next steps:"
    echo "1. Run: go mod tidy"
    echo "2. Run: docker-compose up --build $SERVICE_NAME"
    echo "3. Open: http://localhost:$PORT/swagger/index.html"
else
    echo "❌ Failed to generate Swagger documentation"
    exit 1
fi
EOF

chmod +x "$SERVICE_DIR/generate-docs.sh"

# 4. Update Dockerfile to include Swagger generation
echo "🐳 Updating Dockerfile..."
cat > "$SERVICE_DIR/Dockerfile" << EOF
FROM golang:1.21-alpine AS builder

# Install git (needed for go mod download)
RUN apk add --no-cache git

WORKDIR /app

# Install swag CLI for generating docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger docs
RUN swag init -g cmd/main.go -o docs/

# Build the application
RUN go build -o main ./cmd

FROM alpine:latest
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE $PORT

# Run the application
CMD ["./main"]
EOF

echo ""
echo "✅ Swagger setup complete for $SERVICE_NAME!"
echo ""
echo "📝 Next steps:"
echo "1. cd $SERVICE_DIR"
echo "2. ./generate-docs.sh"
echo "3. go mod tidy"
echo "4. docker-compose up --build $SERVICE_NAME"
echo "5. Open: http://localhost:$PORT/swagger/index.html"
echo ""
echo "🎯 Quick test:"
echo "   curl http://localhost:$PORT/api/$API_PREFIX/hello"