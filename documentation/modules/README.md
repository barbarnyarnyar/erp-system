# Business Modules

The ERP system provides six integrated business modules with REST APIs, in-memory storage, and Kafka event-driven cross-service integration.

> **Storage note**: All modules use **in-memory repositories** (`sync.RWMutex`-protected Go maps). SQL migration files exist for PostgreSQL but no database driver is imported anywhere. Data is lost on service restart.

> **Auth note**: The Auth Service runs on port 8000 but is **not wired** into the deployed API Gateway (`api-gateway/cmd/main.go`). A full JWT+RBAC system exists in `api-gateway/internal/server/server.go` but is not deployed. All endpoints are publicly accessible.

## [Auth Service](../architecture/security-architecture.md)

User identity, authentication, session management, role-based access control, and permission management. Port **8000**.

### Domain Models (7 types)

| Model | Key Fields |
|-------|-----------|
| `User` | ID, Username, Email, PasswordHash, FirstName, LastName, IsActive |
| `Session` | ID, UserID, RefreshToken, IPAddress, UserAgent, ExpiresAt |
| `Role` | ID, Name, Description |
| `Permission` | ID, Code, Description |
| `UserRole` | UserID, RoleID, AssignedBy |
| `UserStore` | UserID, StoreID |
| `RolePermission` | RoleID, PermissionID, AssignedBy |

### Business Services (3)

| Service | Key Methods |
|---------|-------------|
| `AuthService` | `authenticateUser` (returns JWT), `refreshToken`, `revokeToken`, `validateToken` |
| `UserService` | `createUser`, `updateUser`, `updateCredentials`, `deactivateUser`, `assignUserToStore`, `removeUserFromStore` |
| `RBACService` | `createRole`, `createPermission`, `assignPermissionToRole`, `validatePermissions` |

### Kafka Events Published (5 topics)

`auth.user.created`, `auth.user.deactivated`, `auth.user.role.assigned`, `auth.user.store.assigned`, `auth.password.changed`

No consumed events.

**Note**: Auth service runs on port 8000 but is NOT wired into the deployed API Gateway. The gateway reverse-proxies without authentication.

---

## [Financial Management](financial-management/)

General ledger, accounts receivable, accounts payable, cash management, budgeting, and financial reports. Port **8001**.

### Domain Models (17 types)

| Model | Key Fields | Event Triggers |
|-------|-----------|----------------|
| `Account` | ID, Code, Name, Type (Asset/Liability/Equity/Revenue/Expense), NormalSide (Debit/Credit), Level, ParentID, Balance, AllowPosting | `fin.account.created`, `fin.account.balance.changed` |
| `JournalEntry` | ID, Description, Lines[], Status (Draft/Posted/Reversed), SourceModule | `fin.transaction.created` |
| `JournalEntryLine` | AccountID, DebitAmount, CreditAmount, Description | — |
| `Invoice` | ID, CustomerID, Total, Status, DueDate, Lines[] | `fin.invoice.created`, `fin.invoice.paid` |
| `InvoiceLine` | ProductID, Quantity, UnitPrice, Amount | — |
| `Payment` | ID, InvoiceID, Amount, Method, Reference | `fin.payment.received`, `fin.payment.processed` |
| `VendorBill` | ID, VendorID, Amount, Status, Lines[] | — |
| `Budget` | ID, Name, FiscalYear, TotalAmount, SpentAmount, Status | `fin.budget.created`, `fin.budget.updated`, `fin.budget.exceeded`, `fin.budget.approved`, `fin.budget.allocated` |
| `FiscalYear` | ID, Name, StartDate, EndDate, IsClosed | — |
| `CostCenter` | ID, Code, Name, DepartmentID | — |
| `TaxRate` | ID, Name, Rate, Type | — |
| `CurrencyRate` | ID, FromCurrency, ToCurrency, Rate, EffectiveDate | — |
| `BankAccount`, `BankStatement`, `BankStatementLine` | — | — |
| `CustomerCredit` | CustomerID, CreditLimit, CurrentBalance | — |

### Business Services (6)

| Service | Key Methods | Business Logic |
|---------|-------------|---------------|
| `GeneralLedgerService` | `CreateAccount`, `GetAccountBalance`, `CreateJournalEntry`, `PostEntry`, `ReverseEntry`, `GetTrialBalance`, `GetBalanceSheet` | Double-entry balance validation (debits must equal credits), hierarchical account tree, GL posting, trial balance computation, balance sheet by account-type classification |
| `AccountsReceivableService` | `CreateInvoice`, `GetInvoice`, `SendInvoice` | Customer invoice lifecycle |
| `AccountsPayableService` | `CreateVendorBill`, `GetVendorBill` | Vendor bill management |
| `CashManagementService` | `RecordPayment`, `GetPayments` | Payment recording against invoices |
| `BudgetingService` | `CreateBudget`, `MonitorBudget`, `GetBudgetVariance` | Budget vs. actual variance reporting |
| `TaxService` | `CreateTaxRate`, `ListTaxRates`, `GetTaxRate` | Tax rate CRUD — **not wired in code** |

### API Endpoints (25 routes)

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/health` | inline | Health check |
| GET | `/api/v1/accounts` | `accHandler.GetAccounts` | List all accounts |
| POST | `/api/v1/accounts` | `accHandler.CreateAccount` | Create account (validates code uniqueness, level calc) |
| GET | `/api/v1/accounts/:id` | `accHandler.GetAccount` | Get account by ID |
| PUT | `/api/v1/accounts/:id` | `accHandler.UpdateAccount` | Update account properties |
| DELETE | `/api/v1/accounts/:id` | `accHandler.DeleteAccount` | Delete account |
| GET | `/api/v1/accounts/:id/balance` | `accHandler.GetAccountBalance` | Get account balance |
| GET | `/api/v1/journal-entries` | `txHandler.GetTransactions` | List journal entries |
| POST | `/api/v1/journal-entries` | `txHandler.CreateTransaction` | Create journal entry |
| GET | `/api/v1/journal-entries/:id` | `txHandler.GetTransaction` | Get journal entry |
| PUT | `/api/v1/journal-entries/:id` | `txHandler.UpdateTransaction` | Update journal entry |
| DELETE | `/api/v1/journal-entries/:id` | `txHandler.DeleteTransaction` | Delete journal entry |
| GET | `/api/v1/invoices` | `invHandler.GetInvoices` | List invoices |
| POST | `/api/v1/invoices` | `invHandler.CreateInvoice` | Create invoice |
| GET | `/api/v1/invoices/:id` | `invHandler.GetInvoice` | Get invoice |
| PUT | `/api/v1/invoices/:id` | `invHandler.UpdateInvoice` | Update invoice |
| DELETE | `/api/v1/invoices/:id` | `invHandler.DeleteInvoice` | Delete invoice |
| POST | `/api/v1/invoices/:id/send` | `invHandler.SendInvoice` | Send invoice to customer |
| GET | `/api/v1/payments` | `payHandler.GetPayments` | List payments |
| POST | `/api/v1/payments` | `payHandler.RecordPayment` | Record payment |
| GET | `/api/v1/payments/:id` | `payHandler.GetPayment` | Get payment |
| GET | `/api/v1/reports/balance-sheet` | `repHandler.GetBalanceSheet` | Balance sheet report (GL-driven) |
| GET | `/api/v1/reports/income-statement` | `repHandler.GetIncomeStatement` | Income statement |
| GET | `/api/v1/reports/cash-flow` | `repHandler.GetCashFlow` | Cash flow report |

### Kafka Events Published (16 topics, per CDD)

`fin.invoice.created`, `fin.invoice.updated`, `fin.invoice.sent`, `fin.invoice.paid`, `fin.invoice.overdue`, `fin.payment.received`, `fin.payment.processed`, `fin.payment.failed`, `fin.vendor.payment.due`, `fin.budget.created`, `fin.budget.updated`, `fin.budget.exceeded`, `fin.budget.approved`, `fin.account.created`, `fin.account.updated`, `fin.account.balance.changed`

### Kafka Events Consumed (13 topics, per CDD)

Consumed by `EventConsumer` in `internal/business/service/event_consumer.go`:
- **HR → FM**: `hr.employee.created` (track new employee), `hr.payroll.processed` (create salary journal entry), `hr.expense.submitted` (create expense journal entry)
- **SCM → FM**: `scm.purchase.order.created` (create inventory-in-transit entry), `scm.invoice.received` (create AP entry), `scm.inventory.valued` (update inventory GL balance)
- **CRM → FM**: `crm.sale.completed` (create revenue entry), `crm.customer.created` (track new customer)
- **MFG → FM**: `mfg.production.completed` (WIP→finished goods entry), `mfg.material.consumed` (raw material issue entry)
- **PM → FM**: `prj.project.created` (track new project), `prj.time.logged` (create unbilled receivable entry), `prj.expense.incurred` (capitalize project cost)

---

## [Human Resources](human-resources/)

Employee lifecycle, payroll, time tracking, leave management, recruitment, performance reviews, training, and employee documents. Port **8002** (docker-compose: 8003).

### Domain Models (18 types)

| Model | Key Fields |
|-------|-----------|
| `Employee` | ID, EmployeeNo, FirstName, LastName, Email, Phone, DepartmentID, PositionID, HireDate, Salary, Status |
| `Department` | ID, Name, Code, ManagerID, ParentDepartmentID |
| `Position` | ID, Title, JobGrade, MinSalary, MaxSalary |
| `PayrollRecord` | ID, EmployeeID, PeriodStart, PeriodEnd, GrossPay, NetPay, Deductions[], Status |
| `PayrollDeduction` | Type, Amount, Description |
| `AttendanceEntry` | ID, EmployeeID, Date, ClockIn, ClockOut, HoursWorked, Type |
| `LeaveRequest` | ID, EmployeeID, Type, StartDate, EndDate, Status, Reason |
| `LeaveBalance` | ID, EmployeeID, LeaveType, TotalEntitled, Used, Remaining |
| `ExpenseClaim` | ID, EmployeeID, TotalAmount, Status, Lines[] |
| `ExpenseClaimLine` | Date, Category, Amount, Description, ReceiptURL |
| `JobPosting` | ID, Title, DepartmentID, Description, Requirements, Status |
| `JobApplication` | ID, JobPostingID, ApplicantName, Email, Status, ResumeURL |
| `PerformanceReview` | ID, EmployeeID, ReviewerID, Period, Rating, Comments, Status |
| `TrainingProgram` | ID, Name, Description, Duration, MaxParticipants |
| `TrainingEnrollment` | ID, TrainingProgramID, EmployeeID, Status, CompletionDate |
| `EmployeeDocument` | ID, EmployeeID, Type, Name, FileURL |

### Business Services (9)

| Service | Key Responsibilities |
|---------|---------------------|
| `EmployeeManagementService` | CRUD employees, department/position assignment |
| `PayrollService` | Payroll processing, pay slip generation, deduction calculation |
| `TimeAttendanceService` | Clock in/out, timesheet management |
| `LeaveManagementService` | Leave requests, balance tracking, approval workflow |
| `RecruitmentService` | Job postings, applications, hiring pipeline |
| `PerformanceService` | Review creation, rating, feedback |
| `TrainingService` | Program management, enrollment, completion |
| `EmployeeDocumentService` | Document upload, categorization |
| `ReportService` | Headcount, payroll, attendance reports |

### API Endpoints (33 routes)

**Employee Management:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/employees` | List employees |
| POST | `/api/v1/employees` | Create employee |
| GET | `/api/v1/employees/:id` | Get employee |
| PUT | `/api/v1/employees/:id` | Update employee |
| DELETE | `/api/v1/employees/:id` | Delete employee |
| POST | `/api/v1/employees/:id/expenses` | Submit expense claim |

**Payroll:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/payroll` | List payroll records |
| POST | `/api/v1/payroll` | Create payroll record |
| GET | `/api/v1/payroll/:id` | Get payroll record |
| PUT | `/api/v1/payroll/:id` | Update payroll |
| GET | `/api/v1/payroll/employee/:id` | Get employee payroll |

**Time Tracking:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/timesheet` | List timesheet entries |
| POST | `/api/v1/timesheet` | Create entry |
| GET | `/api/v1/timesheet/:id` | Get entry |
| PUT | `/api/v1/timesheet/:id` | Update entry |
| POST | `/api/v1/timesheet/:id/submit` | Submit for approval |
| POST | `/api/v1/timesheet/:id/approve` | Approve timesheet |

**Leave Management:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/leave-requests` | List requests |
| POST | `/api/v1/leave-requests` | Create request |
| GET | `/api/v1/leave-requests/:id` | Get request |
| PUT | `/api/v1/leave-requests/:id` | Update request |
| PUT | `/api/v1/leave-requests/:id/status` | Update status |
| POST | `/api/v1/leave-requests/:id/approve` | Approve |
| POST | `/api/v1/leave-requests/:id/reject` | Reject |

**Recruitment:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/recruitment/jobs` | List job postings |
| POST | `/api/v1/recruitment/jobs` | Create posting |
| GET | `/api/v1/recruitment/jobs/:id` | Get posting |
| PUT | `/api/v1/recruitment/jobs/:id` | Update posting |
| DELETE | `/api/v1/recruitment/jobs/:id` | Delete posting |
| GET | `/api/v1/recruitment/applications` | List applications |
| POST | `/api/v1/recruitment/applications` | Create application |
| GET | `/api/v1/recruitment/applications/:id` | Get application |
| PUT | `/api/v1/recruitment/applications/:id` | Update application |

**Performance:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/performance/reviews` | List reviews |
| POST | `/api/v1/performance/reviews` | Create review |
| GET | `/api/v1/performance/reviews/:id` | Get review |
| PUT | `/api/v1/performance/reviews/:id` | Update review |

**Training:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/training/programs` | List programs |
| POST | `/api/v1/training/programs` | Create program |
| GET | `/api/v1/training/programs/:id` | Get program |
| PUT | `/api/v1/training/programs/:id` | Update program |
| POST | `/api/v1/training/programs/:id/enroll` | Enroll employee |
| POST | `/api/v1/training/enrollments/:id/complete` | Mark completion |

**Documents:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/employees/:id/documents` | List employee documents |
| POST | `/api/v1/employees/:id/documents` | Upload document |
| DELETE | `/api/v1/employees/:id/documents/:docId` | Delete document |

**Reports:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/reports/headcount` | Headcount report |
| GET | `/api/v1/reports/payroll` | Payroll summary report |
| GET | `/api/v1/reports/attendance` | Attendance report |

### Kafka Events Published (22 topics, per CDD)

**Employee:** `hr.employee.created`, `hr.employee.updated`, `hr.employee.terminated`, `hr.employee.promoted`, `hr.employee.available`
**Payroll:** `hr.payroll.processed`, `hr.payroll.failed`, `hr.salary.changed`
**Time:** `hr.timesheet.submitted`, `hr.timesheet.approved`, `hr.overtime.recorded`
**Leave:** `hr.leave.requested`, `hr.leave.approved`, `hr.leave.rejected`
**Training:** `hr.training.completed`, `hr.certification.earned`, `hr.skill.acquired`
**Performance:** `hr.performance.review.completed`, `hr.goal.achieved`, `hr.performance.improvement.needed`
**Other:** `hr.expense.submitted`, `hr.employee.scheduled`, `hr.employee.skills.updated`

### Kafka Events Consumed (5 topics, per CDD)

| Topic | Publisher | Logic |
|-------|-----------|-------|
| `prj.project.created` | PM | Logged only |
| `prj.task.assigned` | PM | Logged only |
| `fin.budget.allocated` | FM | Adjust hiring plans based on budget |
| `mfg.production.scheduled` | MFG | Logged only |
| `scm.training.required` | SCM | Auto-create training program |

---

## [Supply Chain Management](supply-chain-management/)

Product catalog, inventory, procurement, warehouse operations, vendor management, and demand forecasting. Port **8003** (docker-compose: 8006).

### Domain Models (18 types)

| Model | Key Fields |
|-------|-----------|
| `Product` | ID, SKU, Name, Description, CategoryID, UnitPrice, UnitCost, ReorderPoint |
| `ProductCategory` | ID, Name, Description, ParentCategoryID |
| `Supplier` | ID, Code, Name, ContactPerson, Email, PaymentTerms, Status |
| `VendorContract` | ID, SupplierID, StartDate, EndDate, Terms, DiscountRate |
| `PurchaseRequisition` | ID, RequesterID, DepartmentID, Status, Lines[], TotalAmount |
| `PurchaseRequisitionLine` | ProductID, Quantity, EstimatedUnitPrice |
| `PurchaseOrder` | ID, SupplierID, OrderDate, ExpectedDelivery, Status, Lines[] |
| `PurchaseOrderLine` | ProductID, Quantity, UnitPrice, ReceivedQuantity |
| `InventoryItem` | ID, ProductID, LocationID, QuantityOnHand, QuantityReserved, ReorderPoint |
| `InventoryMovement` | ID, ProductID, LocationID, MovementType, Quantity, ReferenceID |
| `StockTransfer` | ID, FromLocationID, ToLocationID, Status, Lines[] |
| `Location` | ID, Name, Code, Type, Address |
| `Receipt` | ID, PurchaseOrderID, ReceivedDate, Lines[], Status |
| `ReceiptLine` | ProductID, QuantityReceived, QuantityAccepted, Notes |
| `Shipment` | ID, CustomerID, ShippedDate, Status, Lines[] |
| `ShipmentLine` | ProductID, QuantityShipped |
| `DemandForecast` | ID, ProductID, PeriodStart, PeriodEnd, ForecastQuantity, ActualQuantity |

### Business Services (7)

| Service | Key Responsibilities |
|---------|---------------------|
| `ProductManagementService` | Product CRUD, category hierarchy |
| `SupplierManagementService` | Supplier CRUD, contract management |
| `PurchaseOrderService` | Requisition→PO lifecycle, approval, send-to-supplier |
| `InventoryService` | Stock tracking, reservations, adjustments, movement recording |
| `WarehouseService` | Receipt processing, shipment processing, stock transfers |
| `DemandPlanningService` | Forecast CRUD, demand data consumption |
| `ReportService` | Inventory levels, vendor performance, procurement metrics, safety stock |

### API Endpoints (49 routes)

**Product Management:**
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

**Supplier Management:**
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

**Procurement:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/purchase-requisitions` | List requisitions |
| POST | `/api/v1/purchase-requisitions` | Create requisition |
| GET | `/api/v1/purchase-requisitions/:id` | Get requisition |
| PUT | `/api/v1/purchase-requisitions/:id` | Update |
| DELETE | `/api/v1/purchase-requisitions/:id` | Delete |
| POST | `/api/v1/purchase-requisitions/:id/approve` | Approve |
| POST | `/api/v1/purchase-requisitions/:id/reject` | Reject |
| GET | `/api/v1/purchase-orders` | List purchase orders |
| POST | `/api/v1/purchase-orders` | Create PO |
| GET | `/api/v1/purchase-orders/:id` | Get PO |
| PUT | `/api/v1/purchase-orders/:id` | Update PO |
| DELETE | `/api/v1/purchase-orders/:id` | Delete PO |
| POST | `/api/v1/purchase-orders/:id/send` | Send PO to supplier |

**Inventory:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/inventory` | List inventory items |
| POST | `/api/v1/inventory` | Create item |
| GET | `/api/v1/inventory/:id` | Get item |
| PUT | `/api/v1/inventory/:id` | Update item |
| DELETE | `/api/v1/inventory/:id` | Delete item |
| POST | `/api/v1/inventory/:id/reserve` | Reserve stock |
| POST | `/api/v1/inventory/:id/release` | Release reservation |

**Stock Transfers:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/stock-transfers` | List transfers |
| POST | `/api/v1/stock-transfers` | Create transfer |
| GET | `/api/v1/stock-transfers/:id` | Get transfer |
| PUT | `/api/v1/stock-transfers/:id` | Update |
| DELETE | `/api/v1/stock-transfers/:id` | Delete |
| POST | `/api/v1/stock-transfers/:id/execute` | Execute transfer |

**Receiving & Shipping:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/receipts` | List receipts |
| POST | `/api/v1/receipts` | Create receipt |
| GET | `/api/v1/receipts/:id` | Get receipt |
| PUT | `/api/v1/receipts/:id` | Update |
| DELETE | `/api/v1/receipts/:id` | Delete |
| GET | `/api/v1/shipments` | List shipments |
| POST | `/api/v1/shipments` | Create shipment |
| GET | `/api/v1/shipments/:id` | Get shipment |
| PUT | `/api/v1/shipments/:id` | Update |
| DELETE | `/api/v1/shipments/:id` | Delete |

**Demand Planning:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/demand-forecasts` | List forecasts |
| POST | `/api/v1/demand-forecasts` | Create forecast |
| GET | `/api/v1/demand-forecasts/:id` | Get forecast |
| PUT | `/api/v1/demand-forecasts/:id` | Update |
| DELETE | `/api/v1/demand-forecasts/:id` | Delete |

**Reports:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/reports/inventory-levels` | Inventory level report |
| GET | `/api/v1/reports/vendor-performance` | Vendor performance report |
| GET | `/api/v1/reports/procurement-metrics` | Procurement metrics report |
| GET | `/api/v1/reports/safety-stock` | Safety stock calculations |

### Kafka Events Published (22 topics)

**Inventory:** `scm.inventory.received`, `scm.inventory.shipped`, `scm.inventory.adjusted`, `scm.inventory.low.stock`, `scm.inventory.out.of.stock`, `scm.inventory.valued`, `scm.inventory.updated`
**Purchase Orders:** `scm.purchase.order.created`, `scm.purchase.order.sent`, `scm.purchase.order.received`, `scm.purchase.order.cancelled`
**Vendors:** `scm.vendor.created`, `scm.vendor.updated`, `scm.vendor.performance.evaluated`
**Shipments:** `scm.shipment.created`, `scm.shipment.dispatched`, `scm.shipment.delivered`, `scm.shipment.delayed`
**Other:** `scm.training.required`, `scm.material.delivered`, `scm.material.received`, `scm.invoice.received`

### Kafka Events Consumed (8 topics, per CDD)

| Topic | Publisher | Logic |
|-------|-----------|-------|
| `crm.sales.order.created` | CRM | Logged only |
| `crm.customer.demand.forecast` | CRM | Create demand forecast record |
| `mfg.material.required` | MFG | Auto-create purchase requisition |
| `mfg.material.consumed` | MFG | Issue raw material from inventory |
| `mfg.production.completed` | MFG | Receive finished goods into inventory |
| `fin.vendor.payment.processed` | FM | Logged only |
| `prj.material.requested` | PM | Issue material from inventory |

---

## [Customer Relationship Management](customer-relationship-management/)

Customer accounts, lead management, opportunity pipeline, sales orders, quotes, service tickets, campaigns, and price lists. Port **8004** (docker-compose: 8002).

### Domain Models (11 types)

| Model | Key Fields |
|-------|-----------|
| `Customer` | ID, Name, Email, Phone, Status, CreditLimit, BillingAddress, ShippingAddress |
| `Lead` | ID, FirstName, LastName, Email, Phone, Source, Status, CompanyName |
| `Opportunity` | ID, CustomerID, Title, Amount, Stage, Probability, ExpectedCloseDate |
| `SalesOrder` | ID, CustomerID, OrderDate, TotalAmount, Status, Items[] |
| `SalesOrderItem` | ProductID, Quantity, UnitPrice, Discount |
| `Quote` | ID, CustomerID, ValidUntil, TotalAmount, Status, Items[] |
| `QuoteLineItem` | ProductID, Quantity, UnitPrice, Discount |
| `ServiceTicket` | ID, CustomerID, Subject, Description, Priority, Status, AssignedTo |
| `Campaign` | ID, Name, Type, StartDate, EndDate, Budget, Status |
| `PriceList` | ID, Name, Currency, Items[] |
| `PriceListItem` | ProductID, UnitPrice, MinQuantity |

### Business Services (8)

| Service | Key Responsibilities |
|---------|---------------------|
| `CustomerService` | Customer CRUD, credit limit management |
| `LeadService` | Lead capture, qualification, conversion to opportunity |
| `OpportunityService` | Pipeline stages, probability tracking, forecasting |
| `SalesOrderService` | Order creation from quotes, order lifecycle |
| `QuoteService` | Quote generation, pricing, send-to-customer |
| `ServiceTicketService` | Ticket creation, assignment, resolution |
| `CampaignService` | Campaign planning, budget tracking |
| `PriceListService` | Price list CRUD, product pricing |

### API Endpoints (35 routes)

**Customers:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/customers` | List customers |
| POST | `/api/v1/customers` | Create customer |
| GET | `/api/v1/customers/:id` | Get customer |
| PUT | `/api/v1/customers/:id` | Update customer |
| DELETE | `/api/v1/customers/:id` | Delete customer |

**Leads:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/leads` | List leads |
| POST | `/api/v1/leads` | Create lead |
| GET | `/api/v1/leads/:id` | Get lead |
| PUT | `/api/v1/leads/:id` | Update lead |
| DELETE | `/api/v1/leads/:id` | Delete lead |
| POST | `/api/v1/leads/:id/convert` | Convert lead to customer+opportunity |

**Opportunities:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/opportunities` | List opportunities |
| POST | `/api/v1/opportunities` | Create opportunity |
| GET | `/api/v1/opportunities/:id` | Get opportunity |
| PUT | `/api/v1/opportunities/:id` | Update opportunity |
| DELETE | `/api/v1/opportunities/:id` | Delete opportunity |

**Sales Orders:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/sales-orders` | List orders |
| POST | `/api/v1/sales-orders` | Create order |
| GET | `/api/v1/sales-orders/:id` | Get order |
| PUT | `/api/v1/sales-orders/:id` | Update order |
| DELETE | `/api/v1/sales-orders/:id` | Delete order |

**Quotes:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/quotes` | List quotes |
| POST | `/api/v1/quotes` | Create quote |
| GET | `/api/v1/quotes/:id` | Get quote |
| PUT | `/api/v1/quotes/:id` | Update quote |
| DELETE | `/api/v1/quotes/:id` | Delete quote |
| POST | `/api/v1/quotes/:id/send` | Send quote to customer |

**Service Tickets:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/service-tickets` | List tickets |
| POST | `/api/v1/service-tickets` | Create ticket |
| GET | `/api/v1/service-tickets/:id` | Get ticket |
| PUT | `/api/v1/service-tickets/:id` | Update ticket |
| DELETE | `/api/v1/service-tickets/:id` | Delete ticket |

**Campaigns:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/campaigns` | List campaigns |
| POST | `/api/v1/campaigns` | Create campaign |
| GET | `/api/v1/campaigns/:id` | Get campaign |
| PUT | `/api/v1/campaigns/:id` | Update campaign |
| DELETE | `/api/v1/campaigns/:id` | Delete campaign |

**Price Lists:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/price-lists` | List price lists |
| POST | `/api/v1/price-lists` | Create price list |
| GET | `/api/v1/price-lists/:id` | Get price list |
| PUT | `/api/v1/price-lists/:id` | Update price list |
| DELETE | `/api/v1/price-lists/:id` | Delete price list |

### Kafka Events Published (28 topics, per CDD)

**Customer:** `crm.customer.created`, `crm.customer.updated`, `crm.customer.activated`, `crm.customer.deactivated`
**Lead:** `crm.lead.created`, `crm.lead.qualified`, `crm.lead.converted`, `crm.lead.lost`
**Opportunity:** `crm.opportunity.created`, `crm.opportunity.updated`, `crm.opportunity.won`, `crm.opportunity.lost`
**Sales Orders:** `crm.sales.order.created`, `crm.sales.order.updated`, `crm.sales.order.confirmed`, `crm.sales.order.cancelled`, `crm.sales.order.shipped`, `crm.sales.order.delivered`, `crm.sales.order.received`
**Service Tickets:** `crm.service.ticket.created`, `crm.service.ticket.updated`, `crm.service.ticket.resolved`, `crm.service.ticket.escalated`
**Campaigns:** `crm.campaign.launched`, `crm.campaign.completed`
**Email:** `crm.email.sent`, `crm.email.opened`, `crm.email.clicked`

### Kafka Events Consumed (7 topics, per CDD)

| Topic | Publisher | Logic |
|-------|-----------|-------|
| `scm.inventory.available` | SCM | Logged only |
| `scm.shipment.delivered` | SCM | Update sales order to DELIVERED |
| `fin.payment.received` | FM | Logged only |
| `fin.credit.check.completed` | FM | Logged only |
| `mfg.production.completed` | MFG | Logged only |
| `prj.project.completed` | PM | Logged only |
| `hr.employee.performance` | HR | Logged only |

---

## [Manufacturing](manufacturing/)

Bill of materials, routings, work orders, production planning, MRP, quality control, work centers, and equipment maintenance. Port **8005** (docker-compose: 8004).

### Domain Models (14 types)

| Model | Key Fields |
|-------|-----------|
| `BillOfMaterials` | ID, ProductID, Name, Version, Status, Components[] |
| `BOMComponent` | ID, BOMID, ComponentProductID, Quantity, ScrapPercentage |
| `RoutingOperation` | ID, BOMID, SequenceNo, WorkCenterID, SetupTime, RunTime |
| `WorkOrder` | ID, ProductionOrderID, OperationID, Status, StartTime, EndTime |
| `ProductionOrder` | ID, ProductID, Quantity, DueDate, Status, Priority |
| `WorkCenter` | ID, Name, Code, Capacity, Efficiency, Status |
| `LaborReport` | ID, WorkOrderID, EmployeeID, HoursWorked, Date |
| `MachineLog` | ID, WorkCenterID, StartTime, EndTime, Status |
| `QualityInspection` | ID, WorkOrderID, InspectedBy, Result, Defects[] |
| `NonConformance` | ID, QualityInspectionID, Description, Severity, Status |
| `Equipment` | ID, WorkCenterID, Name, Model, Status |
| `MaintenanceOrder` | ID, EquipmentID, ScheduleDate, Type, Status |
| `CostingRecord` | ID, ProductionOrderID, MaterialCost, LaborCost, OverheadCost |

### Business Services (5)

| Service | Key Responsibilities |
|---------|---------------------|
| `BOMService` | BOM CRUD with component hierarchy, routing operations, work center management |
| `ProductionService` | Production orders, work orders (start/complete/cancel), labor reporting, quality inspections, maintenance scheduling, MRP execution, costing |
| `QualityService` | Record/list/get/update quality inspections — **wired in code** |
| `MaintenanceService` | Log machine status, create equipment, schedule/complete maintenance — **code exists but routes may be partial** |
| `CostingService` | Get costing record, run MRP — **not wired** |

### API Endpoints (30 routes)

**Bill of Materials:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/boms` | List BOMs |
| POST | `/api/v1/boms` | Create BOM |
| GET | `/api/v1/boms/:id` | Get BOM |
| PUT | `/api/v1/boms/:id` | Update BOM |
| DELETE | `/api/v1/boms/:id` | Delete BOM |

**Routings:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/routings` | List routings |
| POST | `/api/v1/routings` | Create routing |
| GET | `/api/v1/routings/:id` | Get routing |
| PUT | `/api/v1/routings/:id` | Update routing |
| DELETE | `/api/v1/routings/:id` | Delete routing |

**Work Orders:**
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
| POST | `/api/v1/work-orders/:id/inspect` | Perform inspection |

**Production Plans:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/production-plans` | List production plans |
| POST | `/api/v1/production-plans` | Create production plan |
| GET | `/api/v1/production-plans/:id` | Get plan |
| PUT | `/api/v1/production-plans/:id` | Update plan |
| DELETE | `/api/v1/production-plans/:id` | Delete plan |

**MRP:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/mrp/run` | Run MRP (stub) |

**Quality:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/quality-inspections` | List inspections |
| POST | `/api/v1/quality-inspections` | Create inspection |
| GET | `/api/v1/quality-inspections/:id` | Get inspection |
| PUT | `/api/v1/quality-inspections/:id` | Update inspection |
| DELETE | `/api/v1/quality-inspections/:id` | Delete inspection |

**Work Centers:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/work-centers` | List work centers |
| POST | `/api/v1/work-centers` | Create work center |
| GET | `/api/v1/work-centers/:id` | Get work center |
| PUT | `/api/v1/work-centers/:id` | Update work center |
| DELETE | `/api/v1/work-centers/:id` | Delete work center |
| POST | `/api/v1/work-centers/:id/machine-log` | Record machine log |

**Maintenance:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/maintenance-schedules` | List schedules |
| POST | `/api/v1/maintenance-schedules` | Create schedule |
| GET | `/api/v1/maintenance-schedules/:id` | Get schedule |
| PUT | `/api/v1/maintenance-schedules/:id` | Update schedule |
| DELETE | `/api/v1/maintenance-schedules/:id` | Delete schedule |

### Kafka Events Published (19 topics)

**Production:** `mfg.production.scheduled`, `mfg.production.started`, `mfg.production.completed`, `mfg.production.delayed`
**Work Orders:** `mfg.work.order.created`, `mfg.work.order.started`, `mfg.work.order.completed`, `mfg.work.order.cancelled`
**Material:** `mfg.material.consumed`, `mfg.material.wasted`, `mfg.material.required`
**Quality:** `mfg.quality.inspection.passed`, `mfg.quality.inspection.failed`, `mfg.quality.non.conformance.detected`
**Maintenance:** `mfg.maintenance.scheduled`, `mfg.maintenance.completed`
**Equipment:** `mfg.equipment.down`, `mfg.equipment.up`
**Other:** `mfg.custom.production.completed`

### Kafka Events Consumed (6 topics, per CDD)

| Topic | Publisher | Logic |
|-------|-----------|-------|
| `scm.material.received` | SCM | Logged only |
| `scm.inventory.updated` | SCM | Logged only |
| `crm.sales.order.created` | CRM | Auto-schedule production order |
| `fin.cost.budget.allocated` | FM | Adjust production schedules |
| `hr.employee.scheduled` | HR | Logged only |
| `prj.custom.order.created` | PM | Schedule custom production |

---

## [Project Management](project-management/)

Portfolios, projects, tasks, resource allocation, time/expense tracking, documents, issues, change requests, and cross-service integration. Port **8006** (docker-compose: 8005).

### Domain Models (12 types)

| Model | Key Fields |
|-------|-----------|
| `Project` | ID, PortfolioID, Name, Description, StartDate, EndDate, Budget, Status |
| `Portfolio` | ID, Name, Description, Budget, Status |
| `Task` | ID, ProjectID, ParentTaskID, Name, Description, Status, AssignedTo, StartDate, DueDate, Progress |
| `TaskDependency` | ID, TaskID, DependsOnTaskID, DependencyType |
| `ResourceAllocation` | ID, ProjectID, EmployeeID, Role, AllocationPercentage, StartDate, EndDate |
| `ProjectTimeEntry` | ID, ProjectID, EmployeeID, Date, Hours, Description, Status |
| `ProjectExpense` | ID, ProjectID, Category, Amount, Description, Status |
| `ProjectDocument` | ID, ProjectID, Name, Type, FileURL, UploadedBy |
| `ProjectIssue` | ID, ProjectID, Title, Description, Priority, Status, AssignedTo |
| `ChangeRequest` | ID, ProjectID, Title, Description, Impact, Status |
| `ProjectActivity` | ID, ProjectID, ActivityType, Description, PerformedBy |

### Business Services (6)

| Service | Key Responsibilities |
|---------|---------------------|
| `ProjectPlanningService` | Portfolio & project CRUD, status management |
| `TaskManagementService` | Task CRUD, WBS hierarchy, dependency management, assignment, progress tracking |
| `ResourceManagementService` | Resource allocation by role/skill/timeframe |
| `TimeExpenseService` | Time entry, expense tracking, approval workflow |
| `CollaborationService` | Documents, issues, change requests |
| `PortfolioAnalyticsService` | Portfolio summary, project health |

### API Endpoints (23 routes)

**Portfolios:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/portfolios` | List portfolios |
| POST | `/api/v1/projects/portfolios` | Create portfolio |
| GET | `/api/v1/projects/portfolios/:id` | Get portfolio |
| PUT | `/api/v1/projects/portfolios/:id` | Update portfolio |
| DELETE | `/api/v1/projects/portfolios/:id` | Delete portfolio |
| GET | `/api/v1/projects/portfolios/:id/summary` | Portfolio summary |

**Projects:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects` | List projects |
| POST | `/api/v1/projects` | Create project |
| GET | `/api/v1/projects/:id` | Get project |
| PUT | `/api/v1/projects/:id` | Update project |

**Tasks:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/tasks` | List project tasks |
| POST | `/api/v1/projects/:id/tasks` | Create task |
| PUT | `/api/v1/projects/tasks/:task_id/progress` | Update progress |
| PUT | `/api/v1/projects/tasks/:task_id/assign` | Assign task |
| POST | `/api/v1/projects/tasks/:task_id/dependencies` | Add dependency |

**Resource Allocations:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/allocations` | List allocations |
| POST | `/api/v1/projects/:id/allocations` | Create allocation |

**Time & Expenses:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/time` | List time entries |
| POST | `/api/v1/projects/:id/time` | Create time entry |
| PUT | `/api/v1/projects/time/:time_id/approve` | Approve time entry |
| GET | `/api/v1/projects/:id/expenses` | List expenses |
| POST | `/api/v1/projects/:id/expenses` | Create expense |
| PUT | `/api/v1/projects/expenses/:expense_id/approve` | Approve expense |

**Documents:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/documents` | List documents |
| POST | `/api/v1/projects/:id/documents` | Upload document |

**Issues:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/issues` | List issues |
| POST | `/api/v1/projects/:id/issues` | Create issue |
| PUT | `/api/v1/projects/issues/:issue_id/resolve` | Resolve issue |

**Change Requests:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/change-requests` | List CRs |
| POST | `/api/v1/projects/:id/change-requests` | Create CR |
| PUT | `/api/v1/projects/change-requests/:request_id/approve` | Approve CR |

**Cross-Service Integration:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/projects/:id/request-material` | Request material (→ SCM via Kafka) |
| POST | `/api/v1/projects/:id/request-custom-order` | Request custom order (→ MFG via Kafka) |

### Kafka Events Published (25 topics)

**Project:** `prj.project.created`, `prj.project.updated`, `prj.project.started`, `prj.project.completed`, `prj.project.cancelled`, `prj.project.delayed`
**Task:** `prj.task.created`, `prj.task.assigned`, `prj.task.started`, `prj.task.completed`, `prj.task.overdue`
**Resource:** `prj.resource.allocated`, `prj.resource.released`, `prj.resource.overallocated`
**Time/Expense:** `prj.time.logged`, `prj.time.approved`, `prj.time.rejected`, `prj.expense.submitted`, `prj.expense.approved`, `prj.expense.rejected`, `prj.expense.incurred`
**Milestones:** `prj.milestone.achieved`, `prj.milestone.delayed`
**Integration:** `prj.custom.order.created`, `prj.material.requested`

### Kafka Events Consumed (8 topics, per CDD)

| Topic | Publisher | Consumer Logic |
|-------|-----------|----------------|
| `hr.employee.available` | HR | Logged only |
| `hr.employee.skills.updated` | HR | Logged only |
| `fin.budget.approved` | FM | Release project funding upon budget approval |
| `fin.payment.received` | FM | Logged only |
| `crm.sales.order.received` | CRM | Auto-create project + kickoff task |
| `scm.material.delivered` | SCM | Logged only |
| `mfg.custom.production.completed` | MFG | Logged only |

---

## Module Integration Map

### Cross-Service Event Flows

**Make-to-Order:**
```
CRM creates Sales Order
  → publishes crm.sales.order.created
  → MFG consumes: auto-schedules production
  → SCM consumes: triggers fulfillment
```

**Project-Driven Manufacturing:**
```
PM requests Custom Order (POST /api/v1/projects/:id/request-custom-order)
  → publishes prj.custom.order.created
  → MFG consumes: schedules custom production
```

**Production Completion → Financial:**
```
MFG completes Work Order
  → publishes mfg.production.completed, mfg.material.consumed
  → FM consumes: creates WIP→finished goods + material journal entries
```

**Procurement → Financial:**
```
SCM creates Purchase Order
  → publishes scm.purchase.order.created
  → FM consumes: creates inventory-in-transit journal entry
```

**Payroll → Financial:**
```
HR processes Payroll
  → publishes hr.payroll.processed
  → FM consumes: creates salary expense journal entry
```

**Project Time → Financial:**
```
PM logs Time Entry
  → publishes prj.time.logged
  → FM consumes: creates unbilled receivables journal entry
```

**Project Request → Supply Chain:**
```
PM requests Material (POST /api/v1/projects/:id/request-material)
  → publishes prj.material.requested
  → SCM consumes: issues material from inventory
```

**Training Trigger:**
```
SCM requires Training
  → publishes scm.training.required
  → HR consumes: auto-creates training program
```

**Budget Allocation:**
```
FM creates/approves Budget
  → publishes fin.budget.allocated, fin.budget.approved, fin.cost.budget.allocated
  → HR consumes: adjust hiring plans
  → PM consumes: release project funding
  → MFG consumes: adjust production schedules
```

### Data Sharing Matrix

| Module | Shares With | Data Shared | Mechanism |
|--------|-------------|-------------|-----------|
| **CRM** | FM | Customer data, sales completion | Kafka (`crm.sale.completed`, `crm.customer.created`) |
| **CRM** | MFG | Sales order → production trigger | Kafka (`crm.sales.order.created`) |
| **CRM** | PM | Sales order → project trigger | Kafka (`crm.sales.order.received`) |
| **CRM** | SCM | Demand forecast | Kafka (`crm.customer.demand.forecast`) |
| **SCM** | FM | Purchase orders, inventory valuation | Kafka (`scm.purchase.order.created`, `scm.inventory.valued`) |
| **SCM** | MFG | Material requirements | Kafka (`mfg.material.required`) |
| **SCM** | HR | Training needs | Kafka (`scm.training.required`) |
| **HR** | FM | Payroll expenses, employee costs | Kafka (`hr.payroll.processed`, `hr.employee.created`) |
| **HR** | PM | Employee data via API | REST lookup |
| **MFG** | FM | Production costs, material consumption | Kafka (`mfg.production.completed`, `mfg.material.consumed`) |
| **MFG** | SCM | Material requirements | Kafka (`mfg.material.required`) |
| **PM** | FM | Project time, expenses | Kafka (`prj.time.logged`, `prj.expense.incurred`) |
| **PM** | SCM | Material requests | Kafka (`prj.material.requested`) |
| **PM** | MFG | Custom production orders | Kafka (`prj.custom.order.created`) |

---

## API Gateway Routing

The API Gateway (`api-gateway/cmd/main.go`, port 8080) reverse-proxies to all services:

| Gateway Path | Backend URL | Service |
|-------------|-------------|---------|
| `/api/v1/finance/*path` | `http://finance-service:8001/*path` | FM |
| `/api/v1/hr/*path` | `http://hr-service:8002/*path` | HR |
| `/api/v1/scm/*path` | `http://scm-service:8003/*path` | SCM |
| `/api/v1/manufacturing/*path` | `http://manufacturing-service:8004/*path` | MFG |
| `/api/v1/crm/*path` | `http://crm-service:8005/*path` | CRM |
| `/api/v1/projects/*path` | `http://projects-service:8006/*path` | PM |

> **Note**: The gateway backend URLs use **different port numbers** from docker-compose port mappings. The gateway expects services on ports 8001-8006, which matches docker-compose exposed ports.

### Gateway Features

- No authentication middleware
- No rate limiting
- No request/response transformation
- Request path forwarded as-is
- Responses proxied directly from backend services

A full-featured alternative server exists at `api-gateway/internal/server/server.go` with JWT validation, RBAC permission checks, rate limiting, and CORS middleware — but is **not the deployed binary**.

---

## Infrastructure State

| Component | Status |
|-----------|--------|
| **PostgreSQL** | Not connected — no Go database driver imported in any service |
| **Redis** | Not connected — no Redis client code exists |
| **Kafka** | Fully wired — all services produce and consume events via `segmentio/kafka-go` |
| **Auth Service** | Running on port 8000 — issues JWT tokens, but gateway doesn't use them |
| **API Gateway** | Running on port 8080 — simple reverse proxy, no auth |

### Shared Utilities

The `shared/` directory provides:
- `utils/logger.go` — structured logger with Info/Error/Debug/Warn levels and Gin middleware
- `utils/response.go` — standard JSON response helpers (Success, Error, BadRequest, etc.)
- `templates/` — Swagger scaffolding templates for generating `cmd/main.go` per service

These utilities are **available but not used** by any service — all services use `log.Printf` and raw `gin.H` responses.

---

## Known Limitations

| Issue | Details |
|-------|---------|
| **No persistence** | All data in-memory, lost on restart |
| **No auth** | API Gateway exposes all endpoints publicly |
| **Fire-and-forget Kafka** | All event publishes ignore errors (`_ = publisher.Publish(...)`) |
| **No pagination** | List endpoints return all records in one response |
| **Plaintext passwords** | Auth Service stores/compares passwords without hashing |
| **Hardcoded JWT secret** | Default `super-secret-key-123` in source code |
| **Port mismatches** | Gateway backend URLs (8081-8086) differ from docker-compose (8001-8006) |
| **No multi-currency logic** | Domain model exists but no conversion or rate logic |
| **No workflow engine** | All approval logic is custom per-service, no shared engine |
| **No reporting engine** | Reports are simple JSON aggregations, no export/PDF |

## Next Steps

### For Business Users
1. Start with [CRM](customer-relationship-management/) for sales processes
2. Review [Financial Management](financial-management/) for accounting
3. Explore [HR](human-resources/) for employee management

### For Developers
1. Understand [Financial Management](financial-management/) architecture
2. Review [SCM](supply-chain-management/) integration patterns
3. Study [Manufacturing](manufacturing/) workflow automation

### For System Administrators
1. Review [HR](human-resources/) security requirements
2. Understand [Project Management](project-management/) resource needs
3. Plan [CRM](customer-relationship-management/) scaling strategies

## Related Documentation

- [System Architecture](../architecture/services-overview.md) — Service structure and endpoints
- [Event Architecture](../architecture/event-architecture.md) — Full Kafka event catalog with 85+ topics
- [Common Issues](../getting-started/common-issues.md) — 41 known problems across the codebase
- [API Design](../architecture/api-design.md) — Gateway routing and conventions
- [Operations](../operations/README.md) — Deployment, configuration, and troubleshooting
- [Getting Started](../getting-started/README.md) — Setup and development workflow
