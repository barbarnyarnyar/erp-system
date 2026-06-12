# PRD: CDD Enhancement — Event Payloads, Validation, and Type Safety

**Date:** 2026-06-06
**Status:** Active
**Supersedes:** None

## 1. Problem Statement

The project's 7 CDD contract files define entities, components, and event topics, but lack:

1. **Event Payload Definitions** — Events reference undefined types (e.g., `InvoiceEventPayload`, `CustomerCreatedEvent`)
2. **Validation Constraints** — No field-level rules (`@required`, `@min`, `@max`, `@pattern`)
3. **Type Safety Gaps** — Raw strings used for enums, decimal precision issues
4. **Legacy Artifacts** — Duplicate entities (Transaction/TransactionLine alongside JournalEntry)
5. **Optional vs Required Annotations** — Many fields lack `@optional` markers

This PRD addresses these gaps to ensure CDD is a complete source of truth for code generation.

## 2. Gap Dimensions

### 2.1 Missing Event Payload Definitions

Events reference payload types that don't exist in CDD. The code generator cannot produce correct event structs without definitions.

| Service | Event Topic | Referenced Payload | Status |
|---------|-------------|-------------------|--------|
| FM | `fin.invoice.created` | `InvoiceEventPayload` | ❌ Missing |
| FM | `fin.payment.received` | `PaymentEventPayload` | ❌ Missing |
| FM | `fin.budget.exceeded` | `BudgetEventPayload` | ❌ Missing |
| FM | `fin.budget.approved` | `BudgetApprovedEvent` | ❌ Missing |
| HR | `hr.employee.created` | `EmployeeCreatedEvent` | ❌ Missing |
| HR | `hr.payroll.processed` | `PayrollProcessedEvent` | ❌ Missing |
| HR | `hr.leave.requested` | `LeaveRequestedEvent` | ❌ Missing |
| SCM | `scm.product.created` | `ProductCreatedEvent` | ❌ Missing |
| SCM | `scm.inventory.low.stock` | `InventoryLowStockEvent` | ❌ Missing |
| SCM | `scm.purchase.order.created` | `PurchaseOrderCreatedEvent` | ❌ Missing |
| M | `mfg.production.completed` | `ProductionCompletedEvent` | ❌ Missing |
| M | `mfg.material.consumed` | `MaterialConsumedEvent` | ❌ Missing |
| CRM | `crm.customer.created` | `CustomerCreatedEvent` | ❌ Missing |
| CRM | `crm.sales.order.confirmed` | `SalesOrderConfirmedEvent` | ❌ Missing |
| PM | `prj.project.created` | `ProjectCreatedEvent` | ❌ Missing |
| PM | `prj.milestone.achieved` | `MilestoneAchievedEvent` | ❌ Missing |
| Auth | `auth.user.created` | `UserCreatedEvent` | ❌ Missing |

**Total:** 17+ event payloads missing definitions

### 2.2 Missing Validation Constraints

CDD has no mechanism to express field validation rules. This leads to:

| Issue | Example | Impact |
|-------|---------|--------|
| No `@required` | All fields assumed required | Over-validates optional fields |
| No `@min`/`@max` | `rating` in PerformanceReview (1-5) | Invalid values accepted |
| No `@pattern` | Email fields | Malformed emails accepted |
| No `@min_length`/`@max_length` | String fields | Empty strings accepted |
| No `@positive` | `quantity`, `amount` fields | Negative values accepted |

**Fields Affected:**

| Service | Entity | Field | Missing Constraint |
|---------|--------|-------|-------------------|
| HR | PerformanceReview | `rating` | `@min(1) @max(5)` |
| CRM | Customer | `email` | `@pattern(email)` |
| FM | Account | `balance` | `@min(0)` for ASSET |
| SCM | InventoryItem | `quantity_on_hand` | `@min(0)` |
| M | ProductionOrder | `quantity` | `@positive` |
| All | All | `status` | `@enum([...])` |

### 2.3 Raw String Enums (Type Safety)

7+ entities use raw `string` for status/type fields with zero compile-time protection:

| Service | Entity | Field | Valid Values |
|---------|--------|-------|--------------|
| CRM | Customer | `status` | LEAD, PROSPECT, ACTIVE, INACTIVE |
| CRM | Opportunity | `stage` | DISCOVERY, NEGOTIATION, CLOSED_WON, CLOSED_LOST |
| FM | Account | `type` | ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE |
| HR | LeaveRequest | `leave_type` | ANNUAL, SICK, UNPAID |
| HR | LeaveRequest | `status` | PENDING, APPROVED, REJECTED |
| M | ProductionOrder | `status` | DRAFT, PLANNED, IN_PROGRESS, COMPLETED, CANCELLED |
| M | WorkOrder | `status` | PENDING, IN_PROGRESS, COMPLETED, BLOCKED |

### 2.4 Legacy Artifacts

**FM Service — Transaction/TransactionLine**

These entities exist alongside JournalEntry/JournalEntryLine:
- Both serve the same purpose (GL entries)
- Both have full repo + memory implementations
- CDD comment says "New code should prefer JournalEntry"
- No deprecation marker or migration path

**Recommendation:** Mark as deprecated in CDD with migration guidance.

### 2.5 Optional vs Required Annotations

Many fields that should be optional lack `@optional`:

| Service | Entity | Field | Should Be |
|---------|--------|-------|-----------|
| FM | Invoice | `customer_id` | Required (AR) |
| FM | VendorBill | `supplier_id` | Required (AP) |
| FM | Payment | `invoice_id` | Optional (can be AP or AR) |
| FM | Payment | `bill_id` | Optional (can be AP or AR) |
| HR | Employee | `term_date` | Optional (active employees) |
| HR | Employee | `manager_id` | Optional (top-level) |
| SCM | Shipment | `sales_order_id` | Optional (internal transfers) |
| M | ProductionOrder | `sales_order_id` | Optional (MTO vs MTS) |

## 3. Proposed CDD Syntax Extensions

### 3.1 Event Payload Definitions

```cdd
// Add inside service block, before producer_events
event_payload InvoiceEventPayload {
    invoice_id    uuid
    customer_id   uuid
    total_amount  decimal
    status        string
    created_at    timestamp
}

event_payload CustomerCreatedEvent {
    customer_id   uuid
    company_name  string
    email         string
    created_at    timestamp
}
```

### 3.2 Validation Constraints

```cdd
entity PerformanceReview {
    id             uuid      @primary
    employee_id    uuid      @reference(Employee.id)
    reviewer_id    uuid      @reference(Employee.id)
    review_date    date
    period_start   date
    period_end     date
    rating         integer   @min(1) @max(5)
    rating_scale   string    @enum(["1-5", "1-10"])
    feedback       string    @max_length(2000)
    status         string    @enum(["DRAFT", "SUBMITTED", "ACKNOWLEDGED"])
    created_at     timestamp
    updated_at     timestamp
}
```

### 3.3 Enum Type Definitions

```cdd
// Global enum definitions (shared across entities)
enum CustomerStatus {
    LEAD
    PROSPECT
    ACTIVE
    INACTIVE
}

enum AccountType {
    ASSET
    LIABILITY
    EQUITY
    REVENUE
    EXPENSE
}

// Usage in entity
entity Customer {
    id         uuid           @primary
    status     CustomerStatus @default(LEAD)
    // ...
}
```

### 3.4 Deprecation Markers

```cdd
entity Transaction {
    @deprecated("Use JournalEntry instead. Will be removed in v2.0.")
    id          uuid      @primary
    reference   string    @unique
    // ...
}
```

## 4. Definition of Done

- [x] **2.1 resolved**: All 17+ event payload types defined in CDD with field specifications
- [x] **2.2 resolved**: Validation constraint syntax implemented (`@required`, `@min`, `@max`, `@pattern`, `@enum`)
- [x] **2.3 resolved**: All 7+ raw string enum fields replaced with typed enum definitions
- [x] **2.4 resolved**: Legacy entities (Transaction/TransactionLine) marked deprecated with migration path
- [x] **2.5 resolved**: All optional fields annotated with `@optional`
- [x] **2.6 resolved**: Code generator updated to produce Go validation code from CDD constraints
- [x] **2.7 resolved**: All changes verified by `make test` passing

## 5. Priority-Ordered Execution Plan

### P0 — Critical (CDD Incomplete)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 1 | Define event payload structs for all 7 services | 2d | Events reference undefined types; code gen fails |
| 2 | Add `enum` syntax to CDD parser | 1d | Enables type-safe status/type fields |
| 3 | Migrate all raw string enums to typed enums | 1.5d | Compile-time protection for 7+ fields |

### P1 — Validation (Data Integrity)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 4 | Add `@min`, `@max`, `@positive` constraints | 1d | Prevents invalid numeric values |
| 5 | Add `@pattern` for email/phone fields | 0.5d | Format validation |
| 6 | Add `@enum` constraint syntax | 0.5d | Status field validation |
| 7 | Add `@min_length`, `@max_length` constraints | 0.5d | String bounds |
| 8 | Mark all optional fields with `@optional` | 0.5d | Clear required/optional distinction |

### P2 — Code Generation (Automation)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 9 | Update CDD parser to handle new syntax | 2d | Core infrastructure |
| 10 | Update Go code generator for validation | 1.5d | Auto-generate `validate()` functions |
| 11 | Update event struct generator | 1d | Generate typed event payloads |
| 12 | Update enum constant generator | 1d | Generate Go const blocks |

### P3 — Cleanup (Legacy)

| Step | Task | Est. Time | Rationale |
|------|------|-----------|-----------|
| 13 | Mark Transaction/TransactionLine as deprecated | 0.25d | Clear migration path |
| 14 | Add deprecation warnings to code generator | 0.25d | Alert developers |
| 15 | Create migration guide document | 0.5d | Help developers migrate |

**Total Estimate:** ~13 days

## 6. CDD Syntax Reference (Proposed)

### Constraints

| Constraint | Type | Description |
|------------|------|-------------|
| `@required` | Boolean | Field must be provided (default for non-optional) |
| `@optional` | Boolean | Field can be omitted |
| `@min(value)` | Numeric | Minimum value (inclusive) |
| `@max(value)` | Numeric | Maximum value (inclusive) |
| `@positive` | Boolean | Value must be > 0 |
| `@min_length(n)` | String | Minimum string length |
| `@max_length(n)` | String | Maximum string length |
| `@pattern(regex)` | String | Regex pattern match |
| `@enum([...])` | String | Allowed values list |
| `@default(value)` | Any | Default value if not provided |

### Enums

```cdd
enum StatusEnum {
    DRAFT
    ACTIVE
    INACTIVE
}
```

### Event Payloads

```cdd
event_payload EventName {
    field_name  type
    // ...
}
```

### Deprecation

```cdd
entity OldEntity {
    @deprecated("Use NewEntity instead. Will be removed in v2.0.")
    // ...
}
```

## 7. Migration Guide (Future)

For existing code consuming deprecated entities:

| Old Pattern | New Pattern |
|-------------|-------------|
| `Transaction` | `JournalEntry` |
| `TransactionLine` | `JournalEntryLine` |
| Raw string enums | Typed enum constants |
| Missing validation | CDD-generated `validate()` |

## 8. References

- Existing PRD: `2026-06-06-1557-cdd-gap-analysis.md`
- CDD Reference: `documentation/architecture/cdd-reference.md`
- CDD Files: `services/*/contracts/*.cdd`
