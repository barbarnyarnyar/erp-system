# API Reference

REST API endpoints for the ERP system with examples.

> **Note**: The deployed API Gateway has **no authentication**. All endpoints are publicly accessible. See [Authentication](authentication.md) for the inactive JWT auth system.

## Base URLs

### Direct Service Access (Development)

| Service | Base URL |
|---------|----------|
| Auth | `http://localhost:8000/api/v1/auth` |
| Financial Management | `http://localhost:8001/api/v1/finance` |
| Human Resources | `http://localhost:8002/api/v1/hr` |
| Supply Chain | `http://localhost:8003/api/v1/scm` |
| Manufacturing | `http://localhost:8004/api/v1/manufacturing` |
| Customer Relations | `http://localhost:8005/api/v1/crm` |
| Project Management | `http://localhost:8006/api/v1/projects` |

### Via API Gateway

```
http://localhost:8080/api/v1/{service}/*
```

Gateway routes:
- `/api/v1/finance/*` → fm-service:8001
- `/api/v1/hr/*` → hr-service:8002
- `/api/v1/scm/*` → scm-service:8003
- `/api/v1/manufacturing/*` → m-service:8004
- `/api/v1/crm/*` → crm-service:8005
- `/api/v1/projects/*` → pm-service:8006

## Authentication API

### Login

```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "rt_1749267184000000000_usr_...",
  "token_type": "Bearer"
}
```

### Register

```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"newuser","password":"pass123","email":"user@example.com"}'
```

### Refresh Token

```bash
curl -X POST http://localhost:8000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"rt_1749267184000000000_usr_..."}'
```

## Financial Management API

All endpoints under `http://localhost:8001/api/v1/finance` or via gateway.

### Accounts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/accounts` | List all accounts |
| GET | `/accounts/:id` | Get account by ID |
| POST | `/accounts` | Create new account |
| PUT | `/accounts/:id` | Update account |

**List accounts:**
```bash
curl http://localhost:8001/api/v1/finance/accounts
```

**Create account:**
```bash
curl -X POST http://localhost:8001/api/v1/finance/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "account_code": "1100",
    "account_name": "Accounts Receivable",
    "account_type": "ASSET",
    "normal_side": "DEBIT"
  }'
```

### Account Types

Create with valid `account_type`:
- `ASSET`
- `LIABILITY`
- `EQUITY`
- `REVENUE`
- `EXPENSE`

### Journal Entries

| Method | Path | Description |
|--------|------|-------------|
| GET | `/journal-entries` | List all journal entries |
| GET | `/journal-entries/:id` | Get journal entry by ID |
| POST | `/journal-entries` | Create journal entry |
| PUT | `/journal-entries/:id` | Update journal entry |

**Create journal entry:**
```bash
curl -X POST http://localhost:8001/api/v1/finance/journal-entries \
  -H "Content-Type: application/json" \
  -d '{
    "entry_date": "2024-03-15",
    "description": "Monthly rent payment",
    "reference": "RENT-MAR-2024",
    "lines": [
      {"account_id": "acc-500", "description": "Rent expense", "debit_amount": "2500.00", "credit_amount": "0.00"},
      {"account_id": "acc-100", "description": "Cash payment", "debit_amount": "0.00", "credit_amount": "2500.00"}
    ]
  }'
```

### Trial Balance

| Method | Path | Description |
|--------|------|-------------|
| GET | `/reports/trial-balance` | Get trial balance report |

## Human Resources API

All endpoints under `http://localhost:8002/api/v1/hr` or via gateway.

### Employees

| Method | Path | Description |
|--------|------|-------------|
| GET | `/employees` | List all employees |
| GET | `/employees/:id` | Get employee by ID |
| POST | `/employees` | Create new employee |
| PUT | `/employees/:id` | Update employee |

**Create employee:**
```bash
curl -X POST http://localhost:8002/api/v1/hr/employees \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@company.com",
    "department_id": "dept-123",
    "position_id": "pos-456",
    "employment_type": "FULL_TIME",
    "hire_date": "2024-03-15"
  }'
```

### Departments

| Method | Path | Description |
|--------|------|-------------|
| GET | `/departments` | List all departments |
| POST | `/departments` | Create department |

### Positions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/positions` | List all positions |
| POST | `/positions` | Create position |

### Attendance

| Method | Path | Description |
|--------|------|-------------|
| GET | `/attendance` | List attendance records |
| POST | `/attendance` | Record attendance |

### Leave

| Method | Path | Description |
|--------|------|-------------|
| GET | `/leave` | List leave requests |
| POST | `/leave` | Submit leave request |

### Payroll

| Method | Path | Description |
|--------|------|-------------|
| GET | `/payroll` | List payroll records |
| POST | `/payroll/process` | Process payroll |

### Performance

| Method | Path | Description |
|--------|------|-------------|
| GET | `/performance` | List performance reviews |
| POST | `/performance` | Create performance review |

## Supply Chain Management API

All endpoints under `http://localhost:8003/api/v1/scm` or via gateway.

### Products

| Method | Path | Description |
|--------|------|-------------|
| GET | `/products` | List all products |
| GET | `/products/:id` | Get product by ID |
| POST | `/products` | Create product |
| PUT | `/products/:id` | Update product |

### Suppliers

| Method | Path | Description |
|--------|------|-------------|
| GET | `/suppliers` | List all suppliers |
| POST | `/suppliers` | Create supplier |

### Purchase Orders

| Method | Path | Description |
|--------|------|-------------|
| GET | `/purchase-orders` | List purchase orders |
| GET | `/purchase-orders/:id` | Get purchase order by ID |
| POST | `/purchase-orders` | Create purchase order |

### Inventory

| Method | Path | Description |
|--------|------|-------------|
| GET | `/inventory` | List inventory levels |
| POST | `/inventory/movements` | Record inventory movement |

## Manufacturing API

All endpoints under `http://localhost:8004/api/v1/manufacturing` or via gateway.

### Production Orders

| Method | Path | Description |
|--------|------|-------------|
| GET | `/production-orders` | List production orders |
| POST | `/production-orders` | Create production order |

### Bill of Materials

| Method | Path | Description |
|--------|------|-------------|
| GET | `/boms` | List BOMs |
| POST | `/boms` | Create BOM |

### Work Orders

| Method | Path | Description |
|--------|------|-------------|
| GET | `/work-orders` | List work orders |
| POST | `/work-orders` | Create work order |

## Customer Relationship Management API

All endpoints under `http://localhost:8005/api/v1/crm` or via gateway.

### Customers

| Method | Path | Description |
|--------|------|-------------|
| GET | `/customers` | List all customers |
| GET | `/customers/:id` | Get customer by ID |
| POST | `/customers` | Create customer |
| PUT | `/customers/:id` | Update customer |

### Leads

| Method | Path | Description |
|--------|------|-------------|
| GET | `/leads` | List all leads |
| POST | `/leads` | Create lead |

### Sales Orders

| Method | Path | Description |
|--------|------|-------------|
| GET | `/sales-orders` | List sales orders |
| POST | `/sales-orders` | Create sales order |

### Opportunities

| Method | Path | Description |
|--------|------|-------------|
| GET | `/opportunities` | List opportunities |
| POST | `/opportunities` | Create opportunity |

## Project Management API

All endpoints under `http://localhost:8006/api/v1/projects` or via gateway.

### Projects

| Method | Path | Description |
|--------|------|-------------|
| GET | `/projects` | List all projects |
| GET | `/projects/:id` | Get project by ID |
| POST | `/projects` | Create project |
| PUT | `/projects/:id` | Update project |

### Tasks

| Method | Path | Description |
|--------|------|-------------|
| GET | `/tasks` | List tasks |
| POST | `/tasks` | Create task |
| PUT | `/tasks/:id` | Update task |

### Resources

| Method | Path | Description |
|--------|------|-------------|
| GET | `/resources` | List resources |
| POST | `/resources` | Create resource |

### Timesheets

| Method | Path | Description |
|--------|------|-------------|
| GET | `/timesheets` | List timesheet entries |
| POST | `/timesheets` | Create timesheet entry |

## Standard Response Format

All endpoints return JSON:

### Success Response
```json
{
  "id": "acc_1749267184000000000",
  "account_name": "Accounts Receivable",
  ...
}
```

### Error Response
```json
{
  "error": "account not found"
}
```

## HTTP Status Codes

- `200 OK` — Request successful
- `201 Created` — Resource created
- `400 Bad Request` — Invalid request format
- `404 Not Found` — Resource not found
- `500 Internal Server Error` — Server error

> Note: List endpoints return the **full dataset** in a single response — there is no pagination implemented.

## Next Steps

- [Authentication Guide](authentication.md) — Auth service usage
- [Troubleshooting](troubleshooting.md) — Common API issues
