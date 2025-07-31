# ERP System - User Stories

This document translates the functional requirements from the PRD into actionable user stories, categorized by domain.

---

## 1. Financial Management (FIN)

### Epic: General Ledger Management

*   **As a Finance Manager,** I want to manage a flexible Chart of Accounts **so that** I can structure our finances according to business needs.
    *   **AC 1:** I can create, edit, and deactivate GL accounts.
    *   **AC 2:** I can define account hierarchies (parent-child relationships).
    *   **AC 3:** The system prevents deletion of accounts with transaction history.

*   **As an Accountant,** I want to post manual journal entries **so that** I can make adjustments, accruals, and corrections.
    *   **AC 1:** A journal entry must have at least two lines (debit and credit).
    *   **AC 2:** The total debits must equal total credits for the entry to be valid.
    *   **AC 3:** I can attach supporting documents to a journal entry.

### Epic: Accounts Payable (AP)

*   **As an Accountant,** I want to manage vendor profiles **so that** I have a central record of supplier information.
    *   **AC 1:** I can add new vendors with details like name, address, tax ID, and payment terms.
    *   **AC 2:** I can view the complete transaction history for any vendor.
    *   **AC 3:** I can deactivate vendors we no longer work with.

*   **As an Accountant,** I want to process vendor bills by matching them to purchase orders **so that** we only pay for what we ordered and received.
    *   **AC 1:** The system allows matching a bill to one or more POs.
    *   **AC 2:** The system flags discrepancies in quantity or price between the bill and the PO.
    *   **AC 3:** I can approve a bill for payment once it's successfully matched.

### Epic: Accounts Receivable (AR)

*   **As a Finance Manager,** I want to generate and send customer invoices **so that** we can bill for products and services rendered.
    *   **AC 1:** Invoices can be created automatically from sales orders or manually.
    *   **AC 2:** Invoices can be sent to customers via email as a PDF attachment.
    *   **AC 3:** The system tracks the status of each invoice (draft, sent, paid, overdue).

*   **As an Accountant,** I want to record customer payments against invoices **so that** we can keep our receivables up to date.
    *   **AC 1:** I can apply a single payment to multiple invoices.
    *   **AC 2:** I can record partial payments.
    *   **AC 3:** The system automatically updates the invoice status and customer balance.

### Epic: Financial Reporting

*   **As a Business Owner,** I want to view a real-time Profit & Loss statement **so that** I can understand our company's financial performance at a glance.
    *   **AC 1:** The report can be filtered by date range (month, quarter, year).
    *   **AC 2:** The report shows revenue, cost of goods sold, gross profit, expenses, and net income.
    *   **AC 3:** I can drill down from a summary account to see the individual transactions.

---

## 2. Human Resources Management (HRM)

### Epic: Employee Data Management

*   **As an HR Manager,** I want to maintain a central profile for each employee **so that** all employee information is stored securely and accessibly.
    *   **AC 1:** The profile includes personal details, contact info, job role, salary, and emergency contacts.
    *   **AC 2:** I can upload and store documents like contracts and performance reviews.
    *   **AC 3:** Access to sensitive information is restricted based on user roles.

*   **As an Employee,** I want to access a self-service portal **so that** I can view and update my own personal information.
    *   **AC 1:** I can update my address and phone number.
    *   **AC 2:** I can view my payslips and tax documents.
    *   **AC 3:** I can view my remaining leave balance.

### Epic: Payroll & Time Tracking

*   **As an HR Manager,** I want to run the monthly payroll process automatically **so that** employees are paid accurately and on time.
    *   **AC 1:** The system calculates gross pay based on salary or approved time entries.
    *   **AC 2:** The system automatically deducts taxes and benefits.
    *   **AC 3:** I can review and approve the payroll run before payments are processed.

*   **As an Employee,** I want to submit my weekly timesheet for approval **so that** I can be compensated for the hours I've worked.
    *   **AC 1:** I can log hours worked against specific projects or tasks.
    *   **AC 2:** I can submit the timesheet to my manager at the end of the week.
    *   **AC 3:** I receive a notification when my timesheet is approved or rejected.

---

## 3. Supply Chain Management (SCM)

### Epic: Inventory Management

*   **As an Operations Manager,** I want to view real-time inventory levels for all products **so that** I can make informed decisions about stock control.
    *   **AC 1:** The system shows quantity on hand, quantity reserved, and quantity available.
    *   **AC 2:** I can view stock levels across multiple warehouse locations.
    *   **AC 3:** The system automatically updates stock levels when a sale or purchase is made.

*   **As a Warehouse Manager,** I want the system to automatically suggest reorder points for products **so that** we can avoid stockouts.
    *   **AC 1:** The system calculates a suggested reorder point based on historical sales data and lead times.
    *   **AC 2:** I receive an alert when a product's stock level drops to its reorder point.
    *   **AC 3:** I can manually override the suggested reorder point for any product.

### Epic: Procurement

*   **As an Operations Manager,** I want to create and send purchase orders to vendors **so that** I can formally request and track incoming inventory.
    *   **AC 1:** I can select a vendor and add products to the PO from the product catalog.
    *   **AC 2:** The system pre-fills the cost based on the vendor agreement.
    *   **AC 3:** The PO can be emailed to the vendor directly from the system.

*   **As a Warehouse Worker,** I want to record the receipt of goods against a purchase order **so that** inventory levels are updated accurately.
    *   **AC 1:** I can scan product barcodes to receive items.
    *   **AC 2:** I can record partial receipts if the full order doesn't arrive.
    *   **AC 3:** The system flags any discrepancies between the PO and the received quantity.

---

## 4. Sales & Customer Relationship Management (CRM)

### Epic: Customer Management

*   **As a Sales Manager,** I want a 360-degree view of each customer **so that** my team has all the context they need for interactions.
    *   **AC 1:** The customer view shows contact details, communication history, and past orders.
    *   **AC 2:** I can add notes and log calls, meetings, and emails.
    *   **AC 3:** The system links to all related sales orders, invoices, and support tickets.

### Epic: Sales Pipeline Management

*   **As a Salesperson,** I want to manage my sales opportunities in a pipeline view **so that** I can track my deals and prioritize my efforts.
    *   **AC 1:** I can move opportunities through customizable stages (e.g., Prospecting, Proposal, Negotiation).
    *   **AC 2:** Each opportunity tracks the potential deal value and estimated close date.
    *   **AC 3:** The system provides a forecast of my expected sales for the quarter.

*   **As a Salesperson,** I want to generate quotes for potential customers **so that** I can provide them with a formal offer.
    *   **AC 1:** I can add products from the catalog to the quote.
    *   **AC 2:** I can apply discounts and view the final price and margin.
    *   **AC 3:** I can convert a quote into a sales order with one click if the customer accepts.
