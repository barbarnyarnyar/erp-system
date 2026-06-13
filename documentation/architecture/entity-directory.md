# Master Entity & Schema Directory

This document serves as the master directory of all **91 domain entities** across the 7 core ERP microservices. It maps their structures, conceptual descriptions, and cross-service dependencies as defined in the CDD contracts (`*.cdd`).

---

## 1. Auth Service (Authentication & Authorization)
*Defines security credentials, RBAC (Role-Based Access Control) policies, and session tokens.*

| Entity Name | CDD Reference | Description | Relationships & Dependencies |
|-------------|---------------|-------------|------------------------------|
| **User** | `User` | Security principal credentials, username, and email. | Maps 1:1 with HR `Employee`. |
| **Session** | `Session` | Active user login session tracking for token invalidation. | References `User.id` (1:N). |
| **Role** | `Role` | Security role profiles (e.g., Admin, Accountant). | Linked to `Permission` and `User`. |
| **Permission** | `Permission` | Action-level authorizations (e.g., `invoice:create`). | Linked to `Role`. |
| **UserRole** | `UserRole` | Mapping of security roles assigned to users. | Joins `User.id` and `Role.id`. |
| **UserStore** | `UserStore` | Mapping of users to physical/logical stores (location scoping). | Joins `User.id` to an external `store_id` (SCM). |
| **RolePermission**| `RolePermission` | Mapping of permissions granted to security roles. | Joins `Role.id` and `Permission.id`. |

---

## 2. CRM Service (Customer Relationship Management)
*Manages customer profiles, price books, sales pipelines, billing triggers, marketing campaigns, and quoting.*

| CDD Namespace | Entity Name | CDD Reference | Description | Relationships & Dependencies |
|---------------|-------------|---------------|-------------|------------------------------|
| **`erp.crm.core`** | **CustomerProfile** | `CustomerProfile` | Master customer profile with company and contact details. | Linked to `Opportunity`, `Quote`, `SalesOrder`. |
| | **PriceBookHeader** | `PriceBookHeader` | Grouping of pricing rules (e.g. Standard vs. Regional). | Linked to price book entries and pricing strategies. |
| | **PriceBookEntry** | `PriceBookEntry` | Price entry overrides for products in a specific book. | References `PriceBookHeader.id` (1:N), SCM `Product.id`. |
| | **PricingStrategy** | `PricingStrategy` | Rule configuration modifier (markup, temporal, volume splits). | References `PriceBookHeader.id` (1:N). |
| | **SalesOrder** | `SalesOrder` | Confirmed customer sales transaction before billing. | References `CustomerProfile.id` (1:N), `PriceBookHeader.id`. |
| | **SalesOrderLine** | `SalesOrderLine` | Individual items/SKUs included in a sales order. | References `SalesOrder.id` (1:N), SCM `Product.id`. |
| | **BillingTrigger** | `BillingTrigger` | Monthly partitioned records staging billing outputs for AR. | References `SalesOrder.id` (1:N). |
| | **TransactionalOutbox** | `TransactionalOutbox` | Outbox message cache for atomic operations events dispatch. | Outbox pattern integration. |
| | **KafkaEventInbox** | `KafkaEventInbox` | Inbound messaging idempotency checking register. | Idempotent consumer protection. |
| **`erp.crm.operations`** | **Campaign** | `Campaign` | Marketing campaign tracking target audience and budget. | Linked to Lead attribution. |
| | **Lead** | `Lead` | Unqualified contact profiles with score and campaign source. | References `Campaign.id` (1:N, optional). |
| | **Opportunity** | `Opportunity` | Qualified deals in the pipeline with stages and expected values. | References `CustomerProfile.id` (1:N). |
| | **CustomerInteraction** | `CustomerInteraction` | Log of meetings, phone calls, and emails. | References `CustomerProfile.id` (1:N). |
| | **ServiceTicket** | `ServiceTicket` | Support/service request case log with priority. | References `CustomerProfile.id` (1:N). |
| | **Quote** | `Quote` | Customer pricing proposals with validity dates. | References `CustomerProfile.id` (1:N), `Opportunity.id`. |
| | **QuoteLineItem** | `QuoteLineItem` | Proposed items and volumes in a quote. | References `Quote.id` (1:N), SCM `Product.id`. |


---

## 3. Financial Management (FM) Service
*Manages double-entry general ledger, accounting periods, invoices, bills, and cash bank reconciliations.*

| Entity Name | CDD Reference | Description | Relationships & Dependencies |
|-------------|---------------|-------------|------------------------------|
| **CurrencyRate** | `CurrencyRate` | Exchange rates for multi-currency operations. | Core value object. |
| **FiscalYear** | `FiscalYear` | Defined accounting periods and status (Open/Closed). | Controls GL postings. |
| **CostCenter** | `CostCenter` | Departmental tracking of revenues/expenses (e.g., Sales). | References `Account.id`. |
| **TaxRate** | `TaxRate` | Tax codes (e.g., VAT, GST) and percentages. | Applied to `InvoiceLine` and `VendorBillLine`. |
| **BankAccount** | `BankAccount` | Cash general ledger bank account records. | Linked to `Payment` and `BankStatement`. |
| **CustomerCredit**| `CustomerCredit` | Credit limits and active credit utilization checks. | References CRM `Customer.id`. |
| **Account** | `Account` | Chart of accounts (Assets, Liabilities, Revenue, etc.). | References `CostCenter.id` (1:N, optional). |
| **Budget** | `Budget` | Budget limits allocated for cost centers per period. | References `Account.id` (1:N). |
| **JournalEntry** | `JournalEntry` | Header double-entry transaction record. | Parent of `JournalEntryLine`. |
| **JournalEntryLine**| `JournalEntryLine` | Debits/credits matching lines (must balance to 0). | References `JournalEntry.id` (1:N), `Account.id`. |
| **Transaction** | `Transaction` | Higher-level audit logs for financial events. | Parent of `TransactionLine`. |
| **TransactionLine**| `TransactionLine` | Individual line in a transaction tracking impact. | References `Transaction.id` (1:N), `Account.id`. |
| **Invoice** | `Invoice` | Customer receivable billing document. | References CRM `Customer.id`, CRM `SalesOrder.id`. |
| **InvoiceLine** | `InvoiceLine` | Lines specifying billed items/services. | References `Invoice.id` (1:N), SCM `Product.id`. |
| **VendorBill** | `VendorBill` | Supplier payable billing document. | References SCM `Supplier.id`, SCM `PurchaseOrder.id`. |
| **VendorBillLine** | `VendorBillLine` | Lines specifying items/services received. | References `VendorBill.id` (1:N), SCM `Product.id`. |
| **Payment** | `Payment` | Cash receipt/expenditure tracking. | References `BankAccount.id`, `Invoice.id`/`VendorBill.id`. |
| **BankStatement** | `BankStatement` | Header bank reconciliation records. | References `BankAccount.id` (1:N). |
| **BankStatementLine**| `BankStatementLine` | Lines inside bank statement for clearing. | References `BankStatement.id` (1:N). |

---

## 4. Human Resources (HR) Service
*Manages employee directories, job applications, leaves, payroll schedules, and skill sets.*

| Entity Name | CDD Reference | Description | Relationships & Dependencies |
|-------------|---------------|-------------|------------------------------|
| **Department** | `Department` | Business organizational divisions. | Parent of `Position` and `Employee`. |
| **Position** | `Position` | Job title template and salary scale bands. | References `Department.id` (1:N). |
| **Employee** | `Employee` | Master worker record, contact details, status. | References `Department.id`, `Position.id`. |
| **EmployeeCompensationHistory** | `EmployeeCompensationHistory` | Audit trails of salary adjustments. | References `Employee.id` (1:N). |
| **PayrollRecord** | `PayrollRecord` | Individual worker paychecks for a specific period. | References `Employee.id` (1:N). |
| **PayrollDeduction**| `PayrollDeduction` | Tax and health insurance deductions. | References `PayrollRecord.id` (1:N). |
| **AttendanceEntry**| `AttendanceEntry` | Clock-in/clock-out timecard log. | References `Employee.id`, PM `Project.id`/`Task.id`. |
| **LeaveRequest** | `LeaveRequest` | Time-off requests and approval status. | References `Employee.id` (1:N). |
| **LeaveBalance** | `LeaveBalance` | Accrued paid-time-off (PTO) days. | References `Employee.id` (1:N). |
| **JobPosting** | `JobPosting` | Open recruitment vacancy listings. | References `Department.id`, `Position.id`. |
| **JobApplication** | `JobApplication` | Candidate resumes and hiring status. | References `JobPosting.id` (1:N). |
| **PerformanceReview**| `PerformanceReview` | Periodic manager reviews and ratings. | References `Employee.id` (1:N). |
| **TrainingProgram**| `TrainingProgram` | Corporate skill-development listings. | Linked to `TrainingEnrollment`. |
| **TrainingEnrollment**| `TrainingEnrollment` | Employee registrations in training listings. | Joins `TrainingProgram.id` and `Employee.id`. |
| **EmployeeDocument**| `EmployeeDocument` | HR file attachments (passports, contracts). | References `Employee.id` (1:N). |
| **ExpenseClaim** | `ExpenseClaim` | Reimbursement headers for out-of-pocket spend. | References `Employee.id` (1:N). |
| **ExpenseClaimLine**| `ExpenseClaimLine` | Out-of-pocket line items. | References `ExpenseClaim.id` (1:N), PM `Task.id`. |

---

## 5. Manufacturing (M) Service
*Manages bill of materials, work centers, work orders, labor reports, machine logs, and quality checks.*

| Entity Name | CDD Reference | Description | Relationships & Dependencies |
|-------------|---------------|-------------|------------------------------|
| **BillOfMaterials** | `BillOfMaterials` | Assembly recipe for manufactured goods. | References SCM `Product.id`. |
| **BOMComponent** | `BOMComponent` | Raw material lines required for assembly. | References `BillOfMaterials.id` (1:N), SCM `Product.id`. |
| **WorkCenter** | `WorkCenter` | Workspace, machine, or production line cell. | Linked to `RoutingOperation`, `WorkOrder`, `Equipment`. |
| **RoutingOperation**| `RoutingOperation` | Sequence of operations (e.g. cutting, welding). | References `BillOfMaterials.id` (1:N), `WorkCenter.id`. |
| **ProductionOrder** | `ProductionOrder` | Assembly demand header for quantity of goods. | References `BillOfMaterials.id`, SCM `Product.id`, CRM `SalesOrder.id`. |
| **WorkOrder** | `WorkOrder` | Job card for a specific routing step. | References `ProductionOrder.id` (1:N), `WorkCenter.id`. |
| **LaborReport** | `LaborReport` | Operator hours logged on a job card. | References `WorkOrder.id` (1:N), HR `Employee.id`. |
| **MachineLog** | `MachineLog` | Operational and down-time log of work center. | References `WorkCenter.id` (1:N). |
| **QualityInspection**| `QualityInspection` | Pass/Fail audit on a finished job card. | References `WorkOrder.id` (1:N), HR `Employee.id` (Inspector). |
| **NonConformance** | `NonConformance` | Logs defect details and isolation notes. | References `QualityInspection.id` (1:1). |
| **Equipment** | `Equipment` | Asset tracking for shop floor tools/machines. | References `WorkCenter.id` (1:N). |
| **MaintenanceOrder**| `MaintenanceOrder` | Repair and scheduled service logs for tools. | References `Equipment.id` (1:N). |
| **CostingRecord** | `CostingRecord` | Summary of material, labor, and overhead costs. | References `ProductionOrder.id` (1:1). |

---

## 6. Project Management (PM) Service
*Manages project portfolios, WBS (Work Breakdown Structure) tasks, resource allocations, and timesheets.*

| Entity Name | CDD Reference | Description | Relationships & Dependencies |
|-------------|---------------|-------------|------------------------------|
| **Portfolio** | `Portfolio` | Collection of related projects. | References HR `Employee.id` (Manager). |
| **Project** | `Project` | Project profiles tracking start/end dates. | References `Portfolio.id` (1:N), FM `Budget.id` (optional). |
| **Task** | `Task` | WBS tasks tracking title, status, and progress. | References `Project.id` (1:N), HR `Employee.id` (Assignee). |
| **TaskDependency** | `TaskDependency` | Finish-to-Start or Start-to-Start dependencies. | References `Task.id` (joins parent and dependency). |
| **ResourceAllocation**| `ResourceAllocation` | Team assignments to projects. | References `Project.id` (1:N), HR `Employee.id`. |
| **ProjectTimeEntry**| `ProjectTimeEntry` | Timesheet log of hours spent on tasks. | References `Project.id` (1:N), `Task.id`, HR `Employee.id`. |
| **ProjectExpense** | `ProjectExpense` | Expenses incurred directly on projects. | References `Project.id` (1:N), `Task.id`, HR `Employee.id`. |
| **ProjectDocument** | `ProjectDocument` | Project files and blueprint attachments. | References `Project.id` (1:N), HR `Employee.id` (Uploader). |
| **ProjectIssue** | `ProjectIssue` | Bug, issue, or risk trackers on tasks. | References `Project.id` (1:N), HR `Employee.id` (Owner). |
| **ChangeRequest** | `ChangeRequest` | Scope modifications submitted for approval. | References `Project.id` (1:N), HR `Employee.id` (Requestor). |
| **Milestone** | `Milestone` | Key timeline deadlines and milestones. | References `Project.id` (1:N). |

---

## 7. Supply Chain Management (SCM) Service
*Manages warehouse inventory levels, stock transfers, purchasing, receipts, and shipments.*

| Entity Name | CDD Reference | Description | Relationships & Dependencies |
|-------------|---------------|-------------|------------------------------|
| **ProductCategory** | `ProductCategory` | Hierarchy categories for inventory categorization. | Parent of `Product`. |
| **Product** | `Product` | Master catalog SKU details. | References `ProductCategory.id` (1:N). |
| **Location** | `Location` | Physical warehouse, shelf, or bin locations. | Linked to `InventoryItem`, `StockTransfer`. |
| **Supplier** | `Supplier` | Supplier profiles supplying raw materials. | Parent of `PurchaseOrder` and `VendorContract`. |
| **VendorContract** | `VendorContract` | Contractual agreements (pricing/expiry) with suppliers. | References `Supplier.id` (1:N). |
| **InventoryItem** | `InventoryItem` | Active stock levels (Quantity On Hand and Reserved). | References `Product.id` (1:N), `Location.id`. |
| **InventoryMovement**| `InventoryMovement` | Logs increments or decrements (receipts vs issues). | References `Product.id` (1:N), `Location.id`. |
| **StockTransfer** | `StockTransfer` | Inter-warehouse inventory transit tracking. | References `Location.id` (from/to), `Product.id`. |
| **PurchaseRequisition**| `PurchaseRequisition` | Internal request from employee for purchase. | References HR `Employee.id` (Requestor). |
| **PurchaseRequisitionLine**| `PurchaseRequisitionLine` | Line items requesting specific quantities. | References `PurchaseRequisition.id` (1:N), `Product.id`. |
| **PurchaseOrder** | `PurchaseOrder` | Formal buying contract sent to supplier. | References `Supplier.id` (1:N). |
| **PurchaseOrderLine**| `PurchaseOrderLine` | Quantities, unit costs, and tax codes. | References `PurchaseOrder.id` (1:N), `Product.id`. |
| **Receipt** | `Receipt` | Goods receipt verification. | References `PurchaseOrder.id` (optional). |
| **ReceiptLine** | `ReceiptLine` | Verified quantities matching a receipt. | References `Receipt.id` (1:N), `Product.id`. |
| **Shipment** | `Shipment` | Delivery dispatch to customer order. | Parent of `ShipmentLine`. |
| **ShipmentLine** | `ShipmentLine` | Quantities picked and shipped. | References `Shipment.id` (1:N), `Product.id`. |
| **DemandForecast** | `DemandForecast` | Estimates of future demand for replenishment planning. | References `Product.id` (1:N). |

---

## 8. Cross-Service Decoupling Design Patterns

To prevent circular dependency and structural coupling, references crossing service boundaries (e.g. `pm-service` referencing HR `Employee.id`) must strictly observe these rules:

1. **Foreign Key Decoupling:** Databases are completely isolated. Microservices store external entity IDs as strings/UUIDs without physical database checks.
2. **Replication via Events:** When a service needs descriptive information from an external entity (e.g. SCM needing employee names for purchase requisitions), it:
   - Listens to lifecycle events (`hr.employee.created`, `hr.employee.updated`).
   - Caches the minimal required fields (`FirstName`, `LastName`, `Email`) locally in SCM tables.
3. **No Direct Joins:** Cross-service joins are resolved inside API Gateway handlers by making parallel calls to multiple microservices and stitching payloads dynamically.
