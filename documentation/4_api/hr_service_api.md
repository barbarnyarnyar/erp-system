# HR/HCM Service - API Specifications

## Overview

This document defines the REST API specifications for the HR/HCM microservice. The API follows RESTful principles and provides comprehensive employee lifecycle management, time tracking, and leave management capabilities.

## API Base Information

- **Base URL**: `https://api.erp-system.com/api/v1/hr`
- **Version**: v1.0
- **Authentication**: Bearer JWT tokens
- **Content Type**: `application/json`
- **Rate Limiting**: 1000 requests per hour per user

## Authentication

All API endpoints require authentication via JWT Bearer tokens obtained from the Authentication Service.

```http
Authorization: Bearer <jwt_token>
```

### Required Headers

```http
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Request-ID: <unique_request_id>
```

---

## Employee Management API

### 1. Create Employee

Creates a new employee record in the system.

**Endpoint**: `POST /employees`  
**Permission**: HR Admin only

#### Request Body

```json
{
  "employee_id": "EMP001",
  "first_name": "John",
  "last_name": "Doe",
  "middle_name": "Michael",
  "email": "john.doe@company.com",
  "phone": "+1-555-0123",
  "date_of_birth": "1990-01-15",
  "hire_date": "2024-03-01",
  "employment_status": "active",
  "employee_type": "full_time",
  "base_salary": 75000.00,
  "currency": "USD",
  "pay_frequency": "bi_weekly",
  "department_id": "550e8400-e29b-41d4-a716-446655440000",
  "position_id": "550e8400-e29b-41d4-a716-446655440001",
  "manager_id": "550e8400-e29b-41d4-a716-446655440002",
  "address": {
    "line1": "123 Main St",
    "line2": "Apt 4B",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "US"
  }
}
```

#### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "employee_id": "EMP001",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@company.com",
  "hire_date": "2024-03-01",
  "employment_status": "active",
  "department": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Engineering",
    "code": "ENG"
  },
  "position": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "Software Engineer",
    "level": 2
  },
  "created_at": "2024-03-01T10:00:00Z",
  "updated_at": "2024-03-01T10:00:00Z"
}
```

#### Status Codes

- `201 Created` - Employee created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `409 Conflict` - Employee ID or email already exists

### 2. Get Employee

Retrieves employee information by ID.

**Endpoint**: `GET /employees/{id}`  
**Permission**: HR Admin, Manager (direct reports only), Self

#### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "employee_id": "EMP001",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@company.com",
  "phone": "+1-555-0123",
  "hire_date": "2024-03-01",
  "employment_status": "active",
  "employee_type": "full_time",
  "base_salary": 75000.00,
  "department": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Engineering",
    "code": "ENG"
  },
  "position": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "Software Engineer",
    "level": 2
  },
  "manager": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "name": "Jane Smith",
    "email": "jane.smith@company.com"
  },
  "created_at": "2024-03-01T10:00:00Z",
  "updated_at": "2024-03-01T10:00:00Z"
}
```

### 3. Update Employee

Updates employee information.

**Endpoint**: `PUT /employees/{id}`  
**Permission**: HR Admin only

#### Request Body

```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@company.com",
  "phone": "+1-555-0124",
  "employment_status": "active",
  "base_salary": 80000.00,
  "department_id": "550e8400-e29b-41d4-a716-446655440000",
  "position_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

### 4. Search Employees

Searches employees with filtering and pagination.

**Endpoint**: `GET /employees`  
**Permission**: HR Admin, Manager (team members only)

#### Query Parameters

- `search` (string): Search term for name or employee ID
- `department_id` (UUID): Filter by department
- `employment_status` (string): Filter by status
- `manager_id` (UUID): Filter by manager
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 20, max: 100)
- `sort` (string): Sort field (default: last_name)
- `order` (string): Sort order (asc|desc, default: asc)

#### Response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "employee_id": "EMP001",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@company.com",
      "employment_status": "active",
      "department": {
        "name": "Engineering"
      },
      "position": {
        "title": "Software Engineer"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

### 5. Terminate Employee

Terminates an employee's employment.

**Endpoint**: `POST /employees/{id}/terminate`  
**Permission**: HR Admin only

#### Request Body

```json
{
  "termination_date": "2024-12-31",
  "reason": "resignation",
  "notes": "Two weeks notice provided"
}
```

---

## Department Management API

### 1. Create Department

**Endpoint**: `POST /departments`  
**Permission**: HR Admin only

#### Request Body

```json
{
  "department_code": "ENG",
  "department_name": "Engineering",
  "description": "Software development and engineering teams",
  "parent_department_id": "550e8400-e29b-41d4-a716-446655440004",
  "department_manager_id": "550e8400-e29b-41d4-a716-446655440002",
  "cost_center": "CC-ENG-001",
  "budget_amount": 5000000.00
}
```

### 2. Get Department Hierarchy

**Endpoint**: `GET /departments/hierarchy`  
**Permission**: HR Admin, Manager

#### Response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "department_code": "ENG",
      "department_name": "Engineering",
      "manager": {
        "name": "Jane Smith"
      },
      "employee_count": 25,
      "children": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440005",
          "department_code": "ENG-FE",
          "department_name": "Frontend Engineering",
          "employee_count": 12,
          "children": []
        }
      ]
    }
  ]
}
```

---

## Time & Attendance API

### 1. Clock In

Records employee clock-in time.

**Endpoint**: `POST /time/clock-in`  
**Permission**: Employee (self only), HR Admin

#### Request Body

```json
{
  "employee_id": "550e8400-e29b-41d4-a716-446655440003",
  "location": "Office - Main Building",
  "notes": "Starting regular shift"
}
```

#### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440010",
  "employee_id": "550e8400-e29b-41d4-a716-446655440003",
  "entry_date": "2024-03-01",
  "clock_in_time": "2024-03-01T09:00:00Z",
  "location": "Office - Main Building",
  "status": "in_progress"
}
```

### 2. Clock Out

Records employee clock-out time.

**Endpoint**: `POST /time/clock-out`  
**Permission**: Employee (self only), HR Admin

#### Request Body

```json
{
  "employee_id": "550e8400-e29b-41d4-a716-446655440003",
  "notes": "End of regular shift"
}
```

### 3. Get Time Entries

Retrieves time entries for an employee.

**Endpoint**: `GET /time/entries`  
**Permission**: Employee (self only), Manager (direct reports), HR Admin

#### Query Parameters

- `employee_id` (UUID): Employee ID
- `start_date` (date): Start date (YYYY-MM-DD)
- `end_date` (date): End date (YYYY-MM-DD)
- `status` (string): Filter by approval status

#### Response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440010",
      "employee_id": "550e8400-e29b-41d4-a716-446655440003",
      "entry_date": "2024-03-01",
      "clock_in_time": "2024-03-01T09:00:00Z",
      "clock_out_time": "2024-03-01T17:30:00Z",
      "total_hours": 8.5,
      "regular_hours": 8.0,
      "overtime_hours": 0.5,
      "status": "pending",
      "location": "Office - Main Building"
    }
  ],
  "summary": {
    "total_hours": 42.5,
    "regular_hours": 40.0,
    "overtime_hours": 2.5,
    "entries_count": 5
  }
}
```

### 4. Submit Timesheet for Approval

**Endpoint**: `POST /time/timesheets/submit`  
**Permission**: Employee (self only)

#### Request Body

```json
{
  "employee_id": "550e8400-e29b-41d4-a716-446655440003",
  "pay_period_start": "2024-03-01",
  "pay_period_end": "2024-03-14",
  "time_entry_ids": [
    "550e8400-e29b-41d4-a716-446655440010",
    "550e8400-e29b-41d4-a716-446655440011"
  ]
}
```

### 5. Approve Timesheet

**Endpoint**: `POST /time/timesheets/{id}/approve`  
**Permission**: Manager (direct reports only), HR Admin

#### Request Body

```json
{
  "approval_status": "approved",
  "notes": "Timesheet approved for payroll processing"
}
```

---

## Leave Management API

### 1. Get Leave Balances

Retrieves leave balances for an employee.

**Endpoint**: `GET /leave/balances`  
**Permission**: Employee (self only), Manager (direct reports), HR Admin

#### Query Parameters

- `employee_id` (UUID): Employee ID
- `year` (integer): Accrual year (default: current year)

#### Response

```json
{
  "data": [
    {
      "leave_type": "vacation",
      "accrual_year": 2024,
      "annual_allocation": 160.0,
      "accrued_balance": 100.0,
      "used_balance": 40.0,
      "pending_balance": 8.0,
      "available_balance": 52.0
    },
    {
      "leave_type": "sick",
      "accrual_year": 2024,
      "annual_allocation": 80.0,
      "accrued_balance": 50.0,
      "used_balance": 8.0,
      "pending_balance": 0.0,
      "available_balance": 42.0
    }
  ]
}
```

### 2. Submit Leave Request

**Endpoint**: `POST /leave/requests`  
**Permission**: Employee (self only), HR Admin

#### Request Body

```json
{
  "employee_id": "550e8400-e29b-41d4-a716-446655440003",
  "leave_type": "vacation",
  "start_date": "2024-04-15",
  "end_date": "2024-04-19",
  "total_days": 5.0,
  "reason": "Family vacation",
  "is_paid": true
}
```

#### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440020",
  "employee_id": "550e8400-e29b-41d4-a716-446655440003",
  "leave_type": "vacation",
  "start_date": "2024-04-15",
  "end_date": "2024-04-19",
  "total_days": 5.0,
  "reason": "Family vacation",
  "request_status": "pending",
  "requested_at": "2024-03-01T14:30:00Z"
}
```

### 3. Approve/Reject Leave Request

**Endpoint**: `POST /leave/requests/{id}/decision`  
**Permission**: Manager (direct reports only), HR Admin

#### Request Body

```json
{
  "decision": "approved",
  "notes": "Approved - adequate coverage arranged"
}
```

### 4. Get Leave Requests

**Endpoint**: `GET /leave/requests`  
**Permission**: Employee (self only), Manager (direct reports), HR Admin

#### Query Parameters

- `employee_id` (UUID): Filter by employee
- `status` (string): Filter by status
- `start_date` (date): Filter by date range
- `end_date` (date): Filter by date range

#### Response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440020",
      "employee": {
        "name": "John Doe",
        "employee_id": "EMP001"
      },
      "leave_type": "vacation",
      "start_date": "2024-04-15",
      "end_date": "2024-04-19",
      "total_days": 5.0,
      "request_status": "pending",
      "requested_at": "2024-03-01T14:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

---

## Reporting API

### 1. Employee Summary Report

**Endpoint**: `GET /reports/employee-summary`  
**Permission**: HR Admin, Manager (department only)

#### Query Parameters

- `department_id` (UUID): Filter by department
- `employment_status` (string): Filter by status
- `as_of_date` (date): Report as of specific date

#### Response

```json
{
  "report_date": "2024-03-01",
  "summary": {
    "total_employees": 250,
    "active_employees": 235,
    "new_hires_this_month": 8,
    "terminations_this_month": 3
  },
  "by_department": [
    {
      "department": "Engineering",
      "active_count": 45,
      "inactive_count": 2,
      "avg_tenure_months": 28
    }
  ],
  "by_employment_type": {
    "full_time": 220,
    "part_time": 15,
    "contractor": 15
  }
}
```

### 2. Attendance Report

**Endpoint**: `GET /reports/attendance`  
**Permission**: HR Admin, Manager (department only)

#### Query Parameters

- `department_id` (UUID): Filter by department
- `start_date` (date): Report start date
- `end_date` (date): Report end date

#### Response

```json
{
  "report_period": {
    "start_date": "2024-03-01",
    "end_date": "2024-03-31"
  },
  "summary": {
    "total_scheduled_hours": 8800,
    "total_worked_hours": 8560,
    "attendance_rate": 97.3,
    "average_daily_attendance": 235
  },
  "by_employee": [
    {
      "employee_id": "EMP001",
      "employee_name": "John Doe",
      "scheduled_hours": 176,
      "worked_hours": 174,
      "attendance_rate": 98.9,
      "late_days": 1,
      "absent_days": 0
    }
  ]
}
```

---

## Event Integration API

### Webhook Events

The HR service publishes events to the message queue for integration with other ERP services.

#### Employee Created Event

```json
{
  "event_type": "employee.created",
  "event_id": "550e8400-e29b-41d4-a716-446655440030",
  "timestamp": "2024-03-01T10:00:00Z",
  "data": {
    "employee_id": "550e8400-e29b-41d4-a716-446655440003",
    "employee_code": "EMP001",
    "first_name": "John",
    "last_name": "Doe",
    "department_id": "550e8400-e29b-41d4-a716-446655440000",
    "department_code": "ENG",
    "base_salary": 75000.00,
    "currency": "USD",
    "hire_date": "2024-03-01",
    "cost_center": "CC-ENG-001"
  }
}
```

#### Employee Terminated Event

```json
{
  "event_type": "employee.terminated",
  "event_id": "550e8400-e29b-41d4-a716-446655440031",
  "timestamp": "2024-12-31T17:00:00Z",
  "data": {
    "employee_id": "550e8400-e29b-41d4-a716-446655440003",
    "employee_code": "EMP001",
    "termination_date": "2024-12-31",
    "reason": "resignation"
  }
}
```

#### Salary Changed Event

```json
{
  "event_type": "employee.salary_changed",
  "event_id": "550e8400-e29b-41d4-a716-446655440032",
  "timestamp": "2024-06-01T10:00:00Z",
  "data": {
    "employee_id": "550e8400-e29b-41d4-a716-446655440003",
    "employee_code": "EMP001",
    "old_salary": 75000.00,
    "new_salary": 80000.00,
    "effective_date": "2024-06-01",
    "currency": "USD"
  }
}
```

---

## Error Handling

### Error Response Format

All API errors follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": [
      {
        "field": "email",
        "message": "Email address is required"
      },
      {
        "field": "hire_date",
        "message": "Hire date cannot be in the future"
      }
    ],
    "request_id": "550e8400-e29b-41d4-a716-446655440040"
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `AUTHENTICATION_REQUIRED` | 401 | Authentication token required |
| `INSUFFICIENT_PERMISSIONS` | 403 | User lacks required permissions |
| `RESOURCE_NOT_FOUND` | 404 | Requested resource not found |
| `DUPLICATE_RESOURCE` | 409 | Resource already exists |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit exceeded |
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

---

## API Client Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

class HRClient {
  constructor(baseURL, token) {
    this.client = axios.create({
      baseURL,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async createEmployee(employeeData) {
    try {
      const response = await this.client.post('/employees', employeeData);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to create employee: ${error.response.data.error.message}`);
    }
  }

  async getEmployee(employeeId) {
    try {
      const response = await this.client.get(`/employees/${employeeId}`);
      return response.data;
    } catch (error) {
      if (error.response.status === 404) {
        return null;
      }
      throw new Error(`Failed to get employee: ${error.response.data.error.message}`);
    }
  }

  async clockIn(employeeId, location) {
    try {
      const response = await this.client.post('/time/clock-in', {
        employee_id: employeeId,
        location: location
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to clock in: ${error.response.data.error.message}`);
    }
  }
}

// Usage
const hrClient = new HRClient('https://api.erp-system.com/api/v1/hr', 'your-jwt-token');

// Create employee
const newEmployee = await hrClient.createEmployee({
  employee_id: 'EMP001',
  first_name: 'John',
  last_name: 'Doe',
  email: 'john.doe@company.com',
  hire_date: '2024-03-01'
});

// Clock in
await hrClient.clockIn(newEmployee.id, 'Office - Main Building');
```

### Python

```python
import requests
from datetime import datetime
from typing import Optional, Dict, Any

class HRClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }

    def create_employee(self, employee_data: Dict[str, Any]) -> Dict[str, Any]:
        response = requests.post(
            f'{self.base_url}/employees',
            json=employee_data,
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def get_employee(self, employee_id: str) -> Optional[Dict[str, Any]]:
        response = requests.get(
            f'{self.base_url}/employees/{employee_id}',
            headers=self.headers
        )
        
        if response.status_code == 404:
            return None
        
        response.raise_for_status()
        return response.json()

    def submit_leave_request(self, employee_id: str, leave_type: str, 
                           start_date: str, end_date: str, reason: str) -> Dict[str, Any]:
        leave_data = {
            'employee_id': employee_id,
            'leave_type': leave_type,
            'start_date': start_date,
            'end_date': end_date,
            'reason': reason,
            'is_paid': True
        }
        
        response = requests.post(
            f'{self.base_url}/leave/requests',
            json=leave_data,
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

# Usage
hr_client = HRClient('https://api.erp-system.com/api/v1/hr', 'your-jwt-token')

# Create employee
employee = hr_client.create_employee({
    'employee_id': 'EMP001',
    'first_name': 'John',
    'last_name': 'Doe',
    'email': 'john.doe@company.com',
    'hire_date': '2024-03-01'
})

# Submit leave request
leave_request = hr_client.submit_leave_request(
    employee['id'],
    'vacation',
    '2024-04-15',
    '2024-04-19',
    'Family vacation'
)
```

---

## Rate Limiting

The API implements rate limiting to ensure fair usage and system stability:

- **Standard Users**: 1000 requests per hour
- **HR Admins**: 5000 requests per hour
- **System Integration**: 10000 requests per hour

Rate limit headers are included in all responses:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1640995200
```

---

## API Versioning

The API uses URL path versioning:

- **Current Version**: `/api/v1/hr`
- **Beta Version**: `/api/v2/hr` (when available)

Version-specific changes:
- **v1.0**: Initial release with core functionality
- **v1.1**: Added bulk operations and enhanced reporting
- **v2.0**: (Planned) GraphQL support and advanced analytics

This API specification provides a comprehensive foundation for HR system integration and supports the core employee lifecycle, time tracking, and leave management requirements outlined in the business requirements.