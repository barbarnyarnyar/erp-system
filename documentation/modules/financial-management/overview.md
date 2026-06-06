# Financial Management Overview

General ledger, accounts receivable, accounts payable, cash management, and budgeting with double-entry accounting.

## Core Features Summary

### General Ledger and Chart of Accounts
**Purpose**: Central repository for all financial accounts with hierarchical structure and balance tracking.

**Implemented Features:**
- Account CRUD with type classification (ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE)
- Account number and name management
- Parent-child account hierarchy
- Balance tracking with `decimal.Decimal` precision
- Currency field per account (string only — no conversion logic)
- Active/inactive status management
- Account-level balance retrieval
- Journal entries with double-entry balance validation (debits = credits)
- Automatic balance update on journal posting (type-aware: debit-increase for ASSET/EXPENSE)
- Reversal entries with debit/credit swap
- Trial balance report grouped by debit-type vs credit-type accounts
- Balance sheet report grouped by ASSET/LIABILITY/EQUITY types
- Kafka events on account create/update and balance change

### Journal Entry Management
**Purpose**: Record financial transactions with balance validation.

```json
// POST /api/v1/journal-entries
{
  "reference": "JE-2024-001",
  "description": "Monthly rent payment",
  "lines": [
    {"account_id": "acc_123", "debit_amount": "2500.00", "credit_amount": "0", "description": "Rent expense"},
    {"account_id": "acc_456", "debit_amount": "0", "credit_amount": "2500.00", "description": "Cash payment"}
  ]
}
```

**Validation Rules:**
- Minimum 2 lines per entry
- Total debits must equal total credits
- All referenced accounts must exist
- Status: always POSTED on creation (no draft/approval workflow)
- Reversal supported: swaps debit/credit amounts, links original entry

### Accounts Receivable
**Purpose**: Manage customer invoices.

**Implemented Features:**
- Invoice CRUD with line items (description, quantity, unit price)
- Line total auto-calculation (quantity × unit price)
- Invoice send action (status toggle only — no actual email delivery)
- Invoice event publishing

**Not implemented:** Credit limits, payment terms tracking, dunning/collections, aging analysis.

### Accounts Payable
**Purpose**: Manage vendor bills.

A `VendorBill` domain model and `AccountsPayableService` are defined with basic CRUD methods. The `VendorBillHandler` is defined but **not wired** into routes — there are no `/api/v1/vendor-bills` endpoints.

### Cash Management
**Purpose**: Record and track payments.

```json
// POST /api/v1/payments
{
  "invoice_id": "inv_123",
  "amount": "2500.00",
  "payment_method": "bank_transfer"
}
```

**Implemented Features:**
- Payment recording against invoices or bills
- Payment listing and retrieval
- Payment event publishing

**Not implemented:** Bank reconciliation, cash flow forecasting (endpoint returns hardcoded stub), payment batch processing.

### Budgeting
**Purpose**: Budget planning and monitoring.

A `Budget` domain model, `BudgetingService`, and `MemoryBudgetRepo` exist with full CRUD. Supports budget allocation, variance calculation, and cross-service consumption (PM, HR, MFG listen for budget events).

### Financial Reporting

| Report | Implementation |
|--------|---------------|
| **Balance Sheet** | Real — iterates accounts, classifies by type (ASSET/LIABILITY/EQUITY), sums balances |
| **Income Statement** | Stub — returns hardcoded success message, no actual revenue/expense calculation |
| **Cash Flow** | Stub — returns hardcoded success message, no cash flow logic |

## Implementation Details

### Account Type Rules
```
ASSET, EXPENSE     → Debit-increase (debits add, credits subtract)
LIABILITY, EQUITY,
REVENUE            → Credit-increase (debits subtract, credits add)
```

### Trial Balance
```
ASSET + EXPENSE    → debit side
LIABILITY + EQUITY + REVENUE → credit side
```

### Balance Sheet
```
Accounts where Type == "ASSET"     → assets
Accounts where Type == "LIABILITY" → liabilities
Accounts where Type == "EQUITY"    → equity
Revenue and Expense accounts are NOT included in the balance sheet (no income statement integration).
```

## Integration Points

### Internal Module Integration
- **HR Module**: Consumes `hr.payroll.processed` → creates salary journal entries
- **SCM Module**: Consumes `scm.purchase.order.created` → creates inventory-in-transit entries
- **SCM Module**: Consumes `scm.invoice.received` → creates AP entries
- **SCM Module**: Consumes `scm.inventory.valued` → updates inventory GL balance
- **CRM Module**: Consumes `crm.sale.completed` → creates revenue entries
- **MFG Module**: Consumes `mfg.production.completed` → creates WIP→finished goods entries
- **MFG Module**: Consumes `mfg.material.consumed` → creates raw material issue entries
- **PM Module**: Consumes `prj.time.logged` → creates unbilled receivable entries

### Kafka Events Published (16 topics)
`fin.invoice.created`, `fin.invoice.updated`, `fin.invoice.sent`, `fin.invoice.paid`, `fin.invoice.overdue`, `fin.payment.received`, `fin.payment.processed`, `fin.payment.failed`, `fin.vendor.payment.due`, `fin.account.created`, `fin.account.updated`, `fin.account.balance.changed`, `fin.budget.created`, `fin.budget.updated`, `fin.budget.exceeded`, `fin.budget.approved`, `fin.budget.allocated`, `fin.cost.budget.allocated`

## Implementation Status vs Documentation

| Feature Claimed | Actual Status |
|----------------|--------------|
| Multi-currency support | Domain model has `Currency` field — no conversion logic, all USD |
| Accounts Payable (vendor bills) | Domain model + service exist — no routes wired |
| Bank reconciliation | `BankAccount`, `BankStatement` models exist — no logic |
| Income statement | Stub endpoint — no revenue/expense aggregation |
| Cash flow statement | Stub endpoint — no cash flow calculation |
| Budget variance | Implemented — `GetBudgetVariance` with actual vs budget comparison |
| Double-entry accounting | Fully implemented — balance validation, type-aware posting, reversal |
| Trial balance | Implemented — by debit-type vs credit-type classification |
| Balance sheet | Implemented — by ASSET/LIABILITY/EQUITY classification |
| Three-way matching | Not implemented |
| 1099 processing | Not implemented |
| Payment batch processing | Not implemented |
