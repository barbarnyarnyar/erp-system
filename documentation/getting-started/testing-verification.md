# Testing and Verification

How to test, verify, and validate the ERP system.

## Quick Health Check

The fastest way to verify all services are running:

```bash
make health
```

This curls the `/health` endpoint of every service and the API Gateway, reporting which are responding.

Individual health endpoints return:

```bash
curl http://localhost:8001/health
# {"service":"fm-service","status":"healthy","port":"8001"}
```

## Automated Test Suite

The project has a single test file: `services/fm-service/internal/business/service/service_test.go` (102 lines). It tests two scenarios:

- `CreateAccount` publishes the expected Kafka event
- `CreateInvoice` publishes the expected Kafka event

### Run All Tests

```bash
# From fm-service directory
cd services/fm-service
make test
# or
go test ./...

# From root - build test (Docker)
make test
```

### Run Tests with Coverage

```bash
# fm-service only
cd services/fm-service
make test-coverage
```

## API Gateway Verification

### Test Hello World Endpoints

The Makefile provides two test targets:

```bash
# Through the API Gateway (port 8080)
make test

# Directly to each service (bypass gateway)
make test-direct
```

These verify each service responds to its hello endpoint:

| Service | Gateway Route | Direct URL |
|---------|--------------|------------|
| Finance | `http://localhost:8080/api/v1/finance/hello` | `http://localhost:8001/` |
| HR | `http://localhost:8080/api/v1/hr/hello` | `http://localhost:8002/` |
| SCM | `http://localhost:8080/api/v1/scm/hello` | `http://localhost:8003/` |
| Manufacturing | `http://localhost:8080/api/v1/manufacturing/hello` | `http://localhost:8004/` |
| CRM | `http://localhost:8080/api/v1/crm/hello` | `http://localhost:8005/` |
| Projects | `http://localhost:8080/api/v1/projects/hello` | `http://localhost:8006/` |

## Manual API Testing

### Service Discovery (API Gateway)

```bash
curl http://localhost:8080/services
```

### Financial Management (port 8001)

```bash
# List accounts
curl http://localhost:8001/api/v1/accounts

# Create an account
curl -X POST http://localhost:8001/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_number":"1000","name":"Cash","type":"ASSET","currency":"USD"}'

# Create a journal entry
curl -X POST http://localhost:8001/api/v1/journal-entries \
  -H "Content-Type: application/json" \
  -d '{"reference":"JE-001","description":"Test entry","lines":[{"account_id":"acc_...","debit_amount":"100","credit_amount":"0"},{"account_id":"acc_...","debit_amount":"0","credit_amount":"100"}]}'

# Create an invoice
curl -X POST http://localhost:8001/api/v1/invoices \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust_001","issue_date":"2026-06-06T00:00:00Z","due_date":"2026-07-06T00:00:00Z","lines":[{"description":"Widget","quantity":10,"unit_price":"25.00"}]}'

# Balance sheet
curl http://localhost:8001/api/v1/reports/balance-sheet
```

### Human Resources (port 8002)

```bash
# Create an employee
curl -X POST http://localhost:8002/api/v1/employees \
  -H "Content-Type: application/json" \
  -d '{"first_name":"John","last_name":"Doe","email":"john@example.com","department_id":"dept_001","position_id":"pos_001","salary":"60000"}'

# List employees
curl http://localhost:8002/api/v1/employees

# Process payroll
curl -X POST http://localhost:8002/api/v1/payroll \
  -H "Content-Type: application/json" \
  -d '{"employee_id":"emp_...","pay_period_start":"2026-06-01T00:00:00Z","pay_period_end":"2026-06-15T00:00:00Z","regular_hours":"80","overtime_hours":"5"}'

# Submit a leave request
curl -X POST http://localhost:8002/api/v1/leave-requests \
  -H "Content-Type: application/json" \
  -d '{"employee_id":"emp_...","leave_type":"ANNUAL","start_date":"2026-07-01T00:00:00Z","end_date":"2026-07-05T00:00:00Z","reason":"Vacation"}'

# Headcount report
curl http://localhost:8002/api/v1/reports/headcount
```

### Supply Chain Management (port 8003)

```bash
# Create a product
curl -X POST http://localhost:8003/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"product_code":"WIDGET-001","product_name":"Widget Alpha","product_type":"FINISHED_GOOD","unit_of_measure":"EA","standard_cost":"10.00","list_price":"25.00"}'

# List products
curl http://localhost:8003/api/v1/products

# Create a purchase order
curl -X POST http://localhost:8003/api/v1/purchase-orders \
  -H "Content-Type: application/json" \
  -d '{"supplier_id":"supp_...","expected_delivery":"2026-07-01T00:00:00Z","notes":"Rush order","lines":[{"product_id":"prod_...","quantity_ordered":100,"unit_price":"8.00"}]}'

# Check inventory
curl http://localhost:8003/api/v1/inventory

# Run inventory levels report
curl http://localhost:8003/api/v1/reports/inventory-levels
```

### Manufacturing (port 8004)

```bash
# Create a BOM
curl -X POST http://localhost:8004/api/v1/boms \
  -H "Content-Type: application/json" \
  -d '{"product_id":"prod_001","version":"V1.0","description":"Standard assembly"}'

# Create a work center
curl -X POST http://localhost:8004/api/v1/work-centers \
  -H "Content-Type: application/json" \
  -d '{"code":"WC-001","name":"Assembly Line 1","capacity":"160","hourly_rate":"45.00"}'

# Create a production order
curl -X POST http://localhost:8004/api/v1/production-plans \
  -H "Content-Type: application/json" \
  -d '{"bom_id":"bom_default","quantity":100,"scheduled_date":"2026-07-01T00:00:00Z"}'

# List work orders
curl http://localhost:8004/api/v1/work-orders

# Run MRP
curl -X POST http://localhost:8004/api/v1/mrp/run
```

### CRM (port 8005)

```bash
# Create a customer
curl -X POST http://localhost:8005/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{"company_name":"Acme Corp","contact_name":"John Smith","email":"john@acme.com","phone":"555-0100","category":"WHOLESALE"}'

# Create a lead
curl -X POST http://localhost:8005/api/v1/leads \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Jane","last_name":"Doe","company":"TechStart","email":"jane@techstart.com","source":"WEBSITE"}'

# Convert lead to customer + opportunity
curl -X POST http://localhost:8005/api/v1/leads/:id/convert

# Create a sales order
curl -X POST http://localhost:8005/api/v1/sales-orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust_...","items":[{"product_id":"prod_001","quantity":10,"unit_price":"25.00","discount":"0"}]}'

# Create a service ticket
curl -X POST http://localhost:8005/api/v1/service-tickets \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust_...","title":"Login issue","description":"Cannot access dashboard","priority":"HIGH"}'
```

### Project Management (port 8006)

```bash
# Create a portfolio
curl -X POST http://localhost:8006/api/v1/projects/portfolios \
  -H "Content-Type: application/json" \
  -d '{"name":"Digital Transformation","description":"Company-wide digital initiative"}'

# Create a project
curl -X POST http://localhost:8006/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"Warehouse Automation","description":"Automate warehouse operations","start_date":"2026-06-01T00:00:00Z","end_date":"2026-12-31T00:00:00Z"}'

# Create a task
curl -X POST http://localhost:8006/api/v1/projects/:id/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Design phase","description":"Complete system design","assigned_to":"emp_001"}'

# Log time
curl -X POST http://localhost:8006/api/v1/projects/:id/time \
  -H "Content-Type: application/json" \
  -d '{"task_id":"task_...","user_id":"emp_001","entry_date":"2026-06-06T00:00:00Z","hours":"8","description":"System design work"}'

# Portfolio summary
curl http://localhost:8006/api/v1/projects/portfolios/:id/summary
```

### Auth Service (port 8000)

```bash
# Login (default seeded credentials)
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Register a new user
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"jdoe","email":"jdoe@example.com","password":"securepass","first_name":"John","last_name":"Doe"}'

# Validate a permission
curl -X POST http://localhost:8000/api/v1/auth/users/:id/validate-permission \
  -H "Content-Type: application/json" \
  -d '{"permission":"scm:product:read"}'
```

## Verification Checklist

### Infrastructure

- [ ] PostgreSQL is running on port 5432
- [ ] Redis is running on port 6379
- [ ] Zookeeper is running on port 2181
- [ ] Kafka is running on port 9092
- [ ] All Docker containers are healthy: `docker compose ps`

### Services

- [ ] API Gateway accessible on port 8080
- [ ] Auth Service responds on port 8000
- [ ] FM Service responds on port 8001
- [ ] CRM Service responds on port 8005 (code default 8002)
- [ ] HR Service responds on port 8003 (architected port 8002)
- [ ] M Service responds on port 8004
- [ ] PM Service responds on port 8006
- [ ] SCM Service responds on port 8003 (code default 8006)
- [ ] All services return HTTP 200 on `/health`
- [ ] Gateway proxies requests to each backend service

### Known Testing Limitations

- **No database**: All services use in-memory storage. Data is lost on restart. There is no persistence to verify.
- **No auth**: The deployed API gateway has no authentication. All endpoints are publicly accessible.
- **Kafka may be unavailable**: Event publishing errors are silently ignored. The system works without Kafka running.
- **Single test file**: Only `fm-service` has automated tests (2 test cases). All other services have zero test coverage.
- **Stub endpoints**: FM's income statement and cash flow reports return placeholder messages, not real data.
