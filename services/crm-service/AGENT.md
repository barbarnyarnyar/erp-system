# CRM Service - Agent Documentation

## Overview

The CRM (Customer Relationship Management) Service is a microservice within the ERP system that handles customer interactions, service management, and support operations. This service provides comprehensive functionality for managing customer relationships, service offerings, and support tickets.

## Architecture

### Service Structure

```markdown:services/crm-service/AGENT.md
<code_block_to_apply_changes_from>
crm-service/
├── cmd/
│   └── main.go                 # Service entry point
├── internal/
│   ├── config/
│   │   └── database.go         # Database configuration
│   ├── handlers/
│   │   ├── contact_handler.go  # Contact management endpoints
│   │   ├── service_handler.go  # Service management endpoints
│   │   └── support_handler.go  # Support ticket endpoints
│   ├── models/
│   │   ├── contact.go          # Contact data models
│   │   ├── service.go          # Service data models
│   │   └── support.go          # Support ticket models
│   └── repositories/
│       ├── contact_repository.go    # Contact data access
│       ├── service_repository.go    # Service data access
│       └── support_repository.go    # Support data access
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
├── Dockerfile                  # Container configuration
└── AGENT.md                    # This documentation
```

### Technology Stack

- **Language**: Go 1.21
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL (via GORM)
- **Architecture**: Clean Architecture with Repository Pattern
- **Container**: Docker

## Core Features

### 1. Contact Management

The contact management module handles customer contact information and interactions.

#### Key Features:

- **Contact CRUD Operations**: Create, read, update, and delete customer contacts
- **Contact Search**: Search contacts by name, email, company, or other fields
- **Contact Categorization**: Tag and categorize contacts for better organization
- **Lead Source Tracking**: Track how contacts were acquired
- **Contact History**: Maintain contact interaction history

#### Data Model:

```go
type Contact struct {
    ID          uint      `json:"id"`
    FirstName   string    `json:"first_name"`
    LastName    string    `json:"last_name"`
    Email       string    `json:"email"`
    Phone       string    `json:"phone"`
    Company     string    `json:"company"`
    Position    string    `json:"position"`
    Department  string    `json:"department"`
    Address     string    `json:"address"`
    City        string    `json:"city"`
    State       string    `json:"state"`
    Country     string    `json:"country"`
    PostalCode  string    `json:"postal_code"`
    LeadSource  string    `json:"lead_source"`
    Status      string    `json:"status"`
    Notes       string    `json:"notes"`
    Tags        string    `json:"tags"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### 2. Service Management

The service management module handles service offerings and service requests.

#### Key Features:

- **Service Catalog**: Manage available services with pricing and descriptions
- **Service Categories**: Organize services by categories
- **Service Requests**: Handle customer service requests
- **Service Scheduling**: Schedule service appointments
- **Service Status Tracking**: Track service request status and completion

#### Data Models:

```go
type Service struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Category    string    `json:"category"`
    Price       float64   `json:"price"`
    Currency    string    `json:"currency"`
    Duration    int       `json:"duration"`
    Status      string    `json:"status"`
    IsActive    bool      `json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type ServiceRequest struct {
    ID          uint      `json:"id"`
    ContactID   uint      `json:"contact_id"`
    ServiceID   uint      `json:"service_id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Priority    string    `json:"priority"`
    Status      string    `json:"status"`
    RequestedAt time.Time `json:"requested_at"`
    ScheduledAt *time.Time `json:"scheduled_at"`
    CompletedAt *time.Time `json:"completed_at"`
    Notes       string    `json:"notes"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### 3. Support Management

The support management module handles customer support tickets and responses.

#### Key Features:

- **Support Tickets**: Create and manage customer support tickets
- **Ticket Tracking**: Track ticket status and priority
- **Ticket Responses**: Manage internal and external responses
- **Ticket Categories**: Categorize tickets for better organization
- **Ticket Assignment**: Assign tickets to support agents
- **Ticket History**: Maintain complete ticket interaction history

#### Data Models:

```go
type SupportTicket struct {
    ID           uint      `json:"id"`
    ContactID    uint      `json:"contact_id"`
    TicketNumber string    `json:"ticket_number"`
    Subject      string    `json:"subject"`
    Description  string    `json:"description"`
    Category     string    `json:"category"`
    Priority     string    `json:"priority"`
    Status       string    `json:"status"`
    AssignedTo   *uint     `json:"assigned_to"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    ClosedAt     *time.Time `json:"closed_at"`
}

type SupportResponse struct {
    ID         uint      `json:"id"`
    TicketID   uint      `json:"ticket_id"`
    Response   string    `json:"response"`
    IsInternal bool      `json:"is_internal"`
    CreatedAt  time.Time `json:"created_at"`
}
```

## API Endpoints

### Contact Management Endpoints

```
POST   /api/v1/contacts              # Create new contact
GET    /api/v1/contacts              # Get all contacts (with pagination)
GET    /api/v1/contacts/:id          # Get contact by ID
PUT    /api/v1/contacts/:id          # Update contact
DELETE /api/v1/contacts/:id          # Delete contact
GET    /api/v1/contacts/search       # Search contacts
```

### Service Management Endpoints

```
POST   /api/v1/services              # Create new service
GET    /api/v1/services              # Get all services (with pagination)
GET    /api/v1/services/active       # Get active services
GET    /api/v1/services/:id          # Get service by ID
PUT    /api/v1/services/:id          # Update service
DELETE /api/v1/services/:id          # Delete service

POST   /api/v1/service-requests              # Create service request
GET    /api/v1/service-requests              # Get all service requests
GET    /api/v1/service-requests/:id          # Get service request by ID
PUT    /api/v1/service-requests/:id          # Update service request
GET    /api/v1/contacts/:contactId/service-requests  # Get requests by contact
```

### Support Management Endpoints

```
POST   /api/v1/support/tickets              # Create support ticket
GET    /api/v1/support/tickets              # Get all tickets
GET    /api/v1/support/tickets/:id          # Get ticket by ID
GET    /api/v1/support/tickets/number/:ticketNumber  # Get ticket by number
PUT    /api/v1/support/tickets/:id          # Update ticket
PUT    /api/v1/support/tickets/:id/close    # Close ticket
GET    /api/v1/support/tickets/status/:status  # Get tickets by status
GET    /api/v1/contacts/:contactId/tickets  # Get tickets by contact

POST   /api/v1/support/responses            # Create ticket response
GET    /api/v1/support/tickets/:ticketId/responses  # Get responses by ticket
```

## Database Schema

### Tables

1. **contacts** - Customer contact information
2. **services** - Service offerings
3. **service_requests** - Customer service requests
4. **support_tickets** - Support tickets
5. **support_responses** - Ticket responses

### Relationships

- `service_requests.contact_id` → `contacts.id`
- `service_requests.service_id` → `services.id`
- `support_tickets.contact_id` → `contacts.id`
- `support_responses.ticket_id` → `support_tickets.id`

## Configuration

### Environment Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=crm_db
DB_PORT=5432

# Service Configuration
PORT=8002
SERVICE_NAME=crm-service
```

### Database Configuration

The service uses PostgreSQL with GORM as the ORM. Database connection is configured in `internal/config/database.go`.

## Development

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Docker (for containerization)

### Running Locally

```bash
# Set environment variables
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=crm_db
export DB_PORT=5432
export PORT=8002

# Run the service
go run cmd/main.go
```

### Running with Docker

```bash
# Build the image
docker build -t crm-service .

# Run the container
docker run -p 8002:8002 \
  -e DB_HOST=host.docker.internal \
  -e DB_USER=postgres \
  -e DB_PASSWORD=password \
  -e DB_NAME=crm_db \
  crm-service
```

## Testing

### API Testing

Use the provided endpoints to test the service functionality:

1. **Contact Management Test**:

   ```bash
   # Create a contact
   curl -X POST http://localhost:8002/api/v1/contacts \
     -H "Content-Type: application/json" \
     -d '{"first_name":"John","last_name":"Doe","email":"john@example.com"}'
   ```

2. **Service Management Test**:

   ```bash
   # Create a service
   curl -X POST http://localhost:8002/api/v1/services \
     -H "Content-Type: application/json" \
     -d '{"name":"Consultation","description":"Business consultation","price":100.00}'
   ```

3. **Support Management Test**:
   ```bash
   # Create a support ticket
   curl -X POST http://localhost:8002/api/v1/support/tickets \
     -H "Content-Type: application/json" \
     -d '{"contact_id":1,"subject":"Technical Issue","description":"Need help with login"}'
   ```

## Integration

### With Other Services

The CRM service integrates with other ERP services:

1. **Auth Service**: For user authentication and authorization
2. **API Gateway**: For routing and load balancing
3. **Shared Utils**: For common utilities and response formatting

### Message Queue Integration

Future integration with RabbitMQ for:

- Asynchronous processing of service requests
- Notification system for ticket updates
- Integration with external systems

## Monitoring and Logging

### Health Check

```
GET /health
```

Returns service health status including:

- Service name and version
- Database connectivity
- Service uptime

### Logging

The service uses structured logging with:

- Request ID tracking
- Service name identification
- Timestamp and log levels
- Error tracking and debugging information

## Security

### Authentication

- JWT token validation (via auth service)
- Role-based access control
- API key validation for external integrations

### Data Protection

- Input validation and sanitization
- SQL injection prevention (via GORM)
- Sensitive data encryption
- Audit logging for data changes

## Performance

### Optimization Strategies

- Database indexing on frequently queried fields
- Pagination for large datasets
- Connection pooling for database connections
- Caching for frequently accessed data

### Scalability

- Stateless service design
- Horizontal scaling capability
- Load balancing support
- Database read replicas support

## Future Enhancements

### Planned Features

1. **Advanced Analytics**: Customer behavior analysis and reporting
2. **Integration APIs**: Third-party CRM integrations
3. **Workflow Automation**: Automated ticket routing and escalation
4. **Mobile API**: Mobile-optimized endpoints
5. **Real-time Notifications**: WebSocket support for real-time updates

### Technical Improvements

1. **GraphQL API**: Alternative to REST for complex queries
2. **Event Sourcing**: For audit trails and data consistency
3. **Microservice Communication**: gRPC for inter-service communication
4. **Advanced Caching**: Redis integration for performance optimization

## Troubleshooting

### Common Issues

1. **Database Connection**: Check database credentials and connectivity
2. **Port Conflicts**: Ensure port 8002 is available
3. **Environment Variables**: Verify all required environment variables are set

### Debug Mode

Enable debug logging by setting the log level to DEBUG in the configuration.

## Contributing

### Code Standards

- Follow Go coding conventions
- Use meaningful variable and function names
- Add comments for complex logic
- Write unit tests for new features

### Git Workflow

1. Create feature branch from main
2. Implement changes with tests
3. Submit pull request with description
4. Code review and approval
5. Merge to main branch

---

**Last Updated**: December 2024
**Version**: 1.0.0
**Maintainer**: ERP System Team

```

This AGENT.md file provides comprehensive documentation for your CRM service, including:

1. **Architecture Overview** - Clear structure and technology stack
2. **Feature Documentation** - Detailed explanation of Contact, Service, and Support management
3. **API Documentation** - Complete endpoint listing with examples
4. **Database Schema** - Table structure and relationships
5. **Configuration Guide** - Environment variables and setup
6. **Development Instructions** - How to run and test the service
7. **Integration Details** - How it works with other services
8. **Security and Performance** - Best practices and optimization
9. **Future Roadmap** - Planned enhancements and improvements
10. **Troubleshooting** - Common issues and solutions

The documentation follows industry best practices and provides everything needed for developers to understand, maintain, and extend the CRM service.
```
