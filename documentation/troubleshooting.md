# Troubleshooting Guide

This guide helps you quickly resolve common issues when setting up and running the ERP system. Issues are organized by category with step-by-step solutions.

## Quick Diagnosis

Start here to identify your issue category:

### System Won't Start
- [Docker Issues](#docker-issues)
- [Port Conflicts](#port-conflicts)  
- [Database Connection Problems](#database-connection-problems)

### Performance Issues
- [Slow API Responses](#slow-api-responses)
- [Memory Problems](#memory-problems)
- [Database Performance](#database-performance)

### Development Issues
- [Build Failures](#build-failures)
- [Test Failures](#test-failures)
- [Code Generation Issues](#code-generation-issues)

### Production Issues
- [Deployment Failures](#deployment-failures)
- [Service Communication Problems](#service-communication-problems)
- [Monitoring and Logging Issues](#monitoring-and-logging-issues)

---

## Docker Issues

### Problem: Docker Daemon Not Running

**Symptoms:**
```bash
$ make run
Cannot connect to the Docker daemon at unix:///var/run/docker.sock
```

**Solution:**
```bash
# On macOS
open -a Docker

# On Linux (systemd)
sudo systemctl start docker
sudo systemctl enable docker

# On Linux (service)
sudo service docker start

# Verify Docker is running
docker info
```

### Problem: Docker Compose Version Issues

**Symptoms:**
```bash
$ docker-compose --version
docker-compose version 1.29.2
# Services fail to start with networking errors
```

**Solution:**
```bash
# Install Docker Compose V2
sudo apt-get update
sudo apt-get install docker-compose-plugin

# Or use pip
pip install docker-compose>=2.0.0

# Verify version
docker-compose --version
# Should show v2.x.x
```

### Problem: Permission Denied Errors

**Symptoms:**
```bash
$ docker run hello-world
permission denied while trying to connect to Docker daemon
```

**Solution:**
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and log back in, then test
docker run hello-world

# Alternative: Use sudo (not recommended for development)
sudo docker run hello-world
```

---

## Port Conflicts

### Problem: Port Already in Use

**Symptoms:**
```bash
$ make run
ERROR: for api-gateway Cannot start service api-gateway: 
Ports are not available: port 8080 is already allocated
```

**Diagnosis:**
```bash
# Find what's using the port
lsof -i :8080
# Or on some systems
netstat -tulpn | grep :8080
```

**Solutions:**

**Option 1: Kill the conflicting process**
```bash
# Find process ID (PID) from lsof output
kill -9 <PID>

# Example
kill -9 1234
```

**Option 2: Change service ports**
```bash
# Edit docker-compose.yml
services:
  api-gateway:
    ports:
      - "8090:8080"  # Change external port to 8090

# Update your API calls
curl http://localhost:8090/api/v1/finance/hello
```

**Option 3: Stop conflicting services**
```bash
# Stop all Docker containers
docker stop $(docker ps -q)

# Stop specific service
docker-compose stop api-gateway
```

### Problem: Multiple Port Conflicts

**Symptoms:**
Multiple services fail to start due to port conflicts.

**Solution:**
Use our port conflict script:

```bash
#!/bin/bash
# scripts/check-ports.sh

PORTS=(8080 8001 8002 8003 8004 8005 8006 5432 6379 9092)

echo "Checking for port conflicts..."
for port in "${PORTS[@]}"; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        echo "❌ Port $port is in use:"
        lsof -Pi :$port -sTCP:LISTEN
    else
        echo "✅ Port $port is available"
    fi
done
```

Run the script:
```bash
chmod +x scripts/check-ports.sh
./scripts/check-ports.sh
```

---

## Database Connection Problems

### Problem: PostgreSQL Connection Refused

**Symptoms:**
```bash
$ make test
Error connecting to database: dial tcp [::1]:5432: connect: connection refused
```

**Diagnosis:**
```bash
# Check if PostgreSQL container is running
docker-compose ps postgres

# Check container logs
docker-compose logs postgres
```

**Solutions:**

**Solution 1: Start PostgreSQL**
```bash
# Start just the database
docker-compose up -d postgres

# Wait for it to be ready
docker-compose logs -f postgres
# Look for: "database system is ready to accept connections"
```

**Solution 2: Reset database**
```bash
# Stop and remove containers
docker-compose down

# Remove database volume
docker volume rm erp-system_postgres_data

# Start fresh
docker-compose up -d postgres
```

**Solution 3: Check database configuration**
```bash
# Connect to database directly
docker-compose exec postgres psql -U postgres -c "SELECT version();"

# List databases
docker-compose exec postgres psql -U postgres -c "\l"
```

### Problem: Database Migration Failures

**Symptoms:**
```bash
$ cd services/fm-service
$ make migrate-up
Error: migration failed: relation "accounts" already exists
```

**Solution:**
```bash
# Check migration status
make migrate-version

# Force to specific version
make migrate-force version=1

# Re-run migrations
make migrate-up

# Alternative: Reset and re-run
make migrate-down
make migrate-up
```

### Problem: Redis Connection Issues

**Symptoms:**
```bash
Error: dial tcp [::1]:6379: connect: connection refused
```

**Solution:**
```bash
# Start Redis
docker-compose up -d redis

# Test connection
docker-compose exec redis redis-cli ping
# Should return: PONG

# Check Redis logs
docker-compose logs redis
```

---

## Service Communication Problems

### Problem: Services Can't Communicate

**Symptoms:**
```bash
$ make test
Financial Service: ❌ Connection refused
HR Service: ❌ Connection refused
```

**Diagnosis:**
```bash
# Check all services are running
docker-compose ps

# Check service logs
docker-compose logs fm-service
docker-compose logs api-gateway
```

**Solutions:**

**Solution 1: Check service startup order**
```bash
# Services need to start in order: database → services → gateway
docker-compose up -d postgres redis kafka
sleep 10
docker-compose up -d fm-service hr-service scm-service crm-service
sleep 5
docker-compose up -d api-gateway
```

**Solution 2: Verify network connectivity**
```bash
# Test internal service URLs
docker-compose exec api-gateway curl http://fm-service:8001/health

# Check Docker network
docker network ls
docker network inspect erp-system_default
```

**Solution 3: Restart services in correct order**
```bash
# Full restart
make stop
make run

# Or restart specific service
docker-compose restart fm-service
```

---

## Build Failures

### Problem: Go Build Errors

**Symptoms:**
```bash
$ make build
go build: cannot find module providing package
```

**Solutions:**

**Solution 1: Update dependencies**
```bash
cd services/fm-service

# Update go.mod
go mod tidy

# Download dependencies
go mod download

# Verify modules
go mod verify
```

**Solution 2: Clear module cache**
```bash
# Clear Go module cache
go clean -modcache

# Re-download
go mod download
```

**Solution 3: Check Go version**
```bash
# Check Go version
go version
# Should be 1.21 or later

# Update Go if needed (on macOS)
brew install go@1.21
```

### Problem: Docker Build Failures

**Symptoms:**
```bash
$ docker-compose build fm-service
Step 5/10 : RUN go mod download
ERROR: exit status 1
```

**Solution:**
```bash
# Clear Docker build cache
docker system prune -a

# Rebuild with no cache
docker-compose build --no-cache fm-service

# Check Dockerfile syntax
docker build --dry-run -f services/fm-service/Dockerfile services/fm-service/
```

---

## Performance Issues

### Problem: Slow API Responses

**Symptoms:**
API responses taking >5 seconds instead of expected <500ms.

**Diagnosis:**
```bash
# Test API response time
time curl http://localhost:8080/api/v1/finance/accounts

# Check service logs for slow queries
docker-compose logs fm-service | grep -i "slow"
```

**Solutions:**

**Solution 1: Check database performance**
```sql
-- Connect to database
docker-compose exec postgres psql -U postgres -d financial_db

-- Check for missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation 
FROM pg_stats 
WHERE tablename = 'accounts' AND n_distinct > 100;

-- Add missing indexes
CREATE INDEX idx_accounts_type_active ON accounts(account_type, is_active);
```

**Solution 2: Increase service resources**
```yaml
# In docker-compose.yml
services:
  fm-service:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          memory: 1G
```

**Solution 3: Enable caching**
```go
// In service code
func (s *accountService) GetAccount(ctx context.Context, id string) (*Account, error) {
    // Check cache first
    if cached, err := s.cache.Get(ctx, "account:"+id).Result(); err == nil {
        var account Account
        json.Unmarshal([]byte(cached), &account)
        return &account, nil
    }
    
    // Fetch from database
    account, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    data, _ := json.Marshal(account)
    s.cache.Set(ctx, "account:"+id, data, 5*time.Minute)
    
    return account, nil
}
```

---

## Memory Problems

### Problem: Out of Memory Errors

**Symptoms:**
```bash
docker-compose logs fm-service
fatal error: runtime: out of memory
```

**Solutions:**

**Solution 1: Increase container memory**
```yaml
# docker-compose.yml
services:
  fm-service:
    deploy:
      resources:
        limits:
          memory: 2G
    mem_limit: 2g
```

**Solution 2: Fix memory leaks**
```go
// Common memory leak: not closing database connections
func (r *repository) GetAccounts(ctx context.Context) ([]Account, error) {
    rows, err := r.db.QueryContext(ctx, "SELECT * FROM accounts")
    if err != nil {
        return nil, err
    }
    defer rows.Close() // Important: always close rows
    
    var accounts []Account
    for rows.Next() {
        var account Account
        err := rows.Scan(&account.ID, &account.Name)
        if err != nil {
            return nil, err
        }
        accounts = append(accounts, account)
    }
    
    return accounts, rows.Err()
}
```

**Solution 3: Monitor memory usage**
```bash
# Monitor container memory usage
docker stats

# Check Go runtime memory stats
curl http://localhost:8001/debug/pprof/heap
```

---

## Test Failures

### Problem: Unit Tests Failing

**Symptoms:**
```bash
$ make test
FAIL: TestCreateAccount (0.00s)
    account_test.go:25: expected account to be created
```

**Solutions:**

**Solution 1: Check test database**
```bash
# Ensure test database exists
cd services/fm-service
export DB_DATABASE=financial_test_db
make migrate-up

# Run specific test
go test -v ./internal/api/handlers -run TestCreateAccount
```

**Solution 2: Mock dependencies**
```go
// Use testify/mock for testing
func TestAccountService_CreateAccount(t *testing.T) {
    mockRepo := &MockAccountRepository{}
    service := NewAccountService(mockRepo, nil)
    
    expectedAccount := &Account{ID: "1", Name: "Test"}
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedAccount, nil)
    
    result, err := service.CreateAccount(context.Background(), CreateAccountRequest{
        Name: "Test",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, expectedAccount, result)
    mockRepo.AssertExpectations(t)
}
```

---

## Production Issues

### Problem: Service Discovery Failures

**Symptoms:**
Services can't find each other in Kubernetes.

**Solution:**
```yaml
# Ensure services are properly defined
apiVersion: v1
kind: Service
metadata:
  name: financial-service
spec:
  selector:
    app: financial-service
  ports:
    - port: 8001
      targetPort: 8001
---
# Use service names in configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-gateway-config
data:
  FINANCIAL_SERVICE_URL: "http://financial-service:8001"
```

### Problem: LoadBalancer Issues

**Symptoms:**
Uneven request distribution or service timeouts.

**Solution:**
```yaml
# Configure proper health checks
apiVersion: apps/v1
kind: Deployment
metadata:
  name: financial-service
spec:
  template:
    spec:
      containers:
      - name: financial-service
        livenessProbe:
          httpGet:
            path: /health
            port: 8001
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8001
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
```

---

## Emergency Recovery Procedures

### Complete System Reset

When everything is broken, use this nuclear option:

```bash
#!/bin/bash
# scripts/nuclear-reset.sh

echo "🚨 NUCLEAR RESET: This will destroy all data!"
read -p "Are you sure? (yes/no): " confirm

if [ "$confirm" = "yes" ]; then
    echo "Stopping all containers..."
    docker-compose down -v
    
    echo "Removing all containers..."
    docker container prune -f
    
    echo "Removing all images..."
    docker image prune -a -f
    
    echo "Removing all volumes..."
    docker volume prune -f
    
    echo "Removing all networks..."
    docker network prune -f
    
    echo "Starting fresh system..."
    make run
    
    echo "✅ System reset complete!"
else
    echo "Reset cancelled."
fi
```

### Data Recovery

If you've lost data but have backups:

```bash
# Restore database from backup
docker-compose exec postgres pg_restore -U postgres -d financial_db /backup/financial_db.sql

# Restore file data
docker-compose exec file-service cp -r /backup/files/* /app/storage/
```

---

## Getting More Help

### Enable Debug Logging

Add debug logging to any service:

```bash
# Set environment variable
export LOG_LEVEL=debug

# Or in docker-compose.yml
services:
  fm-service:
    environment:
      - LOG_LEVEL=debug
```

### Health Check Commands

Use these commands for systematic diagnosis:

```bash
#!/bin/bash
# scripts/health-check.sh

echo "🔍 ERP System Health Check"
echo "=========================="

# Check Docker
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running"
    exit 1
fi
echo "✅ Docker is running"

# Check containers
echo "📦 Container Status:"
docker-compose ps

# Check ports
echo "🔌 Port Status:"
for port in 8080 8001 8002 8003 8004 8005 8006; do
    if nc -z localhost $port; then
        echo "✅ Port $port is open"
    else
        echo "❌ Port $port is closed"
    fi
done

# Test API endpoints
echo "🌐 API Status:"
for service in finance hr scm crm; do
    if curl -s http://localhost:8080/api/v1/$service/hello >/dev/null; then
        echo "✅ $service API is responding"
    else
        echo "❌ $service API is not responding"
    fi
done

echo "🏁 Health check complete"
```

### Contact and Escalation

1. **Check [FAQ](faq.md)** first
2. **Search existing issues** in the project repository
3. **Collect diagnostic information**:
   ```bash
   # System information
   uname -a
   docker version
   docker-compose version
   go version
   node --version
   
   # Service logs
   docker-compose logs > system-logs.txt
   
   # Configuration
   cat docker-compose.yml
   ```
4. **Create detailed issue report** with:
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details
   - Complete error messages
   - Diagnostic information

Remember: The more detailed your issue report, the faster we can help resolve your problem.

---

**Still having issues?** → Check [📖 FAQ](faq.md)

**Ready to continue development?** → Return to [🏗️ Architecture Overview](architecture-overview.md)