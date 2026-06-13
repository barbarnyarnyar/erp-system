# Customer Relationship Management Module

Customer accounts, lead management, opportunity pipeline, sales orders, quotes, service tickets, campaigns, and price lists. Port **8002** (docker-compose: 8002).

## Module Overview

The CRM Service is architected into two distinct, decoupled domain packages to isolate transactional core business operations from operational and marketing functions, ensuring zero direct dependency coupling ($C_e = 0$) across namespaces:

1. **`erp.crm.core`**: The high-throughput, transactional execution engine that processes customer directories, versioned price books, pricing strategy modifiers, sales orders, and billing triggers.
2. **`erp.crm.operations`**: The operational CRM surface handling lead ingestion, pipelines, opportunity scoring, customer support tickets, campaign attribution, and quoting.

```mermaid
graph TB
    subgraph "erp.crm.core (Transactional Engine)"
        CUST[Customer Profile<br/>Directory & Credit]
        PRICE[Price Books &<br/>Pricing Strategies]
        SO[Sales Orders<br/>Transactional Core]
        BILL[Billing Triggers<br/>Accruals & Purges]
        OUTBOX[Outbox Relay<br/>Event Sourcing]
    end

    subgraph "erp.crm.operations (Operational Surface)"
        LEAD[Lead Management<br/>Conversion Pipeline]
        OPP[Opportunity Tracking<br/>Stages & Value]
        QUOTE[Quote Generation<br/>Negotiation & Ingestion]
        TICKET[Service Tickets<br/>Support Desk]
        CAMP[Campaigns & Leads<br/>Marketing Attribution]
        INTERACT[Customer Interactions<br/>Meetings & Calls]
    end

    subgraph "Kafka Message Bus (C_e = 0 Integration)"
        KAFKA{Kafka Streams<br/>Primitive Token 'uuid'}
    end

    LEAD -.->|Converts & Registers| KAFKA
    KAFKA -.->|crm.core.customer.registered| CUST
    KAFKA -.->|crm.core.customer.status_changed| LEAD
    QUOTE --> OPP
    OPP --> SO
    SO --> BILL
```

## Documentation Structure

### Features Covered

* **[Order-to-Cash Trinity Integration Specification](order-to-cash.md)** — Architectural alignment, sequence flow, and resilience designs between CRM, SCM, and FM.
* **Customer Directory & Profiling** — Accounts, credit limits, and manager assignments (inline).
* **Price Books & Strategy Evaluation** — Volume breaks, temporal discounts, and contract Markups (inline).
* **Sales Order Life-Cycle** — Drafts, state transitions, validation, and confirmation (inline).
* **Billing Accrual & Purging** — Integration and handoff to Accounts Receivable (FM) (inline).
* **Outbox Relay & Event Sourcing** — Reliable transaction replication to Kafka (inline).
* **Marketing Campaign Ingestion** — Campaigns, leads, opportunity stages, and conversions (inline).
* **Customer Interactions & Support** — Meeting summaries, ticket resolutions, and quotes (inline).

---

## Topographical Domain Interaction Map

The diagram below outlines the runtime boundary of the `erp.crm` module, demonstrating how the system boundary remains clean of direct compile-time coupling ($C_e = 0$) by relying entirely on asynchronous event streams and primitive `uuid` tracking identifiers.

```
       [ PLM Core ]               [ SCM Logs ]               [ PRJ Engine ]
            │                          │                          │
            │ plm.material.released    │ scm.order.shipped        │ prj.milestone.achieved
            ▼                          ▼                          ▼
┌───────────────────────────────────────────────────────────────────────────────────────┐
│ erp.crm BOUNDED CONTEXT (Go / Gin / GORM)                                             │
│                                                                                       │
│  ┌─────────────────────────┐     ┌─────────────────────────┐     ┌─────────────────┐  │
│  │   KafkaEventInbox       │     │   CustomerProfile       │     │ PricingStrategy │  │
│  │   (Idempotent Receiver) │     │   (OCC Versioning)      │     │ (JSONB Matrix)  │  │
│  └───────────┬─────────────┘     └─────────────────────────┘     └─────────────────┘  │
│              │                                                                        │
│              ▼                                                                        │
│  ┌─────────────────────────┐             ┌────────────────────────────────────────┐  │
│  │   SalesOrder Engine     │────────────►│   BillingTrigger                       │  │
│  │   (State Machine Gates) │             │   (Monthly Range Partitioned)          │  │
│  └───────────┬─────────────┘             └───────────────────┬────────────────────┘  │
│              │                                               │                        │
│              ▼                                               ▼                        │
│  ┌─────────────────────────┐                                                          │
│  │   TransactionalOutbox   │                                                          │
│  │   (Atomic Event Log)    │                                                          │
│  └───────────┬─────────────┘                                                          │
└──────────────┼───────────────────────────────────────────────┬────────────────────────┘
               │                                               │
               │ crm.order.confirmed / cancelled               │ crm.billing.accrued
               ▼                                               ▼
     [ SCM / MFG Domains ]                               [ FM Ledger Core ]
```

---

## Event Ingress & Egress Pipelines

### 1. Ingress Pipeline: Inbound Message Streams

Inbound streams handle data ingestion via an **Idempotent Consumer** pattern. Processing status is recorded in the `crm_kafka_event_inbox` table to guarantee exactly-once processing semantics before mutating internal business tables.

#### A. `plm.material.released`
* **Source Boundary:** Product Lifecycle Management (PLM)
* **Ingress Execution Pattern:** 
  1. The Kafka consumer intercepts the message payload and checks the `event_id` against the `KafkaEventInbox`.
  2. If the message is unique, the payload is parsed and passed to the domain layer.
  3. The system saves the raw identifier token (`material_id`) into the CRM namespace.
  4. This token becomes immediately available within the `PricingCalculationService` for assigning list prices or configuration matrices via `PriceBookEntry`.

#### B. `scm.order.shipped`
* **Source Boundary:** Supply Chain Management (SCM / Logistics)
* **Ingress Execution Pattern:**
  1. This event indicates physical stock has left the warehouse and ownership has been legally transferred.
  2. The consumer invokes `RevenueBillingService.stageLogisticsBillingEntry`.
  3. The service maps fulfillment data to create a new record in `crm_billing_triggers`.
  4. The record is appended to the appropriate monthly partition based on the `triggered_at` timestamp coordinate, bypassing global lock boundaries.

#### C. `prj.milestone.achieved`
* **Source Boundary:** Project Management (PRJ)
* **Ingress Execution Pattern:**
  1. This event confirms customer sign-off on a fixed-price contract phase or a deliverables milestone.
  2. The consumer converts the project metrics into a financial format.
  3. It writes a billable row directly into the active `crm_billing_triggers` partition, establishing an audit trail for professional services.

---

### 2. Egress Pipeline: Outbound Message Streams

Outbound communication uses the **Transactional Outbox Pattern** to decouple the core system from network conditions. State mutations and event payloads are committed within the same atomic database transaction.

#### A. `crm.order.confirmed`
* **Target Consumers:** `erp.scm` (Logistics Allocation), `erp.mfg` (Factory Production)
* **Egress Trigger:** Fired when `SalesOrderService.processOrderStateTransition` successfully completes financial and credit validation checks, moving the order state to `APPROVED`.
* **Downstream Reactive Mechanics:**
  * **SCM:** Reads the `List<OrderLinePayload>` array to lock down and allocate warehouse stock buffers.
  * **MFG:** Parses line items marked with a `MAKE` procurement flag to initiate production schedules.

#### B. `crm.order.cancelled`
* **Target Consumers:** `erp.scm` (Stock Management), `erp.mfg` (Operations Control)
* **Egress Trigger:** Emitted when a sales contract is voided or terminated.
* **Downstream Reactive Mechanics:**
  * **SCM:** Releases reserved stock allocations back to the available inventory pool.
  * **MFG:** Halts active shop-floor work orders tied to the cancelled sales document.

#### C. `crm.billing.accrued`
* **Target Consumer:** `erp.fm` (Financial Ledger Core)
* **Egress Trigger:** Executed during the worker sweep cycle: `RevenueBillingService.dispatchStagedBillingToAccountsReceivable`.
* **Downstream Reactive Mechanics:**
  * The Financial Management module processes this event to generate a `UniversalJournalEntry`.
  * This action posts directly to the Accounts Receivable sub-ledger, balancing multi-tenant ledger accounts without requiring manual batch reconciliation.

---

## Architectural Trade-off Analysis (ATAM Matrix)

Evaluating these interactions reveals explicit trade-offs between system performance, auditability, and operational maintainability.

| Architectural Decision | Positive Quality Axis (Benefits) | Negative Quality Axis (Risks/Trade-offs) | Mitigation Strategy |
| :--- | :--- | :--- | :--- |
| **Primitive Ref Tokenization** (`uuid` based cross-domain links) | **Maintainability:** Achieves $C_e = 0$. Package updates in SCM/PLM never trigger compilation breakage inside CRM. | **Data Integrity:** The database cannot enforce traditional foreign key constraints across different service boundaries. | Inbound consumer contract validation layers via `ReliableMessagingService` to catch invalid references before database execution. |
| **Transactional Outbox Storage** (`crm_transactional_outbox`) | **Reliability:** Guarantees an RPO of zero. Business state and event entries succeed or fail together. | **Performance:** Double-write penalty. Every transaction requires writing to both the business table and the outbox log. | Utilize high-throughput composite indexing `(status, created_at)` and rapid polling loops on the outbox relay worker. |
| **Data-Driven Strategy Engines** (`jsonb` configuration matrices) | **Modularity:** New pricing policies are added by creating database rows, avoiding code deployments. | **Performance Efficiency:** Parsing complex JSON trees at runtime introduces higher CPU overhead than native code paths. | Apply optimized PostgreSQL indexes to look up active `strategy_version` matrices efficiently. |
| **Time-Bounded Range Partitioning** (`crm_billing_triggers`) | **Capacity:** Prevents large, slow tables. Deleting older records via batch loops maintains index efficiency. | **Operational Complexity:** Requires proactive table maintenance to manage partition creation and data retention policies. | Expose clear maintenance interfaces, such as `purgeProcessedBillingTriggers`, to automate data lifecycle management. |

---

## Domain Models

### 1. Transactional Core Engine (`erp.crm.core`)

| Model | CDD Table Reference | Description |
|-------|---------------------|-------------|
| `CustomerProfile` | `crm_customers` | Master customer account with credit limits, currency, and manager assignments. Includes optional contact info. |
| `PriceBookHeader` | `crm_price_books` | Header representing a pricing policy (Standard, Regional, etc.) with active dates. |
| `PriceBookEntry` | `crm_price_book_entries` | Map of unit list prices and quantity thresholds for materials within a price book. |
| `PricingStrategy` | `crm_pricing_strategies` | Zero-drift historical audit rules (flat markups, tiered breaks) applied dynamically. |
| `SalesOrder` | `crm_sales_orders` | Sales transactions containing total valuation, tax, and tracking state. |
| `SalesOrderLine` | `crm_sales_order_lines` | Detail line items representing quantity ordered/shipped, material IDs, and net amounts. |
| `BillingTrigger` | `crm_billing_triggers` | Monthly partitioned records staging revenue recognition inputs for Accounts Receivable. |
| `TransactionalOutbox` | `crm_transactional_outbox` | Outbox pattern message store ensuring at-least-once message delivery to Kafka. |
| `KafkaEventInbox` | `crm_kafka_event_inbox` | Idempotency log tracking processed Kafka message IDs and execution statuses. |

### 2. Operational CRM Surface (`erp.crm.operations`)

| Model | CDD Table Reference | Description |
|-------|---------------------|-------------|
| `Campaign` | `crm_campaigns` | Marketing campaigns tracking name, type, budget, and schedules. |
| `Lead` | `crm_leads` | Contact profiles representing qualified or unqualified prospective clients. |
| `Opportunity` | `crm_opportunities` | Pipeline tracker for potential deals detailing expected values and stage probabilities. |
| `CustomerInteraction` | `crm_customer_interactions` | Narrative log of customer meetings, phone calls, and email correspondence. |
| `ServiceTicket` | `crm_service_tickets` | Support desk request tickets tracking assignment, priority, and resolution states. |
| `Quote` | `crm_quotes` | Customer pricing proposals capturing overall valuation and validation duration. |
| `QuoteLineItem` | `crm_quote_line_items` | Individual line entries detailing material SKU quotes and pricing. |

---

## Business Services

### Package 1: Transactional Core Services (`erp.crm.core`)

#### CustomerAccountService
- `createProfile`: Provision a new customer profile under a legal entity with an manager employee ID.
- `updateCreditStatus`: Modify customer accounts (Active, Credit Hold, etc.) based on credit reviews.

#### PricingCalculationService
- `createPriceBook`: Define a new pricing matrix list header.
- `assignMaterialPrice`: Set baseline list price overrides for items in a price book.
- `registerStrategyModifier`: Apply modifier percentages or configuration tier rules to a strategy.
- `resolveItemUnitPrice`: Compute net line prices based on customer profiles and tiered pricing rules.

#### SalesOrderService
- `instantiateDraftOrder`: Create a draft order with line-item validation.
- `processOrderStateTransition`: Drive state machine changes (Draft → Pending Check → Approved → Shipped → Delivered).

#### RevenueBillingService
- `stageLogisticsBillingEntry`: Stage accrued billing amounts derived from SCM shipping documents.
- `dispatchStagedBillingToAccountsReceivable`: Batch transmit completed billing triggers to FM.
- `purgeProcessedBillingTriggers`: Clean up old database records using a monthly partition purge.

#### OutboxRelayWorker / ReliableMessagingService
- `getUnsentMessages`: Poll transactional outbox for pending payloads.
- `logProcessingAttempt` / `updateOutboxStatus`: Track delivery retries and success events.
- `executeIdempotentTransaction`: Process incoming event streams inside a transaction log.

---

### Package 2: Operational CRM Services (`erp.crm.operations`)

#### CampaignService
- `createCampaign` / `getCampaign` / `listCampaigns` / `updateCampaign` / `deleteCampaign`: Manage campaign lifecycles.

#### LeadService
- `createLead` / `getLead` / `listLeads` / `updateLead` / `deleteLead`: Perform typical lead maintenance.
- `convertLead`: Execute the lead conversion routine. Emits a registration signal to trigger downstream customer and opportunity profiles.

#### OpportunityService
- `createOpportunity` / `getOpportunity` / `listOpportunities` / `updateOpportunity` / `deleteOpportunity`: Maintain opportunities.

#### QuoteService
- `createQuote` / `getQuote` / `listQuotes` / `updateQuote` / `deleteQuote`: Handle customer quotes and items.
- `sendQuote`: Transition quote status to `SENT` and dispatch events.

#### TicketService
- `createTicket` / `getTicket` / `listTickets` / `updateTicket` / `deleteTicket`: Support ticket lifecycle operations.

#### CustomerInteractionService
- `createInteraction` / `getInteraction` / `listInteractions` / `deleteInteraction`: Manage communication logs.

---

## API Endpoints (35 routes)

### Customers
```http
GET    /api/v1/customers              # List all customers
POST   /api/v1/customers              # Create customer
GET    /api/v1/customers/:id          # Get customer by ID
PUT    /api/v1/customers/:id          # Update customer
DELETE /api/v1/customers/:id          # Delete customer
```

### Leads
```http
GET    /api/v1/leads                  # List all leads
POST   /api/v1/leads                  # Create lead
GET    /api/v1/leads/:id              # Get lead by ID
PUT    /api/v1/leads/:id              # Update lead
DELETE /api/v1/leads/:id              # Delete lead
POST   /api/v1/leads/:id/convert      # Convert lead to customer + opportunity
```

### Opportunities
```http
GET    /api/v1/opportunities          # List all opportunities
POST   /api/v1/opportunities          # Create opportunity
GET    /api/v1/opportunities/:id      # Get opportunity by ID
PUT    /api/v1/opportunities/:id      # Update opportunity
DELETE /api/v1/opportunities/:id      # Delete opportunity
```

### Sales Orders
```http
GET    /api/v1/sales-orders           # List all sales orders
POST   /api/v1/sales-orders           # Create sales order
GET    /api/v1/sales-orders/:id       # Get sales order by ID
PUT    /api/v1/sales-orders/:id       # Update sales order
DELETE /api/v1/sales-orders/:id       # Delete sales order
```

### Quotes
```http
GET    /api/v1/quotes                 # List all quotes
POST   /api/v1/quotes                 # Create quote
GET    /api/v1/quotes/:id             # Get quote by ID
PUT    /api/v1/quotes/:id             # Update quote
DELETE /api/v1/quotes/:id             # Delete quote
POST   /api/v1/quotes/:id/send        # Send quote to customer
```

### Service Tickets
```http
GET    /api/v1/service-tickets        # List all tickets
POST   /api/v1/service-tickets        # Create service ticket
GET    /api/v1/service-tickets/:id    # Get ticket by ID
PUT    /api/v1/service-tickets/:id    # Update ticket
DELETE /api/v1/service-tickets/:id    # Delete ticket
```

### Campaigns
```http
GET    /api/v1/campaigns              # List all campaigns
POST   /api/v1/campaigns              # Create campaign
GET    /api/v1/campaigns/:id          # Get campaign by ID
PUT    /api/v1/campaigns/:id          # Update campaign
DELETE /api/v1/campaigns/:id          # Delete campaign
```

### Price Lists
```http
GET    /api/v1/price-lists            # List all price lists
POST   /api/v1/price-lists            # Create price list
GET    /api/v1/price-lists/:id        # Get price list by ID
PUT    /api/v1/price-lists/:id        # Update price list
DELETE /api/v1/price-lists/:id        # Delete price list
```

---

## Sales Pipeline Flow

### Lead-to-Cash Process
```mermaid
flowchart LR
    A[Lead Capture<br/>Website/Referral/Campaign] --> B[Lead Qualification<br/>Contact & Score]
    B --> C{Qualified?}
    C -->|Yes| D[Convert to<br/>Customer + Opportunity]
    C -->|No| E[Disqualified<br/>Archive]
    D --> F[Opportunity Pipeline<br/>Prospecting → Negotiation]
    F --> G{Closed?}
    G -->|Won| H[Generate Quote]
    H --> I[Create Sales Order]
    I --> J[Order Confirmed<br/>Triggers Production/Fulfillment]
    J --> K[Order Shipped]
    K --> L[Order Delivered]
    L --> M[Invoice → Payment]
    G -->|Lost| N[Lost Reason Analysis]

    style D fill:#c8e6c9
    style I fill:#e1f5fe
    style M fill:#fff3e0
    style N fill:#ffcdd2
```

---

## Kafka Integration

The service decoupling requires all cross-boundary communications between `erp.crm.core` and `erp.crm.operations` (as well as external services) to be processed asynchronously over Kafka streams using primitive identifiers (`uuid`) to guarantee $C_e = 0$.

### Events Published (32 topics, per CDD)

* **Orders & Billing:**
  - `crm.order.confirmed`
  - `crm.order.cancelled`
  - `crm.billing.accrued`
  - `crm.sales.order.created`
  - `crm.sales.order.updated`
  - `crm.sales.order.confirmed`
  - `crm.sales.order.cancelled`
  - `crm.sales.order.shipped`
  - `crm.sales.order.delivered`
  - `crm.sales.order.received`
* **Customers & Profiles:**
  - `crm.customer.created`
  - `crm.customer.updated`
  - `crm.customer.activated`
  - `crm.customer.deactivated`
* **Leads & Opportunities:**
  - `crm.lead.created`
  - `crm.lead.qualified`
  - `crm.lead.converted`
  - `crm.lead.lost`
  - `crm.opportunity.created`
  - `crm.opportunity.updated`
  - `crm.opportunity.won`
  - `crm.opportunity.lost`
* **Marketing & Interactions:**
  - `crm.campaign.launched`
  - `crm.campaign.completed`
  - `crm.customer.interaction.logged`
  - `crm.email.sent`
  - `crm.email.opened`
  - `crm.email.clicked`
* **Service Tickets:**
  - `crm.service.ticket.created`
  - `crm.service.ticket.updated`
  - `crm.service.ticket.resolved`
  - `crm.service.ticket.escalated`

### Events Consumed (12 topics, per CDD)

| Topic | Publisher | Integration / Logic |
|-------|-----------|--------------------|
| `plm.material.released` | PLM | Update material catalogs |
| `scm.order.shipped` | SCM | Mark sales order as SHIPPED |
| `prj.milestone.achieved` | PM | Trigger billing for project milestones |
| `scm.inventory.available` | SCM | Logged only |
| `scm.shipment.delivered` | SCM | Update sales order status to DELIVERED |
| `fm.payment.received` | FM | Mark associated transactions as paid |
| `fm.credit.check.completed` | FM | Trigger credit-hold resolution flows |
| `mfg.production.completed` | MFG | Logged only |
| `prj.project.completed` | PM | Trigger final order reconciliations |
| `hr.employee.performance` | HR | Logged only |
| `crm.core.customer.registered` | Core | Synchronize profile mutations safely |
| `crm.core.customer.status_changed`| Core | Inform operational status views of credit holds |

---

## Seed Data

On startup, the service seeds mock data for development:
- **Customers**: Acme Corporation (Active)
- **Leads**: John Doe from Initech (score 10, New), Jane Smith from Umbrella Corp (score 10, Contacted)
- **Opportunities**: "Enterprise Software Deal" for Acme Corp ($50,000, Stage: Prospecting, 10% probability)

---

## Relation to Other Modules

| Module | Integration | Direction | Topic |
|--------|-------------|-----------|-------|
| **Manufacturing** | Sales order triggers production | Outbound | `crm.sales.order.created` |
| **SCM** | Demand forecast data | Outbound | `crm.customer.demand.forecast` |
| **SCM** | Order fulfillment trigger | Outbound | `crm.sales.order.created` |
| **PM** | Sales order creates project | Outbound | `crm.sales.order.received` |
| **FM** | Completed sale creates revenue entry | Outbound | `crm.sale.completed` |
