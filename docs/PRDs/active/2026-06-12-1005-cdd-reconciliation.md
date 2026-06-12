# PRD: CDD Contract Reconciliation & Legacy Cleanup

**PRD ID**: PRD-2026-06-12-1005  
**Date**: 2026-06-12  
**Status**: Draft (Proposed)  
**Parent Initiative**: ERP Quality & Architecture Alignment  
**Target Coverage**: 100% CDD-to-Go struct alignment, 0 stale domain entities  

---

## 1. Objective & Problem Statement

Contract-Driven Development (CDD) serves as the single source of truth for the ERP system's architecture. However, a comprehensive audit of all 10 microservices reveals a **major structural drift** between the CDD contract files (`*.cdd`) and the actual Go implementations.

While the microservices compile and pass tests, they are doing so using **legacy models and interfaces** that were never cleaned up or migrated. Furthermore:
1. **Parser Limitations**: A parser bug previously choked on single-line enums (like those in `fm.cdd`), which blocked automated code generation for FM.
2. **Missing Core Entities**: Key database tables, outboxes, and event inboxes defined in the CDD do not exist in the codebase.
3. **Obsolete Assets**: Dozens of legacy entities (e.g., Campaign, Opportunity, Lead, Project Time Entry, Job Application) remain in the codebase, inflating complexity and violating architectural constraints.

We need a systematic reconciliation roadmap to clean up stale files, generate the correct Go structures, and refactor service business logic to match the active contracts.

---

## 2. Comprehensive Drift Audit Matrix

The following matrix documents the exact variances per service:

| Service | Active CDD Contract | Missing Entities (Defined in CDD, Missing in Go) | Stale/Legacy Entities (In Go, Missing in CDD) |
| :--- | :--- | :--- | :--- |
| **auth-service** | `auth.cdd` | `transactional_outbox.go`, `kafka_event_inbox.go` | None |
| **crm-service** | `crm.cdd` | `transactional_outbox.go`, `kafka_event_inbox.go`, `customer_profile.go`, `price_book_header.go`, `pricing_strategy.go`, `billing_trigger.go`, `price_book_entry.go`, `sales_order_line.go` | `opportunity.go`, `price_list_item.go`, `service_ticket.go`, `lead.go`, `opportunity_stage_history.go`, `price_list.go`, `quote.go`, `quote_line_item.go`, `sales_order_helpers.go`, `sales_order_item.go`, `campaign.go`, `customer.go`, `customer_interaction.go` |
| **eam-service** | `eam.cdd` | None (100% Synced) | None |
| **fm-service** | `fm.cdd` | `transactional_outbox.go`, `chart_of_accounts.go`, `universal_journal_entry.go`, `ap_vendor_bill.go`, `kafka_event_inbox.go`, `legal_entity.go`, `universal_journal_line.go`, `ar_invoice.go`, `capital_asset.go`, `depreciation_schedule_line.go` | `account.go`, `fiscal_year.go`, `journal_entry.go`, `transaction_line.go`, `vendor_bill.go`, `bank_statement.go`, `budget.go`, `invoice.go`, `invoice_line.go`, `journal_entry_helpers.go`, `journal_entry_line.go`, `vendor_bill_line.go`, `cost_center.go`, `currency_rate.go`, `customer_credit.go`, `transaction.go`, `bank_statement_line.go`, `payment.go`, `tax_rate.go` |
| **hr-service** | `hr.cdd` | `payroll_run.go`, `transactional_outbox.go`, `kafka_event_inbox.go`, `employee_master.go` | `employee.go`, `training_program.go`, `department_history.go`, `employee_compensation_history.go`, `leave_balance.go`, `leave_request.go`, `payroll_record.go`, `job_posting.go`, `performance_review.go`, `position.go`, `position_history.go`, `training_enrollment.go`, `attendance_entry.go`, `employee_document.go`, `job_application.go`, `payroll_deduction.go` |
| **mfg-service** | `mfg.cdd` | `transactional_outbox.go`, `kafka_event_inbox.go`, `routing_station.go`, `work_order_routing_state.go`, `material_consumption_log.go`, `production_yield_log.go` | `production_order.go`, `machine_log.go`, `quality_inspection.go`, `routing_operation.go`, `bill_of_materials.go`, `bomcomponent.go`, `costing_record.go`, `equipment.go`, `labor_report.go`, `maintenance_order.go`, `non_conformance.go` |
| **plm-service** | `plm.cdd` | None | `bom_component_payload.go` |
| **prj-service** | `prj.cdd` | `wbs_node.go`, `time_log.go`, `transactional_outbox.go`, `kafka_event_inbox.go` | `change_request.go`, `project_expense.go`, `project_issue.go`, `resource_allocation.go`, `task.go`, `task_dependency.go`, `milestone.go`, `portfolio.go`, `project_document.go`, `project_time_entry.go` |
| **qms-service** | `qms.cdd` | None | `metric_submission_input.go`, `quality_result_payload.go`, `time_range.go` |
| **scm-service** | `scm.cdd` | `transactional_outbox.go`, `kafka_event_inbox.go`, `warehouse.go`, `stock_balance.go`, `inventory_transaction.go` | `demand_forecast.go`, `inventory_movement.go`, `receipt_line.go`, `shipment.go`, `vendor_contract.go`, `inventory_item.go`, `product_category.go`, `purchase_requisition.go`, `receipt.go`, `shipment_line.go`, `stock_transfer.go`, `location.go`, `product.go`, `purchase_requisition_line.go` |

---

## 3. Scope of Work

### In Scope
1. **Legacy Entity Purge**: Delete all stale Go entity files listed in Section 2 from `internal/business/domain/`.
2. **Schema & Code Generation**: Run `cdd-cli` to generate correct domain structs (`*.go`) and SQL migrations (`schema.sql`) for missing entities in all 10 services.
3. **Service Refactoring**: Refactor business logic service interfaces and methods (e.g., in `services/crm-service/internal/business/service/`) to match the interfaces declared in the CDD.
4. **Repository & Controller Synchronization**: Update GORM SQL/Memory repositories and Gin HTTP handlers to bind and interact with the newly generated CDD structs.
5. **Gateway Routing Verification**: Update routing paths in the API Gateway to route to the correct endpoints matching the CDD interfaces.

### Out of Scope
* Modifying frontend layouts or adding business features that are not defined in the active CDD contracts.

---

## 4. Priority-Ordered Execution Plan

### Phase 1: Clean Slate & Generation (P0)
Establish a clean build environment by removing dead code and auto-generating structures.
* **Step 1**: Run a cleanup script to delete all stale Go files from domain directories.
* **Step 2**: Recompile the parser and run code generation across all services:
  ```bash
  ./scripts/generate-all.sh
  ```
* **Step 3**: Verify that all new CDD-aligned domain model files exist.

### Phase 2: Interface & Repository Refactor (P0)
Implement the core CRUD and domain logic methods.
* **Step 4**: Update repository interfaces in `services/{service}/internal/business/domain/repository.go` to support CDD operations.
* **Step 5**: Write memory database adapters and GORM PostgreSQL migration files for the new schemas.
* **Step 6**: Wire the new repositories in `cmd/main.go` for all services.

### Phase 3: Service Layer Reconciliation (P1)
Update the core business logic handlers to compile.
* **Step 7**: Refactor service files (e.g. `services/crm-service/internal/business/service/*.go`) to use new types and implement new interface methods.
* **Step 8**: Align HTTP controllers to bind to the new inputs/outputs.
* **Step 9**: Re-wire consumer background tasks to accept payload models derived from the CDD.

### Phase 4: Integration Smoke Checks & Gateway Sync (P2)
* **Step 10**: Re-align the API Gateway router middleware, paths, and health check routes to match the new endpoints.
* **Step 11**: Start the docker-compose stack and run `make test` verification tests.

---

## 5. Definition of Done
- [ ] Stale domain files are 100% deleted.
- [ ] Go models and database migrations are generated from the active contracts.
- [ ] Services, repository layers, and handlers compile cleanly using the CDD-aligned structs.
- [ ] Gateway routes match and forward to the active service paths.
- [ ] All unit and integration smoke tests pass.
