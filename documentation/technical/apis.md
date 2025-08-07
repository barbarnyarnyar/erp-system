# API Documentation

This document provides comprehensive API documentation for all ERP microservices, including authentication, endpoints, data models, and integration guidelines.

## Table of Contents

- [API Overview](#api-overview)
- [Authentication and Authorization](#authentication-and-authorization)
- [Common API Patterns](#common-api-patterns)
- [Financial Management API](#financial-management-api)
- [Human Resources API](#human-resources-api)
- [Supply Chain Management API](#supply-chain-management-api)
- [CRM and Sales API](#crm-and-sales-api)
- [Manufacturing API](#manufacturing-api)
- [Project Management API](#project-management-api)
- [Error Handling](#error-handling)
- [Rate Limiting and Performance](#rate-limiting-and-performance)

---

## API Overview

### Base URL Structure

All ERP APIs are accessible through the API Gateway with the following base URLs:

```
Production: https://api.erp-system.com/api/v1
Staging: https://staging-api.erp-system.com/api/v1
Development: https://dev-api.erp-system.com/api/v1
```

### Service Endpoints

| Service | Base Path | Description |
|---------|-----------|-------------|
| **Authentication** | `/auth` | User authentication and authorization |
| **Financial Management** | `/fm` | General ledger, AP, AR, financial reporting |
| **Human Resources** | `/hr` | Employee management, payroll, time tracking |
| **Supply Chain** | `/scm` | Inventory, procurement, supplier management |
| **CRM/Sales** | `/crm` | Customer management, sales pipeline, support |
| **Manufacturing** | `/mfg` | Production planning, BOM, quality control |
| **Project Management** | `/pm` | Project tracking, resource allocation, billing |

### API Versioning

- **Current Version**: v1
- **Versioning Strategy**: URL path versioning (`/api/v1/`, `/api/v2/`)
- **Backward Compatibility**: Previous versions supported for 12 months
- **Deprecation Policy**: 6-month notice for breaking changes

---

## Authentication and Authorization

### Authentication Flow

1. **Initial Authentication**: POST `/auth/login` with credentials
2. **Token Response**: Receive JWT access token and refresh token
3. **API Access**: Include access token in Authorization header
4. **Token Refresh**: Use refresh token when access token expires

### JWT Token Structure

```json
{
  "sub": "user123",
  "email": "user@company.com",
  "roles": ["finance_user", "department_manager"],
  "permissions": ["read:accounts", "write:journal_entries"],
  "exp": 1640995200,
  "iat": 1640991600
}
```

### Required Headers

All authenticated requests must include:

```http
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Request-ID: <unique_request_id>
X-Idempotency-Key: <idempotency_key> (for write operations)
```

### Role-Based Access Control

**System Roles:**
- `system_admin`: Full system access
- `finance_admin`: Financial module administration
- `hr_admin`: HR module administration
- `operations_manager`: SCM and manufacturing access

**Functional Roles:**
- `finance_user`: General financial operations
- `ap_clerk`: Accounts payable operations
- `ar_clerk`: Accounts receivable operations
- `payroll_admin`: Payroll processing
- `sales_rep`: CRM and sales operations
- `warehouse_staff`: Inventory and receiving operations

---

## Common API Patterns

### Standard HTTP Methods

- **GET**: Retrieve resources (safe, idempotent)
- **POST**: Create resources (not idempotent)
- **PUT**: Update/replace resources (idempotent)
- **PATCH**: Partial updates (not idempotent)
- **DELETE**: Remove resources (idempotent)

### Request/Response Format

**Standard Request:**
```json
{
  "data": {
    // Request payload
  },
  "metadata": {
    "request_id": "req_123",
    "timestamp": "2024-03-15T10:30:00Z"
  }
}
```

**Standard Response:**
```json
{
  "data": {
    // Response payload
  },
  "metadata": {
    "request_id": "req_123",
    "timestamp": "2024-03-15T10:30:00Z",
    "processing_time_ms": 150
  },
  "links": {
    "self": "https://api.erp-system.com/api/v1/resource/123",
    "related": "https://api.erp-system.com/api/v1/resource/123/items"
  }
}
```

### Pagination

Large datasets use cursor-based pagination:

**Request:**
```http
GET /api/v1/customers?limit=50&cursor=eyJpZCI6MTIzLCJ0cyI6MTY0MDk5NTIwMH0
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "has_more": true,
    "next_cursor": "eyJpZCI6MTczLCJ0cyI6MTY0MDk5NTQwMH0",
    "limit": 50,
    "total_count": 1250
  }
}
```

### Filtering and Sorting

**Query Parameters:**
- `filter[field]=value`: Filter by field value
- `filter[date_range]=2024-01-01,2024-01-31`: Date range filtering
- `sort=field:desc`: Sort by field (asc/desc)
- `include=field1,field2`: Include related data

**Example:**
```http
GET /api/v1/invoices?filter[status]=pending&filter[amount_gte]=1000&sort=created_at:desc&include=customer,line_items
```

---

## Financial Management API

### Base URL
`/api/v1/fm`

### Core Endpoints

#### General Ledger

**Chart of Accounts:**
- `GET /gl/accounts` - Retrieve chart of accounts
- `POST /gl/accounts` - Create new account
- `GET /gl/accounts/{id}` - Get account details
- `PUT /gl/accounts/{id}` - Update account
- `GET /gl/accounts/{id}/balance` - Get account balance

**Journal Entries:**
- `GET /gl/journal-entries` - List journal entries
- `POST /gl/journal-entries` - Create journal entry
- `GET /gl/journal-entries/{id}` - Get journal entry
- `POST /gl/journal-entries/{id}/post` - Post journal entry
- `POST /gl/journal-entries/{id}/reverse` - Reverse journal entry

**Financial Reporting:**
- `GET /reports/balance-sheet` - Generate balance sheet
- `GET /reports/income-statement` - Generate P&L statement
- `GET /reports/cash-flow` - Generate cash flow statement
- `GET /reports/trial-balance` - Generate trial balance

#### Accounts Payable

**Vendor Management:**
- `GET /ap/vendors` - List vendors
- `POST /ap/vendors` - Create vendor
- `GET /ap/vendors/{id}` - Get vendor details
- `PUT /ap/vendors/{id}` - Update vendor

**Invoice Processing:**
- `GET /ap/invoices` - List AP invoices
- `POST /ap/invoices` - Create AP invoice
- `GET /ap/invoices/{id}` - Get invoice details
- `POST /ap/invoices/{id}/approve` - Approve invoice

**Payment Processing:**
- `GET /ap/payments` - List payments
- `POST /ap/payments` - Process payment
- `GET /ap/aging` - Generate aging report

#### Accounts Receivable

**Customer Management:**
- `GET /ar/customers` - List customers
- `POST /ar/customers` - Create customer
- `GET /ar/customers/{id}` - Get customer details

**Invoice Management:**
- `GET /ar/invoices` - List AR invoices
- `POST /ar/invoices` - Create customer invoice
- `POST /ar/payments` - Record customer payment
- `GET /ar/aging` - Generate AR aging report

### Sample API Calls

**Create Journal Entry:**
```http
POST /api/v1/fm/gl/journal-entries
Content-Type: application/json
Authorization: Bearer <token>

{
  "entry_date": "2024-03-15",
  "description": "Monthly rent payment",
  "reference": "CHECK-001234",
  "lines": [
    {
      "account_id": "acc_rent_expense",
      "debit_amount": 5000.00,
      "description": "Office rent - March 2024"
    },
    {
      "account_id": "acc_cash_operating",
      "credit_amount": 5000.00,
      "description": "Payment from operating cash"
    }
  ]
}
```

---

## Human Resources API

### Base URL
`/api/v1/hr`

### Core Endpoints

#### Employee Management

**Employee Data:**
- `GET /employees` - List employees
- `POST /employees` - Create employee
- `GET /employees/{id}` - Get employee details
- `PUT /employees/{id}` - Update employee
- `POST /employees/{id}/deactivate` - Deactivate employee

**Organizational Structure:**
- `GET /departments` - List departments
- `GET /positions` - List job positions
- `GET /org-chart` - Get organizational chart

#### Payroll Management

**Payroll Processing:**
- `GET /payroll/runs` - List payroll runs
- `POST /payroll/runs` - Create payroll run
- `GET /payroll/runs/{id}` - Get payroll run details
- `POST /payroll/runs/{id}/process` - Process payroll
- `GET /payroll/paystubs/{employee_id}` - Get employee pay stubs

#### Time and Attendance

**Time Tracking:**
- `GET /timesheets` - List timesheets
- `POST /timesheets/entries` - Record time entry
- `GET /employees/{id}/time-summary` - Get time summary

**Leave Management:**
- `GET /leave-requests` - List leave requests
- `POST /leave-requests` - Submit leave request
- `POST /leave-requests/{id}/approve` - Approve leave request

### Sample API Calls

**Create Employee:**
```http
POST /api/v1/hr/employees
Content-Type: application/json
Authorization: Bearer <token>

{
  "employee_id": "EMP001",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@company.com",
  "hire_date": "2024-03-01",
  "department_id": "dept_engineering",
  "position_id": "pos_software_developer",
  "salary": 75000.00,
  "employment_type": "FULL_TIME"
}
```

---

## Supply Chain Management API

### Base URL
`/api/v1/scm`

### Core Endpoints

#### Inventory Management

**Product Data:**
- `GET /products` - List products
- `POST /products` - Create product
- `GET /products/{id}` - Get product details
- `PUT /products/{id}` - Update product

**Inventory Tracking:**
- `GET /inventory` - List inventory items
- `GET /inventory/{product_id}` - Get product inventory
- `POST /inventory/adjustments` - Record inventory adjustment
- `GET /inventory/movements` - List inventory movements

#### Procurement

**Purchase Orders:**
- `GET /purchase-orders` - List purchase orders
- `POST /purchase-orders` - Create purchase order
- `GET /purchase-orders/{id}` - Get PO details
- `POST /purchase-orders/{id}/approve` - Approve PO

**Supplier Management:**
- `GET /suppliers` - List suppliers
- `POST /suppliers` - Create supplier
- `GET /suppliers/{id}/performance` - Get supplier performance

### Sample API Calls

**Create Purchase Order:**
```http
POST /api/v1/scm/purchase-orders
Content-Type: application/json
Authorization: Bearer <token>

{
  "supplier_id": "sup_abc_supplies",
  "order_date": "2024-03-15",
  "delivery_date": "2024-03-22",
  "lines": [
    {
      "product_id": "prod_office_paper",
      "quantity": 100,
      "unit_price": 12.50,
      "line_total": 1250.00
    }
  ]
}
```

---

## CRM and Sales API

### Base URL
`/api/v1/crm`

### Core Endpoints

#### Customer Management

**Customer Data:**
- `GET /customers` - List customers
- `POST /customers` - Create customer
- `GET /customers/{id}` - Get customer details
- `GET /customers/{id}/interactions` - Get interaction history

#### Sales Pipeline

**Opportunities:**
- `GET /opportunities` - List opportunities
- `POST /opportunities` - Create opportunity
- `PUT /opportunities/{id}/stage` - Update opportunity stage
- `GET /pipeline/forecast` - Get sales forecast

#### Marketing and Campaigns

**Campaign Management:**
- `GET /campaigns` - List campaigns
- `POST /campaigns` - Create campaign
- `GET /campaigns/{id}/performance` - Get campaign performance

### Sample API Calls

**Create Opportunity:**
```http
POST /api/v1/crm/opportunities
Content-Type: application/json
Authorization: Bearer <token>

{
  "customer_id": "cust_tech_solutions",
  "name": "Q2 Software Upgrade",
  "stage": "QUALIFIED",
  "value": 50000.00,
  "probability": 60,
  "expected_close_date": "2024-06-30",
  "assigned_to": "sales_rep_jane"
}
```

---

## Manufacturing API

### Base URL
`/api/v1/mfg`

### Core Endpoints

#### Production Planning

**Production Orders:**
- `GET /production-orders` - List production orders
- `POST /production-orders` - Create production order
- `GET /production-orders/{id}` - Get production order details
- `POST /production-orders/{id}/start` - Start production

**Bill of Materials:**
- `GET /bom` - List BOMs
- `POST /bom` - Create BOM
- `GET /bom/{id}` - Get BOM details

#### Quality Management

**Quality Checks:**
- `GET /quality-checks` - List quality checks
- `POST /quality-checks` - Record quality check
- `GET /quality-reports` - Get quality reports

---

## Project Management API

### Base URL
`/api/v1/pm`

### Core Endpoints

#### Project Management

**Projects:**
- `GET /projects` - List projects
- `POST /projects` - Create project
- `GET /projects/{id}` - Get project details
- `PUT /projects/{id}` - Update project

**Tasks and Resources:**
- `GET /projects/{id}/tasks` - List project tasks
- `POST /projects/{id}/tasks` - Create task
- `GET /projects/{id}/resources` - Get project resources

#### Time Tracking and Billing

**Time Entries:**
- `GET /time-entries` - List time entries
- `POST /time-entries` - Record time entry
- `GET /projects/{id}/timesheet` - Get project timesheet

---

## Error Handling

### Standard Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format",
        "value": "invalid-email"
      }
    ],
    "request_id": "req_123",
    "timestamp": "2024-03-15T10:30:00Z",
    "documentation_url": "https://docs.erp-system.com/errors/validation"
  }
}
```

### HTTP Status Codes

| Code | Description | Usage |
|------|-------------|-------|
| 200 | OK | Successful GET, PUT, PATCH |
| 201 | Created | Successful POST |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid request format/validation error |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict (duplicate, etc.) |
| 422 | Unprocessable Entity | Business rule violation |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |

### Common Error Codes

**Authentication Errors:**
- `AUTH_TOKEN_EXPIRED`: JWT token has expired
- `AUTH_INVALID_TOKEN`: Invalid JWT token format
- `AUTH_INSUFFICIENT_PERMISSIONS`: User lacks required permissions

**Validation Errors:**
- `VALIDATION_REQUIRED_FIELD`: Required field missing
- `VALIDATION_INVALID_FORMAT`: Field format invalid
- `VALIDATION_VALUE_OUT_OF_RANGE`: Numeric value out of range

**Business Logic Errors:**
- `BUSINESS_RULE_VIOLATION`: Business rule constraint violated
- `INSUFFICIENT_INVENTORY`: Not enough inventory for operation
- `DUPLICATE_RESOURCE`: Resource already exists

---

## Rate Limiting and Performance

### Rate Limits

**By User Type:**
- **Standard Users**: 1,000 requests per hour
- **Premium Users**: 5,000 requests per hour
- **System Integration**: 50,000 requests per hour

**By Endpoint Category:**
- **Read Operations**: Higher limits (5x multiplier)
- **Write Operations**: Standard limits
- **Report Generation**: Lower limits (0.5x multiplier)

### Rate Limit Headers

```http
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4999
X-RateLimit-Reset: 1640995200
X-RateLimit-Window: 3600
```

### Performance Expectations

**Response Time SLAs:**
- **Simple Queries**: < 100ms (95th percentile)
- **Complex Queries**: < 500ms (95th percentile)
- **Report Generation**: < 30 seconds
- **Bulk Operations**: < 60 seconds

### Caching Strategy

**Cache Headers:**
```http
Cache-Control: public, max-age=300
ETag: "abc123"
Last-Modified: Wed, 15 Mar 2024 10:30:00 GMT
```

**Cache Types:**
- **Static Data**: 1 hour cache (accounts, products)
- **Dynamic Data**: 5 minute cache (balances, inventory)
- **Reports**: 15 minute cache
- **Real-time Data**: No cache (live transactions)

### API Monitoring

**Health Checks:**
- `GET /health` - Basic health check
- `GET /health/deep` - Comprehensive health check including dependencies

**Metrics Exposed:**
- Request count and rate
- Response times (avg, p95, p99)
- Error rates by endpoint
- Cache hit/miss ratios
- Database connection pool status

This comprehensive API documentation provides developers with all the information needed to successfully integrate with the ERP system while maintaining security, performance, and reliability standards.