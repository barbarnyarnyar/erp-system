# PRD: FM Ledger Integrity & Transactional Outbox Alignment

**PRD ID**: PRD-2026-06-12-2255  
**Date**: 2026-06-12  
**Status**: Proposed (Draft)  
**Parent Initiative**: Financial System Hardening & Security Audit  
**Target Coverage**: 100% ACID compliance, 0% float64 currency drift, Outbox pattern enforcement  

---

## 1. Objective & Problem Statement

To support high-frequency, multi-tenant operations, the Financial Management (`fm-service`) microservice must guarantee absolute data consistency and eventual event propagation. A security and architecture audit has highlighted key vulnerabilities in the current implementation:
1. **Precision Vulnerability**: Go structures are susceptible to precision loss due to `float64` usage for financial balances.
2. **ACID Violation**: General Ledger operations simulate transactions in memory via custom rollback functions, creating risks of out-of-sync database balances in the event of microservice container crashes.
3. **Event Loss / Outbox Bypass**: Business logic publishes events synchronously to Kafka, creating tight coupling and potential data drift if brokers are temporarily unavailable.
4. **Constraint Gaps**: Missing composite and unique constraints in `fm.cdd` allow duplicate entries in the ledger database.

This PRD establishes the plan to refactor the FM data model, implement database-level transactions, enforce the Transactional Outbox pattern, and align schema constraints with the CDD contract.

---

## 2. Technical Requirements

### 2.1 Precision Defense & Dependency Pruning
- **Decimals only**: All Go structs representing monetary amounts (functional/transactional) must use `decimal.Decimal` (from `github.com/shopspring/decimal`).
- **Prune Heavy Imports**: Remove `k8s.io/apimachinery` references from data models and use standard library `"encoding/json"`.

### 2.2 Atomic Database Transaction Scope
- **ACID Transactions**: Refactor service operations (such as ledger journal entries, payments, and billings) to run inside standard SQL transactions.
- **Repository Support**: Update Repository interfaces to support a transactional context (passing `tx` or using a transactional database wrapper).
- **Rollback Invariants**: Discard manual in-memory snapshots and rollback closures, delegating all rollbacks to GORM's `tx.Rollback()`.

### 2.3 Transactional Outbox Pattern Integration
- **Outbox Insertion**: In the same transaction as the ledger updates, write event payloads into the `fm_transactional_outbox` table.
- **Outbox Relay**: Implement a background worker (or service loop) that periodically queries pending outbox entries, publishes them to Kafka, and marks them as sent or increments retry counts on failure.

### 2.4 Referential Integrity & Database Constraints
- **Chart of Accounts**: Update `fm.cdd` to declare `@unique_composite(legal_entity_id, account_code)` for `ChartOfAccounts`.
- **Customer Credits**: Update `fm.cdd` to declare `@unique` constraint on `customer_id` for `CustomerCredit`.
- **Foreign Key Actions**: Ensure GORM configuration matches Layer 3 SQL `ON DELETE RESTRICT` actions for ledger and journal entries.

---

## 3. Detailed Scope & Checklist

### Phase 1: CDD Contract updates & Model Regeneration
- [x] Add `@unique_composite(legal_entity_id, account_code)` to `ChartOfAccounts` inside [fm.cdd](file:///Users/sithuhlaing/Projects/erp-system/services/fm-service/contracts/fm.cdd).
- [x] Add `@unique` to `customer_id` in `CustomerCredit` in `fm.cdd`.
- [x] Run the CDD engine generator to update the domain models:
  ```bash
  go run cdd-engine/main.go -cdd services/fm-service/contracts/fm.cdd -go-out services/fm-service/internal/business/domain
  ```
- [x] Ensure `UniversalJournalLine` uses `decimal.Decimal` instead of `float64` in the generated model.

### Phase 2: Schema Migration & Database Constraint Verification
- [x] Generate SQL migrations from the updated `fm.cdd` using the CDD engine:
  ```bash
  go run cdd-engine/main.go -cdd services/fm-service/contracts/fm.cdd -sql-out services/fm-service/internal/data/migrations
  ```
- [x] Update the GORM database connection setup to run these schema migrations on startup.
- [x] Verify database contains the composite index `idx_entity_account` on `fm_chart_of_accounts` and a unique constraint on `fm_customer_credits(customer_id)`.

### Phase 3: Transactional Refactoring of Services
- [x] Refactor [general_ledger_service.go](file:///Users/sithuhlaing/Projects/erp-system/services/fm-service/internal/business/service/general_ledger_service.go) `CreateJournalEntry` to use atomic GORM transactions.
- [x] Refactor [accounts_receivable_service.go](file:///Users/sithuhlaing/Projects/erp-system/services/fm-service/internal/business/service/accounts_receivable_service.go) and [accounts_payable_service.go](file:///Users/sithuhlaing/Projects/erp-system/services/fm-service/internal/business/service/accounts_payable_service.go) to use atomic database transactions.
- [x] Verify that all manual rollback functions are eliminated.

### Phase 4: Outbox Table Implementation & Background Worker
- [x] Implement the `TransactionalOutboxRepository` interface.
- [x] Add logic in service methods to insert event records into the outbox within the database transaction.
- [x] Create an `OutboxRelayWorker` that runs as a background goroutine in `main.go`. It must periodically fetch unsent outbox messages, push them to the Kafka broker, and mark them as `SENT` on success.
- [x] Ensure that failures to publish to Kafka do not rollback the database updates, but trigger outbox retries.

### Phase 5: Verification & Unit Tests
- [x] Update all unit tests to test transactional behaviors (e.g. transaction rollbacks upon balance mismatch errors).
- [x] Verify that FM microservice builds successfully: `go build ./...`.
- [x] Verify that all FM tests pass: `go test ./...`.

---

## 4. Definition of Done
- [x] All database-level unique constraints and composite indexes are applied.
- [x] Manual rollback functions are completely replaced with database transaction blocks.
- [x] Synchronous Kafka publishing is replaced by outbox writes.
- [x] The outbox relay worker runs in a separate thread on startup.
- [x] All microservice components compile cleanly and unit tests pass (100% success rate).
