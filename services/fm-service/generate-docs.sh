#!/bin/bash

# Generate Swagger documentation for FM Service

echo "🔄 Generating Swagger documentation for FM Service..."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "📦 Installing swag CLI..."
    go install github.com/swaggo/swag/cmd/swag@v1.16.1
    
    # Add GOPATH/bin to PATH if not already there
    export PATH=$PATH:$(go env GOPATH)/bin
    
    # Check again if swag is available
    if ! command -v swag &> /dev/null; then
        echo "❌ Failed to install swag. Please install manually:"
        echo "   go install github.com/swaggo/swag/cmd/swag@v1.16.1"
        echo "   export PATH=\$PATH:\$(go env GOPATH)/bin"
        exit 1
    fi
fi

# Ensure we're in the right directory
if [ ! -f "cmd/main.go" ]; then
    echo "❌ Please run this script from the fm-service directory"
    echo "Current directory: $(pwd)"
    exit 1
fi

# Generate docs
echo "📚 Generating docs..."
swag init -g cmd/main.go -o docs/

if [ $? -eq 0 ]; then
    echo "✅ Swagger documentation generated successfully!"
    
    # Run go mod tidy to fix any missing dependencies
    echo "🔧 Running go mod tidy..."
    go mod tidy
    
    echo "📖 Documentation will be available at: http://localhost:8001/swagger/index.html"
    echo ""
    echo "📝 Next steps:"
    echo "1. Run: docker-compose up --build fm-service"
    echo "2. Open: http://localhost:8001/swagger/index.html"
else
    echo "❌ Failed to generate Swagger documentation"
    exit 1
fi