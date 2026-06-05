# Microservices Architecture

Design patterns and principles for the ERP system's microservices architecture.

## Service Design Principles

### Single Responsibility
Each service has a single, well-defined business responsibility:
- **Financial Service**: Accounting and financial operations
- **HR Service**: Employee and payroll management  
- **SCM Service**: Inventory and procurement
- **CRM Service**: Customer relationships and sales
- **Manufacturing Service**: Production planning and execution
- **Project Service**: Project management and billing

### Autonomy and Independence  
Services are designed for maximum autonomy:
- **Independent deployment**: Each service can be deployed separately
- **Independent data**: Each service owns its data and database
- **Independent scaling**: Services scale based on their own load
- **Independent technology**: Services can use different tech stacks if needed

## Service Structure

### Clean Architecture Pattern
Each service follows Clean Architecture principles:

```
services/{service-name}/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/           # HTTP request handlers
│   │   ├── middleware/         # HTTP middleware
│   │   └── routes/            # Route definitions
│   ├── business/
│   │   ├── domain/            # Domain models and business logic
│   │   └── services/          # Business services
│   ├── data/
│   │   ├── repositories/      # Data access layer
│   │   └── migrations/        # Database migrations
│   └── config/               # Service configuration
├── pkg/                      # Public interfaces
├── go.mod
└── Dockerfile
```

### Dependency Direction
Dependencies flow inward following Clean Architecture:
```
API → Business → Data
```

**Example Implementation**:
```go
// Handler depends on business service
type AccountHandler struct {
    accountService *business.AccountService
}

// Business service depends on repository interface
type AccountService struct {
    accountRepo domain.AccountRepository  // Interface
}

// Repository implements the interface
type PostgresAccountRepository struct {
    db *gorm.DB
}
```

## Inter-Service Communication

### Synchronous Communication
**When to Use**: Immediate response needed, strong consistency required

**Pattern**: HTTP/REST APIs with circuit breaker
```go
// Circuit breaker for resilience
type ServiceClient struct {
    httpClient *http.Client
    breaker    *circuit.Breaker
}

func (c *ServiceClient) GetCustomer(id string) (*Customer, error) {
    result, err := c.breaker.Execute(func() (interface{}, error) {
        return c.httpClient.Get(fmt.Sprintf("/customers/%s", id))
    })
    
    if err != nil {
        return nil, err
    }
    
    return result.(*Customer), nil
}
```

### Asynchronous Communication
**When to Use**: Eventual consistency acceptable, decoupling needed

**Pattern**: Event-driven architecture with Kafka
```go
// Publishing domain events
type EventPublisher struct {
    producer kafka.Producer
}

func (e *EventPublisher) PublishAccountCreated(account *Account) error {
    event := AccountCreatedEvent{
        AccountID:   account.ID,
        AccountCode: account.Code,
        CreatedAt:   time.Now(),
    }
    
    return e.producer.Produce("finance.account.created", event)
}

// Consuming events
func (h *CustomerEventHandler) HandleAccountCreated(event AccountCreatedEvent) error {
    // Update customer credit limit based on new account
    return h.customerService.UpdateCreditLimit(event.AccountID)
}
```

## Data Management Patterns

### Database Per Service
Each service owns its data completely:

```go
// Financial service database
type FinancialDB struct {
    Accounts        []Account
    JournalEntries  []JournalEntry
    Transactions    []Transaction
}

// HR service database  
type HRDB struct {
    Employees   []Employee
    Payroll     []PayrollRecord
    Benefits    []Benefit
}
```

### Data Consistency Patterns

**Saga Pattern** for distributed transactions:
```go
type OrderProcessingSaga struct {
    steps []SagaStep
}

type SagaStep struct {
    Execute    func() error
    Compensate func() error
}

func (s *OrderProcessingSaga) Execute() error {
    for i, step := range s.steps {
        if err := step.Execute(); err != nil {
            // Compensate previous steps
            for j := i - 1; j >= 0; j-- {
                s.steps[j].Compensate()
            }
            return err
        }
    }
    return nil
}
```

**Event Sourcing** for audit trails:
```go
type EventStore struct {
    events []DomainEvent
}

type AccountAggregate struct {
    ID      string
    Balance decimal.Decimal
    Version int
}

func (a *AccountAggregate) Debit(amount decimal.Decimal) *AccountDebitedEvent {
    event := &AccountDebitedEvent{
        AccountID: a.ID,
        Amount:    amount,
        Timestamp: time.Now(),
    }
    
    a.Balance = a.Balance.Sub(amount)
    a.Version++
    
    return event
}
```

## Service Discovery and Load Balancing

### Service Registry Pattern
Services register themselves and discover others:

```go
type ServiceRegistry interface {
    Register(service ServiceInfo) error
    Discover(serviceName string) ([]ServiceInstance, error)
    Deregister(serviceID string) error
}

type ServiceInfo struct {
    ID       string
    Name     string
    Address  string
    Port     int
    Health   string
    Metadata map[string]string
}
```

### Load Balancing Strategies
**Round Robin**: Default strategy for even distribution
**Least Connections**: Route to service with fewest active connections  
**Health-based**: Only route to healthy service instances

## Resilience Patterns

### Circuit Breaker Pattern
Prevent cascading failures:

```go
type CircuitBreaker struct {
    state      State
    failures   int
    threshold  int
    timeout    time.Duration
    lastFailure time.Time
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    switch cb.state {
    case Open:
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = HalfOpen
        } else {
            return ErrCircuitBreakerOpen
        }
    case HalfOpen:
        err := fn()
        if err != nil {
            cb.state = Open
            cb.lastFailure = time.Now()
            return err
        }
        cb.state = Closed
        cb.failures = 0
        return nil
    case Closed:
        err := fn()
        if err != nil {
            cb.failures++
            if cb.failures >= cb.threshold {
                cb.state = Open
                cb.lastFailure = time.Now()
            }
        }
        return err
    }
    return nil
}
```

### Retry Pattern with Exponential Backoff
```go
func RetryWithBackoff(fn func() error, maxRetries int) error {
    var err error
    
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(backoff)
    }
    
    return err
}
```

### Bulkhead Pattern
Isolate resources to prevent system-wide failures:

```go
type ResourcePool struct {
    dbPool    *sql.DB  // Database connections
    httpPool  *http.Client  // HTTP connections  
    cachePool *redis.Client  // Cache connections
}

// Separate pools prevent one resource from exhausting others
func (r *ResourcePool) GetDBConnection() *sql.DB {
    return r.dbPool  // Limited to 25 connections
}

func (r *ResourcePool) GetHTTPClient() *http.Client {
    return r.httpPool  // Limited to 50 connections
}
```

## API Gateway Patterns

### Request Routing
Route requests to appropriate services:

```go
type RouteConfig struct {
    Pattern string
    Service string
    Method  string
}

var routes = []RouteConfig{
    {"/api/v1/finance/*", "financial-service", "ANY"},
    {"/api/v1/hr/*", "hr-service", "ANY"},
    {"/api/v1/scm/*", "scm-service", "ANY"},
}

func (g *Gateway) routeRequest(r *http.Request) (string, error) {
    for _, route := range routes {
        if matched, _ := path.Match(route.Pattern, r.URL.Path); matched {
            return route.Service, nil
        }
    }
    return "", ErrNoRouteFound
}
```

### Request Transformation
Transform requests between client and service formats:

```go
type RequestTransformer interface {
    Transform(req *http.Request) (*http.Request, error)
}

type HeaderTransformer struct{}

func (t *HeaderTransformer) Transform(req *http.Request) (*http.Request, error) {
    // Add service-specific headers
    req.Header.Set("X-Service-Version", "v1.0.0")
    req.Header.Set("X-Request-ID", uuid.New().String())
    return req, nil
}
```

## Configuration Management

### Environment-Based Configuration
```go
type ServiceConfig struct {
    Port        int    `env:"PORT" envDefault:"8001"`
    DBHost      string `env:"DB_HOST" envDefault:"localhost"`
    DBPort      int    `env:"DB_PORT" envDefault:"5432"`
    RedisHost   string `env:"REDIS_HOST" envDefault:"localhost"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}

func LoadConfig() *ServiceConfig {
    cfg := &ServiceConfig{}
    env.Parse(cfg)
    return cfg
}
```

### Feature Flags
Control feature rollouts:

```go
type FeatureFlag struct {
    Name    string
    Enabled bool
    Rules   []Rule
}

func (f *FeatureFlag) IsEnabled(context map[string]interface{}) bool {
    if !f.Enabled {
        return false
    }
    
    for _, rule := range f.Rules {
        if !rule.Evaluate(context) {
            return false
        }
    }
    
    return true
}
```

## Testing Microservices

### Contract Testing
Ensure service contracts are maintained:

```go
// Consumer contract test
func TestCustomerServiceContract(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Expected request format
        assert.Equal(t, "GET", r.Method)
        assert.Equal(t, "/customers/123", r.URL.Path)
        
        // Expected response format
        response := CustomerResponse{
            ID:   "123",
            Name: "John Doe",
        }
        json.NewEncoder(w).Encode(response)
    }))
    
    client := NewCustomerClient(mockServer.URL)
    customer, err := client.GetCustomer("123")
    
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", customer.Name)
}
```

### Integration Testing
Test service interactions:

```go
func TestOrderProcessingIntegration(t *testing.T) {
    // Start test containers
    postgres := startPostgresContainer(t)
    redis := startRedisContainer(t)
    kafka := startKafkaContainer(t)
    
    // Start services
    orderService := startOrderService(t, postgres, redis, kafka)
    inventoryService := startInventoryService(t, postgres, kafka)
    
    // Test integration
    order := createTestOrder()
    err := orderService.ProcessOrder(order)
    
    assert.NoError(t, err)
    
    // Verify inventory was updated
    inventory, err := inventoryService.GetInventory(order.ProductID)
    assert.NoError(t, err)
    assert.Equal(t, expectedQuantity, inventory.Quantity)
}
```

## Next Steps

Learn more about specific aspects:
- [Database Design](database-design.md) - Data modeling for microservices
- [Event-Driven Architecture](event-architecture.md) - Asynchronous communication
- [Security Architecture](security-architecture.md) - Securing microservices