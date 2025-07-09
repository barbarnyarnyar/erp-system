#!/bin/bash

# Setup Swagger for all services

echo "🔄 Setting up Swagger for all ERP services..."

# Define all services
SERVICES=("auth-service" "fm-service" "hr-service" "scm-service" "m-service" "crm-service" "pm-service")

# Setup each service
for SERVICE in "${SERVICES[@]}"; do
    echo ""
    echo "🔧 Setting up $SERVICE..."
    ./shared/setup-swagger.sh "$SERVICE"
    
    if [ $? -eq 0 ]; then
        echo "✅ $SERVICE setup complete"
    else
        echo "❌ Failed to setup $SERVICE"
    fi
done

echo ""
echo "🎉 Swagger setup complete for all services!"
echo ""
echo "📝 Next steps:"
echo "1. Generate docs for each service:"
echo "   cd services/auth-service && ./generate-docs.sh && cd ../.."
echo "   cd services/fm-service && ./generate-docs.sh && cd ../.."
echo "   # ... repeat for all services"
echo ""
echo "2. Or use this one-liner to generate all docs:"
echo '   for service in auth-service fm-service hr-service scm-service m-service crm-service pm-service; do cd services/$service && ./generate-docs.sh && cd ../..; done'
echo ""
echo "3. Run all services:"
echo "   docker-compose up --build"
echo ""
echo "📖 Access Swagger documentation:"
echo "   Auth Service:    http://localhost:8000/swagger/index.html"
echo "   FM Service:      http://localhost:8001/swagger/index.html"
echo "   HR Service:      http://localhost:8002/swagger/index.html"
echo "   SCM Service:     http://localhost:8003/swagger/index.html"
echo "   Manufacturing:   http://localhost:8004/swagger/index.html"
echo "   CRM Service:     http://localhost:8005/swagger/index.html"
echo "   PM Service:      http://localhost:8006/swagger/index.html"