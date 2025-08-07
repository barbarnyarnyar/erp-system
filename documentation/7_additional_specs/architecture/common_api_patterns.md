# Common API Patterns

## Overview
This document defines the standardized API patterns used across all ERP microservices to ensure consistency, maintainability, and ease of integration.

## REST API Standards

### Base URL Structure
```
https://api.erp-system.com/api/v1/{service}/{resource}
```

**Examples:**
- Financial: `/api/v1/fm/accounts`
- HR: `/api/v1/hr/employees`
- CRM: `/api/v1/crm/customers`
- SCM: `/api/v1/scm/products`

### Standard CRUD Operations

#### Resource Collection Operations
```go
// List resources with pagination and filtering
GET    /api/v1/{service}/{resource}
Query Parameters:
  - page: int (default: 1)
  - limit: int (default: 20, max: 100)
  - sort: string (e.g., "created_at:desc")
  - filter: string (service-specific filtering)

// Create new resource
POST   /api/v1/{service}/{resource}
Body: Resource creation payload
```

#### Individual Resource Operations
```go
// Get resource by ID
GET    /api/v1/{service}/{resource}/{id}

// Update resource (full update)
PUT    /api/v1/{service}/{resource}/{id}
Body: Complete resource payload

// Partial update resource
PATCH  /api/v1/{service}/{resource}/{id}
Body: Partial resource payload

// Delete resource
DELETE /api/v1/{service}/{resource}/{id}
```

#### Resource Actions
```go
// Perform action on resource
POST   /api/v1/{service}/{resource}/{id}/{action}
Body: Action-specific payload

Examples:
POST   /api/v1/crm/orders/123/confirm
POST   /api/v1/hr/employees/456/promote
POST   /api/v1/fm/invoices/789/send
```

### Health and Status Endpoints

#### Health Check
```go
GET    /health
Response: {
  "status": "healthy|unhealthy",
  "service": "service-name",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z",
  "dependencies": {
    "database": "healthy",
    "message_queue": "healthy",
    "cache": "healthy"
  }
}
```

#### Service Info
```go
GET    /info
Response: {
  "service": "service-name",
  "version": "1.0.0",
  "build": "abc123",
  "environment": "production",
  "uptime": "24h30m15s"
}
```

## Standard HTTP Response Format

### Success Response
```json
{
  "success": true,
  "data": {
    // Resource data or response payload
  },
  "metadata": {
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req-123",
    "total_count": 100,      // For paginated responses
    "page": 1,               // For paginated responses  
    "limit": 20              // For paginated responses
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  },
  "metadata": {
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req-123"
  }
}
```

## Standard HTTP Status Codes

### Success Codes
- `200 OK`: Successful GET, PUT, PATCH, DELETE
- `201 Created`: Successful POST (resource creation)
- `202 Accepted`: Request accepted for asynchronous processing
- `204 No Content`: Successful DELETE with no response body

### Client Error Codes
- `400 Bad Request`: Invalid request format or data
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Authentication valid but insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., duplicate)
- `422 Unprocessable Entity`: Validation errors

### Server Error Codes
- `500 Internal Server Error`: Unexpected server error
- `502 Bad Gateway`: Upstream service error
- `503 Service Unavailable`: Service temporarily unavailable
- `504 Gateway Timeout`: Upstream service timeout

## Pagination

### Query Parameters
```
GET /api/v1/service/resource?page=2&limit=50&sort=created_at:desc
```

### Response Format
```json
{
  "success": true,
  "data": [
    // Array of resources
  ],
  "metadata": {
    "pagination": {
      "current_page": 2,
      "per_page": 50,
      "total_items": 1250,
      "total_pages": 25,
      "has_next": true,
      "has_previous": true
    }
  }
}
```

## Filtering and Searching

### Query Parameters
```
GET /api/v1/service/resource?filter=status:active,type:premium&search=john
```

### Standard Filter Operators
- `eq`: Equal (default)
- `ne`: Not equal
- `gt`: Greater than
- `gte`: Greater than or equal
- `lt`: Less than
- `lte`: Less than or equal
- `in`: In list
- `like`: Text search (partial match)

### Examples
```
filter=created_at:gte:2024-01-01
filter=status:in:active,pending
filter=name:like:john
```

## Authentication and Authorization

### Authentication Header
```
Authorization: Bearer <jwt-token>
```

### API Key Authentication (for service-to-service)
```
X-API-Key: <api-key>
```

### Request Context Headers
```
X-Request-ID: <unique-request-id>
X-User-ID: <authenticated-user-id>
X-Tenant-ID: <tenant-id> (for multi-tenant systems)
```

## Versioning

### URL Versioning (Current Standard)
```
/api/v1/service/resource
/api/v2/service/resource
```

### Header Versioning (Alternative)
```
Accept: application/vnd.erp-system.v1+json
```

## Rate Limiting

### Headers
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1609459200
```

### Rate Limit Exceeded Response
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again later.",
    "details": {
      "limit": 1000,
      "remaining": 0,
      "reset_time": "2024-01-01T13:00:00Z"
    }
  }
}
```

## Content-Type and Accept Headers

### Standard Headers
```
Content-Type: application/json
Accept: application/json
```

### File Upload
```
Content-Type: multipart/form-data
```

### File Download
```
Accept: application/octet-stream
Content-Disposition: attachment; filename="report.pdf"
```

## Validation Patterns

### Request Validation
- All incoming data must be validated
- Use JSON Schema for complex validation
- Return detailed validation errors
- Sanitize input data

### Example Validation Error
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "code": "INVALID_FORMAT",
        "message": "Email format is invalid",
        "value": "invalid-email"
      },
      {
        "field": "password",
        "code": "TOO_SHORT",
        "message": "Password must be at least 8 characters",
        "value": "***"
      }
    ]
  }
}
```

## Async Operations

### Long-Running Operations
For operations that take significant time:

1. Return `202 Accepted` immediately
2. Provide a way to check status
3. Use webhooks or polling for completion

```go
// Initial request
POST /api/v1/service/resource/{id}/process
Response: 202 Accepted
{
  "success": true,
  "data": {
    "operation_id": "op-123",
    "status": "processing",
    "status_url": "/api/v1/service/operations/op-123"
  }
}

// Status check
GET /api/v1/service/operations/op-123
Response: 200 OK
{
  "success": true,
  "data": {
    "operation_id": "op-123",
    "status": "completed|failed|processing",
    "progress": 75,
    "result": { /* operation result */ },
    "error": { /* error details if failed */ }
  }
}
```

## Security Best Practices

### Input Validation
- Validate all input data
- Use parameterized queries
- Sanitize user input
- Implement size limits

### Authentication
- Use JWT tokens with short expiration
- Implement token refresh mechanism
- Use secure session management
- Support multi-factor authentication

### Authorization
- Implement role-based access control (RBAC)
- Use principle of least privilege
- Validate permissions on every request
- Support resource-level permissions

### Data Protection
- Encrypt sensitive data
- Use HTTPS for all communications
- Implement audit logging
- Follow data privacy regulations

This standardization ensures consistency across all ERP microservices and simplifies integration, testing, and maintenance.