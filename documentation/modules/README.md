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

General ledger, accounts receivable, accounts payable, cash management, budgeting, fixed assets, and financial reports. Port **8001**.

### Domain Models

| Model | Key Fields | Event Triggers |
|-------|-----------|----------------|
| `LegalEntity` | ID, CompanyCode, CompanyName, FunctionalCurrency, TaxRegistrationNumber | — |
| `ChartOfAccounts` | ID, LegalEntityID, AccountCode, AccountName, Type (ASSET/LIABILITY/EQUITY/REVENUE/EXPENSE), IsActive | `fm.account.created`, `fm.account.updated`, `fm.account.balance.changed` |
| `UniversalJournalEntry` | ID, LegalEntityID, SourceModule, SourceDocumentID, PostingDate, FinancialPeriod, Status (DRAFT/POSTED/REVERSED) | — |
| `UniversalJournalLine` | ID, JournalEntryID, AccountID, AmountFunctional, AmountTransactional, CurrencyTransactional | `fm.account.balance.changed` (via outbox event) |
| `ArInvoice` | ID, LegalEntityID, InvoiceNumber, CustomerID, SalesOrderID, TotalAmount, TaxAmount, DueDate, Status | `fm.invoice.created`, `fm.invoice.updated`, `fm.invoice.sent`, `fm.invoice.paid` |
| `ApVendorBill` | ID, LegalEntityID, BillNumber, VendorID, PurchaseOrderID, TotalAmount, TaxAmount, DueDate, Status | `fm.vendor.paid` |
| `CapitalAsset` | ID, LegalEntityID, AssetTag, EamEquipmentID, AcquisitionCost, AccumulatedDepreciation, UsefulLifeMonths, CapitalizationDate, Status | — |
| `DepreciationScheduleLine` | ID, FixedAssetID, FiscalYear, PeriodNumber, DepreciationAmount, IsPosted | — |
| `BankAccount` | ID, LegalEntityID, AccountNumber, Currency, LiquidBalance | — |
| `Payment` | ID, InvoiceID, BillID, BankAccountID, PaymentNumber, PaymentDate, Amount, PaymentMethod, Status | `fm.payment.received`, `fm.payment.processed` |
| `BankStatement` | ID, BankAccountID, StatementDate, EndingBalance, IsReconciled | — |
| `BankStatementLine` | ID, StatementID, TransactionDate, Description, Amount, IsMatched | — |
| `Budget` | ID, AccountID, CostCenterID, FiscalYear, Period, AllocatedAmount, SpentAmount | `fm.budget.created`, `fm.budget.updated`, `fm.budget.exceeded`, `fm.budget.approved` |

### Business Services (6)

| Service | Key Methods | Business Logic |
|---------|-------------|---------------|
| `GeneralLedgerService` | `CreateAccount`, `GetAccountBalance`, `CreateJournalEntry`, `ReverseJournalEntry`, `GetBalanceSheet`, `GetIncomeStatement`, `GetCashFlow` | Multi-tenant chart of accounts, balanced universal double-entry validation, reports from live database lines. |
| `AccountsReceivableService` | `CreateInvoice`, `GetInvoice`, `SendInvoice` | Customer invoice lifecycle, flat schemas. |
| `AccountsPayableService` | `CreateVendorBill`, `GetVendorBill` | Vendor bill lifecycle. |
| `CashManagementService` | `RecordPayment`, `GetPayments`, `GetBankStatement` | Record payment against AR/AP, bank statement line tracking. |
| `BudgetingService` | `CreateBudget`, `GetBudgetVariance` | Period budgeting and actual vs budget variance comparison. |
| `CapitalAssetService` | `CapitalizeAsset`, `GenerateDepreciationSchedule`, `PostMonthlyStraightLineDepreciation` | Straight-line depreciation scheduling and GL posting. |

### API Endpoints (33 routes)

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/health` | inline | Health check |
| GET | `/api/v1/legal-entities` | `leHandler.GetLegalEntities` | List legal entities |
| POST | `/api/v1/legal-entities` | `leHandler.CreateLegalEntity` | Create legal entity |
| GET | `/api/v1/legal-entities/:id` | `leHandler.GetLegalEntity` | Get legal entity by ID |
| GET | `/api/v1/accounts` | `accHandler.GetAccounts` | List all accounts |
| POST | `/api/v1/accounts` | `accHandler.CreateAccount` | Create account |
| GET | `/api/v1/accounts/:id` | `accHandler.GetAccount` | Get account by ID |
| PUT | `/api/v1/accounts/:id` | `accHandler.UpdateAccount` | Update account properties |
| DELETE | `/api/v1/accounts/:id` | `accHandler.DeleteAccount` | Delete account |
| GET | `/api/v1/accounts/:id/balance` | `accHandler.GetAccountBalance` | Get account balance |
| GET | `/api/v1/journal-entries` | `txHandler.GetTransactions` | List journal entries |
| POST | `/api/v1/journal-entries` | `txHandler.CreateTransaction` | Create journal entry |
| GET | `/api/v1/journal-entries/:id` | `txHandler.GetTransaction` | Get journal entry with lines |
| PUT | `/api/v1/journal-entries/:id` | `txHandler.UpdateTransaction` | Update journal entry |
| DELETE | `/api/v1/journal-entries/:id` | `txHandler.DeleteTransaction` | Delete journal entry |
| GET | `/api/v1/invoices` | `invHandler.GetInvoices` | List invoices |
| POST | `/api/v1/invoices` | `invHandler.CreateInvoice` | Create invoice |
| GET | `/api/v1/invoices/:id` | `invHandler.GetInvoice` | Get invoice details |
| PUT | `/api/v1/invoices/:id` | `invHandler.UpdateInvoice` | Update invoice |
| DELETE | `/api/v1/invoices/:id` | `invHandler.DeleteInvoice` | Delete invoice |
| POST | `/api/v1/invoices/:id/send` | `invHandler.SendInvoice` | Send invoice to customer |
| GET | `/api/v1/invoices/:id/lines` | `invHandler.GetInvoiceLines` | Get empty lines list for compatibility |
| GET | `/api/v1/vendor-bills` | `billHandler.GetVendorBills` | List vendor bills |
| POST | `/api/v1/vendor-bills` | `billHandler.CreateVendorBill` | Create vendor bill |
| GET | `/api/v1/vendor-bills/:id/lines` | `billHandler.GetVendorBillLines` | Get empty lines list for compatibility |
| GET | `/api/v1/payments` | `payHandler.GetPayments` | List payments |
| POST | `/api/v1/payments` | `payHandler.RecordPayment` | Record payment |
| GET | `/api/v1/payments/:id` | `payHandler.GetPayment` | Get payment details |
| GET | `/api/v1/bank-statements/:id/lines` | `payHandler.GetBankStatementLines` | Get bank statement lines |
| GET | `/api/v1/assets` | `assetHandler.GetAssets` | List capitalized assets |
| POST | `/api/v1/assets/capitalize` | `assetHandler.CapitalizeAsset` | Capitalize asset |
| GET | `/api/v1/assets/:id` | `assetHandler.GetAsset` | Get asset details |
| POST | `/api/v1/assets/:id/depreciation-schedule` | `assetHandler.GenerateDepreciationSchedule` | Generate straight-line depreciation schedule |
| POST | `/api/v1/assets/depreciate` | `assetHandler.PostMonthlyDepreciation` | Post monthly depreciation to GL |
| GET | `/api/v1/reports/balance-sheet` | `repHandler.GetBalanceSheet` | Balance sheet report (GL live ledger-driven) |
| GET | `/api/v1/reports/income-statement` | `repHandler.GetIncomeStatement` | Income statement report |
| GET | `/api/v1/reports/cash-flow` | `repHandler.GetCashFlow` | Cash flow report |

### Kafka Events Published (18 topics)

`fm.invoice.created`, `fm.invoice.updated`, `fm.invoice.sent`, `fm.invoice.paid`, `fm.invoice.overdue`, `fm.payment.received`, `fm.payment.processed`, `fm.payment.failed`, `fm.vendor.payment.due`, `fm.vendor.paid`, `fm.customer.credit_status.updated`, `fm.account.created`, `fm.account.updated`, `fm.account.balance.changed`, `fm.budget.created`, `fm.budget.updated`, `fm.budget.exceeded`, `fm.budget.approved`

### Kafka Events Consumed (17 topics)

Consumed transactionally via Inbox deduplication checks:
- **HR → FM**: `hr.payroll.processed` (create salary journal entry), `hr.employee.created` (track employee), `hr.expense.submitted` (create expense entry)
- **SCM → FM**: `scm.receipt.staged` (staged receipt), `scm.order.shipped` (record COGS), `scm.purchase.order.created` (reference PO), `scm.invoice.received` (create AP entry), `scm.inventory.valued` (update inventory GL balance)
- **CRM → FM**: `crm.order.confirmed` (create receivable invoice), `crm.customer.created` (track new customer)
- **MFG → FM**: `mfg.yield.produced` (material yield), `mfg.production.completed` (WIP→finished goods entry), `mfg.material.consumed` (raw material issue entry)
- **PM → FM**: `prj.milestone.achieved` (billing milestone), `prj.project.created` (track project), `prj.time.logged` (create unbilled receivable entry), `prj.expense.incurred` (capitalize project cost)

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

### Domain Models

| Model | Key Fields |
|-------|-----------|
| `Product` | ID, ProductCode, ProductName, Description, ProductType, CategoryID, UnitOfMeasure, StandardCost, ListPrice, IsActive |
| `ProductCategory` | ID, Code, Name, Description |
| `Location` | ID, LocationCode, LocationName, LocationType, IsActive |
| `Supplier` | ID, SupplierCode, SupplierName, ContactName, Email, Phone, IsActive |
| `VendorContract` | ID, ContractNumber, SupplierID, StartDate, EndDate, Terms, Status |
| `PurchaseRequisition` | ID, ReqNumber, RequesterID, RequestDate, Status, TotalAmount, Notes |
| `PurchaseRequisitionLine` | ID, PurchaseRequisitionID, ProductID, QuantityRequested, EstimatedUnitPrice, LineTotal |
| `PurchaseOrder` | ID, PoNumber, SupplierID, OrderDate, ExpectedDelivery, Status, TotalAmount, Notes |
| `PurchaseOrderLine` | ID, PurchaseOrderID, ProductID, QuantityOrdered, QuantityReceived, UnitPrice, LineTotal, Description |
| `InventoryItem` | ID, ProductID, LocationID, QuantityOnHand, QuantityReserved, QuantityAvailable, ReorderPoint, MaximumStock, UnitCost |
| `InventoryMovement` | ID, ProductID, LocationID, MovementType, Quantity, UnitCost, ReferenceType, ReferenceID, Notes |
| `StockTransfer` | ID, FromLocationID, ToLocationID, ProductID, Quantity, Status, TransferredAt |
| `Receipt` | ID, PurchaseOrderID, ReceivedDate, Status, Notes |
| `ReceiptLine` | ID, ReceiptID, ProductID, QuantityReceived, LocationID |
| `Shipment` | ID, Carrier, TrackingNumber, EstimatedDelivery, Status, Notes |
| `ShipmentLine` | ID, ShipmentID, ProductID, QuantityShipped, LocationID |
| `DemandForecast` | ID, ProductID, ForecastDate, ForecastQuantity, ConfidenceLevel, Notes |

### Business Services (7)

| Service | Key Responsibilities |
|---------|---------------------|
| `ProductManagementService` | Product CRUD, category classification, locations CRUD |
| `SupplierManagementService` | Supplier CRUD, contract management |
| `PurchaseOrderService` | Requisition→PO lifecycle, approvals, send PO to vendor |
| `InventoryService` | Stock tracking, reservations, movements recording, transfers |
| `WarehouseService` | Goods receipts processing, outbound shipments |
| `DemandPlanningService` | Demand forecasts management |

### API Endpoints (47 routes)

**Product Categories:**
- `GET /api/v1/product-categories` — List categories
- `POST /api/v1/product-categories` — Create category
- `GET /api/v1/product-categories/:id` — Get category details
- `PUT /api/v1/product-categories/:id` — Update category
- `DELETE /api/v1/product-categories/:id` — Delete category

**Products:**
- `GET /api/v1/products` — List products
- `POST /api/v1/products` — Create product
- `GET /api/v1/products/:id` — Get product details
- `PUT /api/v1/products/:id` — Update product
- `DELETE /api/v1/products/:id` — Delete product

**Locations:**
- `GET /api/v1/locations` — List locations
- `POST /api/v1/locations` — Create location
- `GET /api/v1/locations/:id` — Get location details
- `PUT /api/v1/locations/:id` — Update location
- `DELETE /api/v1/locations/:id` — Delete location

**Supplier Management:**
- `GET /api/v1/vendors` — List vendors
- `POST /api/v1/vendors` — Create vendor
- `GET /api/v1/vendors/:id` — Get vendor details
- `PUT /api/v1/vendors/:id` — Update vendor details
- `DELETE /api/v1/vendors/:id` — Delete vendor
- `GET /api/v1/vendor-contracts` — List contracts
- `POST /api/v1/vendor-contracts` — Create contract
- `GET /api/v1/vendor-contracts/:id` — Get contract details
- `PUT /api/v1/vendor-contracts/:id` — Update contract details
- `DELETE /api/v1/vendor-contracts/:id` — Delete contract

**Purchase Requisitions:**
- `GET /api/v1/purchase-requisitions` — List requisitions
- `POST /api/v1/purchase-requisitions` — Create requisition
- `GET /api/v1/purchase-requisitions/:id` — Get requisition details
- `PUT /api/v1/purchase-requisitions/:id` — Update requisition details
- `DELETE /api/v1/purchase-requisitions/:id` — Delete requisition
- `POST /api/v1/purchase-requisitions/:id/approve` — Approve requisition
- `POST /api/v1/purchase-requisitions/:id/reject` — Reject requisition
- `GET /api/v1/purchase-requisitions/:id/lines` — Get requisition line items

**Purchase Orders:**
- `GET /api/v1/purchase-orders` — List purchase orders
- `POST /api/v1/purchase-orders` — Create purchase order
- `GET /api/v1/purchase-orders/:id` — Get purchase order details
- `PUT /api/v1/purchase-orders/:id` — Update purchase order details
- `DELETE /api/v1/purchase-orders/:id` — Delete purchase order
- `POST /api/v1/purchase-orders/:id/send` — Send PO to supplier
- `GET /api/v1/purchase-orders/:id/lines` — Get PO line items

**Inventory & Transfers:**
- `GET /api/v1/inventory` — List inventory items
- `POST /api/v1/inventory` — Create inventory item
- `GET /api/v1/inventory/:id` — Get inventory item details
- `PUT /api/v1/inventory/:id` — Update inventory item details
- `DELETE /api/v1/inventory/:id` — Delete inventory item
- `POST /api/v1/inventory/reserve` — Reserve stock
- `POST /api/v1/inventory/release` — Release stock reservation
- `GET /api/v1/inventory/movements` — List inventory movements
- `GET /api/v1/stock-transfers` — List transfers
- `POST /api/v1/stock-transfers` — Create stock transfer
- `GET /api/v1/stock-transfers/:id` - Get transfer details
- `POST /api/v1/stock-transfers/:id/execute` — Execute stock transfer

**Warehouse Operations:**
- `GET /api/v1/receipts` — List receipts
- `POST /api/v1/receipts` — Create goods receipt
- `GET /api/v1/receipts/:id` — Get receipt details
- `PUT /api/v1/receipts/:id` — Update receipt details
- `GET /api/v1/receipts/:id/lines` — Get receipt line items
- `GET /api/v1/shipments` — List shipments
- `POST /api/v1/shipments` — Create shipment
- `GET /api/v1/shipments/:id` — Get shipment details
- `PUT /api/v1/shipments/:id` — Update shipment details
- `GET /api/v1/shipments/:id/lines` — Get shipment line items

**Demand Planning:**
- `GET /api/v1/demand-forecasts` — List forecasts
- `POST /api/v1/demand-forecasts` — Create forecast
- `GET /api/v1/demand-forecasts/:id` — Get forecast details
- `PUT /api/v1/demand-forecasts/:id` — Update forecast details

**Reports:**
- `GET /api/v1/reports/inventory-levels` — Inventory levels report
- `GET /api/v1/reports/vendor-performance` — Vendor performance metrics
- `GET /api/v1/reports/procurement-metrics` — Procurement metrics
- `GET /api/v1/reports/safety-stock` — Safety stock report

### Kafka Events Published (22 topics)

- **Inventory**: `scm.inventory.received`, `scm.inventory.shipped`, `scm.inventory.adjusted`, `scm.inventory.low.stock`, `scm.inventory.out.of.stock`, `scm.inventory.valued`, `scm.inventory.updated`
- **Purchase Orders**: `scm.purchase.order.created`, `scm.purchase.order.sent`, `scm.purchase.order.received`, `scm.purchase.order.cancelled`
- **Vendors**: `scm.vendor.created`, `scm.vendor.updated`, `scm.vendor.performance.evaluated`
- **Shipments**: `scm.shipment.created`, `scm.shipment.dispatched`, `scm.shipment.delivered`, `scm.shipment.delayed`
- **Other**: `scm.training.required`, `scm.material.delivered`

### Kafka Events Consumed (7 topics)

| Topic | Publisher | Logic |
|-------|-----------|-------|
| `crm.sales.order.created` | CRM | Logged for metrics |
| `crm.customer.demand.forecast` | CRM | Create demand forecast record |
| `mfg.material.required` | MFG | Auto-create purchase requisition |
| `mfg.material.consumed` | MFG | Issue raw material from inventory |
| `mfg.production.completed` | MFG | Receive finished goods into inventory |
| `fin.vendor.payment.processed` | FM | Logged for status sync |
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

A highly optimized, multi-tenant Shop Floor Execution core. Port **8004** (docker-compose: 8004).

### Domain Models (8 types)

| Model | Key Fields | Description |
|-------|------------|-------------|
| `WorkCenter` | ID, LegalEntityID, WorkCenterCode, Name | Core shop floor work area |
| `RoutingStation` | ID, WorkCenterID, RoutingCode, SetupTime, RunTime | Step inside a work center |
| `WorkOrder` | ID, LegalEntityID, WorkOrderNumber, MaterialID, BomHeaderID, QuantityTarget, Status | Production execution task |
| `WorkOrderRoutingState` | ID, WorkOrderID, CurrentStationID, Status, SequenceNumber | State machine tracker |
| `MaterialConsumptionLog` | ID, LegalEntityID, WorkOrderID, StationID, MaterialID, QuantityConsumed | Physical material usage |
| `ProductionYieldLog` | ID, LegalEntityID, WorkOrderID, StationID, QuantityGood, QuantityScrap | Scrap vs good production count |
| `TransactionalOutbox` | ID, EventType, AggregateID, Payload, Status, CreatedAt | Outbox event delivery |
| `KafkaEventInbox` | EventID, EventType, ProcessedAt, ProcessingStatus, Payload | Idempotent event receiver |

### Business Services (5)

| Service | Key Responsibilities |
|---------|---------------------|
| `FloorConfigurationService` | Work center setups, station assignments |
| `WorkOrderExecutionService` | Work order creation, state transitions, rerouting |
| `ShopFloorTelemetryService` | Material consumption, scrap/good yield logs |
| `OutboxRelayWorker` | Dispatch outbox events to Kafka |
| `ReliableMessagingService` | Idempotent transaction receiver |

### API Endpoints (11 routes)

**Work Centers & Routing:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/mfg/work-centers` | Establish work center |
| POST | `/api/v1/mfg/work-centers/:id/stations` | Append station to work center |

**Work Orders:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/mfg/work-orders` | Instantiate work order |
| PUT | `/api/v1/mfg/work-orders/:id/state` | Transition work order state |
| PUT | `/api/v1/mfg/work-orders/:id/reroute` | Reroute work order station |
| GET | `/api/v1/mfg/work-orders/:id` | Fetch work order detail |
| GET | `/api/v1/mfg/work-orders` | List work orders |

**Telemetry & Logs:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/mfg/telemetry/consumption` | Record material consumption |
| POST | `/api/v1/mfg/telemetry/yield` | Commit production yield |
| GET | `/api/v1/mfg/work-orders/:id/state` | Get routing state timeline |

### Kafka Events Published (3 topics)

`mfg.production.started`, `mfg.material.consumed`, `mfg.yield.produced`

### Kafka Events Consumed (3 topics)

| Topic | Publisher | Logic |
|-------|-----------|-------|
| `plm.bom.released` | PLM | Logged only |
| `qms.inspection.passed/failed` | QMS | Logged only |
| `eam.offline` | EAM | Logged only |

## [Product Lifecycle Management](product-lifecycle-management/)

Material specifications, Bill of Materials (BOM), Engineering Change Orders (ECO), and design revision tracking. Port **8008**.

### Domain Models (4 types)

| Model | Key Fields |
|-------|-----------|
| `MaterialMaster` | ID, LegalEntityID, SKU, Description, UOM, ProcurementType, Status, TechnicalSpecifications |
| `BomHeader` | ID, LegalEntityID, MaterialID, EcoID, VersionString, Status |
| `BomLine` | ID, BomHeaderID, ComponentMaterialID, SequenceNumber, QuantityRequired, UOM, ScrapPercentage |
| `EngineeringChangeOrder` | ID, LegalEntityID, TargetMaterialID, EcoNumber, Title, Description, Status, RequestedByHrID, ApprovedByHrID |

### Business Services (3)

| Service | Key Methods |
|---------|-------------|
| `MaterialService` | `createMaterial`, `updateTechnicalSpecs`, `transitionStatus` |
| `BomService` | `establishBomHeader`, `releaseBom`, `explodeBillOfMaterials` |
| `EngineeringChangeService` | `initiateChangeRequest`, `processApprovalAction` |

### Kafka Events Published (4 topics)

`plm.material.released`, `plm.material.obsoleted`, `plm.bom.released`, `plm.eco.implemented`

### Kafka Events Consumed (5 topics)

`scm.receipt.staged`, `mfg.material.consumed`, `hr.employee.created`, `qms.inspection.failed`, `eam.machine.offline`

---

## [Project Management](project-management/)

Projects, Work Breakdown Structure (WBS), and timesheet validation. Port **8005** (docker-compose: 8005).

### Domain Models (5 types)

| Model | Key Fields | Description |
|-------|------------|-------------|
| `Project` | ID, LegalEntityID, CustomerID, ProjectCode, Name, Status, BillingMethod, StartDate, EndDate, Version | Core project tracking metadata |
| `WbsNode` | ID, ProjectID, ParentNodeID, WbsDepthLevel, NodeCode, Title, NodeType, EstimatedHours, BudgetRevenueFunctional, IsCompleted, Version | Work breakdown structure elements |
| `TimeLog` | ID, LegalEntityID, WbsNodeID, EmployeeID, WorkDate, HoursSpent, InternalCostRate, BillingRate, IsBillable, IsApproved, ApprovedByHrID | Timesheet logs |
| `TransactionalOutbox` | ID, EventType, AggregateID, Payload, Status, CreatedAt | Outbox event delivery |
| `KafkaEventInbox` | EventID, EventType, ProcessedAt, ProcessingStatus, Payload | Idempotent event receiver |

### Business Services (5)

| Service | Key Responsibilities |
|---------|---------------------|
| `ProjectTrackingService` | Project initialization and status transitions |
| `WbsStructureService` | WBS node creation, completion, and tree fetch |
| `TimeTrackingService` | Time logs, bulk submission, and timesheet approvals |
| `OutboxRelayWorker` | Dispatch outbox events to Kafka |
| `ReliableMessagingService` | Idempotent transaction receiver |

### API Endpoints (9 routes)

**Projects:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/projects` | Initialize project |
| PUT | `/api/v1/projects/:id/status` | Transition project status |
| GET | `/api/v1/projects` | List projects |
| GET | `/api/v1/projects/:id` | Get project detail |

**WBS Structure:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/projects/:id/wbs` | Append WBS node |
| PUT | `/api/v1/wbs/:node_id/complete` | Declare WBS node completion |
| GET | `/api/v1/projects/:id/wbs` | Fetch project WBS tree |

**Time Tracking:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/time-logs/bulk` | Log bulk operational hours |
| POST | `/api/v1/time-logs/approve` | Process timesheet approvals |

### Kafka Events Published (2 topics)

`prj.time.logged`, `prj.milestone.achieved`

### Kafka Events Consumed (3 topics)

| Topic | Publisher | Logic |
|-------|-----------|-------|
| `hr.employee.created` | HR | Logged only |
| `hr.employee.terminated` | HR | Logged only |
| `crm.sales.order.confirmed` | CRM | Auto-create project stub |

---

## [Enterprise Asset Management](enterprise-asset-management/)

Physical plant locations, equipment registry, soft deletes, reactive and preventative maintenance schedules, and machine telemetry. Port **8007** (docker-compose: 8007).

### Domain Models (7 types)

| Model | Key Fields | Description |
|-------|------------|-------------|
| `Facility` | ID, LegalEntityID, Name, PhysicalAddress, IsActive | Plant location metadata |
| `Equipment` | ID, LegalEntityID, FacilityID, AssetTag, Name, Manufacturer, SerialNumber, FinancialAssetID, Status, InstallationDate, TechnicalSpecifications, DeletedAt | Plant machinery with soft-delete support |
| `MaintenanceWorkOrder` | ID, LegalEntityID, EquipmentID, TicketNumber, Title, Description, Category, Priority, Status, ReportedByHrID, AssignedTechHrID, ReportedAt, StartedAt, ResolvedAt, ResolutionNotes | Work orders for maintenance |
| `PreventativeSchedule` | ID, LegalEntityID, EquipmentID, Title, InstructionSet, IntervalDays, LastExecutedAt, NextDueDate, IsActive | Preventative maintenance rules |
| `TelemetryIngestBuffer` | ID, LegalEntityID, EquipmentID, SensorKey, ReadingValue, RecordedAt | Staged telemetry metrics |
| `TransactionalOutbox` | ID, EventType, AggregateID, Payload, Status, CreatedAt | Outbox event delivery |
| `KafkaEventInbox` | EventID, EventType, ProcessedAt, ProcessingStatus, Payload | Idempotent event receiver |

### Business Services (5)

| Service | Key Responsibilities |
|---------|---------------------|
| `EquipmentService` | Infrastructure config, plant asset registry, and manual status override |
| `MaintenanceService` | Incident ticketing, tech routing, PM scheduling loop, and spares requests |
| `TelemetryIngestionService` | Sensor metrics staging and skipped-lock transaction flushes |
| `OutboxRelayWorker` | Dispatch outbox events to Kafka |
| `ReliableMessagingService` | Idempotent transaction receiver |

### API Endpoints (11 routes)

**Infrastructure & Registry:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/eam/facilities` | Create new facility |
| POST | `/api/v1/eam/equipment` | Register equipment |
| GET | `/api/v1/eam/equipment` | List tenant equipment |
| PUT | `/api/v1/eam/equipment/:id/status` | Override equipment status |
| PUT | `/api/v1/eam/equipment/:id/finance-asset` | Associate financial asset |

**Maintenance Work Orders:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/eam/work-orders` | File incident ticket |
| PUT | `/api/v1/eam/work-orders/:id/route` | Assign technician |
| POST | `/api/v1/eam/work-orders/:id/start` | Start work order |
| POST | `/api/v1/eam/work-orders/:id/resolve` | Resolve work order |

**Telemetry Logs:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/eam/telemetry/sensor-metrics` | Queue telemetry logs |
| POST | `/api/v1/eam/telemetry/flush` | Flush telemetry logs |

### Kafka Events Published (3 topics)

`eam.machine.offline`, `eam.machine.online`, `eam.workorder.spares_requested`

### Kafka Events Consumed (3 topics)

| Topic | Publisher | Logic |
|-------|-----------|-------|
| `scm.asset.received` | SCM | Auto-registers equipment inside EAM |
| `fm.asset.capitalized` | FM | Links equipment to financial capital asset |
| `hr.employee.created` | HR | Syncs technician profile |

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
