#!/bin/bash
# Script to compile all service CDD contracts and generate domain models & migrations

set -e

# Base directory
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CDD_CLI="${BASE_DIR}/cdd-engine/cdd-cli"

# Ensure cdd-cli is built
echo "🔨 Building CDD CLI engine..."
(cd "${BASE_DIR}/cdd-engine" && go build -o cdd-cli main.go)

# Array of services and their contract name
# format: "service_dir:contract_file"
SERVICES=(
    "auth-service:auth"
    "crm-service:crm"
    "fm-service:fm"
    "hr-service:hr"
    "mfg-service:mfg"
    "prj-service:prj"
    "scm-service:scm"
    "eam-service:eam"
    "plm-service:plm"
    "qms-service:qms"
)

echo "🚀 Starting contract-driven code generation..."
echo ""

for item in "${SERVICES[@]}"; do
    IFS=":" read -r service contract <<< "$item"
    
    cdd_file="${BASE_DIR}/services/${service}/contracts/${contract}.cdd"
    go_out="${BASE_DIR}/services/${service}/internal/business/domain"
    sql_out="${BASE_DIR}/services/${service}/internal/data/migrations"
    
    if [ -f "$cdd_file" ]; then
        echo "=========================================="
        echo "📦 Processing Service: ${service}"
        echo "📄 Contract: ${cdd_file}"
        echo "=========================================="
        
        # Run generator
        "$CDD_CLI" -cdd "$cdd_file" -go-out "$go_out" -sql-out "$sql_out"
        echo ""
    else
        echo "⚠️  Warning: Contract file not found for ${service} at ${cdd_file}"
    fi
done

echo "=========================================="
echo "📝 Generating Unified OpenAPI Specification"
echo "=========================================="
"$CDD_CLI" -openapi-out "${BASE_DIR}/api-gateway/openapi.yaml"
echo ""

echo "✅ Code generation complete!"
