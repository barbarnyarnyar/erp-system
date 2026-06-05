# Development Environment

Set up your local development environment for efficient coding and testing.

## IDE Configuration

### Visual Studio Code (Recommended)
Install essential extensions for Go and ERP development:

```bash
# Install VS Code extensions
code --install-extension golang.Go
code --install-extension ms-vscode.vscode-typescript-next
code --install-extension bradlc.vscode-tailwindcss
code --install-extension ms-vscode.vscode-json
```

**Workspace Settings (`.vscode/settings.json`):**
```json
{
  "go.gopath": "${workspaceFolder}",
  "go.goroot": "/usr/local/go",
  "go.toolsGopath": "${workspaceFolder}/.vscode/go-tools",
  "go.useCodeSnippetsOnFunctionSuggest": true,
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

### Go Module Setup
Initialize Go modules for each service:

```bash
# Initialize Go modules for each service if not done
for service in fm hr scm crm mfg pm; do
  cd services/${service}-service
  go mod tidy
  go mod download
  cd ../..
done
```

## Local Development Without Docker

For faster development cycles, run services locally while keeping infrastructure in Docker:

### Step 1: Start Infrastructure Only
```bash
# Start only databases and message queue
docker-compose up -d postgres redis kafka
```

### Step 2: Run Services Locally
```bash
# Terminal 1: Financial Service
cd services/fm-service
export DB_HOST=localhost
export REDIS_HOST=localhost
export KAFKA_BROKERS=localhost:9092
go run cmd/server/main.go

# Terminal 2: HR Service
cd services/hr-service
export DB_HOST=localhost
export DB_PORT=5433  # Different database
go run cmd/main.go

# Continue for other services...
```

### Step 3: API Gateway Configuration
```bash
# Update API Gateway to point to localhost services
cd api-gateway
export FM_SERVICE_URL=http://localhost:8001
export HR_SERVICE_URL=http://localhost:8002
export SCM_SERVICE_URL=http://localhost:8003
export CRM_SERVICE_URL=http://localhost:8004
export MFG_SERVICE_URL=http://localhost:8005
export PM_SERVICE_URL=http://localhost:8006
go run main.go
```

## Hot Reload Development

Enable hot reload for faster development cycles:

### Using Air (Go Hot Reload)
```bash
# Install air globally
go install github.com/cosmtrek/air@latest

# Run service with hot reload
cd services/fm-service
air
```

**Air Configuration (`.air.toml`):**
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main cmd/server/main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

## Database Development Tools

### Direct Database Access
```bash
# Connect to PostgreSQL directly
docker-compose exec postgres psql -U postgres

# Connect to specific service database
docker-compose exec postgres psql -U postgres -d financial_db

# Run SQL queries
docker-compose exec postgres psql -U postgres -c "SELECT * FROM accounts LIMIT 5;"
```

### Database Migrations
```bash
# Create new migration
cd services/fm-service
make migrate-create name=add_customer_table

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-version
```

## Code Quality Tools

### Pre-commit Hooks
```bash
# Install pre-commit hooks (optional)
pip install pre-commit
pre-commit install

# Run manually
pre-commit run --all-files
```

**.pre-commit-config.yaml:**
```yaml
repos:
- repo: https://github.com/golangci/golangci-lint
  rev: v1.54.2
  hooks:
    - id: golangci-lint
- repo: https://github.com/pre-commit/pre-commit-hooks
  rev: v4.4.0
  hooks:
    - id: trailing-whitespace
    - id: end-of-file-fixer
    - id: check-yaml
    - id: check-added-large-files
```

### Go Linting and Formatting
```bash
# Run linter on all services
make lint-all

# Run linter on specific service
cd services/fm-service
make lint

# Format code
gofmt -w .
go mod tidy
```

## Development Scripts

### Quick Development Setup
Create a development script for rapid setup:

```bash
#!/bin/bash
# scripts/dev-setup.sh

echo "Setting up development environment..."

# Start infrastructure
docker-compose up -d postgres redis kafka
sleep 20

# Run migrations
for service in fm hr scm crm mfg pm; do
  cd services/${service}-service
  make migrate-up
  cd ../..
done

# Start services with hot reload
echo "Development environment ready!"
echo "Start your services with: cd services/fm-service && air"
```

### Development Workflow Script
```bash
#!/bin/bash
# scripts/dev-workflow.sh

SERVICE=$1
if [ -z "$SERVICE" ]; then
  echo "Usage: $0 <service-name>"
  echo "Available services: fm, hr, scm, crm, mfg, pm"
  exit 1
fi

echo "Starting development workflow for $SERVICE service..."

# Navigate to service directory
cd services/${SERVICE}-service

# Install dependencies
go mod tidy

# Run tests
make test

# Run linting
make lint

# Start with hot reload
echo "Starting service with hot reload..."
air
```

## Testing in Development

### Unit Testing
```bash
# Run tests for all services
make test-all

# Run tests for specific service
cd services/fm-service
make test

# Run tests with coverage
make test-coverage
```

### Integration Testing
```bash
# Run integration tests
make test-integration

# Test API endpoints manually
curl http://localhost:8080/api/v1/finance/accounts
```

## Debugging

### Go Debugging with Delve
```bash
# Install Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Start service with debugger
cd services/fm-service
dlv debug cmd/server/main.go
```

### VS Code Debugging
Add debug configuration to `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug FM Service",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/services/fm-service/cmd/server/main.go",
      "env": {
        "DB_HOST": "localhost",
        "REDIS_HOST": "localhost"
      }
    }
  ]
}
```

## Performance Optimization

### Development Performance Tips
- Use local services instead of Docker for active development
- Enable Go module proxy: `export GOPROXY=https://proxy.golang.org,direct`
- Use `air` for hot reloading to avoid manual restarts
- Run infrastructure services only in Docker
- Use `make` commands for consistent workflows

## Next Steps

With your development environment set up:
- [Development Workflow](development-workflow.md) - Learn daily development practices
- [Testing and Verification](testing-verification.md) - Ensure code quality
- [System Architecture](../architecture/README.md) - Understand the codebase structure