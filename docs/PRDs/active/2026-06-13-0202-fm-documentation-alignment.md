# PRD: FM Service Documentation Alignment & API Reference Update

**PRD ID**: PRD-2026-06-13-0202  
**Date**: 2026-06-13  
**Status**: Implemented  
**Parent Initiative**: Technical Documentation Sync & API Standard  
**Target Alignment**: 100% parity between implementation schemas, CDD contract definitions, and user-facing modules documentation  

---

## 1. Objective & Problem Statement

A detailed comparison between the Financial Management module's documentation and its actual codebase implementation reveals significant variations:
1. **Model & Schema Mismatches**: The API reference documents legacy payload structures (e.g. `account_number`, `parent_id`, `balance`, `currency` under Accounts) that do not match the GORM structures or CDD contract fields (`account_code`, `account_name`, `legal_entity_id`, etc.).
2. **Journal Entry Schema drift**: Journal entries are documented with `reference`, `description`, `debit_amount`, and `credit_amount`, but the actual handler binds `source_module`, `source_document_id`, `amount_functional`, `amount_transactional`, and `currency_transactional` lines.
3. **Flat Invoice schema**: Invoices are documented with a nested array of lines, whereas `fm.cdd` and the Go handler utilize a flat subledger invoice schema.
4. **Missing Endpoint Specifications**: The recently added `LegalEntity`, `CapitalAsset`, and `VendorBill` endpoints are completely missing from the API reference.
5. **Stub vs Real Reports**: Income statement and Cash flow reports are documented as stubs, but they are fully functional in the general ledger service implementation.

This PRD defines the scope to synchronize all documentation under `documentation/modules/financial-management/` with the current Go implementation and `fm.cdd` specification.

---

## 2. Alignment Matrix (Variations to Resolve)

| Documentation API Endpoint | Current Document Schema | Actual Code Binding Schema |
| :--- | :--- | :--- |
| **POST /accounts** | Uses `account_number`, `name`, `parent_id`, `currency`. | Binds `legal_entity_id`, `account_code`, `account_name`, `type`. |
| **POST /journal-entries** | Uses `reference`, `description`, and lines with `debit_amount`, `credit_amount`, `description`. | Binds `legal_entity_id`, `source_module`, `source_document_id`, `posting_date`, and lines with `account_id`, `amount_functional`, `amount_transactional`, `currency_transactional`. |
| **POST /invoices** | Uses nested `lines` array. | Binds flat `legal_entity_id`, `customer_id`, `sales_order_id`, `total_amount`, `tax_amount`, `due_date`. |
| **Vendor Bills** | Missing completely. | Implemented as `/api/v1/vendor-bills` and `/api/v1/vendor-bills/:id/lines`. |
| **Legal Entities** | Missing completely. | Implemented as `/api/v1/legal-entities` and `/api/v1/legal-entities/:id`. |
| **Assets & Depreciation** | Missing completely. | Implemented as `/api/v1/assets/capitalize`, `/api/v1/assets/:id/depreciation-schedule`, `/api/v1/assets/depreciate`. |

---

## 3. Scope & Checklist

### Phase 1: API Reference Sync
- [x] Update [api-reference.md](file:///Users/sithuhlaing/Projects/erp-system/documentation/modules/financial-management/api-reference.md) to correct the Account payload requests and responses (updating number/name fields, removing parent/currency/balance).
- [x] Update the Journal Entries section to match the transactional ledger line schemas (`amount_functional`, etc.).
- [x] Add request and response examples for **Vendor Bills** (`/api/v1/vendor-bills`).
- [x] Add request and response examples for **Legal Entities** (`/api/v1/legal-entities`).
- [x] Add request and response examples for **Assets & Depreciation** (`/api/v1/assets/capitalize`, `depreciate`).
- [x] Update report endpoint descriptions to reflect the real database aggregation output instead of stubs.

### Phase 2: Overview & Concept Update
- [x] Update [overview.md](file:///Users/sithuhlaing/Projects/erp-system/documentation/modules/financial-management/overview.md) and [general-ledger.md](file:///Users/sithuhlaing/Projects/erp-system/documentation/modules/financial-management/general-ledger.md) to documents:
  - The Transactional Outbox pattern background relay.
  - Event inbox deduplication checks.
  - Straight-line depreciation rules.

---

## 4. Definition of Done
- [x] Zero outdated properties (like `account_number` or `parent_id`) remain in the API Reference.
- [x] All 33 routes and subroutes in `routes.go` have matching API examples.
- [x] Documentation explains the new security mechanisms (outbox relay and inbox idempotency).
