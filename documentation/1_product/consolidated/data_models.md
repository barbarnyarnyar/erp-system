# Data Models

## Overview

This document defines the essential data models for the ERP system, designed around the core principles of accounting and HR. The models follow double-entry accounting standards and are optimized for microservices architecture with complete audit trails.

## Design Principles

### 1. **Core Model Foundation**
- **ACCOUNTS** - Where money is categorized (Chart of Accounts)
- **TRANSACTIONS** - What financial events occurred (Journal Entries)
- **ENTITIES** - Who we do business with (Vendors and Customers)
- **EMPLOYEES** - The people who work for the company

### 2. **Double-Entry Accounting**
- Every transaction must balance (Debits = Credits)
- Complete audit trail for all financial movements
- Real-time account balance calculations
- Financial statement integrity guaranteed

### 3. **Microservices-Friendly**
- UUID primary keys for distributed systems
- Event-driven data synchronization
- Minimal cross-service dependencies
- Independent deployment capabilities

### 4. **Financial Precision**
- Decimal data types for all monetary amounts
- Currency-aware calculations
- Rounding and precision controls
- Exchange rate handling

### 5. **Audit & Compliance**
- Soft deletes with audit trails
- Change tracking for all sensitive data
- GDPR and data retention compliance ready

---

## Abstract Business Models (The Conceptual Foundation)

### 1. Account Model (WHERE money is categorized)

```go
type Account struct {
    ID            uuid.UUID       `json:"id"`
    Code          string          `json:"code"`          // "1000", "2000", etc.
    Name          string          `json:"name"`          // "Cash", "Accounts Payable"
    Type          AccountType     `json:"type"`          // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
    Balance       decimal.Decimal `json:"balance"`       // Current balance
    NormalSide    DebitCredit     `json:"normal_side"`   // DEBIT or CREDIT
}
```

### 2. Transaction Model (WHAT financial events occurred)

```go
type Transaction struct {
    ID          uuid.UUID       `json:"id"`
    Date        time.Time       `json:"date"`
    Reference   string          `json:"reference"`
    Description string          `json:"description"`
    Lines       []TransactionLine `json:"lines"`     // Double-entry lines
    Status      TransactionStatus `json:"status"`
}
```

### 3. Entity Model (WHO we do business with)

```go
type Entity struct {
    ID      uuid.UUID       `json:"id"`
    Name    string          `json:"name"`
    Type    EntityType      `json:"type"`      // VENDOR or CUSTOMER
    Email   string          `json:"email"`
    Address string          `json:"address"`
    Balance decimal.Decimal `json:"balance"`   // What we owe them or they owe us
}
```

### 4. Employee Model (WHO works for us)

```go
type Employee struct {
    ID            uuid.UUID       `json:"id"`
    EmployeeID    string          `json:"employee_id"`
    FirstName     string          `json:"first_name"`
    LastName      string          `json:"last_name"`
    Email         string          `json:"email"`
    HireDate      time.Time       `json:"hire_date"`
    EmploymentStatus string       `json:"employment_status"`
}
```

---

## Complete Database Models (The Implementation Reality)

### Core Foundation Tables

#### 1. ACCOUNTS (Chart of Accounts)

```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_code VARCHAR(20) UNIQUE NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(20) NOT NULL 
        CHECK (account_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),
    parent_account_id UUID REFERENCES accounts(id),
    normal_side VARCHAR(6) NOT NULL CHECK (normal_side IN ('DEBIT', 'CREDIT')),
    current_balance DECIMAL(15,2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. JOURNAL_ENTRIES (Transaction Headers)

```sql
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_number VARCHAR(50) UNIQUE NOT NULL,
    entry_date DATE NOT NULL,
    posting_date DATE,
    description TEXT NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'DRAFT' 
        CHECK (status IN ('DRAFT', 'POSTED', 'REVERSED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. JOURNAL_LINES (Transaction Details - Double-Entry)

```sql
CREATE TABLE journal_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    debit_amount DECIMAL(15,2) DEFAULT 0.00,
    credit_amount DECIMAL(15,2) DEFAULT 0.00,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 4. EMPLOYEES (Employee Master Data)

```sql
CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    hire_date DATE NOT NULL,
    employment_status VARCHAR(20) NOT NULL DEFAULT 'active',
    department_id UUID REFERENCES departments(id),
    position_id UUID REFERENCES positions(id),
    manager_id UUID REFERENCES employees(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL -- Soft delete
);
```

### Entity Management Tables

#### 5. VENDORS (Supplier Master Data)

```sql
CREATE TABLE vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_code VARCHAR(20) UNIQUE NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    current_balance DECIMAL(15,2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 6. CUSTOMERS (Customer Master Data)

```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_code VARCHAR(20) UNIQUE NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    current_balance DECIMAL(15,2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Operational Transaction Tables

#### 7. AP_INVOICES (Accounts Payable Invoices)

```sql
CREATE TABLE ap_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    vendor_id UUID NOT NULL REFERENCES vendors(id),
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'PENDING' 
        CHECK (status IN ('PENDING', 'APPROVED', 'PAID', 'CANCELLED', 'ON_HOLD')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 8. AR_INVOICES (Accounts Receivable Invoices)

```sql
CREATE TABLE ar_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'PENDING' 
        CHECK (status IN ('PENDING', 'SENT', 'PAID', 'OVERDUE', 'CANCELLED', 'DISPUTED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 9. PAYMENTS (Unified Payment Processing)

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_number VARCHAR(50) UNIQUE NOT NULL,
    payment_date DATE NOT NULL,
    payment_type VARCHAR(20) NOT NULL 
        CHECK (payment_type IN ('AP_PAYMENT', 'AR_RECEIPT', 'GENERAL')),
    amount DECIMAL(15,2) NOT NULL,
    payment_method VARCHAR(20) NOT NULL 
        CHECK (payment_method IN ('CHECK', 'ACH', 'WIRE', 'CREDIT_CARD', 'CASH')),
    status VARCHAR(20) DEFAULT 'PENDING' 
        CHECK (status IN ('PENDING', 'PROCESSED', 'CLEARED', 'FAILED', 'CANCELLED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 10. TIME_ENTRIES (Time & Attendance)

```sql
CREATE TABLE time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    entry_date DATE NOT NULL,
    total_hours DECIMAL(4,2),
    approval_status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 11. LEAVE_REQUESTS (Time Off Management)

```sql
CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    leave_type VARCHAR(20) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_days DECIMAL(3,1),
    request_status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Supporting Tables

#### 12. DEPARTMENTS (Organizational Structure)

```sql
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    department_code VARCHAR(20) UNIQUE NOT NULL,
    department_name VARCHAR(100) NOT NULL,
    parent_department_id UUID REFERENCES departments(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 13. POSITIONS (Job Roles & Compensation)

```sql
CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    position_code VARCHAR(20) UNIQUE NOT NULL,
    position_title VARCHAR(100) NOT NULL,
    min_salary DECIMAL(12,2),
    max_salary DECIMAL(12,2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 14. AUDIT_LOG (Complete Change Tracking)

```sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE', 'SELECT')),
    old_values JSONB,
    new_values JSONB,
    user_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
