#!/bin/bash

###############################################################################
# ERP System - Secure Secrets Setup Script
# 
# This script generates strong credentials for your ERP system deployment.
# Use this BEFORE your first deployment and for production environments.
#
# Usage:
#   ./scripts/setup-secrets.sh              # Generate with prompts
#   ./scripts/setup-secrets.sh --auto       # Auto-generate all secrets
#   ./scripts/setup-secrets.sh --help       # Show this help message
#
###############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env"
ENV_EXAMPLE="$PROJECT_ROOT/.env.example"

###############################################################################
# Helper Functions
###############################################################################

print_header() {
    echo -e "\n${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC} $1"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Generate random password
generate_password() {
    local length=${1:-32}
    if command -v openssl &> /dev/null; then
        openssl rand -base64 "$length" | tr -d '\n'
    elif command -v python3 &> /dev/null; then
        python3 -c "import secrets; print(secrets.token_urlsafe($length))"
    else
        # Fallback: use /dev/urandom
        head -c "$length" /dev/urandom | base64 | tr -d '\n'
    fi
}

# Generate JWT secret (32+ character hex string)
generate_jwt_secret() {
    if command -v openssl &> /dev/null; then
        openssl rand -hex 32
    elif command -v python3 &> /dev/null; then
        python3 -c "import secrets; print(secrets.token_hex(32))"
    else
        head -c 32 /dev/urandom | base64 | tr -d '\n'
    fi
}

# Validate password strength
validate_password() {
    local password=$1
    local min_length=${2:-16}
    
    if [ ${#password} -lt $min_length ]; then
        print_error "Password must be at least $min_length characters (got ${#password})"
        return 1
    fi
    return 0
}

###############################################################################
# Interactive Setup
###############################################################################

setup_interactive() {
    print_header "ERP System - Interactive Secrets Setup"
    
    print_info "This script will guide you through setting up secure credentials."
    print_info "Keep your .env file secure and never commit it to version control.\n"
    
    # PostgreSQL User
    read -p "PostgreSQL username (default: postgres): " postgres_user
    postgres_user=${postgres_user:-postgres}
    
    # PostgreSQL Password
    while true; do
        read -sp "PostgreSQL password (min 16 chars): " postgres_password
        echo
        if validate_password "$postgres_password" 16; then
            break
        fi
    done
    
    # Redis Password
    while true; do
        read -sp "Redis password (min 16 chars): " redis_password
        echo
        if validate_password "$redis_password" 16; then
            break
        fi
    done
    
    # JWT Secret
    print_info "\nGenerating JWT secret..."
    jwt_secret=$(generate_jwt_secret)
    print_success "JWT secret generated: ${jwt_secret:0:16}..."
    
    # Admin Credentials
    read -p "Admin username (default: admin): " admin_username
    admin_username=${admin_username:-admin}
    
    while true; do
        read -sp "Admin password (min 12 chars): " admin_password
        echo
        if validate_password "$admin_password" 12; then
            break
        fi
    done
    
    read -p "Admin email (default: admin@erp.local): " admin_email
    admin_email=${admin_email:-admin@erp.local}
    
    # Summary
    print_header "Configuration Summary"
    echo "Database:"
    echo "  POSTGRES_USER: $postgres_user"
    echo "  POSTGRES_PASSWORD: ****"
    echo ""
    echo "Cache:"
    echo "  REDIS_PASSWORD: ****"
    echo ""
    echo "Security:"
    echo "  JWT_SECRET: ${jwt_secret:0:16}..."
    echo ""
    echo "Admin User:"
    echo "  ADMIN_USERNAME: $admin_username"
    echo "  ADMIN_PASSWORD: ****"
    echo "  ADMIN_EMAIL: $admin_email"
    
    read -p "Create .env file with these values? (y/n): " confirm
    
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        print_warning "Setup cancelled."
        exit 0
    fi
    
    create_env_file "$postgres_user" "$postgres_password" "$redis_password" \
                    "$jwt_secret" "$admin_username" "$admin_password" "$admin_email"
}

###############################################################################
# Auto Setup (Non-interactive)
###############################################################################

setup_auto() {
    print_header "ERP System - Auto Secrets Generation"
    
    print_info "Generating secure credentials automatically...\n"
    
    # Generate all credentials
    postgres_user="postgres"
    postgres_password=$(generate_password 32)
    redis_password=$(generate_password 32)
    jwt_secret=$(generate_jwt_secret)
    admin_username="admin"
    admin_password=$(generate_password 20)
    admin_email="admin@erp.local"
    
    print_success "PostgreSQL user: $postgres_user"
    print_success "PostgreSQL password: Generated (32 chars)"
    print_success "Redis password: Generated (32 chars)"
    print_success "JWT secret: Generated (64 hex chars)"
    print_success "Admin username: $admin_username"
    print_success "Admin password: Generated (20 chars)"
    print_success "Admin email: $admin_email"
    
    create_env_file "$postgres_user" "$postgres_password" "$redis_password" \
                    "$jwt_secret" "$admin_username" "$admin_password" "$admin_email"
}

###############################################################################
# Create .env File
###############################################################################

create_env_file() {
    local postgres_user=$1
    local postgres_password=$2
    local redis_password=$3
    local jwt_secret=$4
    local admin_username=$5
    local admin_password=$6
    local admin_email=$7
    
    if [ -f "$ENV_FILE" ]; then
        print_warning ".env already exists!"
        read -p "Overwrite existing .env? (y/n): " overwrite
        if [[ "$overwrite" != "y" && "$overwrite" != "Y" ]]; then
            print_info "Keeping existing .env file"
            return
        fi
    fi
    
    # Create .env file with secrets
    cat > "$ENV_FILE" << EOF
# ⚠️  SECURITY WARNING: This file contains sensitive credentials!
# ⚠️  NEVER commit this file to version control!
# ⚠️  Generated on $(date)

# ============================================================================
# DATABASE CREDENTIALS
# ============================================================================
POSTGRES_USER=$postgres_user
POSTGRES_PASSWORD=$postgres_password
POSTGRES_DB=erp_db

# ============================================================================
# MESSAGE QUEUE
# ============================================================================
KAFKA_BROKERS=kafka:9092

# ============================================================================
# CACHE LAYER
# ============================================================================
REDIS_PASSWORD=$redis_password

# ============================================================================
# AUTHENTICATION & SECURITY
# ============================================================================
# JWT secret for token signing (generated via openssl rand)
JWT_SECRET=$jwt_secret

# ============================================================================
# INITIAL ADMIN USER
# ============================================================================
# IMPORTANT: Change these immediately after first login!
ADMIN_USERNAME=$admin_username
ADMIN_PASSWORD=$admin_password
ADMIN_EMAIL=$admin_email

# ============================================================================
# APPLICATION ENVIRONMENT
# ============================================================================
ENVIRONMENT=development
LOG_LEVEL=info

# ============================================================================
# SECURITY CHECKLIST
# ============================================================================
# Production deployment checklist:
# - [ ] Use strong credentials (this script generates them)
# - [ ] Change admin password immediately after first login
# - [ ] Store .env in a secure secrets manager (Vault, AWS Secrets Manager)
# - [ ] Enable TLS/HTTPS for all services
# - [ ] Set up a proper authentication provider (OAuth2, SAML)
# - [ ] Enable rate limiting and DDoS protection
# - [ ] Set up audit logging
# - [ ] Run regular security scans
# - [ ] Implement secrets rotation policy
# ============================================================================
EOF
    
    # Set restrictive permissions
    chmod 600 "$ENV_FILE"
    
    print_success ".env file created with secure credentials"
    print_info "File permissions: 600 (readable only by owner)"
    print_warning "IMPORTANT: Keep this file secure and never commit it!"
}

###############################################################################
# Show Help
###############################################################################

show_help() {
    cat << EOF

${BLUE}ERP System - Secure Secrets Setup${NC}

${BLUE}USAGE:${NC}
    $0 [OPTION]

${BLUE}OPTIONS:${NC}
    --auto          Generate all secrets automatically (non-interactive)
    --interactive   Interactive setup with prompts (default)
    --help          Show this help message
    --validate      Validate existing .env file

${BLUE}EXAMPLES:${NC}
    # Auto-generate all secrets
    $0 --auto

    # Interactive setup
    $0 --interactive

    # Interactive setup (default)
    $0

${BLUE}SECURITY NOTES:${NC}
    • All passwords are generated with high entropy
    • The .env file is created with restrictive permissions (600)
    • Never commit .env to version control (it's in .gitignore)
    • For production, use a dedicated secrets manager
    • Change admin password immediately after first login

${BLUE}PRODUCTION DEPLOYMENT:${NC}
    1. Run this script: ./scripts/setup-secrets.sh --auto
    2. Review the generated .env file
    3. Store it in your secrets manager (Vault, AWS Secrets Manager, etc.)
    4. Use environment variables to inject credentials at runtime
    5. Never hardcode secrets in code or configuration

EOF
}

###############################################################################
# Main
###############################################################################

main() {
    case "${1:-}" in
        --auto)
            setup_auto
            ;;
        --interactive)
            setup_interactive
            ;;
        --validate)
            # TODO: Implement validation
            print_error "Validation not yet implemented"
            exit 1
            ;;
        --help)
            show_help
            exit 0
            ;;
        "")
            setup_interactive
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
    
    print_header "Setup Complete!"
    print_info "Your .env file is ready for use"
    print_info "To start the services:"
    echo ""
    echo "  make build && make run"
    echo ""
    print_warning "Remember to change admin password after first login!"
}

main "$@"
