# Financial Management API Reference

Complete REST API documentation for the Financial Management module. Port **8001**.

## Base URL
```
http://localhost:8001/api/v1
```

## Response Format

All endpoints return JSON:
```json
{
  "data": { ... },
  "error": "..."  // only on errors
}
```

Error responses include HTTP status codes:
- `400 Bad Request` — validation error
- `404 Not Found` — resource not found
- `500 Internal Server Error` — server error

> **Note**: There is no authentication on any endpoint. No rate limiting is applied.

---

## Legal Entities

Multi-tenant partition shield for all universal ledger entities.

### List Legal Entities
```http
GET /api/v1/legal-entities
```

Response:
```json
{
  "data": [
    {
      "id": "le_1234567890",
      "company_code": "CORP_DE",
      "company_name": "Acme Germany GmbH",
      "functional_currency": "EUR",
      "tax_registration_number": "DE123456789",
      "created_at": "2026-01-15T10:30:00Z",
      "updated_at": "2026-01-15T10:30:00Z"
    }
  ]
}
```

### Create Legal Entity
```http
POST /api/v1/legal-entities
Content-Type: application/json

{
  "company_code": "CORP_DE",
  "company_name": "Acme Germany GmbH",
  "functional_currency": "EUR",
  "tax_registration_number": "DE123456789"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "le_1234567890",
    "company_code": "CORP_DE",
    "company_name": "Acme Germany GmbH",
    "functional_currency": "EUR",
    "tax_registration_number": "DE123456789",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Legal Entity
```http
GET /api/v1/legal-entities/:id
```

Response:
```json
{
  "data": {
    "id": "le_1234567890",
    "company_code": "CORP_DE",
    "company_name": "Acme Germany GmbH",
    "functional_currency": "EUR",
    "tax_registration_number": "DE123456789",
    "created_at": "2026-01-15T10:30:00Z",
    "updated_at": "2026-01-15T10:30:00Z"
  }
}
```

---

## Account Management

Chart of accounts definitions for legal entities.

### List Accounts
```http
GET /api/v1/accounts
```

Response:
```json
{
  "data": [
    {
      "id": "acc_1234567890",
      "legal_entity_id": "le_1234567890",
      "account_code": "1100",
      "account_name": "Cash - Operating",
      "type": "ASSET",
      "is_active": true,
      "created_at": "2026-01-15T10:30:00Z",
      "updated_at": "2026-03-15T14:25:00Z"
    }
  ]
}
```

### Create Account
```http
POST /api/v1/accounts
Content-Type: application/json

{
  "legal_entity_id": "le_1234567890",
  "account_code": "1100",
  "account_name": "Cash - Operating",
  "type": "ASSET"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "acc_1234567890",
    "legal_entity_id": "le_1234567890",
    "account_code": "1100",
    "account_name": "Cash - Operating",
    "type": "ASSET",
    "is_active": true,
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Account
```http
GET /api/v1/accounts/:id
```

Response:
```json
{
  "data": {
    "id": "acc_1234567890",
    "legal_entity_id": "le_1234567890",
    "account_code": "1100",
    "account_name": "Cash - Operating",
    "type": "ASSET",
    "is_active": true,
    "created_at": "2026-01-15T10:30:00Z",
    "updated_at": "2026-03-15T14:25:00Z"
  }
}
```

### Update Account
```http
PUT /api/v1/accounts/:id
Content-Type: application/json

{
  "account_name": "Cash - Operating Account",
  "type": "ASSET",
  "is_active": true
}
```

Response `200 OK`:
```json
{
  "data": {
    "id": "acc_1234567890",
    "legal_entity_id": "le_1234567890",
    "account_code": "1100",
    "account_name": "Cash - Operating Account",
    "type": "ASSET",
    "is_active": true,
    "created_at": "2026-01-15T10:30:00Z",
    "updated_at": "2026-06-13T02:05:00Z"
  }
}
```

### Delete Account
```http
DELETE /api/v1/accounts/:id
```

Response:
```json
{
  "message": "account deleted"
}
```

### Get Account Balance
```http
GET /api/v1/accounts/:id/balance
```

Response:
```json
{
  "balance": "50000.0000"
}
```

---

## Journal Entries

Universal Ledger transaction management.

### List Journal Entries
```http
GET /api/v1/journal-entries
```

Response:
```json
{
  "data": [
    {
      "id": "je_1234567890",
      "legal_entity_id": "le_1234567890",
      "source_module": "FM",
      "source_document_id": "doc_9876543210",
      "posting_date": "2026-06-13T02:00:00Z",
      "financial_period": "2026-06",
      "status": "POSTED",
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Create Journal Entry
```http
POST /api/v1/journal-entries
Content-Type: application/json

{
  "legal_entity_id": "le_1234567890",
  "source_module": "FM",
  "source_document_id": "doc_9876543210",
  "posting_date": "2026-06-13T02:00:00Z",
  "lines": [
    {
      "account_id": "acc_1234567890",
      "amount_functional": "-2500.00",
      "amount_transactional": "-2500.00",
      "currency_transactional": "EUR"
    },
    {
      "account_id": "acc_0987654321",
      "amount_functional": "2500.00",
      "amount_transactional": "2500.00",
      "currency_transactional": "EUR"
    }
  ]
}
```

Validation rules:
- Minimum 2 lines
- The sum of `amount_functional` across all lines must equal exactly zero (balanced journal)
- All referenced account IDs must exist

Response `201 Created`:
```json
{
  "data": {
    "id": "je_1234567890",
    "legal_entity_id": "le_1234567890",
    "source_module": "FM",
    "source_document_id": "doc_9876543210",
    "posting_date": "2026-06-13T02:00:00Z",
    "financial_period": "2026-06",
    "status": "POSTED",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Journal Entry
```http
GET /api/v1/journal-entries/:id
```

Response:
```json
{
  "data": {
    "id": "je_1234567890",
    "legal_entity_id": "le_1234567890",
    "source_module": "FM",
    "source_document_id": "doc_9876543210",
    "posting_date": "2026-06-13T02:00:00Z",
    "financial_period": "2026-06",
    "status": "POSTED",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  },
  "lines": [
    {
      "id": "jel_1111111111",
      "journal_entry_id": "je_1234567890",
      "account_id": "acc_1234567890",
      "amount_functional": "-2500.00",
      "amount_transactional": "-2500.00",
      "currency_transactional": "EUR",
      "tracking_dimensions": null
    },
    {
      "id": "jel_2222222222",
      "journal_entry_id": "je_1234567890",
      "account_id": "acc_0987654321",
      "amount_functional": "2500.00",
      "amount_transactional": "2500.00",
      "currency_transactional": "EUR",
      "tracking_dimensions": null
    }
  ]
}
```

### Update Journal Entry
```http
PUT /api/v1/journal-entries/:id
Content-Type: application/json

{
  "legal_entity_id": "le_1234567890",
  "source_module": "FM",
  "source_document_id": "doc_9876543210-updated",
  "posting_date": "2026-06-13T02:00:00Z",
  "lines": [
    {
      "account_id": "acc_1234567890",
      "amount_functional": "-3000.00",
      "amount_transactional": "-3000.00",
      "currency_transactional": "EUR"
    },
    {
      "account_id": "acc_0987654321",
      "amount_functional": "3000.00",
      "amount_transactional": "3000.00",
      "currency_transactional": "EUR"
    }
  ]
}
```

Response `200 OK`:
```json
{
  "data": {
    "id": "je_1234567890",
    "legal_entity_id": "le_1234567890",
    "source_module": "FM",
    "source_document_id": "doc_9876543210-updated",
    "posting_date": "2026-06-13T02:00:00Z",
    "financial_period": "2026-06",
    "status": "POSTED",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:05:00Z"
  }
}
```

### Delete Journal Entry
```http
DELETE /api/v1/journal-entries/:id
```

Response:
```json
{
  "message": "journal entry deleted successfully"
}
```

---

## Invoices (Accounts Receivable)

Flat-schema customer invoices representing receivables.

### List Invoices
```http
GET /api/v1/invoices
```

Response:
```json
{
  "data": [
    {
      "id": "inv_1234567890",
      "legal_entity_id": "le_1234567890",
      "invoice_number": "INV-2026-001",
      "customer_id": "cust_0010000000",
      "sales_order_id": "so_5555555555",
      "total_amount": "2500.0000",
      "tax_amount": "125.0000",
      "due_date": "2026-07-13T02:00:00Z",
      "status": "OPEN",
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Create Invoice
```http
POST /api/v1/invoices
Content-Type: application/json

{
  "legal_entity_id": "le_1234567890",
  "customer_id": "cust_0010000000",
  "sales_order_id": "so_5555555555",
  "total_amount": "2500.00",
  "tax_amount": "125.00",
  "due_date": "2026-07-13T02:00:00Z"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "inv_1234567890",
    "legal_entity_id": "le_1234567890",
    "invoice_number": "INV-2026-001",
    "customer_id": "cust_0010000000",
    "sales_order_id": "so_5555555555",
    "total_amount": "2500.0000",
    "tax_amount": "125.0000",
    "due_date": "2026-07-13T02:00:00Z",
    "status": "OPEN",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Invoice
```http
GET /api/v1/invoices/:id
```

Response:
```json
{
  "data": {
    "id": "inv_1234567890",
    "legal_entity_id": "le_1234567890",
    "invoice_number": "INV-2026-001",
    "customer_id": "cust_0010000000",
    "sales_order_id": "so_5555555555",
    "total_amount": "2500.0000",
    "tax_amount": "125.0000",
    "due_date": "2026-07-13T02:00:00Z",
    "status": "OPEN",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Update Invoice
```http
PUT /api/v1/invoices/:id
Content-Type: application/json

{
  "status": "PAID"
}
```

Response `200 OK`:
```json
{
  "data": {
    "id": "inv_1234567890",
    "legal_entity_id": "le_1234567890",
    "invoice_number": "INV-2026-001",
    "customer_id": "cust_0010000000",
    "sales_order_id": "so_5555555555",
    "total_amount": "2500.0000",
    "tax_amount": "125.0000",
    "due_date": "2026-07-13T02:00:00Z",
    "status": "PAID",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:05:00Z"
  }
}
```

### Delete Invoice
```http
DELETE /api/v1/invoices/:id
```

Response:
```json
{
  "message": "invoice deleted successfully"
}
```

### Send Invoice
```http
POST /api/v1/invoices/:id/send
```

Response:
```json
{
  "message": "invoice sent successfully"
}
```

### Get Invoice Lines
```http
GET /api/v1/invoices/:id/lines
```

Response:
```json
{
  "data": []
}
```

---

## Vendor Bills (Accounts Payable)

Flat-schema vendor bills representing payables.

### List Vendor Bills
```http
GET /api/v1/vendor-bills
```

Response:
```json
{
  "data": [
    {
      "id": "bill_1234567890",
      "legal_entity_id": "le_1234567890",
      "bill_number": "BILL-2026-088",
      "vendor_id": "vend_9999999999",
      "purchase_order_id": "po_8888888888",
      "total_amount": "4800.0000",
      "tax_amount": "240.0000",
      "due_date": "2026-07-20T00:00:00Z",
      "status": "OPEN",
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Create Vendor Bill
```http
POST /api/v1/vendor-bills
Content-Type: application/json

{
  "legal_entity_id": "le_1234567890",
  "vendor_id": "vend_9999999999",
  "bill_number": "BILL-2026-088",
  "purchase_order_id": "po_8888888888",
  "due_date": "2026-07-20T00:00:00Z",
  "total_amount": "4800.00",
  "tax_amount": "240.00"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "bill_1234567890",
    "legal_entity_id": "le_1234567890",
    "bill_number": "BILL-2026-088",
    "vendor_id": "vend_9999999999",
    "purchase_order_id": "po_8888888888",
    "total_amount": "4800.0000",
    "tax_amount": "240.0000",
    "due_date": "2026-07-20T00:00:00Z",
    "status": "OPEN",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Vendor Bill Lines
```http
GET /api/v1/vendor-bills/:id/lines
```

Response:
```json
{
  "data": []
}
```

---

## Payments & Banking

Recording outgoing/incoming payments and retrieving bank transactions.

### List Payments
```http
GET /api/v1/payments
```

Response:
```json
{
  "data": [
    {
      "id": "pay_1234567890",
      "invoice_id": "inv_1234567890",
      "bill_id": null,
      "bank_account_id": "bank_account_888",
      "payment_number": "PAY-100021",
      "payment_date": "2026-06-13T02:00:00Z",
      "amount": "2500.0000",
      "payment_method": "bank_transfer",
      "status": "PAID",
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Record Payment
```http
POST /api/v1/payments
Content-Type: application/json

{
  "invoice_id": "inv_1234567890",
  "bill_id": "",
  "bank_account_id": "bank_account_888",
  "amount": "2500.00",
  "payment_method": "bank_transfer"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "pay_1234567890",
    "invoice_id": "inv_1234567890",
    "bill_id": null,
    "bank_account_id": "bank_account_888",
    "payment_number": "PAY-100021",
    "payment_date": "2026-06-13T02:00:00Z",
    "amount": "2500.0000",
    "payment_method": "bank_transfer",
    "status": "PAID",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Payment
```http
GET /api/v1/payments/:id
```

Response:
```json
{
  "data": {
    "id": "pay_1234567890",
    "invoice_id": "inv_1234567890",
    "bill_id": null,
    "bank_account_id": "bank_account_888",
    "payment_number": "PAY-100021",
    "payment_date": "2026-06-13T02:00:00Z",
    "amount": "2500.0000",
    "payment_method": "bank_transfer",
    "status": "PAID",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Bank Statement Lines
```http
GET /api/v1/bank-statements/:id/lines
```

Response:
```json
{
  "data": [
    {
      "id": "bsl_999999",
      "statement_id": "bs_123",
      "transaction_date": "2026-06-12T10:00:00Z",
      "description": "ACH Transfer Inbound",
      "amount": "2500.0000",
      "is_matched": true
    }
  ]
}
```

---

## Assets & Depreciation

Fixed assets tracking and straight-line depreciation schedules.

### List Assets
```http
GET /api/v1/assets
```

Response:
```json
{
  "data": [
    {
      "id": "asset_1234567890",
      "legal_entity_id": "le_1234567890",
      "asset_tag": "EQ-SERVER-001",
      "eam_equipment_id": "equip_555",
      "acquisition_cost": "12000.0000",
      "accumulated_depreciation": "2000.0000",
      "useful_life_months": 60,
      "capitalization_date": "2026-01-01T00:00:00Z",
      "status": "ACTIVE",
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Capitalize Asset
```http
POST /api/v1/assets/capitalize
Content-Type: application/json

{
  "legal_entity_id": "le_1234567890",
  "asset_tag": "EQ-SERVER-001",
  "acquisition_cost": "12000.00",
  "useful_life_months": 60,
  "eam_equipment_id": "equip_555"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "asset_1234567890",
    "legal_entity_id": "le_1234567890",
    "asset_tag": "EQ-SERVER-001",
    "eam_equipment_id": "equip_555",
    "acquisition_cost": "12000.0000",
    "accumulated_depreciation": "0.0000",
    "useful_life_months": 60,
    "capitalization_date": "2026-06-13T02:00:00Z",
    "status": "ACTIVE",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Asset
```http
GET /api/v1/assets/:id
```

Response:
```json
{
  "data": {
    "id": "asset_1234567890",
    "legal_entity_id": "le_1234567890",
    "asset_tag": "EQ-SERVER-001",
    "eam_equipment_id": "equip_555",
    "acquisition_cost": "12000.0000",
    "accumulated_depreciation": "2000.0000",
    "useful_life_months": 60,
    "capitalization_date": "2026-01-01T00:00:00Z",
    "status": "ACTIVE",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Generate Depreciation Schedule
```http
POST /api/v1/assets/:id/depreciation-schedule
```

Response `200 OK`:
```json
{
  "data": [
    {
      "id": "dsl_1",
      "fixed_asset_id": "asset_1234567890",
      "fiscal_year": 2026,
      "period_number": 1,
      "depreciation_amount": "200.0000",
      "is_posted": true,
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    },
    {
      "id": "dsl_2",
      "fixed_asset_id": "asset_1234567890",
      "fiscal_year": 2026,
      "period_number": 2,
      "depreciation_amount": "200.0000",
      "is_posted": false,
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Post Monthly Depreciation
```http
POST /api/v1/assets/depreciate
Content-Type: application/json

{
  "legal_entity_id": "le_1234567890",
  "fiscal_year": 2026,
  "period_number": 2
}
```

Response `200 OK`:
```json
{
  "message": "depreciation posted successfully"
}
```

---

## Reports

Real aggregation queries over the multi-tenant general ledger lines database.

### Balance Sheet
```http
GET /api/v1/reports/balance-sheet
```

Response:
```json
{
  "report": {
    "assets": {
      "Cash - Operating": "48000.0000",
      "Equipment - Servers": "12000.0000"
    },
    "total_assets": "60000.0000",
    "liabilities": {
      "Accounts Payable": "15000.0000"
    },
    "total_liabilities": "15000.0000",
    "equity": {
      "Retained Earnings": "45000.0000"
    },
    "total_equity": "45000.0000"
  }
}
```

### Income Statement
```http
GET /api/v1/reports/income-statement
```

Response:
```json
{
  "report": {
    "revenues": {
      "Product Sales": "12000.0000",
      "Services Revenue": "3500.0000"
    },
    "total_revenue": "15500.0000",
    "expenses": {
      "Rent Expense": "2500.0000",
      "Depreciation Expense": "400.0000"
    },
    "total_expense": "2900.0000",
    "net_income": "12600.0000"
  }
}
```

### Cash Flow Report
```http
GET /api/v1/reports/cash-flow
```

Response:
```json
{
  "report": {
    "operating_inflows": {
      "Cash - Operating": "50000.0000"
    },
    "total_inflows": "50000.0000",
    "operating_outflows": {
      "Cash - Operating": "15000.0000"
    },
    "total_outflows": "15000.0000",
    "net_cash_flow": "35000.0000"
  }
}
```

---

## Health Check

```http
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "fm-service"
}
```

---

## Webhook/Event Subscriptions

Events are published to Kafka topics only. See the Overview for Kafka topics.

## Error Responses

```json
{
  "error": "account not found"
}
```

```json
{
  "error": "journal entry is unbalanced: functional sum=500.00"
}
```

```json
{
  "error": "a journal entry must have at least 2 lines"
}
```

---

## Related Documentation

- [Overview](overview.md) — Module features and capabilities
- [General Ledger](general-ledger.md) — Account management details
- [Service Architecture](../../architecture/services-overview.md) — Integration with other services
