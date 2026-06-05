# Financial Management API Reference

Complete REST API documentation for the Financial Management module.

## Base URL
```
/api/v1/finance
```

## Authentication
All endpoints require JWT authentication via Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Account Management API

### List Accounts
```http
GET /api/v1/finance/accounts
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Results per page (default: 20, max: 100)
- `type` (optional): Filter by account type (ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE)
- `active` (optional): Filter by active status (true/false)
- `parent_id` (optional): Filter by parent account
- `include_balances` (optional): Include current balances (true/false)

**Response:**
```json
{
  "data": [
    {
      "id": "acc-123",
      "account_code": "1000",
      "account_name": "Cash - Operating",
      "account_type": "ASSET",
      "parent_account_id": null,
      "account_level": 1,
      "normal_side": "DEBIT",
      "current_balance": "25000.00",
      "is_active": true,
      "allow_posting": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-03-15T14:25:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "per_page": 20,
    "total_pages": 5,
    "total_items": 95
  }
}
```

### Create Account
```http
POST /api/v1/finance/accounts
Content-Type: application/json

{
  "account_code": "1100",
  "account_name": "Accounts Receivable",
  "account_type": "ASSET",
  "parent_account_id": "acc-123",
  "normal_side": "DEBIT",
  "allow_posting": true,
  "description": "Customer receivables from sales"
}
```

**Response:**
```json
{
  "id": "acc-124",
  "account_code": "1100",
  "account_name": "Accounts Receivable",
  "account_type": "ASSET",
  "parent_account_id": "acc-123",
  "account_level": 2,
  "normal_side": "DEBIT",
  "current_balance": "0.00",
  "is_active": true,
  "allow_posting": true,
  "created_at": "2024-03-15T14:25:00Z",
  "updated_at": "2024-03-15T14:25:00Z"
}
```

### Get Account Details
```http
GET /api/v1/finance/accounts/{account_id}
```

**Response:**
```json
{
  "id": "acc-124",
  "account_code": "1100",
  "account_name": "Accounts Receivable",
  "account_type": "ASSET",
  "parent_account_id": "acc-123",
  "account_level": 2,
  "normal_side": "DEBIT",
  "current_balance": "15000.00",
  "is_active": true,
  "allow_posting": true,
  "child_accounts": [
    {
      "id": "acc-125",
      "account_code": "1110",
      "account_name": "Trade Receivables"
    }
  ],
  "recent_transactions": [
    {
      "id": "je-456",
      "date": "2024-03-15",
      "description": "Customer Invoice #1001",
      "debit_amount": "1000.00",
      "credit_amount": "0.00"
    }
  ]
}
```

### Update Account
```http
PUT /api/v1/finance/accounts/{account_id}
Content-Type: application/json

{
  "account_name": "Accounts Receivable - Trade",
  "description": "Trade receivables from customers",
  "is_active": true
}
```

### Get Account Balance
```http
GET /api/v1/finance/accounts/{account_id}/balance
```

**Query Parameters:**
- `as_of_date` (optional): Balance as of specific date (YYYY-MM-DD)
- `include_pending` (optional): Include unposted transactions (true/false)

**Response:**
```json
{
  "account_id": "acc-124",
  "account_code": "1100",
  "account_name": "Accounts Receivable",
  "current_balance": "15000.00",
  "as_of_date": "2024-03-15",
  "balance_breakdown": {
    "beginning_balance": "12000.00",
    "period_debits": "5000.00",
    "period_credits": "2000.00",
    "ending_balance": "15000.00"
  },
  "currency": "USD"
}
```

## Journal Entry API

### List Journal Entries
```http
GET /api/v1/finance/journal-entries
```

**Query Parameters:**
- `date_from` (optional): Start date filter (YYYY-MM-DD)
- `date_to` (optional): End date filter (YYYY-MM-DD)
- `status` (optional): Filter by status (DRAFT, POSTED, REVERSED)
- `account_id` (optional): Filter by account
- `source_module` (optional): Filter by originating module
- `page` (optional): Page number
- `limit` (optional): Results per page

**Response:**
```json
{
  "data": [
    {
      "id": "je-456",
      "entry_number": "JE-2024-001",
      "entry_date": "2024-03-15",
      "posting_date": "2024-03-15T14:30:00Z",
      "description": "Monthly rent payment",
      "reference": "RENT-MAR-2024",
      "source_module": "finance",
      "total_amount": "2500.00",
      "status": "POSTED",
      "created_by": "user-123",
      "lines": [
        {
          "account_code": "5000",
          "account_name": "Rent Expense",
          "debit_amount": "2500.00",
          "credit_amount": "0.00"
        },
        {
          "account_code": "1000",
          "account_name": "Cash",
          "debit_amount": "0.00",
          "credit_amount": "2500.00"
        }
      ]
    }
  ],
  "pagination": {
    "current_page": 1,
    "per_page": 20,
    "total_pages": 12,
    "total_items": 235
  }
}
```

### Create Journal Entry
```http
POST /api/v1/finance/journal-entries
Content-Type: application/json

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
      "credit_amount": "0.00",
      "department_code": "ADMIN",
      "cost_center": "CC001"
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

**Response:**
```json
{
  "id": "je-457",
  "entry_number": "JE-2024-002",
  "entry_date": "2024-03-15",
  "posting_date": null,
  "description": "Monthly rent payment",
  "reference": "RENT-MAR-2024",
  "source_module": "finance",
  "total_amount": "2500.00",
  "status": "DRAFT",
  "requires_approval": false,
  "created_by": "user-123",
  "created_at": "2024-03-15T14:30:00Z",
  "lines": [
    {
      "id": "jel-789",
      "account_id": "acc-500",
      "account_code": "5000",
      "description": "Rent expense",
      "debit_amount": "2500.00",
      "credit_amount": "0.00",
      "department_code": "ADMIN",
      "cost_center": "CC001"
    },
    {
      "id": "jel-790",
      "account_id": "acc-100",
      "account_code": "1000",
      "description": "Cash payment",
      "debit_amount": "0.00",
      "credit_amount": "2500.00"
    }
  ]
}
```

### Post Journal Entry
```http
POST /api/v1/finance/journal-entries/{entry_id}/post
Content-Type: application/json

{
  "posting_date": "2024-03-15T14:30:00Z",
  "notes": "Posting approved by manager"
}
```

### Reverse Journal Entry
```http
POST /api/v1/finance/journal-entries/{entry_id}/reverse
Content-Type: application/json

{
  "reversal_date": "2024-03-16",
  "reason": "Correction required",
  "notes": "Incorrect amount posted"
}
```

## Financial Reports API

### Balance Sheet
```http
GET /api/v1/finance/reports/balance-sheet
```

**Query Parameters:**
- `as_of_date`: Balance sheet date (YYYY-MM-DD) - Required
- `format` (optional): Response format (json, pdf, excel) - Default: json
- `include_zero_balances` (optional): Include accounts with zero balances (true/false)
- `consolidate_subsidiaries` (optional): Include subsidiary data (true/false)

**Response:**
```json
{
  "report_title": "Balance Sheet",
  "as_of_date": "2024-03-31",
  "currency": "USD",
  "assets": {
    "current_assets": {
      "cash_and_equivalents": "50000.00",
      "accounts_receivable": "75000.00",
      "inventory": "100000.00",
      "prepaid_expenses": "10000.00",
      "total_current_assets": "235000.00"
    },
    "fixed_assets": {
      "property_plant_equipment": "500000.00",
      "accumulated_depreciation": "-150000.00",
      "net_ppe": "350000.00",
      "intangible_assets": "25000.00",
      "total_fixed_assets": "375000.00"
    },
    "total_assets": "610000.00"
  },
  "liabilities": {
    "current_liabilities": {
      "accounts_payable": "45000.00",
      "accrued_expenses": "15000.00",
      "current_portion_ltd": "20000.00",
      "total_current_liabilities": "80000.00"
    },
    "long_term_liabilities": {
      "long_term_debt": "150000.00",
      "total_long_term_liabilities": "150000.00"
    },
    "total_liabilities": "230000.00"
  },
  "equity": {
    "share_capital": "200000.00",
    "retained_earnings": "180000.00",
    "total_equity": "380000.00"
  },
  "total_liabilities_and_equity": "610000.00"
}
```

### Income Statement
```http
GET /api/v1/finance/reports/income-statement
```

**Query Parameters:**
- `period_start`: Start of period (YYYY-MM-DD) - Required
- `period_end`: End of period (YYYY-MM-DD) - Required
- `format` (optional): Response format (json, pdf, excel) - Default: json
- `comparison_period` (optional): Include prior period comparison (true/false)

### Trial Balance
```http
GET /api/v1/finance/reports/trial-balance
```

**Query Parameters:**
- `as_of_date`: Trial balance date (YYYY-MM-DD) - Required
- `include_zero_balances` (optional): Include zero balance accounts (true/false)
- `account_type` (optional): Filter by account type
- `format` (optional): Response format (json, pdf, excel)

### General Ledger
```http
GET /api/v1/finance/reports/general-ledger
```

**Query Parameters:**
- `account_id`: Specific account ID - Required
- `date_from`: Start date (YYYY-MM-DD) - Required
- `date_to`: End date (YYYY-MM-DD) - Required
- `include_beginning_balance` (optional): Include opening balance (true/false)

## Vendor Management API

### List Vendors
```http
GET /api/v1/finance/vendors
```

### Create Vendor
```http
POST /api/v1/finance/vendors
Content-Type: application/json

{
  "vendor_code": "VEN001",
  "vendor_name": "Office Supplies Inc",
  "contact_name": "John Smith",
  "email": "john@officesupplies.com",
  "phone": "+1-555-123-4567",
  "address": {
    "street": "123 Business St",
    "city": "Business City",
    "state": "CA",
    "postal_code": "12345",
    "country": "US"
  },
  "payment_terms": "NET_30",
  "credit_limit": "50000.00"
}
```

## Error Responses

### Validation Error
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input provided",
    "details": [
      {
        "field": "account_code",
        "message": "Account code must be unique"
      },
      {
        "field": "debit_amount",
        "message": "Debit amount must be positive"
      }
    ]
  },
  "timestamp": "2024-03-15T14:30:00Z",
  "request_id": "req-abc123"
}
```

### Business Rule Error
```json
{
  "error": {
    "code": "BUSINESS_RULE_VIOLATION",
    "message": "Journal entry does not balance",
    "details": [
      {
        "rule": "DEBIT_CREDIT_BALANCE",
        "message": "Total debits (2500.00) must equal total credits (2000.00)"
      }
    ]
  },
  "timestamp": "2024-03-15T14:30:00Z",
  "request_id": "req-def456"
}
```

## Rate Limits

- **Standard Users**: 1000 requests per hour
- **Premium Users**: 5000 requests per hour
- **Reporting Endpoints**: 100 requests per hour (due to processing overhead)

Rate limit headers included in responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Next Steps

- [Overview](overview.md) - Module features and capabilities
- [General Ledger](general-ledger.md) - Account management details
- [Journal Entries](journal-entries.md) - Transaction processing
- [Database Schema](database-schema.md) - Data model implementation