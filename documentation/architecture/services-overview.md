# Services Overview

Detailed documentation for all 6 ERP microservices, including API endpoints, domain models, event contracts, and architecture notes.

## Service Map

| Service | Gateway Route | Direct Port | Entry Point | Default Port |
|---------|--------------|-------------|-------------|-------------|
| Financial Management (FM) | `/api/v1/fm/*` | 8001 | `cmd/server/main.go` | 8001 |
| Human Resources (HR) | `/api/v1/hr/*` | 8002 | `cmd/main.go` | 8003 |
| Supply Chain Management (SCM) | `/api/v1/scm/*` | 8003 | `cmd/main.go` | 8006 |
| Manufacturing (M) | `/api/v1/m/*` | 8004 | `cmd/main.go` | 8004 |
| Customer Relationship Management (CRM) | `/api/v1/crm/*` | 8005 | `cmd/main.go` | 8002 |
| Project Management (PM) | `/api/v1/pm/*` | 8006 | `cmd/main.go` | 8006 |

> **Port Note**: Several services have code defaults that differ from the documented architecture ports (see each service section for details).

---

## Financial Management Service (fm-service)

**Purpose**: Complete financial management and accounting — general ledger, accounts receivable/payable, cash management, budgeting, and tax.

**Port**: `8001` (default) — consistent with architecture docs.

### API Endpoints

All under `/api/v1`. Total: **22 endpoints** (1 health + 21 API).

#### Accounts
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/accounts` | List all accounts |
| POST | `/api/v1/accounts` | Create an account |
| GET | `/api/v1/accounts/:id` | Get account details |
| PUT | `/api/v1/accounts/:id` | Update an account |
| DELETE | `/api/v1/accounts/:id` | Delete an account |
| GET | `/api/v1/accounts/:id/balance` | Get account balance |

#### Journal Entries
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/journal-entries` | List journal entries |
| POST | `/api/v1/journal-entries` | Create a journal entry (auto-posted) |
| GET | `/api/v1/journal-entries/:id` | Get journal entry details |
| PUT | `/api/v1/journal-entries/:id` | Update a journal entry |
| DELETE | `/api/v1/journal-entries/:id` | Delete a journal entry |

#### Invoices
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/invoices` | List invoices |
| POST | `/api/v1/invoices` | Create an invoice |
| GET | `/api/v1/invoices/:id` | Get invoice details |
| PUT | `/api/v1/invoices/:id` | Update an invoice |
| DELETE | `/api/v1/invoices/:id` | Delete an invoice |
| POST | `/api/v1/invoices/:id/send` | Mark invoice as sent |

#### Payments
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/payments` | List payments |
| POST | `/api/v1/payments` | Record a payment |
| GET | `/api/v1/payments/:id` | Get payment details |

#### Reports
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/reports/balance-sheet` | Get balance sheet |
| GET | `/api/v1/reports/income-statement` | Get income statement (stub) |
| GET | `/api/v1/reports/cash-flow` | Get cash flow report (stub) |

### Domain Models (17 entities)

Account, BankAccount, BankStatement, BankStatementLine, Budget, CostCenter, CurrencyRate, CustomerCredit, FiscalYear, Invoice, InvoiceLine, JournalEntry, JournalEntryLine, Payment, TaxRate, Transaction (with TransactionLine), VendorBill, VendorBillLine.

The `Transaction` domain has rich behavior: `Post()`, `Reverse()`, `IsBalanced()`, `AddLine()` — following a PENDING → POSTED → REVERSED state machine.

### Subdomain Services

| Service | Responsibilities | Events Published |
|---------|-----------------|-----------------|
| `GeneralLedgerService` | Accounts, journal entries, trial balance, balance sheet | `fin.account.created`, `fin.account.updated`, `fin.account.balance.changed` |
| `AccountsReceivableService` | Customer invoices, credit checks | `fin.invoice.created`, `fin.invoice.updated`, `fin.invoice.sent`, `fin.invoice.overdue` |
| `AccountsPayableService` | Vendor bills, PO matching | `fin.vendor.payment.due` |
| `CashManagementService` | Payments, bank reconciliation, cash forecasting | `fin.payment.received`, `fin.payment.processed`, `fin.payment.failed`, `fin.invoice.paid` |
| `BudgetingService` | Budgets, variance reports, expense tracking | `fin.budget.created`, `fin.budget.updated`, `fin.budget.exceeded` |
| `TaxService` | Tax rate CRUD — **NOT wired in main.go** | None |

### Events

**Produced (17 topics)**: `fin.invoice.*`, `fin.payment.*`, `fin.budget.*`, `fin.account.*`, `fin.vendor.payment.due`

**Consumed (13 topics)**: Processes events from HR (employee created, payroll, expenses), SCM (PO created, invoice received, inventory valued), CRM (sale completed, customer created), Manufacturing (production completed, material consumed), PM (project created, time logged, expense incurred) — creates corresponding journal entries and accounts automatically.

### Key Observations

- `TaxService` is **dead code** — defined but never wired.
- RabbitMQ config exists but **only Kafka is used**.
- Income statement and cash flow reports are **unimplemented stubs**.
- `MarkInvoiceOverdue` service method exists but has **no HTTP route**.
- Journal entries are always created as `POSTED` (no pending workflow).
- Has a `Makefile` and tests (`service_test.go`).

---

## Human Resources Service (hr-service)

**Purpose**: Employee lifecycle management — from recruitment and onboarding through payroll, leave, training, performance, and time tracking.

**Port**: `8003` (default in code) — architecture docs list HR as `8002`.

### API Endpoints

All under `/api/v1`. Total: **41 endpoints** (1 health + 40 API).

#### Employee Management
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/employees` | List all employees |
| POST | `/api/v1/employees` | Create an employee |
| GET | `/api/v1/employees/:id` | Get employee details |
| PUT | `/api/v1/employees/:id` | Update an employee |
| DELETE | `/api/v1/employees/:id` | Delete (terminate) an employee |
| POST | `/api/v1/employees/:id/expenses` | Submit expense claim |

#### Payroll
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/payroll` | List payroll records |
| POST | `/api/v1/payroll` | Process payroll |
| GET | `/api/v1/payroll/:id` | Get payroll record |
| PUT | `/api/v1/payroll/:id` | Update payroll record |
| GET | `/api/v1/payroll/employee/:id` | Get employee payroll history |

#### Time & Attendance
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/timesheet` | List timesheets |
| POST | `/api/v1/timesheet` | Create timesheet entry |
| GET | `/api/v1/timesheet/:id` | Get timesheet entry |
| PUT | `/api/v1/timesheet/:id` | Update timesheet entry |
| POST | `/api/v1/timesheet/:id/submit` | Submit timesheet |
| POST | `/api/v1/timesheet/:id/approve` | Approve timesheet |

#### Leave Management
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/leave-requests` | List leave requests |
| POST | `/api/v1/leave-requests` | Create leave request |
| GET | `/api/v1/leave-requests/:id` | Get leave request |
| PUT | `/api/v1/leave-requests/:id` | Update leave request |
| POST | `/api/v1/leave-requests/:id/approve` | Approve leave |
| POST | `/api/v1/leave-requests/:id/reject` | Reject leave |
| PUT | `/api/v1/leave-requests/:id/status` | Update leave status directly |

#### Recruitment
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/recruitment/jobs` | List job postings |
| POST | `/api/v1/recruitment/jobs` | Create job posting |
| GET | `/api/v1/recruitment/jobs/:id` | Get job posting |
| PUT | `/api/v1/recruitment/jobs/:id` | Update job posting |
| DELETE | `/api/v1/recruitment/jobs/:id` | Delete job posting |
| GET | `/api/v1/recruitment/applications` | List applications |
| POST | `/api/v1/recruitment/applications` | Create application |
| GET | `/api/v1/recruitment/applications/:id` | Get application |
| PUT | `/api/v1/recruitment/applications/:id` | Update application |

#### Performance
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/performance/reviews` | List reviews |
| POST | `/api/v1/performance/reviews` | Create review |
| GET | `/api/v1/performance/reviews/:id` | Get review |
| PUT | `/api/v1/performance/reviews/:id` | Update review |

#### Training
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/training/programs` | List programs |
| POST | `/api/v1/training/programs` | Create program |
| GET | `/api/v1/training/programs/:id` | Get program |
| PUT | `/api/v1/training/programs/:id` | Update program |
| POST | `/api/v1/training/programs/:id/enroll` | Enroll employee |
| POST | `/api/v1/training/enrollments/:enrollmentId/complete` | Complete training |

#### Documents
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/employees/:id/documents` | List employee documents |
| POST | `/api/v1/employees/:id/documents` | Upload document |
| DELETE | `/api/v1/employees/:id/documents/:docId` | Delete document |

#### Reports
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/reports/headcount` | Headcount by department |
| GET | `/api/v1/reports/payroll` | Payroll summary |
| GET | `/api/v1/reports/attendance` | Attendance metrics |

### Domain Models (16 entities)

Department, Position, Employee, PayrollRecord, PayrollDeduction, AttendanceEntry, LeaveRequest, LeaveBalance, JobPosting, JobApplication, PerformanceReview, TrainingProgram, TrainingEnrollment, EmployeeDocument, ExpenseClaim, ExpenseClaimLine.

### Subdomain Services (9 services)

| Service | Responsibilities | Events Published |
|---------|-----------------|-----------------|
| `EmployeeManagementService` | Employee CRUD, expense claims | `hr.employee.created/updated/terminated/promoted`, `hr.salary.changed`, `hr.expense.submitted` |
| `PayrollService` | Payroll processing with deductions | `hr.payroll.processed` |
| `TimeAttendanceService` | Timesheets, overtime detection | `hr.timesheet.submitted/approved`, `hr.overtime.recorded` |
| `LeaveManagementService` | Leave requests, balance validation | `hr.leave.requested/approved/rejected` |
| `RecruitmentService` | Job postings, applications | None |
| `PerformanceService` | Performance reviews | `hr.performance.review.completed`, `hr.performance.improvement.needed` |
| `TrainingService` | Training programs, enrollment | `hr.training.completed` |
| `EmployeeDocumentService` | Document management | None |
| `ReportService` | Headcount/payroll/attendance reports | None |

### Events

**Produced (22 topics)**: `hr.employee.*`, `hr.payroll.*`, `hr.timesheet.*`, `hr.overtime.*`, `hr.leave.*`, `hr.training.*`, `hr.performance.*`, `hr.expense.*`. Five topics are defined but never published (`hr.certification.earned`, `hr.skill.acquired`, `hr.goal.achieved`, `hr.employee.available`, `hr.employee.skills.updated`, `hr.payroll.failed`).

**Consumed (5 topics)**: `prj.project.created`, `prj.task.assigned`, `fin.budget.allocated`, `mfg.production.scheduled` — logged only; `scm.training.required` — auto-creates training program.

### Key Observations

- **Department and Position** repos/interfaces exist but are **never instantiated** in main.go.
- Recruitment and document services publish **no events**.
- Payroll hardcodes: 160h/month, 1.5x overtime, 15% income tax, 5% social security.
- Employee IDs are generated as `EMP-` + unix timestamp.
- Leave balance auto-initializes to 15 days for new leave types.
- No Makefile.

---

## Supply Chain Management Service (scm-service)

**Purpose**: Inventory and procurement management — products, suppliers, purchase orders, warehouse receipts/shipments, stock transfers, demand forecasting.

**Port**: `8006` (default in code) — architecture docs list SCM as `8003`.

### API Endpoints

All under `/api/v1`. Total: **50 endpoints** (1 health + 49 API).

#### Products
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/product-categories` | List categories |
| POST | `/api/v1/product-categories` | Create category |
| GET | `/api/v1/product-categories/:id` | Get category |
| PUT | `/api/v1/product-categories/:id` | Update category |
| DELETE | `/api/v1/product-categories/:id` | Delete category |
| GET | `/api/v1/products` | List products |
| POST | `/api/v1/products` | Create product |
| GET | `/api/v1/products/:id` | Get product |
| PUT | `/api/v1/products/:id` | Update product |
| DELETE | `/api/v1/products/:id` | Delete product |

#### Vendors & Contracts
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/vendors` | List vendors |
| POST | `/api/v1/vendors` | Create vendor |
| GET | `/api/v1/vendors/:id` | Get vendor |
| PUT | `/api/v1/vendors/:id` | Update vendor |
| DELETE | `/api/v1/vendors/:id` | Delete vendor |
| GET | `/api/v1/vendor-contracts` | List contracts |
| POST | `/api/v1/vendor-contracts` | Create contract |
| GET | `/api/v1/vendor-contracts/:id` | Get contract |
| PUT | `/api/v1/vendor-contracts/:id` | Update contract |
| DELETE | `/api/v1/vendor-contracts/:id` | Delete contract |

#### Purchase Requisitions & Orders
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/purchase-requisitions` | List requisitions |
| POST | `/api/v1/purchase-requisitions` | Create requisition |
| GET | `/api/v1/purchase-requisitions/:id` | Get requisition |
| PUT | `/api/v1/purchase-requisitions/:id` | Update requisition |
| DELETE | `/api/v1/purchase-requisitions/:id` | Delete requisition |
| POST | `/api/v1/purchase-requisitions/:id/approve` | Approve requisition |
| POST | `/api/v1/purchase-requisitions/:id/reject` | Reject requisition |
| GET | `/api/v1/purchase-orders` | List purchase orders |
| POST | `/api/v1/purchase-orders` | Create purchase order |
| GET | `/api/v1/purchase-orders/:id` | Get purchase order |
| PUT | `/api/v1/purchase-orders/:id` | Update purchase order |
| DELETE | `/api/v1/purchase-orders/:id` | Delete purchase order |
| POST | `/api/v1/purchase-orders/:id/send` | Send purchase order |

#### Inventory
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/inventory` | List inventory items |
| POST | `/api/v1/inventory` | Create inventory item |
| GET | `/api/v1/inventory/:id` | Get inventory item |
| PUT | `/api/v1/inventory/:id` | Update inventory item |
| DELETE | `/api/v1/inventory/:id` | Delete inventory item |
| POST | `/api/v1/inventory/reserve` | Reserve stock |
| POST | `/api/v1/inventory/release` | Release reservation |

#### Stock Transfers
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/stock-transfers` | List transfers |
| POST | `/api/v1/stock-transfers` | Create transfer |
| GET | `/api/v1/stock-transfers/:id` | Get transfer |
| POST | `/api/v1/stock-transfers/:id/execute` | Execute transfer |

#### Warehouse
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/receipts` | List receipts |
| POST | `/api/v1/receipts` | Create receipt |
| GET | `/api/v1/receipts/:id` | Get receipt |
| PUT | `/api/v1/receipts/:id` | Update receipt |
| GET | `/api/v1/shipments` | List shipments |
| POST | `/api/v1/shipments` | Create shipment |
| GET | `/api/v1/shipments/:id` | Get shipment |
| PUT | `/api/v1/shipments/:id` | Update shipment |

#### Forecasting
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/demand-forecasts` | List forecasts |
| POST | `/api/v1/demand-forecasts` | Create forecast |
| GET | `/api/v1/demand-forecasts/:id` | Get forecast |
| PUT | `/api/v1/demand-forecasts/:id` | Update forecast |

#### Reports
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/reports/inventory-levels` | Stock levels with valuation |
| GET | `/api/v1/reports/vendor-performance` | Supplier completion rates |
| GET | `/api/v1/reports/procurement-metrics` | Total spend, avg order |
| GET | `/api/v1/reports/safety-stock` | Safety stock recommendations |

### Domain Models (17 entities)

Product, ProductCategory, Location, Supplier, VendorContract, InventoryItem, InventoryMovement, StockTransfer, PurchaseOrder, PurchaseOrderLine, PurchaseRequisition, PurchaseRequisitionLine, Receipt, ReceiptLine, Shipment, ShipmentLine, DemandForecast.

### Subdomain Services (7 services)

| Service | Responsibilities | Events Published |
|---------|-----------------|-----------------|
| `ProductManagementService` | Products + categories CRUD | None |
| `SupplierManagementService` | Suppliers + contracts CRUD | None |
| `PurchaseOrderService` | POs + requisitions lifecycle | `scm.purchase.order.created` |
| `InventoryService` | Stock tracking, reservations, transfers, valuation | `scm.inventory.valued` |
| `WarehouseService` | Receipts, shipments, PO fulfillment | None |
| `DemandPlanningService` | Forecast CRUD | None |
| `ReportService` | 4 business reports | None |

### Events

**Produced (2 topics published)**: `scm.purchase.order.created`, `scm.inventory.valued`. Many additional topics are defined but not published.

**Consumed (7 topics)**: `crm.sales.order.created` (log), `crm.customer.demand.forecast` (auto-create forecast), `mfg.material.required` (auto-create purchase requisition), `mfg.material.consumed` (issue inventory), `mfg.production.completed` (receive finished goods), `fin.vendor.payment.processed` (log), `prj.material.requested` (issue inventory).

### Key Observations

- **Largest endpoint count** at 50 endpoints across 7 handler files.
- **Port mismatch**: defaults to 8006, documented as 8003.
- `WarehouseService` depends on `InventoryService` directly (cross-service composition).
- `DeleteInventoryItem` handler returns success without actually deleting.
- `CreateReceipt` has a bug where PO line update uses `Create` instead of `Update`.
- Reports use computed metrics (safety stock: `avgForecast * 1.25 + 10`).
- No Makefile.

---

## Manufacturing Service (m-service)

**Purpose**: Production planning and execution — BOMs, work centers, routing, work orders, quality control, equipment maintenance, cost variance analysis.

**Port**: `8004` (default) — consistent with architecture docs.

### API Endpoints

All under `/api/v1`. Total: **39 endpoints** (2 utility + 37 API).

#### Bill of Materials
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/boms` | List all BOMs |
| POST | `/api/v1/boms` | Create a BOM |
| GET | `/api/v1/boms/:id` | Get BOM by ID |
| PUT | `/api/v1/boms/:id` | Update a BOM |
| DELETE | `/api/v1/boms/:id` | Delete a BOM |

#### Work Centers
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/work-centers` | List work centers |
| POST | `/api/v1/work-centers` | Create work center |
| GET | `/api/v1/work-centers/:id` | Get work center |
| PUT | `/api/v1/work-centers/:id` | Update work center |
| DELETE | `/api/v1/work-centers/:id` | Delete work center |
| POST | `/api/v1/work-centers/:id/machine-log` | Log machine status |

#### Routing
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/routings` | List routings |
| POST | `/api/v1/routings` | Create routing |
| GET | `/api/v1/routings/:id` | Get routing |
| PUT | `/api/v1/routings/:id` | Update routing |
| DELETE | `/api/v1/routings/:id` | Delete routing |

#### Production Planning
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/production-plans` | List plans |
| POST | `/api/v1/production-plans` | Create production order |
| GET | `/api/v1/production-plans/:id` | Get plan |
| PUT | `/api/v1/production-plans/:id` | Update plan |
| POST | `/api/v1/mrp/run` | Run MRP |

#### Work Orders
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/work-orders` | List work orders |
| POST | `/api/v1/work-orders` | Create work order |
| GET | `/api/v1/work-orders/:id` | Get work order |
| PUT | `/api/v1/work-orders/:id` | Update work order |
| DELETE | `/api/v1/work-orders/:id` | Delete work order |
| POST | `/api/v1/work-orders/:id/start` | Start work order |
| POST | `/api/v1/work-orders/:id/complete` | Complete work order |
| POST | `/api/v1/work-orders/:id/labor` | Report labor |
| POST | `/api/v1/work-orders/:id/inspect` | Record inspection |

#### Quality
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/quality-inspections` | List inspections |
| POST | `/api/v1/quality-inspections` | Record inspection |
| GET | `/api/v1/quality-inspections/:id` | Get inspection |
| PUT | `/api/v1/quality-inspections/:id` | Update inspection |

#### Maintenance
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/maintenance-schedules` | List schedules |
| POST | `/api/v1/maintenance-schedules` | Schedule maintenance |
| GET | `/api/v1/maintenance-schedules/:id` | Get schedule |
| PUT | `/api/v1/maintenance-schedules/:id` | Update schedule |

### Domain Models (13 entities)

BillOfMaterials, BOMComponent, WorkCenter, RoutingOperation, ProductionOrder, WorkOrder, LaborReport, MachineLog, QualityInspection, NonConformance, Equipment, MaintenanceOrder, CostingRecord.

### Subdomain Services

| Service | Responsibilities | Events Published |
|---------|-----------------|-----------------|
| `BOMService` | BOMs, components, work centers, routings CRUD | None (publisher injected but unused) |
| `ProductionService` | Production orders, work orders, quality, maintenance, costing, MRP | All production/quality/maintenance events |
| `QualityService` | Standalone quality component — **NOT wired** | — |
| `CostingService` | Standalone costing component — **NOT wired** | — |

### Events

**Produced (18 topics)**: `mfg.production.scheduled/started/completed/delayed`, `mfg.work.order.created/started/completed/cancelled`, `mfg.material.consumed/wasted/required`, `mfg.quality.inspection.passed/failed`, `mfg.quality.non.conformance.detected`, `mfg.maintenance.scheduled/completed`, `mfg.equipment.down/up`.

**Consumed (6 topics)**: `crm.sales.order.created` (auto-schedule production), `scm.material.received` (log), `scm.inventory.updated` (log), `fin.cost.budget.allocated` (log), `hr.employee.scheduled` (log), `prj.custom.order.created` (auto-schedule production).

### Key Business Flows

**Make-to-Stock**: Create PO → MRP → auto-create work orders → start → labor → quality → complete → costing record.

**Make-to-Order**: Consumed `crm.sales.order.created` event triggers auto-scheduling via default BOM.

**Production Completion**: Auto-triggered when all work orders are completed + all inspections passed. Calculates standard vs actual labor/material variance.

### Key Observations

- `QualityService` and `CostingService` are defined as standalone components but **unused** — `ProductionService` handles all logic.
- Has a `Makefile`.
- Dockerfile exposes 8001 but service defaults to 8004.
- Schema SQL has a typo: `bill_of_materialss` (double `s`).

---

## Customer Relationship Management Service (crm-service)

**Purpose**: Customer lifecycle and sales management — leads, opportunities, quotes, sales orders, service tickets, campaigns, price lists.

**Port**: `8002` (default in code) — architecture docs list CRM as `8005`.

### API Endpoints

All under `/api/v1`. Total: **42 endpoints** (1 health + 41 API).

#### Customers
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/customers` | List customers |
| POST | `/api/v1/customers` | Create customer |
| GET | `/api/v1/customers/:id` | Get customer |
| PUT | `/api/v1/customers/:id` | Update customer |
| DELETE | `/api/v1/customers/:id` | Delete customer |

#### Leads
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/leads` | List leads |
| POST | `/api/v1/leads` | Create lead |
| GET | `/api/v1/leads/:id` | Get lead |
| PUT | `/api/v1/leads/:id` | Update lead |
| DELETE | `/api/v1/leads/:id` | Delete lead |
| POST | `/api/v1/leads/:id/convert` | Convert lead → customer + opportunity |

#### Opportunities
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/opportunities` | List opportunities |
| POST | `/api/v1/opportunities` | Create opportunity |
| GET | `/api/v1/opportunities/:id` | Get opportunity |
| PUT | `/api/v1/opportunities/:id` | Update opportunity |
| DELETE | `/api/v1/opportunities/:id` | Delete opportunity |

#### Sales Orders
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/sales-orders` | List orders |
| POST | `/api/v1/sales-orders` | Create order |
| GET | `/api/v1/sales-orders/:id` | Get order |
| PUT | `/api/v1/sales-orders/:id` | Update order |
| DELETE | `/api/v1/sales-orders/:id` | Delete order |

#### Quotes
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/quotes` | List quotes |
| POST | `/api/v1/quotes` | Create quote |
| GET | `/api/v1/quotes/:id` | Get quote |
| PUT | `/api/v1/quotes/:id` | Update quote |
| DELETE | `/api/v1/quotes/:id` | Delete quote |
| POST | `/api/v1/quotes/:id/send` | Send quote (publishes email event) |

#### Service Tickets
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/service-tickets` | List tickets |
| POST | `/api/v1/service-tickets` | Create ticket |
| GET | `/api/v1/service-tickets/:id` | Get ticket |
| PUT | `/api/v1/service-tickets/:id` | Update ticket |
| DELETE | `/api/v1/service-tickets/:id` | Delete ticket |

#### Campaigns
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/campaigns` | List campaigns |
| POST | `/api/v1/campaigns` | Create campaign |
| GET | `/api/v1/campaigns/:id` | Get campaign |
| PUT | `/api/v1/campaigns/:id` | Update campaign |
| DELETE | `/api/v1/campaigns/:id` | Delete campaign |

#### Price Lists
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/price-lists` | List price lists |
| POST | `/api/v1/price-lists` | Create price list |
| GET | `/api/v1/price-lists/:id` | Get price list |
| PUT | `/api/v1/price-lists/:id` | Update price list |
| DELETE | `/api/v1/price-lists/:id` | Delete price list |

### Domain Models (12 entities)

Customer, Lead, Opportunity, SalesOrder, SalesOrderItem, Quote, QuoteLineItem, PriceList, PriceListItem, ServiceTicket, Campaign.

### Subdomain Services (8 services)

| Service | Responsibilities | Events Published |
|---------|-----------------|-----------------|
| `CustomerService` | Customer CRUD | `crm.customer.created/updated/activated/deactivated` |
| `LeadService` | Lead CRUD + convert to customer/opportunity | `crm.lead.created/qualified/converted/lost` |
| `OpportunityService` | Opportunity pipeline | `crm.opportunity.created/updated/won/lost` |
| `SalesOrderService` | Sales order lifecycle | `crm.sales.order.created/updated/confirmed/shipped/delivered/cancelled` |
| `QuoteService` | Quotes + send | `crm.email.sent` (on send) |
| `ServiceTicketService` | Support tickets | `crm.service.ticket.created/updated/resolved/escalated` |
| `CampaignService` | Marketing campaigns | `crm.campaign.launched/completed` |
| `PriceListService` | Price list CRUD — **No event publisher** | None |

### Events

**Produced (21 topics)**: Full lifecycle events for customers, leads, opportunities, sales orders, service tickets, campaigns.

**Consumed (7 topics)**: `scm.inventory.available` (log), `scm.shipment.delivered` (update order to DELIVERED), `fin.payment.received` (log), `fin.credit.check.completed` (log), `mfg.production.completed` (log), `prj.project.completed` (log), `hr.employee.performance` (log).

### Key Observations

- **Lead conversion** is a cross-service orchestration: creates Customer + Opportunity simultaneously.
- Only `scm.shipment.delivered` has a real side-effect among consumed events — all others just log.
- `PriceListService` is the only service with **no event publishing**.
- Has graceful shutdown (SIGINT/SIGTERM with 5s timeout).
- Seeds mock data: Acme Corporation, Alice Smith/Bob Johnson leads, $45K opportunity.
- No Makefile, no tests.

---

## Project Management Service (pm-service)

**Purpose**: Project planning and resource management — portfolios, projects, tasks, resource allocation, time tracking, expenses, issues, change requests.

**Port**: `8006` (default) — consistent with architecture docs.

### API Endpoints

Under `/api/v1/projects/*` prefix. Total: **31 endpoints** (3 utility + 28 API).

#### Portfolios
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/portfolios` | List portfolios |
| POST | `/api/v1/projects/portfolios` | Create portfolio |
| GET | `/api/v1/projects/portfolios/:id` | Get portfolio |
| GET | `/api/v1/projects/portfolios/:id/summary` | Portfolio analytics |

#### Projects
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects` | List projects |
| POST | `/api/v1/projects` | Create project |
| GET | `/api/v1/projects/:id` | Get project |
| PUT | `/api/v1/projects/:id/status` | Update project status |

#### Tasks
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/tasks` | List tasks |
| POST | `/api/v1/projects/:id/tasks` | Create task |
| PUT | `/api/v1/projects/tasks/:task_id/progress` | Update progress |
| PUT | `/api/v1/projects/tasks/:task_id/assign` | Assign task |
| POST | `/api/v1/projects/tasks/:task_id/dependencies` | Add dependency |

#### Resource Allocations
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/allocations` | List allocations |
| POST | `/api/v1/projects/:id/allocations` | Allocate resource |

#### Time Tracking
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/time` | List time entries |
| POST | `/api/v1/projects/:id/time` | Log time |
| PUT | `/api/v1/projects/time/:time_id/approve` | Approve time |

#### Expenses
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/expenses` | List expenses |
| POST | `/api/v1/projects/:id/expenses` | Log expense |
| PUT | `/api/v1/projects/expenses/:expense_id/approve` | Approve expense |

#### Documents
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/documents` | List documents |
| POST | `/api/v1/projects/:id/documents` | Upload document |

#### Issues
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/issues` | List issues |
| POST | `/api/v1/projects/:id/issues` | Log issue |
| PUT | `/api/v1/projects/issues/:issue_id/resolve` | Resolve issue |

#### Change Requests
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/change-requests` | List change requests |
| POST | `/api/v1/projects/:id/change-requests` | Create change request |
| PUT | `/api/v1/projects/change-requests/:request_id/approve` | Approve change request |

#### Cross-Service Triggers
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/projects/:id/request-material` | Request material from SCM |
| POST | `/api/v1/projects/:id/request-custom-order` | Request custom order from Manufacturing |

### Domain Models (10 entities)

Portfolio, Project, Task, TaskDependency, ResourceAllocation, ProjectTimeEntry, ProjectExpense, ProjectDocument, ProjectIssue, ChangeRequest.

### Subdomain Services (6 services)

| Service | Responsibilities | Events Published |
|---------|-----------------|-----------------|
| `ProjectPlanningService` | Portfolios, projects, custom orders | `prj.project.created/started/completed/cancelled/updated`, `prj.custom.order.created` |
| `TaskManagementService` | Tasks, dependencies, material requests | `prj.task.created/assigned/started/completed`, `prj.material.requested` |
| `ResourceManagementService` | Resource allocations | `prj.resource.allocated` |
| `TimeExpenseService` | Time tracking, expenses | `prj.time.logged/approved`, `prj.expense.submitted/approved/incurred` |
| `CollaborationService` | Documents, issues, change requests | None |
| `PortfolioAnalyticsService` | Portfolio summaries | None (read-only) |

### Events

**Produced (22 topics)**: Project lifecycle, task lifecycle, resource allocation, time tracking, expense management, and cross-service triggers (`prj.material.requested`, `prj.custom.order.created`).

**Consumed (7 topics)**: `hr.employee.available` (log), `hr.employee.skills.updated` (log), `fin.budget.approved` (log), `fin.payment.received` (log), `crm.sales.order.received` (auto-create project + kickoff task), `scm.material.delivered` (log), `mfg.custom.production.completed` (log).

### Key Observations

- **Nested routing**: All endpoints under `/api/v1/projects/*`.
- Only **2 direct dependencies** (gin + decimal) — kafka-go is indirect.
- Rich seed data: portfolio, project, tasks with WBS hierarchy, time entries, expenses, documents, issues, change requests.
- `crm.sales.order.received` consumer auto-creates projects — the only consumed event with real side-effects.
- `PortfolioAnalyticsService` is read-only with no event publishing.
- Collaboration service (documents, issues, change requests) publishes **no events**.
- No Makefile, no tests.

---

## Common Patterns Across Services

### Architecture
All 6 services follow the same Clean Architecture pattern:

```
cmd/main.go          — Entry point / DI wiring
internal/
  api/handlers/      — HTTP handlers
  api/routes/        — Route definitions
  business/domain/   — Entities + repository interfaces + events
  business/service/  — Business logic
  config/            — Env-based configuration
  data/memory/       — In-memory repository implementations
  data/kafka/        — Kafka producer + consumer
  data/migrations/   — PostgreSQL schema (code-generated)
```

### Shared Characteristics
- **Gin** HTTP framework
- **In-memory storage** with `sync.RWMutex`-protected maps (no database at runtime)
- **Kafka** event publishing via `segmentio/kafka-go`
- **`shopspring/decimal`** for monetary/numeric values
- **Nanosecond-timestamp IDs** (`prefix_` + `time.Now().UnixNano()`) instead of UUIDs
- CDD-generated domain models and SQL schemas
- All events are **fire-and-forget** (errors silently discarded with `_ =`)
- No authentication/authorization middleware
- No Makefile in most services (except FM and M)

### Known Issues
1. **Port inconsistencies**: Multiple services have code defaults differing from documented architecture ports.
2. **Dead code**: `TaxService` (FM), `QualityService`/`CostingService` (M) are defined but never wired.
3. **Unused events**: Many event topics are defined in constants but never published.
4. **Stub endpoints**: Income statement and cash flow reports (FM) return placeholder messages.
5. **Bug**: SCM's `CreateReceipt` calls `Create` instead of `Update` on PO line repository.
6. **Unused utilities**: The `common-utils/` symlink provides shared logger/response helpers but no service uses them.
