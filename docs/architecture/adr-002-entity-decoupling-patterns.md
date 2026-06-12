# ADR-002: Entity Decoupling Patterns

## Status
Accepted

## Context
As our ERP scales to 9 core business modules (10 total microservices), avoiding tight coupling is critical to maintaining a scalable, resilient system. Without strict boundaries, the application risks falling into a "distributed monolith" where database schemas, transaction locks, and sync API dependencies propagate throughout the services.

Specifically, we found overlaps where:
1. **User / Employee details** are needed across Auth, HR, Project Management, and Manufacturing (for labor/incident logging).
2. **Product / SKU definitions** from SCM are referenced in Manufacturing BOMs and CRM Sales Orders.
3. **Financial details** (Invoices, Budgets) are coupled to project spending and customer credits.

We need to formalize patterns for cross-service entity references, data caching/replication, and eventual invariant validation.

## Decision

We enforce three main architectural patterns across all services to ensure low coupling and high coherence:

### 1. No Shared Databases
* Under no circumstances shall one microservice query or write to another service's database tables. 
* Direct database foreign key constraints crossing service boundaries are strictly forbidden. All cross-service references must be represented as primitive values (e.g., string or UUID fields such as `EmployeeID`, `ProductID`, or `BudgetID`).

### 2. Asynchronous Replication & Cache
* When a service needs details of an entity owned by another service (for instance, the name of a product on a sales order line), the consumer service must either:
  * Replicate and cache the minimal required fields locally by subscribing to Kafka change events (e.g., `scm.inventory.updated`, `hr.employee.created`).
  * Or fetch them dynamically at the API Gateway level (or via backend-for-frontend parallel queries) to populate user interface elements.
* Services must handle the eventual consistency of this replicated data gracefully.

```
          [ Auth ]      [ CRM ]      [ FM ]
             │             │           │
             ▼             ▼           ▼
        EmployeeID    SalesOrderID  BudgetID
             │             │           │
             ▼             ▼           ▼
    [ HR Service ] ──► [ PM Service ] ◄── [ SCM Service ]
                            │
                            ▼
                        ProductID
```

### 3. Eventual Invariant Validation
* Business constraints crossing service boundaries must be validated asynchronously. For example, if a project task allocation must not exceed the general ledger budget:
  1. The Project Management (PM) service registers the local project spending.
  2. The PM service publishes a spending update event.
  3. The Financial Management (FM) service consumes this event and validates it against the current budget ceiling.
  4. If the budget is exceeded, the FM service publishes a warning or compensation event (`fin.budget.exceeded`), triggering a remediation workflow (e.g., suspending project tasks or alerting management).
* Blocking synchronous RPC calls to check invariants during transaction processing should be avoided to prevent system-wide latency and single-point-of-failure vulnerabilities.

## Consequences
* **Service Autonomy**: Services can deploy and schema-migrate independently because database dependencies are reduced to event contracts.
* **Network Partition Tolerance**: If a dependency service is down, dependent services can still run, read from their local cache, and capture requests to process later.
* **Eventual Consistency**: Developers must design user interfaces and workflow engines to tolerate small delay windows (eventual consistency) in data synchronization.
