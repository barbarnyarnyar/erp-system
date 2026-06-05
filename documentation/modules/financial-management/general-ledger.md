# General Ledger

Central repository for all financial transactions with hierarchical chart of accounts and real-time balance tracking.

## Chart of Accounts Structure

### Account Hierarchy Design
```mermaid
graph TD
    ROOT[Chart of Accounts<br/>Company Level]
    
    subgraph "Level 1 - Account Types"
        A1[1000 - Assets]
        A2[2000 - Liabilities] 
        A3[3000 - Equity]
        A4[4000 - Revenue]
        A5[5000 - Expenses]
    end
    
    subgraph "Level 2 - Account Categories"
        A11[1100 - Current Assets]
        A12[1400 - Fixed Assets]
        A21[2100 - Current Liabilities]
        A22[2400 - Long-term Debt]
    end
    
    subgraph "Level 3 - Account Subcategories"
        A111[1110 - Cash Accounts]
        A112[1120 - Receivables]
        A113[1130 - Inventory]
        A211[2110 - Payables]
        A212[2120 - Accrued Liabilities]
    end
    
    subgraph "Level 4 - Detailed Accounts"
        A1111[1111 - Operating Cash]
        A1112[1112 - Petty Cash]
        A1121[1121 - Trade Receivables]
        A1122[1122 - Employee Advances]
    end
    
    ROOT --> A1
    ROOT --> A2
    ROOT --> A3
    ROOT --> A4
    ROOT --> A5
    
    A1 --> A11
    A1 --> A12
    A2 --> A21
    A2 --> A22
    
    A11 --> A111
    A11 --> A112
    A11 --> A113
    A21 --> A211
    A21 --> A212
    
    A111 --> A1111
    A111 --> A1112
    A112 --> A1121
    A112 --> A1122
```

## Standard Chart of Accounts

### Asset Accounts (1000-1999)
```mermaid
graph TB
    subgraph "Current Assets (1000-1399)"
        CA1[1000-1099<br/>💰 Cash & Cash Equivalents<br/>- Operating Cash<br/>- Savings Accounts<br/>- Money Market<br/>- Petty Cash]
        
        CA2[1100-1199<br/>📋 Accounts Receivable<br/>- Trade Receivables<br/>- Employee Advances<br/>- Other Receivables<br/>- Allowance for Doubtful Debts]
        
        CA3[1200-1299<br/>📦 Inventory<br/>- Raw Materials<br/>- Work in Progress<br/>- Finished Goods<br/>- Supplies]
        
        CA4[1300-1399<br/>⏰ Prepaid Expenses<br/>- Prepaid Insurance<br/>- Prepaid Rent<br/>- Prepaid Software<br/>- Deposits]
    end
    
    subgraph "Fixed Assets (1400-1999)"
        FA1[1400-1499<br/>🏢 Property, Plant & Equipment<br/>- Land<br/>- Buildings<br/>- Equipment<br/>- Vehicles]
        
        FA2[1500-1599<br/>📉 Accumulated Depreciation<br/>- Buildings Depreciation<br/>- Equipment Depreciation<br/>- Vehicles Depreciation]
        
        FA3[1600-1699<br/>💡 Intangible Assets<br/>- Software<br/>- Patents<br/>- Trademarks<br/>- Goodwill]
        
        FA4[1700-1999<br/>📈 Investments<br/>- Long-term Investments<br/>- Investment Securities<br/>- Subsidiary Investments]
    end
```

### Liability Accounts (2000-2999)
```mermaid
graph TB
    subgraph "Current Liabilities (2000-2399)"
        CL1[2000-2099<br/>💳 Accounts Payable<br/>- Trade Payables<br/>- Accrued Expenses<br/>- Employee Payables<br/>- Tax Payables]
        
        CL2[2100-2199<br/>💼 Payroll Liabilities<br/>- Salaries Payable<br/>- Payroll Tax Payable<br/>- Benefits Payable<br/>- Vacation Accrual]
        
        CL3[2200-2299<br/>🏦 Short-term Debt<br/>- Bank Loans<br/>- Credit Lines<br/>- Current Portion LTD<br/>- Notes Payable]
        
        CL4[2300-2399<br/>📊 Other Current Liabilities<br/>- Customer Deposits<br/>- Unearned Revenue<br/>- Warranty Reserves<br/>- Accrued Interest]
    end
    
    subgraph "Long-term Liabilities (2400-2999)"
        LL1[2400-2499<br/>🏛️ Long-term Debt<br/>- Term Loans<br/>- Mortgages<br/>- Bonds Payable<br/>- Equipment Financing]
        
        LL2[2500-2599<br/>📋 Lease Liabilities<br/>- Operating Leases<br/>- Finance Leases<br/>- Equipment Leases<br/>- Property Leases]
        
        LL3[2600-2699<br/>👥 Employee Benefits<br/>- Pension Obligations<br/>- OPEB Liabilities<br/>- Deferred Compensation<br/>- Stock Options]
        
        LL4[2700-2999<br/>⚖️ Other Long-term Liabilities<br/>- Deferred Tax<br/>- Asset Retirement<br/>- Environmental<br/>- Legal Reserves]
    end
```

## Account Management Features

### Account Creation Workflow
```mermaid
flowchart TD
    A[Request New Account] --> B[Validate Account Code<br/>Check Uniqueness]
    B --> C{Code Available?}
    C -->|No| D[Generate Alternative<br/>Suggest Similar Codes]
    C -->|Yes| E[Validate Account Type<br/>Check Business Rules]
    E --> F{Type Valid?}
    F -->|No| G[Return Error<br/>Invalid Account Type]
    F -->|Yes| H[Set Account Level<br/>Based on Parent]
    H --> I[Configure Properties<br/>Normal Side, Posting Rules]
    I --> J[Create Account<br/>Generate ID]
    J --> K[Update Hierarchy<br/>Parent-Child Links]
    K --> L[Publish Event<br/>Account Created]
    L --> M[Account Active<br/>Ready for Use]
    
    D --> A
    G --> A
    
    style M fill:#c8e6c9
    style G fill:#ffcdd2
    style D fill:#fff3e0
```

### Account Properties and Rules
```mermaid
graph TD
    subgraph "Account Properties"
        AP1[Account Code<br/>Unique Identifier<br/>Format: ####]
        AP2[Account Name<br/>Descriptive Title<br/>Max 100 chars]
        AP3[Account Type<br/>ASSET/LIABILITY<br/>EQUITY/REVENUE/EXPENSE]
        AP4[Normal Side<br/>DEBIT/CREDIT<br/>Natural Balance]
        AP5[Parent Account<br/>Hierarchical Structure<br/>Optional Reference]
    end
    
    subgraph "Business Rules"
        BR1[Allow Posting<br/>Can record transactions<br/>Leaf accounts only]
        BR2[Require Department<br/>Departmental tracking<br/>Cost center allocation]
        BR3[Require Project<br/>Project tracking<br/>Job costing]
        BR4[Active Status<br/>Available for use<br/>Can be deactivated]
        BR5[Currency Enabled<br/>Multi-currency support<br/>Foreign transactions]
    end
    
    subgraph "Control Features"
        CF1[Budget Control<br/>Spending limits<br/>Approval thresholds]
        CF2[Approval Required<br/>Transaction approval<br/>Amount-based rules]
        CF3[Statistical Account<br/>Quantity tracking<br/>Non-monetary units]
        CF4[Tax Account<br/>Tax calculations<br/>Compliance reporting]
    end
    
    AP1 -.-> BR1
    AP3 -.-> BR2
    AP4 -.-> CF1
    AP5 -.-> CF2
```

## Balance Calculation and Tracking

### Real-time Balance Updates
```mermaid
sequenceDiagram
    participant JE as Journal Entry
    participant GL as General Ledger
    participant BAL as Balance Calculator
    participant CACHE as Cache Layer
    participant DB as Database
    
    JE->>GL: Post Transaction
    GL->>BAL: Calculate Impact
    BAL->>BAL: Determine Debit/Credit Effect
    BAL->>DB: Update Account Balances
    DB->>BAL: Confirm Update
    BAL->>CACHE: Update Balance Cache
    CACHE->>BAL: Confirm Cache Update
    BAL->>GL: Balance Updated
    GL->>JE: Transaction Posted
    
    Note over BAL: Real-time balance calculation<br/>Immediate availability
    Note over CACHE: High-performance access<br/>Sub-second response
```

### Balance History Tracking
```mermaid
graph TD
    subgraph "Balance Types"
        BT1[Current Balance<br/>Real-time Amount<br/>Latest Transaction]
        BT2[Beginning Balance<br/>Period Start<br/>Opening Amount]
        BT3[Period Activity<br/>Debits & Credits<br/>Net Movement]
        BT4[Ending Balance<br/>Period End<br/>Closing Amount]
    end
    
    subgraph "Historical Tracking"
        HT1[Daily Balances<br/>End-of-day snapshots<br/>Audit trail]
        HT2[Monthly Balances<br/>Period-end closing<br/>Reporting basis]
        HT3[Yearly Balances<br/>Annual closing<br/>Comparative analysis]
        HT4[Transaction History<br/>Complete audit trail<br/>Detailed movements]
    end
    
    subgraph "Reporting Views"
        RV1[Trial Balance<br/>All account balances<br/>Debit/Credit totals]
        RV2[Balance Sheet<br/>Asset/Liability/Equity<br/>Financial position]
        RV3[Account Analysis<br/>Detailed movements<br/>Variance analysis]
        RV4[Comparative Reports<br/>Period-over-period<br/>Trend analysis]
    end
    
    BT1 --> HT1
    BT2 --> HT2
    BT3 --> HT3
    BT4 --> HT4
    
    HT1 --> RV1
    HT2 --> RV2
    HT3 --> RV3
    HT4 --> RV4
```

## Multi-Dimensional Analysis

### Cost Center Integration
```mermaid
graph TB
    subgraph "Account Dimensions"
        AD1[Account Code<br/>Primary Classification]
        AD2[Department<br/>Organizational Unit]
        AD3[Cost Center<br/>Responsibility Center]
        AD4[Project<br/>Job/Contract]
        AD5[Location<br/>Geographic Site]
    end
    
    subgraph "Analysis Views"
        AV1[By Department<br/>Departmental P&L<br/>Resource allocation]
        AV2[By Cost Center<br/>Cost control<br/>Budget variance]
        AV3[By Project<br/>Project profitability<br/>Job costing]
        AV4[By Location<br/>Geographic performance<br/>Site analysis]
    end
    
    subgraph "Reporting Combinations"
        RC1[Account + Department<br/>Departmental account detail]
        RC2[Account + Project<br/>Project account analysis]
        RC3[Department + Cost Center<br/>Detailed cost tracking]
        RC4[All Dimensions<br/>Complete analysis cube]
    end
    
    AD1 --> AV1
    AD2 --> AV2
    AD3 --> AV3
    AD4 --> AV4
    AD5 --> AV1
    
    AV1 --> RC1
    AV2 --> RC2
    AV3 --> RC3
    AV4 --> RC4
```

## Account Reconciliation

### Reconciliation Process
```mermaid
flowchart TD
    A[Select Account<br/>Choose period] --> B[Load GL Balance<br/>System balance]
    B --> C[Load External Source<br/>Bank, subsidiary, etc.]
    C --> D[Compare Balances<br/>Identify differences]
    D --> E{Balances Match?}
    E -->|Yes| F[Mark Reconciled<br/>Complete process]
    E -->|No| G[Identify Differences<br/>Analyze variances]
    G --> H[Research Items<br/>Find explanations]
    H --> I[Make Adjustments<br/>Create journal entries]
    I --> J[Re-run Comparison<br/>Verify balance]
    J --> K{Now Match?}
    K -->|Yes| F
    K -->|No| L[Document Exception<br/>Escalate if needed]
    L --> M[Management Review<br/>Approve variance]
    M --> F
    
    style F fill:#c8e6c9
    style L fill:#fff3e0
    style M fill:#ffecb3
```

### Reconciliation Types
```mermaid
graph TD
    subgraph "Bank Reconciliation"
        BR1[Cash Accounts<br/>vs Bank Statements]
        BR2[Outstanding Checks<br/>In-transit deposits]
        BR3[Bank Charges<br/>Interest income]
        BR4[Automated Matching<br/>Transaction recognition]
    end
    
    subgraph "Intercompany Reconciliation"
        IR1[Subsidiary Balances<br/>vs Parent records]
        IR2[Intercompany Transactions<br/>Reciprocal entries]
        IR3[Currency Translation<br/>Foreign subsidiaries]
        IR4[Elimination Entries<br/>Consolidated reporting]
    end
    
    subgraph "Sub-Ledger Reconciliation"
        SR1[AR Sub-ledger<br/>vs GL control account]
        SR2[AP Sub-ledger<br/>vs GL control account]
        SR3[Fixed Assets<br/>vs GL balances]
        SR4[Inventory<br/>vs GL valuation]
    end
    
    subgraph "Balance Sheet Reconciliation"
        BSR1[Asset Accounts<br/>Supporting details]
        BSR2[Liability Accounts<br/>Aging analysis]
        BSR3[Equity Accounts<br/>Movement analysis]
        BSR4[Accrual Analysis<br/>Cut-off testing]
    end
```

## Performance Optimization

### Indexing Strategy
```mermaid
graph TB
    subgraph "Primary Indexes"
        PI1[Account ID<br/>Clustered index<br/>Unique identifier]
        PI2[Account Code<br/>Unique index<br/>Business key]
        PI3[Account Type<br/>Non-unique<br/>Filtering]
    end
    
    subgraph "Secondary Indexes"
        SI1[Parent Account<br/>Hierarchy queries<br/>Tree navigation]
        SI2[Department<br/>Dimensional analysis<br/>Cost center reporting]
        SI3[Active Status<br/>Active accounts<br/>Current operations]
        SI4[Created Date<br/>Audit queries<br/>Historical analysis]
    end
    
    subgraph "Composite Indexes"
        CI1[Type + Active<br/>Active accounts by type<br/>Common filter combination]
        CI2[Parent + Level<br/>Hierarchy navigation<br/>Tree operations]
        CI3[Code + Department<br/>Departmental accounts<br/>Analysis queries]
    end
    
    PI1 -.-> SI1
    PI2 -.-> CI1
    PI3 -.-> CI2
    SI2 -.-> CI3
```

## API Examples

### Create Account
```http
POST /api/v1/finance/accounts
Content-Type: application/json
Authorization: Bearer <token>

{
  "account_code": "1150",
  "account_name": "Accounts Receivable - Trade",
  "account_type": "ASSET",
  "normal_side": "DEBIT",
  "parent_account_id": "acc-1100",
  "allow_posting": true,
  "require_department": true,
  "active": true,
  "description": "Customer receivables from normal business operations"
}
```

### Get Account Balance
```http
GET /api/v1/finance/accounts/acc-1150/balance
Authorization: Bearer <token>

Query Parameters:
- as_of_date: 2024-03-31 (optional, defaults to current date)
- include_pending: true (include unposted transactions)
- currency: USD (for multi-currency accounts)
```

### Account Hierarchy
```http
GET /api/v1/finance/accounts/hierarchy
Authorization: Bearer <token>

Query Parameters:
- account_type: ASSET (filter by account type)
- active_only: true (exclude inactive accounts)
- max_level: 4 (limit hierarchy depth)
- include_balances: true (include current balances)
```

## Next Steps

- [Journal Entries](journal-entries.md) - Recording transactions against accounts
- [Financial Reporting](financial-reporting.md) - Using accounts in financial statements
- [Database Schema](database-schema.md) - Technical implementation details