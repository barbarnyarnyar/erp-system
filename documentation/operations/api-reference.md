# API Reference

Complete REST API documentation for the ERP system with examples and usage patterns.

## Base URLs and Routing

### Production API Gateway
```
Base URL: https://erp.company.com/api/v1
```

### Service Routes
All requests go through the API Gateway which routes to appropriate services:
- **Financial Management**: `/api/v1/finance/*` → fm-service:8001
- **Human Resources**: `/api/v1/hr/*` → hr-service:8002
- **Supply Chain**: `/api/v1/scm/*` → scm-service:8003
- **Customer Relations**: `/api/v1/crm/*` → crm-service:8004
- **Manufacturing**: `/api/v1/manufacturing/*` → mfg-service:8005
- **Project Management**: `/api/v1/projects/*` → pm-service:8006

## Authentication

### Login and Token Management

**Login**
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@company.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "id": "user-123",
    "email": "user@company.com",
    "roles": ["finance_user", "hr_viewer"]
  }
}
```

**Using JWT Tokens**
Include the JWT token in the Authorization header for all API requests:
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Refresh Token**
```http
POST /auth/refresh
Content-Type: application/json
Authorization: Bearer <refresh_token>

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Financial Management API

### Account Management

**List Accounts**
```http
GET /api/v1/finance/accounts
Authorization: Bearer <token>
```

Query Parameters:
- `page` (optional): Page number (default: 1)
- `limit` (optional): Results per page (default: 20, max: 100)
- `type` (optional): Filter by account type (ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE)
- `active` (optional): Filter by active status (true/false)

**Create Account**
```http
POST /api/v1/finance/accounts
Content-Type: application/json
Authorization: Bearer <token>

{
  "account_code": "1100",
  "account_name": "Accounts Receivable", 
  "account_type": "ASSET",
  "parent_account_id": "acc-123",
  "normal_side": "DEBIT",
  "allow_posting": true
}
```

**Get Account Details**
```http
GET /api/v1/finance/accounts/{account_id}
Authorization: Bearer <token>
```

### Journal Entries

**Create Journal Entry**
```http
POST /api/v1/finance/journal-entries
Content-Type: application/json
Authorization: Bearer <token>

{
  "entry_date": "2024-03-15",
  "description": "Monthly rent payment",
  "reference": "RENT-MAR-2024",
  "source_module": "finance",
  "lines": [
    {
      "account_id": "acc-500",
      "description": "Rent expense",
      "debit_amount": "2500.00",
      "credit_amount": "0.00"
    },
    {
      "account_id": "acc-100", 
      "description": "Cash payment",
      "debit_amount": "0.00",
      "credit_amount": "2500.00"
    }
  ]
}
```

**List Journal Entries**
```http
GET /api/v1/finance/journal-entries
Authorization: Bearer <token>
```

Query Parameters:
- `date_from` (optional): Start date filter (YYYY-MM-DD)
- `date_to` (optional): End date filter (YYYY-MM-DD)
- `status` (optional): Filter by entry status (DRAFT, POSTED, REVERSED)
- `account_id` (optional): Filter by account

### Financial Reports

**Balance Sheet**
```http
GET /api/v1/finance/reports/balance-sheet
Authorization: Bearer <token>
```

Query Parameters:
- `as_of_date`: Balance sheet date (YYYY-MM-DD)
- `format` (optional): Response format (json, pdf, excel) - default: json

**Income Statement**
```http
GET /api/v1/finance/reports/income-statement
Authorization: Bearer <token>
```

Query Parameters:
- `period_start`: Start of period (YYYY-MM-DD)
- `period_end`: End of period (YYYY-MM-DD)
- `format` (optional): Response format (json, pdf, excel) - default: json

## Human Resources API

### Employee Management

**List Employees**
```http
GET /api/v1/hr/employees
Authorization: Bearer <token>
```

Query Parameters:
- `department_id` (optional): Filter by department
- `status` (optional): Filter by employment status
- `page` (optional): Page number
- `limit` (optional): Results per page

**Create Employee**
```http
POST /api/v1/hr/employees
Content-Type: application/json
Authorization: Bearer <token>

{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@company.com",
  "phone": "+1-555-123-4567",
  "hire_date": "2024-03-15",
  "department_id": "dept-123",
  "position_id": "pos-456",
  "employment_type": "FULL_TIME",
  "salary": "75000.00",
  "pay_frequency": "MONTHLY"
}
```

**Get Employee Details**
```http
GET /api/v1/hr/employees/{employee_id}
Authorization: Bearer <token>
```

### Payroll Processing

**Calculate Payroll**
```http
POST /api/v1/hr/payroll/calculate
Content-Type: application/json
Authorization: Bearer <token>

{
  "employee_ids": ["emp-123", "emp-456"],
  "pay_period_start": "2024-03-01",
  "pay_period_end": "2024-03-31"
}
```

**Get Pay Stub**
```http
GET /api/v1/hr/payroll/{payroll_id}/paystub
Authorization: Bearer <token>
```

Query Parameters:
- `format` (optional): Response format (pdf, html) - default: html

## Supply Chain Management API

### Inventory Management

**Get Inventory Levels**
```http
GET /api/v1/scm/inventory
Authorization: Bearer <token>
```

Query Parameters:
- `location_id` (optional): Filter by location
- `product_id` (optional): Filter by product
- `low_stock` (optional): Show only low stock items (true/false)

**Process Inventory Movement**
```http
POST /api/v1/scm/inventory/movements
Content-Type: application/json
Authorization: Bearer <token>

{
  "product_id": "prod-123",
  "location_id": "loc-456",
  "movement_type": "RECEIPT",
  "quantity": 100,
  "unit_cost": "25.50",
  "reference_type": "PURCHASE_ORDER",
  "reference_id": "po-789",
  "notes": "Weekly inventory receipt"
}
```

### Purchase Order Management

**Create Purchase Order**
```http
POST /api/v1/scm/purchase-orders
Content-Type: application/json
Authorization: Bearer <token>

{
  "supplier_id": "sup-123",
  "order_date": "2024-03-15",
  "expected_delivery": "2024-03-25",
  "items": [
    {
      "product_id": "prod-456",
      "quantity": 100,
      "unit_price": "25.00",
      "description": "Widget Component A"
    }
  ]
}
```

**List Purchase Orders**
```http
GET /api/v1/scm/purchase-orders
Authorization: Bearer <token>
```

Query Parameters:
- `status` (optional): Filter by order status
- `supplier_id` (optional): Filter by supplier
- `date_from` (optional): Start date filter
- `date_to` (optional): End date filter

## Customer Relationship Management API

### Lead Management

**Create Lead**
```http
POST /api/v1/crm/leads
Content-Type: application/json
Authorization: Bearer <token>

{
  "first_name": "Jane",
  "last_name": "Smith",
  "company": "ABC Corp",
  "email": "jane.smith@abccorp.com",
  "phone": "+1-555-987-6543",
  "source": "WEBSITE",
  "estimated_value": "50000.00"
}
```

**Convert Lead to Customer**
```http
POST /api/v1/crm/leads/{lead_id}/convert
Content-Type: application/json
Authorization: Bearer <token>

{
  "create_opportunity": true,
  "opportunity_name": "ABC Corp - ERP Implementation",
  "opportunity_value": "50000.00",
  "close_date": "2024-06-30"
}
```

### Customer Management

**List Customers**
```http
GET /api/v1/crm/customers
Authorization: Bearer <token>
```

Query Parameters:
- `status` (optional): Filter by customer status
- `segment` (optional): Filter by customer segment
- `page` (optional): Page number
- `limit` (optional): Results per page

## Standard Response Formats

### Success Response
```json
{
  "data": {
    "id": "resource-123",
    "name": "Resource Name",
    "status": "active"
  },
  "meta": {
    "request_id": "req-abc123",
    "timestamp": "2024-03-15T14:30:00Z"
  }
}
```

### Paginated Response
```json
{
  "data": [
    {
      "id": "item-1",
      "name": "Item 1"
    },
    {
      "id": "item-2", 
      "name": "Item 2"
    }
  ],
  "pagination": {
    "current_page": 1,
    "per_page": 20,
    "total_pages": 5,
    "total_items": 95,
    "has_next": true,
    "has_prev": false
  },
  "meta": {
    "request_id": "req-def456",
    "timestamp": "2024-03-15T14:30:00Z"
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input provided",
    "details": [
      {
        "field": "email",
        "message": "Email format is invalid"
      },
      {
        "field": "hire_date",
        "message": "Hire date cannot be in the future"
      }
    ]
  },
  "meta": {
    "request_id": "req-xyz789",
    "timestamp": "2024-03-15T14:30:00Z"
  }
}
```

## HTTP Status Codes

### Success Codes
- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `202 Accepted` - Request accepted, processing asynchronously
- `204 No Content` - Successful request with no response body

### Client Error Codes
- `400 Bad Request` - Invalid request format or parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (duplicate, etc.)
- `422 Unprocessable Entity` - Validation errors
- `429 Too Many Requests` - Rate limit exceeded

### Server Error Codes
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Service unavailable
- `503 Service Unavailable` - Temporary service unavailability
- `504 Gateway Timeout` - Request timeout

## Rate Limiting

### Rate Limits
- **Default**: 1000 requests per hour per authenticated user
- **Unauthenticated**: 100 requests per hour per IP address
- **Premium accounts**: 5000 requests per hour

### Rate Limit Headers
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

### Rate Limit Exceeded Response
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests",
    "retry_after": 3600
  }
}
```

## Webhooks and Real-time Updates

### Webhook Configuration
```http
POST /api/v1/webhooks
Content-Type: application/json
Authorization: Bearer <token>

{
  "url": "https://your-app.com/webhooks/erp",
  "events": ["account.created", "employee.updated", "order.completed"],
  "secret": "your-webhook-secret"
}
```

### Webhook Payload Example
```json
{
  "event": "account.created",
  "data": {
    "id": "acc-123",
    "account_code": "1100",
    "account_name": "Accounts Receivable"
  },
  "timestamp": "2024-03-15T14:30:00Z",
  "webhook_id": "wh-456"
}
```

## SDK and Integration Examples

### cURL Examples
```bash
# Create account with cURL
curl -X POST https://erp.company.com/api/v1/finance/accounts \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "account_code": "1100",
    "account_name": "Accounts Receivable",
    "account_type": "ASSET"
  }'
```

### JavaScript/Node.js Example
```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'https://erp.company.com/api/v1',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});

// Create account
const account = await client.post('/finance/accounts', {
  account_code: '1100',
  account_name: 'Accounts Receivable',
  account_type: 'ASSET'
});
```

### Python Example
```python
import requests

class ERPClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def create_account(self, account_data):
        response = requests.post(
            f'{self.base_url}/finance/accounts',
            json=account_data,
            headers=self.headers
        )
        return response.json()

# Usage
client = ERPClient('https://erp.company.com/api/v1', 'your-jwt-token')
account = client.create_account({
    'account_code': '1100',
    'account_name': 'Accounts Receivable',
    'account_type': 'ASSET'
})
```

## Next Steps

- [Authentication Guide](authentication.md) - Detailed authentication setup
- [Integration Patterns](integration-patterns.md) - Best practices for integration
- [Troubleshooting](troubleshooting.md) - Common API issues and solutions
- [Performance Optimization](performance.md) - Optimize your API usage