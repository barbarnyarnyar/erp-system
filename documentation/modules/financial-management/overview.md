# Financial Management Overview

General ledger, accounts receivable, accounts payable, cash management, budgeting, and asset depreciation with multi-tenant double-entry accounting.

## Core Features Summary

### General Ledger and Chart of Accounts
**Purpose**: Central repository for all financial accounts with multi-tenant legal entity partitioning.

**Implemented Features:**
- Account CRUD with type classification (ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE).
- Account code and name management mapped per legal entity.
- Unique composite constraint on `(legal_entity_id, account_code)`.
- Balance tracking with `decimal.Decimal` precision (calculated dynamically from journal lines).
- Active/inactive status management.
- Account-level balance retrieval.
- Journal entries with double-entry balance validation (the sum of functional amount across lines must equal exactly zero).
- Automatic balance calculation by summing all posted journal lines.
- Reversal entries with swapped debit/credit amounts.
- Trial balance report grouped by debit-type vs credit-type accounts.
- Balance sheet report grouped by ASSET/LIABILITY/EQUITY types.
- Kafka events on account create/update and balance changes.

### Journal Entry Management (Universal Ledger)
**Purpose**: Record financial transactions with strict multi-tenant partitioning and balance validation.

```json
// POST /api/v1/journal-entries
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

**Validation Rules:**
- Minimum 2 lines per entry.
- Sum of `amount_functional` across all lines must equal zero.
- All referenced accounts must exist.
- Status: defaults to `POSTED` on creation.
- Reversal supported: swaps amount values, creates reversing entry, and sets the original entry's status to `REVERSED`.

### Accounts Receivable (AR Sub-ledger)
**Purpose**: Manage customer invoices.

**Implemented Features:**
- Invoice CRUD with flat-schema properties (TotalAmount, TaxAmount, CustomerID, SalesOrderID).
- Invoice send action (toggles sent status and publishes Kafka event).
- Invoice event publishing.

### Accounts Payable (AP Sub-ledger)
**Purpose**: Manage vendor bills.

**Implemented Features:**
- Vendor bill CRUD with flat-schema properties (TotalAmount, TaxAmount, VendorID, PurchaseOrderID).
- Endpoints fully wired at `/api/v1/vendor-bills` and `/api/v1/vendor-bills/:id/lines`.
- Vendor bill event publishing.

### Cash Management & Payments
**Purpose**: Record payments and reconciliation statements.

**Implemented Features:**
- Record incoming and outgoing payments against invoices or bills.
- Retrieve bank statement transactions at `/api/v1/bank-statements/:id/lines`.
- Real-time payment event publishing.
- Cash Flow Report dynamically aggregates cash inflows/outflows from bank accounts.

### Budgeting
**Purpose**: Budget planning and monitoring.

**Implemented Features:**
- Full budget CRUD and allocation.
- Variance calculation (Budget vs Actual comparison).
- Publishes budget created, updated, approved, and exceeded events.

### Fixed Assets & Depreciation
**Purpose**: Manage capitalized assets and generate depreciation schedules.

**Implemented Features:**
- Capitalize assets with useful life, cost, and EAM equipment tracking.
- Generate straight-line monthly depreciation schedules.
- Post monthly depreciation entries to the General Ledger.

### Financial Reporting

| Report | Implementation |
|--------|---------------|
| **Balance Sheet** | Real — groups and sums balances by ASSET, LIABILITY, and EQUITY. |
| **Income Statement** | Real — aggregates REVENUE and EXPENSE lines, calculates net income. |
| **Cash Flow** | Real — aggregates inflows/outflows from asset accounts containing "cash" or "bank" in their name. |

---

## Technical & Security Mechanisms

### Transactional Outbox Pattern
The service guarantees **at-least-once delivery** of domain events by storing them inside the `TransactionalOutbox` table within the same database transaction as the business state change. A background relay worker reads pending outbox messages, publishes them to Kafka, and marks them as sent.

### Event Inbox Idempotency
To prevent duplicate processing of Kafka messages, the service employs a `KafkaEventInbox` table. Every incoming consumer event check verifies if `event_id` is already processed. If it is, the message is ignored (idempotent deduplication).

### Straight-Line Depreciation Rules
The asset service calculates monthly depreciation as `AcquisitionCost / UsefulLifeMonths`. On calling `/assets/depreciate`, the system posts balancing journal entries adjusting accumulated depreciation and depreciation expense accounts.

---

## Integration Points

### Internal Module Integration
- **HR Module**: Consumes `hr.payroll.processed` → creates salary journal entries.
- **SCM Module**: Consumes `scm.purchase.order.created` → creates inventory journal entries.
- **SCM Module**: Consumes `scm.invoice.received` → creates vendor bills/AP entries.
- **SCM Module**: Consumes `scm.inventory.valued` → updates inventory GL balance.
- **CRM Module**: Consumes `crm.sale.completed` (or `crm.order.confirmed`) → creates revenue journal entries.
- **MFG Module**: Consumes `mfg.production.completed` → creates finished goods entries.
- **MFG Module**: Consumes `mfg.material.consumed` → creates raw material issue entries.
- **PM Module**: Consumes `prj.time.logged` → creates unbilled receivable entries.

### Kafka Events Published (18 topics)
All topics use the `fm.*` namespace:
- `fm.invoice.created`, `fm.invoice.updated`, `fm.invoice.sent`, `fm.invoice.paid`, `fm.invoice.overdue`
- `fm.payment.received`, `fm.payment.processed`, `fm.payment.failed`
- `fm.vendor.payment.due`, `fm.vendor.paid`
- `fm.customer.credit_status.updated`
- `fm.account.created`, `fm.account.updated`, `fm.account.balance.changed`
- `fm.budget.created`, `fm.budget.updated`, `fm.budget.exceeded`, `fm.budget.approved`

### Kafka Events Consumed (13 topics)
All events are processed via the Kafka Event Inbox:
- `hr.payroll.processed`, `hr.employee.created`, `hr.expense.submitted`
- `scm.receipt.staged`, `scm.order.shipped`, `scm.purchase.order.created`, `scm.invoice.received`, `scm.inventory.valued`
- `crm.order.confirmed`, `crm.customer.created`
- `mfg.yield.produced`, `mfg.production.completed`, `mfg.material.consumed`
- `prj.milestone.achieved`, `prj.project.created`, `prj.time.logged`, `prj.expense.incurred`

---

## Implementation Status vs Documentation

| Feature Claimed | Implementation Status |
|----------------|-----------------------|
| Multi-currency support | Supported via functional vs transactional amounts in Ledger Lines |
| Accounts Payable (vendor bills) | Fully wired and integrated (`/api/v1/vendor-bills`) |
| Bank reconciliation | `BankAccount` and `BankStatement` entities active, lines query supported |
| Income statement | Fully implemented |
| Cash flow statement | Fully implemented |
| Budget variance | Implemented |
| Double-entry accounting | Fully implemented |
| Trial balance | Implemented |
| Balance sheet | Implemented |
| Transactional Outbox | Active background worker |
| Idempotency Inbox | Active event deduplication |
| Straight-line depreciation | Capitalize, Generate Schedule, Post Monthly |
