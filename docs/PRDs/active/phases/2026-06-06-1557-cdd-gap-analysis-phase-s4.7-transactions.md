# ERP System CDD Gap Analysis — Phase S4.7: Transaction & Atomicity Enforcements (FM & CRM)

**Source PRD**: [cdd-gap-analysis.md](file:///Users/sithuhlaing/Projects/erp-system/docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md)
**PRD ID**: PRD-2026-06-06-1557
**Phase**: S4.7
**Status**: Completed
**Created**: June 06, 2026

---

## Objective

Ensure write-safety, data integrity, and atomic execution for General Ledger modifications and Lead Conversion processes. Specifically:
1. **FM (General Ledger)**: `UpdateJournalEntry` must validate that debits equal credits, and update GL account balances atomically with application-level rollback support on failure.
2. **CRM (Lead Conversion)**: `ConvertLead` must run atomically, rolling back lead status updates, customer creation, and opportunity creation if any subsequent step fails.

## Rationale

* **FM (2.9)**: Currently, `UpdateJournalEntry` updates the entry record directly without verifying if the debits equal credits for the new lines. More importantly, it does not update the General Ledger account balances at all when the journal entry lines change, leading to severe desynchronization between account balances and journal entries.
* **CRM (2.11)**: Lead conversion involves updating the lead, creating a customer, creating an opportunity, and publishing a Kafka event. If any step fails midway, the system is left in a corrupted or orphaned state (e.g., a customer created without a corresponding opportunity, or a converted lead without a successfully published event).

---

## Scope

### In Scope
1. **FM Service**:
   - Refactor `UpdateJournalEntry` to validate `lines` debit/credit equality.
   - Fetch the old journal entry lines.
   - Snapshot all affected GL accounts (both old and new accounts).
   - Reverse the balances from the old lines, apply the balances from the new lines, and update all affected account balances.
   - Save the updated journal entry and new lines.
   - If any step fails, restore the snapshots to perform an application-level rollback.
   - Write unit tests in `service_test.go` covering successful journal updates, unbalanced updates rejection, and rollback on repository errors.

2. **CRM Service**:
   - Refactor `ConvertLead` to use a rollback/transaction mechanism.
   - Snapshot the old lead state.
   - Track if Customer and Opportunity were created.
   - If any step fails (customer creation, opportunity creation, event publishing), delete the created customer/opportunity and restore the lead status back to its original state.
   - Write unit tests verifying that a failure at the opportunity creation level successfully rolls back the customer creation and lead status.

### Out of Scope
* Introducing real relational database SQL transactions (the storage layer uses in-memory maps; application-level rollbacks are the system standard here).
* Handling currency conversion during journal entry updates (assumed to be uniform).

---

## Implementation Tasks

### Task 1: Refactor `UpdateJournalEntry` (FM Service)
* **File**: `services/fm-service/internal/business/service/general_ledger_service.go`
* **Logic**:
  1. Retrieve existing journal entry and lines.
  2. Validate that the new `lines` have `debits == credits`.
  3. Prepare snapshots of accounts.
  4. Deduct/reverse the old line values from their respective accounts.
  5. Add/apply the new line values to their respective accounts.
  6. Save the new entry lines and update the entry record.
  7. Implement a `rollback()` callback to revert account states and entry status if any DB/repository update fails.

### Task 2: Refactor `ConvertLead` (CRM Service)
* **File**: `services/crm-service/internal/business/service/lead_service.go`
* **Logic**:
  1. Keep a backup of the original `Lead` object.
  2. Perform lead status update.
  3. Create customer via `CustomerService`.
  4. Create opportunity via `OpportunityService`.
  5. Publish lead converted event.
  6. If any step fails, run `rollback()` to delete the created customer (via `DeleteCustomer`) and opportunity, and revert the lead status in repository.

### Task 3: Unit Tests (FM Service)
* **File**: `services/fm-service/internal/business/service/service_test.go`
* **Add Tests**:
  - `TestGeneralLedgerService_UpdateJournalEntry_Success`: Verifies GL account balances are adjusted correctly after updating lines.
  - `TestGeneralLedgerService_UpdateJournalEntry_Unbalanced`: Verifies unbalanced line updates are rejected.
  - `TestGeneralLedgerService_UpdateJournalEntry_Rollback`: Verifies account balances are fully rolled back if the final save fails.

### Task 4: Unit Tests (CRM Service)
* **File**: `services/crm-service/internal/business/service/lead_transaction_test.go` (create new test file)
* **Add Tests**:
  - `TestConvertLead_Success`: Verifies successful lead conversion.
  - `TestConvertLead_Rollback`: Mocks a failure in opportunity creation and verifies customer deletion and lead status reversion.

---

## Verification

```bash
# Verify FM Service
cd services/fm-service
go test -v ./internal/business/service/...

# Verify CRM Service
cd services/crm-service
go test -v ./internal/business/service/...

# Ensure full compilation
cd ../..
make build
make run
make test
```

---

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
| --- | --- | --- |
| Concurrent modifications corrupting snapshot states | Medium | In-memory repositories utilize mutex locks (`sync.RWMutex`) during writes, preventing dirty reads/writes. |
| Event publish failure triggers full database rollback | Low | Aligning transactional state with events is required; rolling back when publish fails ensures downstream services don't desynchronize. |

---

## Definition of Done

- [x] `UpdateJournalEntry` rejects unbalanced entries.
- [x] `UpdateJournalEntry` correctly reverses old lines and applies new lines to GL accounts.
- [x] `UpdateJournalEntry` successfully rolls back account balances on save error.
- [x] `ConvertLead` rolls back lead status and deletes created customers if opportunity creation fails.
- [x] FM and CRM unit tests pass.
- [x] Full ERP system builds and compiles cleanly.
- [x] PRD gap analysis checklist items `2.9` and `2.11` are checked off.
