# Technology Stack

Detailed overview of technologies, frameworks, and tools used in the ERP system.

## Backend Technologies

### Go Programming Language
**Version**: 1.21+
**Why Go**: 
- Excellent performance for concurrent operations
- Strong standard library for web services
- Simple deployment with single binary
- Great tooling and testing support
- Strong typing and memory safety

**Key Libraries**:
- `gin-gonic/gin` - HTTP web framework
- `gorm.io/gorm` - ORM for database operations  
- `golang-jwt/jwt` - JWT token handling
- `go-redis/redis` - Redis client
- `shopspring/decimal` - Precise decimal arithmetic
- `google/uuid` - UUID generation

### Web Framework: Gin
**Purpose**: HTTP routing and middleware
**Features**:
- Fast HTTP router with radix tree
- Middleware support for cross-cutting concerns
- JSON binding and validation
- HTTP/2 support
- Minimal memory footprint

**Example Usage**:
```go
r := gin.Default()
r.Use(AuthMiddleware())
r.GET("/accounts", accountHandler.GetAccounts)
r.POST("/accounts", accountHandler.CreateAccount)
```

## Database Technologies

### PostgreSQL
**Version**: 15+
**Why PostgreSQL**:
- ACID compliance for financial data
- Rich data types including JSON
- Excellent performance for complex queries
- Strong consistency guarantees
- Extensive indexing capabilities

**Configuration**:
- Separate database per service
- Connection pooling with PgBouncer
- Read replicas for reporting
- Automated backups with WAL archiving

### Database Per Service Pattern
Each service has its own database:
- `financial_db` - Financial Management
- `hr_db` - Human Resources  
- `scm_db` - Supply Chain Management
- `crm_db` - Customer Relations
- `manufacturing_db` - Manufacturing
- `project_db` - Project Management

## Caching and Session Management

### Redis
**Version**: 7+
**Why Redis**:
- In-memory performance for caching
- Rich data structures (strings, hashes, sets, lists)
- Pub/sub capabilities for real-time features
- Persistence options for durability
- Clustering support for high availability

**Usage Patterns**:
```go
// Caching
cache.Set("account:123", accountData, 15*time.Minute)

// Session storage  
cache.HSet("session:abc", "user_id", "user123")

// Rate limiting
cache.Incr("rate_limit:user123")
```

## Message Queue and Event Streaming

### Apache Kafka
**Why Kafka**:
- High-throughput event streaming
- Durable message storage
- Horizontal scalability
- Stream processing capabilities
- Strong ordering guarantees

**Event Patterns**:
```go
// Domain events
type AccountCreatedEvent struct {
    AccountID   string    `json:"account_id"`
    AccountCode string    `json:"account_code"`
    CreatedAt   time.Time `json:"created_at"`
}

// Publishing events
producer.Publish("finance.account.created", event)
```

## API and Communication

### REST API Standards
**Standards Used**:
- RESTful resource design
- JSON request/response format
- HTTP status codes for error handling
- OpenAPI 3.0 specification
- Consistent URL patterns

**URL Patterns**:
```
GET    /api/v1/finance/accounts
POST   /api/v1/finance/accounts
GET    /api/v1/finance/accounts/{id}
PUT    /api/v1/finance/accounts/{id}
DELETE /api/v1/finance/accounts/{id}
```

### Authentication: JWT
**Token Structure**:
```json
{
  "iss": "erp-system",
  "sub": "user123", 
  "aud": "erp-api",
  "exp": 1640995200,
  "roles": ["finance_user", "hr_viewer"],
  "permissions": ["accounts:read", "accounts:write"]
}
```

## Frontend Technologies

### Web Application
**Framework**: React 18+ with TypeScript
**Why React + TypeScript**:
- Component-based architecture
- Strong typing for better development experience
- Large ecosystem and community
- Server-side rendering support
- Excellent tooling and debugging

**Key Libraries**:
- `@mui/material` - Material-UI components
- `react-query` - Server state management
- `react-hook-form` - Form handling
- `recharts` - Data visualization
- `axios` - HTTP client

### Mobile Application
**Framework**: React Native
**Why React Native**:
- Code sharing between iOS and Android
- Leverages existing React skills
- Native performance with JavaScript
- Hot reloading for fast development

## Containerization and Orchestration

### Docker
**Base Images**:
- `golang:1.21-alpine` - Go services
- `postgres:15-alpine` - Database
- `redis:7-alpine` - Cache
- `node:18-alpine` - Frontend build

**Multi-stage Build Example**:
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/server/main.go

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/main /main
EXPOSE 8001
CMD ["./main"]
```

### Kubernetes
**Deployment Strategy**:
- Deployment objects for stateless services
- StatefulSets for databases
- Services for internal communication
- Ingress for external access
- ConfigMaps and Secrets for configuration

## Development and Build Tools

### Go Toolchain
- `go mod` - Dependency management
- `go test` - Testing framework
- `gofmt` - Code formatting
- `golangci-lint` - Code linting
- `air` - Hot reloading for development

### Build Automation
**Make**: Task runner for common operations
```makefile
.PHONY: build test lint

build:
	go build -o bin/main cmd/server/main.go

test:
	go test ./...

lint:
	golangci-lint run
```

## Monitoring and Observability

### Metrics: Prometheus
**Why Prometheus**:
- Multi-dimensional data model
- Powerful query language (PromQL)
- Service discovery integration
- Alerting capabilities
- Grafana integration

**Custom Metrics**:
```go
var (
    httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

### Visualization: Grafana
**Dashboards**:
- Application performance metrics
- Business KPIs and analytics
- Infrastructure monitoring
- Error tracking and alerting

### Logging: Structured Logging
**Format**: JSON for production, text for development
```go
logger.WithFields(logrus.Fields{
    "user_id": "user123",
    "account_id": "acc456", 
    "operation": "create_account",
}).Info("Account created successfully")
```

## Security Technologies

### Encryption
- **TLS 1.2+** for all HTTP communication
- **AES-256** for data at rest encryption
- **bcrypt** for password hashing
- **HMAC-SHA256** for JWT signing

### Security Headers
```go
// Security middleware
r.Use(func(c *gin.Context) {
    c.Header("X-Frame-Options", "DENY")
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("X-XSS-Protection", "1; mode=block")
    c.Header("Strict-Transport-Security", "max-age=31536000")
    c.Next()
})
```

## Testing Technologies

### Testing Frameworks
- **Go testing** - Built-in testing
- **Testify** - Assertion library
- **Mockery** - Mock generation
- **Docker Compose** - Integration testing

### Testing Strategy
```go
// Unit test example
func TestAccountService_CreateAccount(t *testing.T) {
    // Arrange
    mockRepo := &MockAccountRepository{}
    service := NewAccountService(mockRepo)
    
    // Act & Assert
    account, err := service.CreateAccount(ctx, request)
    assert.NoError(t, err)
    assert.Equal(t, "1000", account.Code)
}
```

## CI/CD Technologies

### GitHub Actions
**Pipeline Stages**:
1. Code checkout
2. Dependency installation
3. Linting and formatting
4. Unit testing with coverage
5. Integration testing
6. Security scanning
7. Docker image building
8. Deployment to staging

## Cloud and Infrastructure

### Supported Platforms
- **Kubernetes** - Primary orchestration platform
- **Docker Swarm** - Lightweight alternative
- **AWS ECS** - Managed container service
- **Cloud Foundry** - Platform-as-a-Service

### Infrastructure as Code
- **Terraform** - Infrastructure provisioning
- **Helm** - Kubernetes package management
- **Docker Compose** - Local development

## Development Environment

### Required Tools
- Go 1.21+
- Docker and Docker Compose
- Git
- Make
- Node.js 18+ (for frontend)

### Recommended IDE Setup
**VS Code Extensions**:
- Go extension
- Docker extension  
- GitLens
- REST Client
- Kubernetes extension

## Performance and Optimization

### Performance Tools
- **pprof** - Go profiling
- **Apache Bench** - Load testing
- **wrk** - HTTP benchmarking
- **Grafana** - Performance monitoring

### Optimization Techniques
- Connection pooling for databases
- HTTP/2 for client communication
- Gzip compression for responses
- CDN for static assets
- Database query optimization

## Next Steps

Learn more about specific architectural aspects:
- [Microservices Architecture](microservices-architecture.md) - Service design patterns
- [Database Design](database-design.md) - Data modeling approach
- [Security Architecture](security-architecture.md) - Security implementation details