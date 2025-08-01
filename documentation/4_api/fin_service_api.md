# Financial Management (FIN) Service - API Specifications

## Overview

This document defines the REST API specifications for the Financial Management (FIN) microservice. The API provides comprehensive financial management capabilities including general ledger, accounts payable, accounts receivable, and financial reporting.

## API Base Information

- **Base URL**: `https://api.erp-system.com/api/v1/fin`
- **Version**: v1.0
- **Authentication**: Bearer JWT tokens
- **Content Type**: `application/json`
- **Rate Limiting**: 5000 requests per hour per user (financial operations)

## Authentication

All API endpoints require authentication via JWT Bearer tokens with appropriate financial permissions.

```http
Authorization: Bearer <jwt_token>
```

### Required Headers

```http
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Request-ID: <unique_request_id>
X-Idempotency-Key: <idempotency_key> (for write operations)
```

---

## General Ledger API

### 1. Chart of Accounts Management

#### Create Account

Creates a new account in the chart of accounts.

**Endpoint**: `POST /gl/accounts`  
**Permission**: Finance Admin, Controller

##### Request Body

```json
{
  "account_code": "1000",
  "account_name": "Cash - Operating",
  "account_type": "ASSET",
  "parent_account_id": "550e8400-e29b-41d4-a716-446655440000",
  "normal_side": "DEBIT",
  "is_active": true,
  "allow_posting": true,
  "is_control_account": false,
  "description": "Primary operating cash account"
}
```

##### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "account_code": "1000",
  "account_name": "Cash - Operating",
  "account_type": "ASSET",
  "parent_account_id": "550e8400-e29b-41d4-a716-446655440000",
  "account_level": 2,
  "normal_side": "DEBIT",
  "current_balance": 0.00,
  "debit_balance": 0.00,
  "credit_balance": 0.00,
  "is_active": true,
  "allow_posting": true,
  "is_control_account": false,
  "created_at": "2024-03-01T10:00:00Z",
  "updated_at": "2024-03-01T10:00:00Z"
}
```

#### Get Account Balance

Retrieves current balance for a specific account.

**Endpoint**: `GET /gl/accounts/{accountId}/balance`  
**Permission**: Finance User, Department Manager (restricted)

##### Query Parameters

- `as_of_date` (date, optional): Balance as of specific date (YYYY-MM-DD)
- `include_unposted` (boolean, default: false): Include draft journal entries

##### Response

```json
{
  "account_id": "550e8400-e29b-41d4-a716-446655440001",
  "account_code": "1000",
  "account_name": "Cash - Operating",
  "account_type": "ASSET",
  "normal_side": "DEBIT",
  "current_balance": 45250.75,
  "debit_balance": 125750.50,
  "credit_balance": 80499.75,
  "as_of_date": "2024-03-15",
  "last_updated": "2024-03-15T14:30:00Z"
}
```

#### Get Chart of Accounts

Retrieves the complete chart of accounts with hierarchy.

**Endpoint**: `GET /gl/accounts`  
**Permission**: Finance User

##### Query Parameters

- `account_type` (string): Filter by account type (ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE)
- `active_only` (boolean, default: true): Only return active accounts
- `include_balances` (boolean, default: false): Include current balances
- `hierarchy` (boolean, default: true): Return hierarchical structure

##### Response

```json
{
  "accounts": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "account_code": "1000",
      "account_name": "Assets",
      "account_type": "ASSET",
      "account_level": 1,
      "is_control_account": true,
      "current_balance": 285750.25,
      "children": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440001", 
          "account_code": "1100",
          "account_name": "Current Assets",
          "account_level": 2,
          "current_balance": 185250.50,
          "children": [
            {
              "account_code": "1110",
              "account_name": "Cash - Operating",
              "current_balance": 45250.75
            }
          ]
        }
      ]
    }
  ],
  "summary": {
    "total_accounts": 45,
    "by_type": {
      "ASSET": 15,
      "LIABILITY": 8,
      "EQUITY": 3,
      "REVENUE": 12,
      "EXPENSE": 7
    }
  }
}
```

### 2. Journal Entry Management

#### Create Journal Entry

Creates a new journal entry with multiple lines.

**Endpoint**: `POST /gl/journal-entries`  
**Permission**: Finance User, Controller

##### Request Body

```json
{
  "entry_date": "2024-03-15",
  "description": "Monthly office rent payment",
  "reference": "CHECK-001234",
  "source_module": "MANUAL",
  "requires_approval": false,
  "lines": [
    {
      "line_number": 1,
      "account_id": "550e8400-e29b-41d4-a716-446655440010",
      "debit_amount": 5000.00,
      "credit_amount": 0.00,
      "description": "March 2024 office rent",
      "department_id": "550e8400-e29b-41d4-a716-446655440020"
    },
    {
      "line_number": 2,
      "account_id": "550e8400-e29b-41d4-a716-446655440001",
      "debit_amount": 0.00,
      "credit_amount": 5000.00,
      "description": "Payment from operating cash"
    }
  ]
}
```

##### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440030",
  "entry_number": "JE-2024-001234",
  "entry_date": "2024-03-15",
  "posting_date": null,
  "description": "Monthly office rent payment",
  "reference": "CHECK-001234",
  "source_module": "MANUAL",
  "total_amount": 5000.00,
  "status": "DRAFT",
  "requires_approval": false,
  "lines": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440031",
      "line_number": 1,
      "account": {
        "id": "550e8400-e29b-41d4-a716-446655440010",
        "account_code": "6100",
        "account_name": "Rent Expense"
      },
      "debit_amount": 5000.00,
      "credit_amount": 0.00,
      "description": "March 2024 office rent"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440032",
      "line_number": 2,
      "account": {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "account_code": "1000",
        "account_name": "Cash - Operating"
      },
      "debit_amount": 0.00,
      "credit_amount": 5000.00,
      "description": "Payment from operating cash"
    }
  ],
  "created_at": "2024-03-15T10:30:00Z",
  "created_by": "user@company.com"
}
```

#### Post Journal Entry

Posts a draft journal entry to update account balances.

**Endpoint**: `POST /gl/journal-entries/{entryId}/post`  
**Permission**: Finance User, Controller

##### Request Body

```json
{
  "posting_date": "2024-03-15",
  "notes": "Month-end posting approved by controller"
}
```

##### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440030",
  "entry_number": "JE-2024-001234",
  "status": "POSTED",
  "posting_date": "2024-03-15",
  "posted_at": "2024-03-15T15:45:00Z",
  "posted_by": "controller@company.com",
  "account_balances_updated": [
    {
      "account_id": "550e8400-e29b-41d4-a716-446655440010",
      "account_code": "6100",
      "old_balance": 45000.00,
      "new_balance": 50000.00,
      "change": 5000.00
    },
    {
      "account_id": "550e8400-e29b-41d4-a716-446655440001",
      "account_code": "1000",
      "old_balance": 50250.75,
      "new_balance": 45250.75,
      "change": -5000.00
    }
  ]
}
```

#### Get Trial Balance

Generates trial balance report for specified date.

**Endpoint**: `GET /gl/trial-balance`  
**Permission**: Finance User, Controller, CFO

##### Query Parameters

- `as_of_date` (date, required): Trial balance date (YYYY-MM-DD)
- `account_level` (integer, default: all): Limit to specific account level
- `include_zero_balances` (boolean, default: false): Include accounts with zero balance

##### Response

```json
{
  "trial_balance": {
    "as_of_date": "2024-03-15",
    "generated_at": "2024-03-15T16:00:00Z",
    "accounts": [
      {
        "account_id": "550e8400-e29b-41d4-a716-446655440001",
        "account_code": "1000",
        "account_name": "Cash - Operating",
        "account_type": "ASSET",
        "debit_balance": 45250.75,
        "credit_balance": 0.00
      },
      {
        "account_id": "550e8400-e29b-41d4-a716-446655440010",
        "account_code": "6100", 
        "account_name": "Rent Expense",
        "account_type": "EXPENSE",
        "debit_balance": 50000.00,
        "credit_balance": 0.00
      }
    ],
    "totals": {
      "total_debits": 285750.25,
      "total_credits": 285750.25,
      "difference": 0.00,
      "is_balanced": true
    }
  }
}
```

---

## Accounts Payable API

### 1. Vendor Management

#### Create Vendor

Creates a new vendor record.

**Endpoint**: `POST /ap/vendors`  
**Permission**: Finance User, AP Clerk

##### Request Body

```json
{
  "vendor_code": "VEN-001",
  "vendor_name": "ABC Office Supplies Inc",
  "email": "ap@abcoffice.com",
  "phone": "+1-555-0123",
  "address_line1": "123 Business St",
  "city": "New York",
  "state": "NY",
  "postal_code": "10001",
  "country": "US",
  "tax_id": "12-3456789",
  "payment_terms": "NET30",
  "credit_limit": 50000.00,
  "vendor_category": "OFFICE_SUPPLIES",
  "is_1099_vendor": true,
  "form_1099_type": "NEC"
}
```

##### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440040",
  "vendor_code": "VEN-001",
  "vendor_name": "ABC Office Supplies Inc",
  "email": "ap@abcoffice.com",
  "payment_terms": "NET30",
  "credit_limit": 50000.00,
  "current_balance": 0.00,
  "ytd_purchases": 0.00,
  "vendor_category": "OFFICE_SUPPLIES",
  "is_1099_vendor": true,
  "is_active": true,
  "created_at": "2024-03-01T10:00:00Z"
}
```

### 2. Invoice Management

#### Create AP Invoice

Records a new vendor invoice.

**Endpoint**: `POST /ap/invoices`  
**Permission**: Finance User, AP Clerk

##### Request Body

```json
{
  "vendor_id": "550e8400-e29b-41d4-a716-446655440040",
  "vendor_invoice_number": "INV-2024-001",
  "invoice_date": "2024-03-15",
  "due_date": "2024-04-14",
  "subtotal_amount": 1200.00,
  "tax_amount": 96.00,
  "total_amount": 1296.00,
  "description": "Office supplies - March 2024",
  "purchase_order_id": "550e8400-e29b-41d4-a716-446655440050",
  "lines": [
    {
      "line_number": 1,
      "description": "Printer paper - 10 reams",
      "quantity": 10,
      "unit_price": 45.00,
      "line_amount": 450.00,
      "account_id": "550e8400-e29b-41d4-a716-446655440060"
    },
    {
      "line_number": 2,
      "description": "Office pens - 5 boxes",
      "quantity": 5,
      "unit_price": 150.00,
      "line_amount": 750.00,
      "account_id": "550e8400-e29b-41d4-a716-446655440060"
    }
  ]
}
```

##### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440070",
  "invoice_number": "AP-2024-001234",
  "vendor_invoice_number": "INV-2024-001",
  "vendor": {
    "id": "550e8400-e29b-41d4-a716-446655440040",
    "vendor_name": "ABC Office Supplies Inc",
    "payment_terms": "NET30"
  },
  "invoice_date": "2024-03-15",
  "due_date": "2024-04-14",
  "subtotal_amount": 1200.00,
  "tax_amount": 96.00,
  "total_amount": 1296.00,
  "paid_amount": 0.00,
  "outstanding_amount": 1296.00,
  "status": "PENDING",
  "journal_entry_id": "550e8400-e29b-41d4-a716-446655440080",
  "created_at": "2024-03-15T11:00:00Z"
}
```

### 3. Payment Processing

#### Process Vendor Payment

Processes payment to one or more vendors.

**Endpoint**: `POST /ap/payments`  
**Permission**: Finance User, AP Clerk

##### Request Body

```json
{
  "payment_date": "2024-04-14",
  "payment_method": "ACH",
  "bank_account_id": "550e8400-e29b-41d4-a716-446655440090",
  "description": "Weekly vendor payments - ACH batch",
  "invoices": [
    {
      "invoice_id": "550e8400-e29b-41d4-a716-446655440070",
      "payment_amount": 1296.00,
      "discount_taken": 0.00
    }
  ]
}
```

##### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440100",
  "payment_number": "PAY-2024-001234",
  "payment_date": "2024-04-14",
  "payment_method": "ACH",
  "total_amount": 1296.00,
  "status": "PROCESSED",
  "bank_account": {
    "id": "550e8400-e29b-41d4-a716-446655440090",
    "account_name": "Operating Checking"
  },
  "vendor_payments": [
    {
      "vendor_id": "550e8400-e29b-41d4-a716-446655440040",
      "vendor_name": "ABC Office Supplies Inc",
      "payment_amount": 1296.00,
      "invoices_paid": [
        {
          "invoice_id": "550e8400-e29b-41d4-a716-446655440070",
          "invoice_number": "AP-2024-001234",
          "amount_paid": 1296.00,
          "discount_taken": 0.00
        }
      ]
    }
  ],
  "journal_entry_id": "550e8400-e29b-41d4-a716-446655440110",
  "created_at": "2024-04-14T10:30:00Z"
}
```

#### Get AP Aging Report

Generates accounts payable aging report.

**Endpoint**: `GET /ap/aging`  
**Permission**: Finance User, AP Clerk, Controller

##### Query Parameters

- `as_of_date` (date, required): Aging report date (YYYY-MM-DD)
- `vendor_id` (UUID, optional): Specific vendor aging
- `include_zero_balances` (boolean, default: false): Include vendors with zero balance

##### Response

```json
{
  "ap_aging": {
    "as_of_date": "2024-04-15",
    "generated_at": "2024-04-15T09:00:00Z",
    "vendors": [
      {
        "vendor_id": "550e8400-e29b-41d4-a716-446655440040",
        "vendor_name": "ABC Office Supplies Inc",
        "total_balance": 2500.00,
        "aging_buckets": {
          "current": 1200.00,
          "30_days": 800.00,
          "60_days": 500.00,
          "90_days": 0.00,
          "over_90_days": 0.00
        },
        "overdue_amount": 1300.00,
        "payment_terms": "NET30"
      }
    ],
    "summary": {
      "total_outstanding": 15750.50,
      "current": 8250.25,
      "30_days": 4500.00,
      "60_days": 2000.25,
      "90_days": 1000.00,
      "over_90_days": 0.00,
      "total_overdue": 7500.25
    }
  }
}
```

---

## Accounts Receivable API

### 1. Customer Management

#### Create Customer

Creates a new customer record.

**Endpoint**: `POST /ar/customers`  
**Permission**: Finance User, AR Clerk

##### Request Body

```json
{
  "customer_code": "CUS-001",
  "customer_name": "Tech Solutions LLC",
  "email": "accounting@techsolutions.com",
  "phone": "+1-555-0987",
  "billing_address_line1": "456 Innovation Dr",
  "billing_city": "San Francisco",
  "billing_state": "CA",
  "billing_postal_code": "94105",
  "billing_country": "US",
  "credit_limit": 100000.00,
  "payment_terms": "NET30",
  "customer_category": "ENTERPRISE",
  "tax_exempt": false
}
```

### 2. Invoice Management

#### Create AR Invoice

Generates a customer invoice.

**Endpoint**: `POST /ar/invoices`  
**Permission**: Finance User, AR Clerk, Sales Rep

##### Request Body

```json
{
  "customer_id": "550e8400-e29b-41d4-a716-446655440120",
  "invoice_date": "2024-03-15",
  "due_date": "2024-04-14",
  "sales_order_id": "550e8400-e29b-41d4-a716-446655440130",
  "subtotal_amount": 15000.00,
  "tax_amount": 1200.00,
  "total_amount": 16200.00,
  "description": "Software development services - March 2024", 
  "lines": [
    {
      "line_number": 1,
      "description": "Development hours - 100 hrs @ $150/hr",
      "quantity": 100,
      "unit_price": 150.00,
      "line_amount": 15000.00,
      "revenue_account_id": "550e8400-e29b-41d4-a716-446655440140"
    }
  ]
}
```

##### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440150",
  "invoice_number": "INV-2024-001234",
  "customer": {
    "id": "550e8400-e29b-41d4-a716-446655440120",
    "customer_name": "Tech Solutions LLC",
    "payment_terms": "NET30"
  },
  "invoice_date": "2024-03-15",
  "due_date": "2024-04-14",
  "subtotal_amount": 15000.00,
  "tax_amount": 1200.00,
  "total_amount": 16200.00,
  "paid_amount": 0.00,
  "outstanding_amount": 16200.00,
  "status": "SENT",
  "journal_entry_id": "550e8400-e29b-41d4-a716-446655440160",
  "created_at": "2024-03-15T14:00:00Z"
}
```

### 3. Payment Processing

#### Record Customer Payment

Records a payment received from customer.

**Endpoint**: `POST /ar/payments`  
**Permission**: Finance User, AR Clerk

##### Request Body

```json
{
  "customer_id": "550e8400-e29b-41d4-a716-446655440120",
  "payment_date": "2024-04-10",
  "payment_method": "ACH",
  "amount": 16200.00,
  "reference_number": "ACH-789012",
  "description": "Payment for INV-2024-001234",
  "bank_account_id": "550e8400-e29b-41d4-a716-446655440090",
  "invoice_applications": [
    {
      "invoice_id": "550e8400-e29b-41d4-a716-446655440150",
      "applied_amount": 16200.00,
      "discount_taken": 0.00
    }
  ]
}
```

##### Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440170",
  "payment_number": "REC-2024-001234",
  "customer": {
    "id": "550e8400-e29b-41d4-a716-446655440120",
    "customer_name": "Tech Solutions LLC"
  },
  "payment_date": "2024-04-10",
  "payment_method": "ACH",
  "amount": 16200.00,
  "reference_number": "ACH-789012",
  "status": "PROCESSED",
  "invoice_applications": [
    {
      "invoice_id": "550e8400-e29b-41d4-a716-446655440150",
      "invoice_number": "INV-2024-001234",
      "applied_amount": 16200.00,
      "remaining_balance": 0.00
    }
  ],
  "journal_entry_id": "550e8400-e29b-41d4-a716-446655440180",
  "created_at": "2024-04-10T15:30:00Z"
}
```

#### Get AR Aging Report

Generates accounts receivable aging report.

**Endpoint**: `GET /ar/aging`  
**Permission**: Finance User, AR Clerk, Controller, Sales Manager

##### Response

```json
{
  "ar_aging": {
    "as_of_date": "2024-04-15",
    "generated_at": "2024-04-15T09:00:00Z",
    "customers": [
      {
        "customer_id": "550e8400-e29b-41d4-a716-446655440120",
        "customer_name": "Tech Solutions LLC",
        "total_balance": 25000.00,
        "aging_buckets": {
          "current": 16200.00,
          "30_days": 8800.00,
          "60_days": 0.00,
          "90_days": 0.00,
          "over_90_days": 0.00
        },
        "overdue_amount": 8800.00,
        "credit_limit": 100000.00,
        "available_credit": 75000.00
      }
    ],
    "summary": {
      "total_outstanding": 125750.00,
      "current": 89250.00,
      "30_days": 25500.00,
      "60_days": 8000.00,
      "90_days": 3000.00,
      "over_90_days": 0.00,
      "total_overdue": 36500.00
    }
  }
}
```

---

## Financial Reporting API

### 1. Standard Financial Statements

#### Generate Balance Sheet

Creates balance sheet report.

**Endpoint**: `GET /reports/balance-sheet`  
**Permission**: Finance User, Controller, CFO

##### Query Parameters

- `as_of_date` (date, required): Balance sheet date (YYYY-MM-DD)
- `comparison_date` (date, optional): Comparison period date
- `format` (string, default: "summary"): Detail level (summary, detail)

##### Response

```json
{
  "balance_sheet": {
    "as_of_date": "2024-03-31",
    "company_name": "ERP Company Inc",
    "generated_at": "2024-04-01T09:00:00Z",
    "assets": {
      "current_assets": {
        "cash_and_equivalents": 125750.50,
        "accounts_receivable": 89250.00,
        "inventory": 65000.00,
        "prepaid_expenses": 12500.00,
        "total_current_assets": 292500.50
      },
      "fixed_assets": {
        "equipment": 150000.00,
        "accumulated_depreciation": -45000.00,
        "net_fixed_assets": 105000.00
      },
      "total_assets": 397500.50
    },
    "liabilities": {
      "current_liabilities": {
        "accounts_payable": 45250.25,
        "accrued_expenses": 15750.00,
        "current_portion_long_term_debt": 10000.00,
        "total_current_liabilities": 71000.25
      },
      "long_term_liabilities": {
        "long_term_debt": 125000.00,
        "total_long_term_liabilities": 125000.00
      },
      "total_liabilities": 196000.25
    },
    "equity": {
      "common_stock": 100000.00,
      "retained_earnings": 101500.25,
      "total_equity": 201500.25
    },
    "total_liabilities_and_equity": 397500.50,
    "balance_check": {
      "is_balanced": true,
      "difference": 0.00
    }
  }
}
```

#### Generate Income Statement

Creates income statement report.

**Endpoint**: `GET /reports/income-statement`  
**Permission**: Finance User, Controller, CFO

##### Query Parameters

- `start_date` (date, required): Period start date (YYYY-MM-DD)
- `end_date` (date, required): Period end date (YYYY-MM-DD)
- `comparison_start_date` (date, optional): Comparison period start
- `comparison_end_date` (date, optional): Comparison period end
- `format` (string, default: "summary"): Detail level

##### Response

```json
{
  "income_statement": {
    "period": {
      "start_date": "2024-03-01",
      "end_date": "2024-03-31"
    },
    "company_name": "ERP Company Inc",
    "generated_at": "2024-04-01T09:00:00Z",
    "revenue": {
      "service_revenue": 156250.00,
      "product_revenue": 89750.00,
      "other_revenue": 2500.00,
      "total_revenue": 248500.00
    },
    "cost_of_goods_sold": {
      "materials": 45000.00,
      "direct_labor": 32500.00,
      "manufacturing_overhead": 15750.00,
      "total_cogs": 93250.00
    },
    "gross_profit": 155250.00,
    "operating_expenses": {
      "salaries_and_wages": 85000.00,
      "rent": 15000.00,
      "utilities": 3500.00,
      "insurance": 2500.00,
      "office_supplies": 1250.00,
      "depreciation": 3750.00,
      "other_expenses": 5500.00,
      "total_operating_expenses": 116500.00
    },
    "operating_income": 38750.00,
    "other_income_expense": {
      "interest_expense": -2500.00,
      "other_income": 750.00,
      "total_other": -1750.00
    },
    "net_income_before_taxes": 37000.00,
    "income_tax_expense": 9250.00,
    "net_income": 27750.00,
    "earnings_per_share": {
      "basic": 2.78,
      "diluted": 2.78
    }
  }
}
```

#### Generate Cash Flow Statement

Creates cash flow statement report.

**Endpoint**: `GET /reports/cash-flow`  
**Permission**: Finance User, Controller, CFO

##### Response

```json
{
  "cash_flow_statement": {
    "period": {
      "start_date": "2024-03-01", 
      "end_date": "2024-03-31"
    },
    "operating_activities": {
      "net_income": 27750.00,
      "adjustments": {
        "depreciation": 3750.00,
        "accounts_receivable_change": -15250.00,
        "inventory_change": -8500.00,
        "accounts_payable_change": 12250.50,
        "accrued_expenses_change": 2500.00
      },
      "net_cash_from_operations": 22500.50
    },
    "investing_activities": {
      "equipment_purchases": -25000.00,
      "net_cash_from_investing": -25000.00
    },
    "financing_activities": {
      "loan_proceeds": 15000.00,
      "loan_payments": -5000.00,
      "dividends_paid": -10000.00,
      "net_cash_from_financing": 0.00
    },
    "net_change_in_cash": -2499.50,
    "cash_beginning_of_period": 128250.00,
    "cash_end_of_period": 125750.50
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
    "code": "FINANCIAL_VALIDATION_ERROR",
    "message": "Journal entry does not balance",
    "details": [
      {
        "field": "lines",
        "message": "Total debits (5000.00) do not equal total credits (4500.00)",
        "debit_total": 5000.00,
        "credit_total": 4500.00,
        "difference": 500.00
      }
    ],
    "request_id": "550e8400-e29b-41d4-a716-446655440200",
    "timestamp": "2024-03-15T10:30:00Z"
  }
}
```

### Common Financial Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `JOURNAL_ENTRY_NOT_BALANCED` | 400 | Journal entry debits do not equal credits |
| `ACCOUNT_NOT_FOUND` | 404 | Referenced account does not exist |
| `ACCOUNT_INACTIVE` | 400 | Cannot post to inactive account |
| `INSUFFICIENT_PERMISSIONS` | 403 | User lacks required financial permissions |
| `PERIOD_CLOSED` | 400 | Cannot post to closed fiscal period |
| `DUPLICATE_TRANSACTION` | 409 | Transaction with same reference already exists |
| `NEGATIVE_BALANCE_NOT_ALLOWED` | 400 | Transaction would create invalid negative balance |
| `CREDIT_LIMIT_EXCEEDED` | 400 | Transaction exceeds customer credit limit |
| `VENDOR_ON_HOLD` | 400 | Cannot process payment to vendor on hold |

---

## Webhook Events

### Financial Event Notifications

The FIN service publishes events for integration with other ERP modules:

#### Journal Entry Posted Event

```json
{
  "event_type": "fin.journal_entry.posted",
  "event_id": "550e8400-e29b-41d4-a716-446655440300",
  "timestamp": "2024-03-15T15:45:00Z",
  "data": {
    "journal_entry_id": "550e8400-e29b-41d4-a716-446655440030",
    "entry_number": "JE-2024-001234",
    "posting_date": "2024-03-15",
    "total_amount": 5000.00,
    "source_module": "MANUAL",
    "account_impacts": [
      {
        "account_id": "550e8400-e29b-41d4-a716-446655440010",
        "account_code": "6100",
        "impact_amount": 5000.00,
        "impact_type": "DEBIT"
      }
    ]
  }
}
```

#### Payment Processed Event

```json
{
  "event_type": "fin.payment.processed",
  "event_id": "550e8400-e29b-41d4-a716-446655440301",
  "timestamp": "2024-04-14T10:30:00Z",
  "data": {
    "payment_id": "550e8400-e29b-41d4-a716-446655440100",
    "payment_number": "PAY-2024-001234",
    "payment_type": "AP_PAYMENT",
    "entity_id": "550e8400-e29b-41d4-a716-446655440040",
    "entity_name": "ABC Office Supplies Inc",
    "amount": 1296.00,
    "payment_method": "ACH",
    "payment_date": "2024-04-14"
  }
}
```

---

## Rate Limiting & Performance

### API Rate Limits
- **Standard Users**: 5000 requests per hour
- **Finance Users**: 10000 requests per hour  
- **System Integration**: 50000 requests per hour

### Performance Expectations
- **Account Balance Queries**: <100ms response time
- **Journal Entry Creation**: <500ms response time
- **Financial Report Generation**: <30 seconds for standard reports
- **Complex Aging Reports**: <60 seconds for large datasets

### Caching Strategy
- **Account Balances**: Cached for 5 minutes with real-time invalidation
- **Chart of Accounts**: Cached for 1 hour
- **Financial Reports**: Cached for 15 minutes
- **Exchange Rates**: Cached for 1 hour

This comprehensive API specification provides the foundation for building robust financial applications that integrate seamlessly with the ERP ecosystem while maintaining the highest standards of accuracy and compliance.