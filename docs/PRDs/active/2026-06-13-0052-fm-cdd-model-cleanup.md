# PRD: FM CDD Naming Alignment & Legacy Model Cleanup

**PRD ID**: PRD-2026-06-13-0052  
**Date**: 2026-06-13  
**Status**: Approved (Implemented)  
**Parent Initiative**: Codebase Standardization & CDD Alignment  
**Target Coverage**: 100% CDD naming compliance, 0% duplicate or redundant entities  

---

## 1. Objective & Context

The Financial Management (`fm-service`) microservice was originally built using legacy entity names (`Account`, `Invoice`, `VendorBill`, `JournalEntry`, `JournalEntryLine`, `Transaction`, `TransactionLine`). 

When the `fm.cdd` contract-driven design specification was defined, it introduced standard enterprise names (`ChartOfAccounts`, `ArInvoice`, `ApVendorBill`, `UniversalJournalEntry`, `UniversalJournalLine`). However, because the active business service layer and HTTP handlers were never refactored to use the newly generated models, the codebase currently carries **two duplicate sets of structures**:
1. **Active Legacy Models**: `Account`, `Invoice`, `VendorBill`, `JournalEntry`, `JournalEntryLine` (used by services and handlers).
2. **Unused CDD-Generated Models**: `ChartOfAccounts`, `ArInvoice`, `ApVendorBill`, `UniversalJournalEntry`, `UniversalJournalLine` (generated from `fm.cdd` but unused).

To achieve strict contract alignment, eliminate dead code, and reduce database schema bloat, this PRD outlines the task of refactoring all active business logic to use the CDD-compliant generated models and removing the redundant legacy code files.

---

## 2. Refactoring Scope

### 2.1 Entity Mapping & Renaming Plan

The following table defines the target migration path:

| Legacy Entity (To Be Deleted) | CDD Entity (To Be Adopted) | Target Domain File | Notes |
| --- | --- | --- | --- |
| `Account` | `ChartOfAccounts` | `chart_of_accounts.go` | Rename `AccountNumber` to `AccountCode`, `Name` to `AccountName`. |
| `Invoice`, `InvoiceLine` | `ArInvoice`, `UniversalJournalLine` | `ar_invoice.go` | Subledger billing alignment. |
| `VendorBill`, `VendorBillLine` | `ApVendorBill`, `UniversalJournalLine` | `ap_vendor_bill.go` | Subledger accounts payable alignment. |
| `JournalEntry`, `JournalEntryLine` | `UniversalJournalEntry`, `UniversalJournalLine` | `universal_journal_entry.go` | Universal GL Ledger alignment. |
| `Transaction`, `TransactionLine` | `UniversalJournalEntry`, `UniversalJournalLine` | N/A | Consolidate and delete transaction files. |

### 2.2 Schema & Database Cleanup
* **Auto-Migration Update**: Remove legacy models from GORM's `db.AutoMigrate` statement in `db.go`.
* **GORM Struct Mappings**: Update `models.go` to remove mappings for `Account`, `Invoice`, `VendorBill`, `JournalEntry`, etc., keeping only the CDD-compliant structs.

### 2.3 Service & Handler Refactoring
* Update all references in `service/` and `handlers/` packages to use the new names and field mappings.
* Update unit and integration tests to align with the new model structures.

---

## 3. Scope & Checklist

### Phase 1: File Deletion & Clean-up
- [x] Delete legacy files under `services/fm-service/internal/business/domain/`:
  - `account.go`
  - `invoice.go`
  - `invoice_line.go`
  - `vendor_bill.go`
  - `vendor_bill_line.go`
  - `journal_entry.go`
  - `journal_entry_line.go`
  - `journal_entry_helpers.go`
  - `transaction.go`
  - `transaction_line.go`
  - `cost_center.go` (if not defined in CDD)
  - `fiscal_year.go` (if not defined in CDD)
- [x] Remove deleted legacy tables from `db.AutoMigrate` inside `db.go`.
- [x] Remove legacy structures and their mapping helpers from `models.go`.

### Phase 2: Service & API Layer Refactoring
- [x] Refactor `GeneralLedgerService` to use `ChartOfAccounts` and `UniversalJournalEntry` instead of `Account` and `JournalEntry`.
- [x] Refactor `AccountsReceivableService` to use `ArInvoice` instead of `Invoice`.
- [x] Refactor `AccountsPayableService` to use `ApVendorBill` instead of `VendorBill`.
- [x] Refactor `CashManagementService` to use `ArInvoice` / `ApVendorBill` in payments.
- [x] Refactor all Gin HTTP Handlers to bind requests and return responses mapping to the new model structures.

### Phase 3: Repository Implementation Consolidation
- [x] Rewrite repository interfaces in `repository.go` to use the CDD types:
  - `AccountRepository` becomes `ChartOfAccounts` operations.
  - Remove redundant `TransactionRepository`.
- [x] Update `sql_repos.go` and `memory_repos.go` to implement these updated repositories.

### Phase 4: Verification & Test Alignment
- [x] Update mock structures in `memory_repos.go` to use new entities.
- [x] Refactor `service_test.go`, `journal_entry_audit_test.go`, and `handlers_all_test.go` to verify the refactored code compiles and passes successfully.
- [x] Ensure `go build ./...` compiles cleanly and `go test -cover ./...` meets 80%+ test coverage.

---

## 4. Definition of Done
- [x] Zero references to legacy `Account`, `Invoice`, `VendorBill`, `JournalEntry` models in the microservice codebase.
- [x] Standardized CDD domain files are the sole representation of the data layer.
- [x] GORM database migrations run successfully with the refined model list.
- [x] Service package builds cleanly, and 100% of unit tests pass successfully.
