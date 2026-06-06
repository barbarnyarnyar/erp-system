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

## Account Management

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
      "account_number": "1100",
      "name": "Cash - Operating",
      "type": "ASSET",
      "parent_id": null,
      "balance": "50000.00",
      "currency": "USD",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-03-15T14:25:00Z"
    }
  ]
}
```

### Create Account
```http
POST /api/v1/accounts
Content-Type: application/json

{
  "account_number": "1100",
  "name": "Cash - Operating",
  "type": "ASSET",
  "parent_id": "",
  "currency": "USD"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "acc_1234567890",
    "account_number": "1100",
    "name": "Cash - Operating",
    "type": "ASSET",
    "parent_id": null,
    "balance": "0",
    "currency": "USD",
    "is_active": true,
    "created_at": "2024-03-15T14:25:00Z",
    "updated_at": "2024-03-15T14:25:00Z"
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
    "account_number": "1100",
    "name": "Cash - Operating",
    "type": "ASSET",
    "parent_id": null,
    "balance": "50000.00",
    "currency": "USD",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-03-15T14:25:00Z"
  }
}
```

### Update Account
```http
PUT /api/v1/accounts/:id
Content-Type: application/json

{
  "name": "Cash - Operating Account",
  "type": "ASSET",
  "parent_id": "",
  "is_active": true
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
  "balance": "50000.00"
}
```

---

## Journal Entries

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
      "reference": "JE-2024-001",
      "date": "2024-03-15T14:30:00Z",
      "description": "Monthly rent payment",
      "status": "POSTED",
      "created_by": "system",
      "reversed_by": null,
      "created_at": "2024-03-15T14:30:00Z",
      "updated_at": "2024-03-15T14:30:00Z"
    }
  ]
}
```

### Create Journal Entry
```http
POST /api/v1/journal-entries
Content-Type: application/json

{
  "reference": "JE-2024-001",
  "description": "Monthly rent payment",
  "lines": [
    {
      "account_id": "acc_500",
      "debit_amount": "2500.00",
      "credit_amount": "0",
      "description": "Rent expense"
    },
    {
      "account_id": "acc_100",
      "debit_amount": "0",
      "credit_amount": "2500.00",
      "description": "Cash payment"
    }
  ]
}
```

Validation rules:
- Minimum 2 lines
- Total debits must equal total credits
- All account IDs must exist

Response `201 Created`:
```json
{
  "data": {
    "id": "je_1234567890",
    "reference": "JE-2024-001",
    "date": "2024-03-15T14:30:00Z",
    "description": "Monthly rent payment",
    "status": "POSTED",
    "created_by": "system",
    "reversed_by": null,
    "created_at": "2024-03-15T14:30:00Z",
    "updated_at": "2024-03-15T14:30:00Z"
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
    "reference": "JE-2024-001",
    "date": "2024-03-15T14:30:00Z",
    "description": "Monthly rent payment",
    "status": "POSTED",
    "created_by": "system",
    "reversed_by": null,
    "created_at": "2024-03-15T14:30:00Z",
    "updated_at": "2024-03-15T14:30:00Z"
  },
  "lines": [
    {
      "id": "jel_0_1234567890",
      "entry_id": "je_1234567890",
      "account_id": "acc_500",
      "debit_amount": "2500.00",
      "credit_amount": "0",
      "description": "Rent expense"
    },
    {
      "id": "jel_1_1234567890",
      "entry_id": "je_1234567890",
      "account_id": "acc_100",
      "debit_amount": "0",
      "credit_amount": "2500.00",
      "description": "Cash payment"
    }
  ]
}
```

### Update Journal Entry
```http
PUT /api/v1/journal-entries/:id
Content-Type: application/json

{
  "reference": "JE-2024-001-UPDATED",
  "description": "Updated description",
  "lines": [...]
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

### Reverse Journal Entry
There is no dedicated endpoint. Reversal is performed through service logic — see general ledger implementation.

---

## Invoices

### List Invoices
```http
GET /api/v1/invoices
```

### Create Invoice
```http
POST /api/v1/invoices
Content-Type: application/json

{
  "customer_id": "cust_001",
  "issue_date": "2024-03-15T00:00:00Z",
  "due_date": "2024-04-15T00:00:00Z",
  "lines": [
    {
      "description": "Widget A",
      "quantity": 10,
      "unit_price": "25.00"
    }
  ]
}
```

### Get Invoice
```http
GET /api/v1/invoices/:id
```

### Update Invoice
```http
PUT /api/v1/invoices/:id
Content-Type: application/json

{
  "customer_id": "cust_002"
}
```

### Delete Invoice
```http
DELETE /api/v1/invoices/:id
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
> **Note**: No email is actually sent — this toggles `is_sent` status and publishes a Kafka event.

---

## Payments

### List Payments
```http
GET /api/v1/payments
```

### Record Payment
```http
POST /api/v1/payments
Content-Type: application/json

{
  "invoice_id": "inv_123",
  "bill_id": "",
  "bank_account_id": "",
  "amount": "2500.00",
  "payment_method": "bank_transfer"
}
```

### Get Payment
```http
GET /api/v1/payments/:id
```

---

## Reports

### Balance Sheet
```http
GET /api/v1/reports/balance-sheet
```

Response:
```json
{
  "report": {
    "assets": {"Cash - Operating": "50000.00"},
    "total_assets": "50000.00",
    "liabilities": {"Accounts Payable": "15000.00"},
    "total_liabilities": "15000.00",
    "equity": {"Retained Earnings": "35000.00"},
    "total_equity": "35000.00"
  }
}
```

### Income Statement (Stub)
```http
GET /api/v1/reports/income-statement
```

Returns hardcoded message — no actual revenue/expense aggregation.

### Cash Flow Report (Stub)
```http
GET /api/v1/reports/cash-flow
```

Returns hardcoded message — no actual cash flow calculation.

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

No webhook support. Events are published to Kafka topics only.

## Error Responses

```json
{
  "error": "account not found"
}
```

```json
{
  "error": "journal entry is unbalanced: debits=2500.00, credits=2000.00"
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
