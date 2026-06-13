# PRD: CRM Service Package Boundary Splitting & CDD Alignment

**PRD ID**: PRD-2026-06-13-0920  
**Date**: 2026-06-13  
**Status**: Approved (Implemented)  
**Parent Initiative**: Technical Documentation Sync & Domain Parity  
**Target Coverage**: 100% CDD namespace compliance, zero architectural drift  

---

## 1. Objective & Architectural Context

To align the CRM Service architecture with enterprise standards, the system boundary is split into two distinct, decoupled domain packages under the contract definition:
1. **`erp.crm.core`**: High-throughput transactional core engine managing Customer Profiles, Price Books, Sales Orders, and Billing Triggers.
2. **`erp.crm.operations`**: Operational CRM surface managing Campaigns, Leads, Opportunities, Quotes, Service Tickets, and Customer Interactions.

We must reconcile the CDD contracts (`crm.cdd`) to cleanly specify these two packages and regenerate all domain models while maintaining absolute parity with the Go codebase implementation and ensuring Health Check and REST API endpoints remain 100% functional.

---

## 2. Technical Scope & Parity Matrix

### A. CDD Namespace Structure
* The single contract file `contracts/crm.cdd` will be reorganized to define both namespaces (`erp.crm.core` and `erp.crm.operations`).
* Keep the optional properties on `CustomerProfile` (`ContactName`, `Email`, `Phone`, `Category`, `ParentCustomerID`) that are actively consumed by handler and service logic to prevent runtime compilation breaks.

### B. Emitters & Consumers
* Consolidate and preserve all 32 producer events and 10 consumer events in the CDD event blocks to maintain constant mapping definitions in `event_topics.go`.

---

## 3. Scope & Implementation Checklist

### Phase 1: Reconcile CDD Contract
- [x] Overwrite `services/crm-service/contracts/crm.cdd` to include the dual namespace boundaries (`erp.crm.core` and `erp.crm.operations`) and merge active field properties.
- [x] Define the complete set of producer and consumer events matching `event_topics.go`.

### Phase 2: Domain Model Code Generation
- [x] Run the CDD generator on `crm.cdd` to rebuild domain model files.
- [x] Ensure that manual repository interfaces (`CustomerRepository`, `LeadRepository`, etc.) are cleanly decoupled or consolidated inside `domain/repository.go`.

### Phase 3: Code Compilation & Verification
- [x] Verify that `go build ./...` compiles cleanly without any errors.
- [x] Run the test suite using `go test ./...` to ensure all tests pass.

### Phase 4: Documentation Alignment
- [x] Update the CRM README and module documentation to reflect the new decoupled architecture.

---

## 4. Definition of Done
- [x] `crm.cdd` fully represents both `core` and `operations` namespace contracts.
- [x] Go models compile cleanly with zero modifications to hand-crafted handlers.
- [x] `go test ./...` passes successfully.
