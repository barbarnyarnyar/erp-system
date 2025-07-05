#!/bin/bash

# Sync utilities script for ERP services
# This script copies shared utilities from templates to all services

echo "🔄 Syncing shared utilities to all services..."

# Define services
SERVICES=("auth-service" "fm-service" "hr-service" "scm-service" "m-service" "crm-service" "pm-service")

# Source directory
SOURCE_DIR="shared/templates/utils"

# Check if source directory exists
if [ ! -d "$SOURCE_DIR" ]; then
    echo "❌ Source directory $SOURCE_DIR not found!"
    exit 1
fi

# Loop through each service
for SERVICE in "${SERVICES[@]}"; do
    TARGET_DIR="services/$SERVICE/utils"
    
    echo "📁 Syncing utils to $SERVICE..."
    
    # Create target directory if it doesn't exist
    mkdir -p "$TARGET_DIR"
    
    # Copy files
    cp "$SOURCE_DIR"/*.go "$TARGET_DIR/"
    
    if [ $? -eq 0 ]; then
        echo "✅ Successfully synced utils to $SERVICE"
    else
        echo "❌ Failed to sync utils to $SERVICE"
    fi
done

echo ""
echo "🎉 Sync complete! All services now have updated utilities."
echo ""
echo "📝 Next steps:"
echo "1. Review the synced files in each service"
echo "2. Run: docker-compose up --build"
echo "3. Test your services"
echo ""
echo "💡 To sync again after making changes to shared/templates/utils/:"
echo "   ./shared/sync-utils.sh"