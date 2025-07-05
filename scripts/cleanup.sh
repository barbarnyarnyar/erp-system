# scripts/cleanup.sh
#!/bin/bash

echo "🧹 Cleaning up ERP Microservices..."

# Stop and remove containers
docker-compose down -v --rmi all --remove-orphans

# Remove unused Docker resources
docker system prune -af --volumes

# Remove any dangling images
docker image prune -af

echo "✅ Cleanup completed!"
