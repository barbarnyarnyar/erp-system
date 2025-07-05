#!/bin/bash

# Sync utilities to a single service
# Usage: ./shared/sync-single.sh SERVICE_NAME

if [ $# -eq 0 ]; then
    echo "❌ Please provide a service name"
    echo "Usage: ./shared/sync-single.sh SERVICE_NAME"
    echo "Example: ./shared/sync-single.sh auth-service"
    exit 1
fi

SERVICE=$1
SOURCE_DIR="shared/templates/utils"
TARGET_DIR="services/$SERVICE/utils"

echo "🔄 Syncing shared utilities to $SERVICE..."

# Check if source directory exists
if [ ! -d "$SOURCE_DIR" ]; then
    echo "❌ Source directory $SOURCE_DIR not found!"
    exit 1
fi

# Check if service directory exists
if [ ! -d "services/$SERVICE" ]; then
    echo "❌ Service directory services/$SERVICE not found!"
    exit 1
fi

# Create target directory if it doesn't exist
mkdir -p "$TARGET_DIR"

# Copy files
cp "$SOURCE_DIR"/*.go "$TARGET_DIR/"

if [ $? -eq 0 ]; then
    echo "✅ Successfully synced utils to $SERVICE"
    echo ""
    echo "📝 Next steps:"
    echo "1. Review the synced files in services/$SERVICE/utils/"
    echo "2. Update services/$SERVICE/cmd/main.go to import \"$SERVICE/utils\""
    echo "3. Run: cd services/$SERVICE && go mod tidy"
    echo "4. Test: docker-compose up --build $SERVICE"
else
    echo "❌ Failed to sync utils to $SERVICE"
    exit 1
fi