# Database Design

Comprehensive data modeling and database architecture for the ERP system's microservices.

## Database Architecture Overview

### Database-Per-Service Pattern
```mermaid
graph TB
    subgraph "Service Databases"
        FM_DB[(Financial Management<br/>financial_db<br/>PostgreSQL)]
        HR_DB[(Human Resources<br/>hr_db<br/>PostgreSQL)]
        SCM_DB[(Supply Chain<br/>scm_db<br/>PostgreSQL)]
        CRM_DB[(Customer Relations<br/>crm_db<br/>PostgreSQL)]
        MFG_DB[(Manufacturing<br/>manufacturing_db<br/>PostgreSQL)]
        PM_DB[(Project Management<br/>project_db<br/>PostgreSQL)]
    end
    
    subgraph "Shared Infrastructure"
        CACHE[(Redis Cache<br/>Session & Cache Data)]
        MQ[(Kafka<br/>Event Streaming)]
        BACKUP[(Backup Storage<br/>Point-in-time Recovery)]
    end
    
    subgraph "Cross-Service Data"
        EVENTS[Domain Events<br/>Eventual Consistency]
        APIS[Service APIs<br/>Synchronous Queries]
        SHARED[Shared Reference Data<br/>Users, Companies]
    end
    
    FM_DB --> EVENTS
    HR_DB --> EVENTS
    SCM_DB --> EVENTS
    CRM_DB --> EVENTS
    MFG_DB --> EVENTS
    PM_DB --> EVENTS
    
    EVENTS --> CACHE
    APIS --> CACHE
    
    FM_DB --> BACKUP
    HR_DB --> BACKUP
    SCM_DB --> BACKUP
    CRM_DB --> BACKUP
    MFG_DB --> BACKUP
    PM_DB --> BACKUP
```

## Data Consistency Patterns

### Eventual Consistency Model
```mermaid
sequenceDiagram
    participant CRM as CRM Service
    participant Events as Event Bus
    participant Finance as Finance Service
    participant HR as HR Service
    participant Cache as Redis Cache
    
    CRM->>Events: Customer Created Event
    Events->>Finance: Process Customer Event
    Events->>HR: Process Customer Event
    
    Finance->>Finance: Create Customer Record
    Finance->>Events: Customer Account Created
    
    HR->>HR: Create Employee Customer Link
    HR->>Cache: Update Customer Cache
    
    Events->>CRM: Confirmation Events
    CRM->>Cache: Update Customer Status
    
    Note over CRM,Cache: Eventual consistency achieved<br/>All services synchronized
```

### Saga Pattern for Distributed Transactions
```mermaid
flowchart TD
    A[Order Processing Saga] --> B[Reserve Inventory<br/>SCM Service]
    B --> C{Inventory Available?}
    C -->|Yes| D[Create Production Order<br/>Manufacturing Service]
    C -->|No| E[Cancel Order<br/>Compensating Action]
    D --> F{Production Scheduled?}
    F -->|Yes| G[Generate Invoice<br/>Finance Service]
    F -->|No| H[Release Inventory<br/>Compensating Action]
    G --> I{Invoice Created?}
    I -->|Yes| J[Complete Order<br/>Success]
    I -->|No| K[Cancel Production<br/>Compensating Action]
    
    E --> L[Order Cancelled]
    H --> M[Order Failed]
    K --> N[Order Failed]
    
    style J fill:#c8e6c9
    style L fill:#ffcdd2
    style M fill:#ffcdd2
    style N fill:#ffcdd2
```

## Core Data Models

### Financial Management Schema
```mermaid
erDiagram
    ACCOUNTS {
        uuid id PK
        string account_code UK
        string account_name
        enum account_type
        uuid parent_account_id FK
        int account_level
        enum normal_side
        decimal current_balance
        boolean is_active
        boolean allow_posting
        timestamp created_at
        timestamp updated_at
    }
    
    JOURNAL_ENTRIES {
        uuid id PK
        string entry_number UK
        date entry_date
        timestamp posting_date
        string description
        string reference
        string source_module
        decimal total_amount
        enum status
        boolean requires_approval
        uuid created_by FK
        timestamp created_at
        timestamp updated_at
    }
    
    JOURNAL_ENTRY_LINES {
        uuid id PK
        uuid journal_entry_id FK
        uuid account_id FK
        string description
        decimal debit_amount
        decimal credit_amount
        string department_code
        string cost_center
        uuid project_id FK
        timestamp created_at
    }
    
    VENDORS {
        uuid id PK
        string vendor_code UK
        string vendor_name
        string contact_name
        string email
        string phone
        json address
        enum payment_terms
        decimal credit_limit
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }
    
    ACCOUNTS ||--o{ ACCOUNTS : "parent_child"
    JOURNAL_ENTRIES ||--o{ JOURNAL_ENTRY_LINES : "has_lines"
    ACCOUNTS ||--o{ JOURNAL_ENTRY_LINES : "posts_to"
    VENDORS ||--o{ JOURNAL_ENTRIES : "vendor_invoices"
```

### Human Resources Schema
```mermaid
erDiagram
    EMPLOYEES {
        uuid id PK
        string employee_id UK
        string first_name
        string last_name
        string email UK
        string phone
        date hire_date
        date termination_date
        uuid department_id FK
        uuid position_id FK
        uuid manager_id FK
        enum employment_status
        enum employment_type
        decimal salary
        decimal hourly_rate
        enum pay_frequency
        json address
        json emergency_contacts
        timestamp created_at
        timestamp updated_at
    }
    
    DEPARTMENTS {
        uuid id PK
        string department_code UK
        string department_name
        string description
        uuid manager_id FK
        uuid parent_department_id FK
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }
    
    POSITIONS {
        uuid id PK
        string position_code UK
        string position_title
        string description
        uuid department_id FK
        decimal min_salary
        decimal max_salary
        json requirements
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }
    
    PAYROLL_RECORDS {
        uuid id PK
        uuid employee_id FK
        date pay_period_start
        date pay_period_end
        decimal regular_hours
        decimal overtime_hours
        decimal gross_pay
        decimal net_pay
        json tax_details
        json deduction_details
        enum status
        timestamp created_at
        timestamp updated_at
    }
    
    TIME_ENTRIES {
        uuid id PK
        uuid employee_id FK
        date entry_date
        timestamp clock_in
        timestamp clock_out
        int break_minutes
        decimal total_hours
        uuid project_id FK
        string notes
        enum status
        timestamp created_at
        timestamp updated_at
    }
    
    DEPARTMENTS ||--o{ EMPLOYEES : "belongs_to"
    POSITIONS ||--o{ EMPLOYEES : "assigned_to"
    EMPLOYEES ||--o{ EMPLOYEES : "manager_of"
    EMPLOYEES ||--o{ PAYROLL_RECORDS : "has_payroll"
    EMPLOYEES ||--o{ TIME_ENTRIES : "logs_time"
    DEPARTMENTS ||--o{ DEPARTMENTS : "parent_child"
```

### Supply Chain Management Schema
```mermaid
erDiagram
    PRODUCTS {
        uuid id PK
        string product_code UK
        string product_name
        string description
        enum product_type
        string unit_of_measure
        decimal standard_cost
        decimal list_price
        boolean is_active
        json specifications
        timestamp created_at
        timestamp updated_at
    }
    
    LOCATIONS {
        uuid id PK
        string location_code UK
        string location_name
        enum location_type
        json address
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }
    
    INVENTORY_ITEMS {
        uuid id PK
        uuid product_id FK
        uuid location_id FK
        int quantity_on_hand
        int quantity_reserved
        int quantity_available
        int reorder_point
        int maximum_stock
        decimal unit_cost
        timestamp last_received
        timestamp last_sold
        timestamp created_at
        timestamp updated_at
    }
    
    INVENTORY_MOVEMENTS {
        uuid id PK
        uuid product_id FK
        uuid location_id FK
        enum movement_type
        int quantity
        decimal unit_cost
        string reference_type
        string reference_id
        string notes
        uuid created_by FK
        timestamp created_at
    }
    
    SUPPLIERS {
        uuid id PK
        string supplier_code UK
        string supplier_name
        string contact_name
        string email
        string phone
        json address
        enum payment_terms
        json categories
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }
    
    PURCHASE_ORDERS {
        uuid id PK
        string po_number UK
        uuid supplier_id FK
        date order_date
        date expected_delivery
        enum status
        decimal total_amount
        string notes
        uuid created_by FK
        timestamp created_at
        timestamp updated_at
    }
    
    PURCHASE_ORDER_LINES {
        uuid id PK
        uuid purchase_order_id FK
        uuid product_id FK
        int quantity_ordered
        int quantity_received
        decimal unit_price
        decimal line_total
        string description
        timestamp created_at
    }
    
    PRODUCTS ||--o{ INVENTORY_ITEMS : "tracked_at"
    LOCATIONS ||--o{ INVENTORY_ITEMS : "stored_at"
    PRODUCTS ||--o{ INVENTORY_MOVEMENTS : "moves"
    LOCATIONS ||--o{ INVENTORY_MOVEMENTS : "location"
    SUPPLIERS ||--o{ PURCHASE_ORDERS : "supplies"
    PURCHASE_ORDERS ||--o{ PURCHASE_ORDER_LINES : "has_lines"
    PRODUCTS ||--o{ PURCHASE_ORDER_LINES : "orders"
```

## Indexing Strategy

### Primary Indexes
```mermaid
graph TB
    subgraph "Clustered Indexes (Physical Ordering)"
        CI1[Primary Keys<br/>UUID columns<br/>Unique identification]
        CI2[Time-series Data<br/>created_at columns<br/>Chronological access]
        CI3[Sequential Access<br/>Order numbers<br/>Natural progression]
    end
    
    subgraph "Unique Indexes (Business Keys)"
        UI1[Business Codes<br/>account_code, employee_id<br/>Human-readable keys]
        UI2[Email Addresses<br/>User identification<br/>Login credentials]
        UI3[External References<br/>Invoice numbers<br/>Document tracking]
    end
    
    subgraph "Foreign Key Indexes"
        FI1[Reference Columns<br/>department_id, product_id<br/>Join optimization]
        FI2[Hierarchical References<br/>parent_account_id<br/>Tree navigation]
        FI3[Cross-table References<br/>Service boundaries<br/>API lookups]
    end
    
    subgraph "Composite Indexes"
        CMI1[Filter + Sort<br/>status + created_at<br/>List queries]
        CMI2[Partition Keys<br/>tenant_id + entity_id<br/>Multi-tenancy]
        CMI3[Covering Indexes<br/>Include frequently accessed<br/>Index-only queries]
    end
```

### Query Optimization Patterns
```mermaid
flowchart TD
    A[Query Request] --> B{Query Type}
    
    B -->|Point Lookup| C[Use Primary Key Index<br/>O(1) access time]
    B -->|Range Query| D[Use Composite Index<br/>Optimized range scan]
    B -->|Join Query| E[Use Foreign Key Indexes<br/>Nested loop optimization]
    B -->|Aggregation| F[Use Covering Index<br/>Index-only scan]
    
    C --> G[Return Single Row<br/>Sub-millisecond response]
    D --> H[Return Range Result<br/>Paginated response]
    E --> I[Return Joined Data<br/>Denormalized view]
    F --> J[Return Aggregated Data<br/>Pre-calculated results]
    
    subgraph "Performance Targets"
        PT1[Point Queries: < 1ms]
        PT2[Range Queries: < 10ms]
        PT3[Join Queries: < 50ms]
        PT4[Aggregations: < 100ms]
    end
    
    G -.-> PT1
    H -.-> PT2
    I -.-> PT3
    J -.-> PT4
```

## Data Partitioning

### Horizontal Partitioning Strategy
```mermaid
graph TB
    subgraph "Time-Based Partitioning"
        TBP1[Transaction Tables<br/>Partition by month<br/>journal_entries_2024_01]
        TBP2[Log Tables<br/>Partition by day<br/>audit_logs_2024_03_15]
        TBP3[Historical Data<br/>Archive old partitions<br/>Cold storage migration]
    end
    
    subgraph "Range-Based Partitioning"
        RBP1[Account Ranges<br/>1000-1999, 2000-2999<br/>Balance Sheet grouping]
        RBP2[Employee ID Ranges<br/>000-999, 1000-1999<br/>Department grouping]
        RBP3[Geographic Ranges<br/>US, EU, APAC<br/>Regional compliance]
    end
    
    subgraph "Hash-Based Partitioning"
        HBP1[Customer Data<br/>Hash customer ID<br/>Even distribution]
        HBP2[Product Catalog<br/>Hash product code<br/>Load balancing]
        HBP3[Order Processing<br/>Hash order number<br/>Parallel processing]
    end
    
    subgraph "Benefits"
        B1[Query Performance<br/>Partition elimination<br/>Reduced scan time]
        B2[Maintenance Operations<br/>Parallel execution<br/>Faster backups]
        B3[Storage Optimization<br/>Compression by partition<br/>Archival strategies]
        B4[Scalability<br/>Add/drop partitions<br/>Dynamic growth]
    end
    
    TBP1 --> B1
    RBP1 --> B2
    HBP1 --> B3
    TBP3 --> B4
```

## Backup and Recovery

### Backup Strategy
```mermaid
graph TB
    subgraph "Backup Types"
        BT1[Full Backup<br/>Complete database<br/>Weekly schedule]
        BT2[Incremental Backup<br/>Changed data only<br/>Daily schedule]
        BT3[Transaction Log Backup<br/>WAL segments<br/>15-minute intervals]
        BT4[Snapshot Backup<br/>Point-in-time copy<br/>Before major changes]
    end
    
    subgraph "Storage Locations"
        SL1[Local Storage<br/>Fast recovery<br/>Same datacenter]
        SL2[Remote Storage<br/>Disaster recovery<br/>Geographic separation]
        SL3[Cloud Storage<br/>Long-term retention<br/>Cost-effective archival]
        SL4[Tape Storage<br/>Compliance archive<br/>Air-gapped security]
    end
    
    subgraph "Recovery Scenarios"
        RS1[Point-in-time Recovery<br/>Restore to specific moment<br/>Transaction precision]
        RS2[Database Corruption<br/>Full restoration<br/>Last known good state]
        RS3[Disaster Recovery<br/>Complete site failure<br/>Geographic failover]
        RS4[Partial Recovery<br/>Table-level restore<br/>Selective restoration]
    end
    
    BT1 --> SL1
    BT2 --> SL2
    BT3 --> SL3
    BT4 --> SL4
    
    SL1 --> RS1
    SL2 --> RS2
    SL3 --> RS3
    SL4 --> RS4
```

### Recovery Time Objectives
```mermaid
gantt
    title Database Recovery Time Objectives
    dateFormat X
    axisFormat %M:%S
    
    section Critical Systems
    Financial DB    :crit, financial, 0, 5
    Customer DB     :crit, customer, 0, 10
    
    section Business Systems
    HR DB          :active, hr, 0, 15
    SCM DB         :active, scm, 0, 20
    
    section Supporting Systems
    Project DB     :project, 0, 30
    Manufacturing DB :manufacturing, 0, 30
    
    section Recovery Targets
    RTO Target     :milestone, rto, 15, 0
    Maximum RTO    :milestone, max_rto, 30, 0
```

## Performance Monitoring

### Database Metrics
```mermaid
graph TB
    subgraph "Performance Metrics"
        PM1[Query Response Time<br/>Average: < 10ms<br/>95th percentile: < 50ms]
        PM2[Connection Pool Usage<br/>Active connections<br/>Pool efficiency]
        PM3[Transaction Throughput<br/>TPS measurement<br/>Peak load handling]
        PM4[Index Usage Statistics<br/>Index effectiveness<br/>Query optimization]
    end
    
    subgraph "Resource Metrics"
        RM1[CPU Utilization<br/>Database server load<br/>Query processing]
        RM2[Memory Usage<br/>Buffer cache hit ratio<br/>Working set size]
        RM3[Disk I/O Patterns<br/>Read/write operations<br/>Storage performance]
        RM4[Network Bandwidth<br/>Data transfer rates<br/>Connection overhead]
    end
    
    subgraph "Health Indicators"
        HI1[Slow Query Log<br/>Queries > 1 second<br/>Optimization candidates]
        HI2[Lock Contention<br/>Blocking queries<br/>Concurrency issues]
        HI3[Replication Lag<br/>Standby synchronization<br/>Data consistency]
        HI4[Backup Success Rate<br/>Backup completion<br/>Recovery readiness]
    end
    
    subgraph "Alerting Thresholds"
        AT1[Critical: > 100ms avg<br/>Query performance degradation]
        AT2[Warning: > 80% pool<br/>Connection exhaustion risk]
        AT3[Critical: > 90% CPU<br/>Resource constraint]
        AT4[Warning: > 1 min lag<br/>Replication delay]
    end
    
    PM1 --> AT1
    PM2 --> AT2
    RM1 --> AT3
    HI3 --> AT4
```

## Data Security and Compliance

### Encryption at Rest and Transit
```mermaid
graph TB
    subgraph "Encryption at Rest"
        EAR1[Database Files<br/>AES-256 encryption<br/>Transparent encryption]
        EAR2[Backup Files<br/>Encrypted backups<br/>Key rotation]
        EAR3[Log Files<br/>Encrypted WAL<br/>Audit trail protection]
        EAR4[Temporary Files<br/>Sort/join operations<br/>Memory protection]
    end
    
    subgraph "Encryption in Transit"
        EIT1[Client Connections<br/>TLS 1.2+ mandatory<br/>Certificate validation]
        EIT2[Replication Streams<br/>Encrypted WAL shipping<br/>Standby security]
        EIT3[Backup Transfers<br/>Secure file transfer<br/>Network protection]
        EIT4[Service Communication<br/>Inter-service calls<br/>mTLS authentication]
    end
    
    subgraph "Key Management"
        KM1[Key Generation<br/>Hardware security modules<br/>Cryptographic standards]
        KM2[Key Rotation<br/>Automatic rotation<br/>Zero-downtime updates]
        KM3[Key Storage<br/>Secure key vault<br/>Access control]
        KM4[Key Recovery<br/>Escrow procedures<br/>Business continuity]
    end
    
    subgraph "Compliance Features"
        CF1[Data Masking<br/>PII protection<br/>Non-production environments]
        CF2[Audit Logging<br/>Access tracking<br/>Compliance reporting]
        CF3[Data Retention<br/>Automated purging<br/>Legal requirements]
        CF4[Access Controls<br/>Role-based permissions<br/>Principle of least privilege]
    end
    
    EAR1 --> KM1
    EIT1 --> KM2
    EAR2 --> KM3
    EIT2 --> KM4
    
    KM1 --> CF1
    KM2 --> CF2
    KM3 --> CF3
    KM4 --> CF4
```

## Next Steps

- [Event-Driven Architecture](event-architecture.md) - Inter-service communication patterns
- [Security Architecture](security-architecture.md) - Authentication and authorization
- [Performance Architecture](performance-architecture.md) - Caching and optimization strategies