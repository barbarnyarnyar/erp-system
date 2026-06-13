# Order-to-Cash Trinity: CRM-SCM-FM Integration Alignment

The **Order-to-Cash (O2C)** process forms the primary transactional engine of the ERP system. It connects three core business units—Sales (CRM), Logistics (SCM), and Accounting (FM)—into a unified, eventual-consistent integration loop.

```
       ┌──────────────────┐               ┌──────────────────┐
       │   CRM Service    ├──────────────►│   SCM Service    │
       │  (Sales Orders)  │ crm.sales.    │   (Logistics)    │
       └────────▲─────────┘ order.conf    └────────┬─────────┘
                │                                  │
                │ fm.payment.                      │ scm.shipment.
                │ received                         │ delivered
                │                                  ▼
       ┌────────┴─────────┐               ┌──────────────────┐
       │    FM Service    │◄──────────────┤   CRM Service    │
       │ (Ledger/Invoice) │  crm.billing. │ (Billing Stage)  │
       └──────────────────┘    accrued    └──────────────────┘
```

---

## 1. The Trinity Dynamics

The relationship between the three services maps directly to the flow of commitments, physical goods, and financial recognition:

| Service Boundary | Responsibility | Key Domain Entity | Input Trigger | Output Event |
| :--- | :--- | :--- | :--- | :--- |
| **CRM** (`erp.crm.core`) | Captures customer agreement, verifies credit limits, and processes billing coordinates. | `SalesOrder` | User interaction / credit clearance | `crm.sales.order.confirmed` |
| **SCM** (`erp.scm`) | Reserves stock, coordinates shipping logistics, and transfers legal ownership of goods. | `Shipment` | `crm.sales.order.confirmed` | `scm.shipment.delivered` |
| **CRM** (Billing Core) | Receives delivery proof, creates partitioned billing triggers, and calculates revenue. | `BillingTrigger` | `scm.shipment.delivered` | `crm.billing.accrued` |
| **FM** (`erp.fm`) | Creates accounts receivable subledger invoices, matches payments, and posts journal entries. | `ArInvoice` | `crm.billing.accrued` | `fm.payment.received` |

---

## 2. Event-Driven State Transitions

### Step 1: Order Confirmation (CRM $\rightarrow$ SCM)
* **Initiation**: A sales order is confirmed in CRM (moving state from `DRAFT` to `CONFIRMED`).
* **Publishing**: CRM commits the state change and writes a `SalesOrderConfirmedEvent` payload atomically to the transactional outbox (`crm_transactional_outbox`).
* **Event**: `crm.sales.order.confirmed`
* **Consumption**: SCM consumes this event, registers the order in its local system, locks down the requested product stock, and generates a warehouse picking ticket.

### Step 2: Physical Fulfillment (SCM $\rightarrow$ CRM)
* **Initiation**: SCM completes the picking, packing, and dispatch cycles, and receives delivery confirmation from logistics carriers.
* **Publishing**: SCM writes a `ShipmentDeliveredEvent` payload atomically to its outbox.
* **Event**: `scm.shipment.delivered` (or `scm.order.shipped`)
* **Consumption**: CRM intercepts the event, updates the sales order state to `DELIVERED`, and stages a billing entry.

### Step 3: Billing Accrual (CRM $\rightarrow$ FM)
* **Initiation**: The CRM billing worker scans active billing triggers (such as range-partitioned `crm_billing_triggers`) and dispatches them.
* **Publishing**: CRM writes a `BillingAccruedEvent` payload to the outbox.
* **Event**: `crm.billing.accrued`
* **Consumption**: FM receives the billing accrue details, generates an Accounts Receivable subledger invoice (`ArInvoice`), and posts ledger lines (`UniversalJournalEntry`) debiting Accounts Receivable and crediting Revenue.

### Step 4: Payment Recognition (FM $\rightarrow$ CRM)
* **Initiation**: The customer payment is received (wire, credit card, etc.) and recorded in FM.
* **Publishing**: FM writes a `PaymentReceivedEvent` to its outbox.
* **Event**: `fm.payment.received`
* **Consumption**: CRM consumes the payment event to update the customer's credit record and flag the sales order as paid.

---

## 3. Resilience & Failure Scenarios

Because this trinity spans three separate network nodes, failures in event delivery represent critical business risks. The architecture mitigates these risks using two main patterns:

### Scenario A: CRM $\rightarrow$ SCM Link Failure
* **Risk**: A sales order is paid/confirmed in CRM, but the SCM service never receives the event. SCM does not allocate stock, and the order is never shipped.
* **Impact**: Customer dissatisfaction, breach of service contracts.
* **Mitigation**:
  1. **At-Least-Once Delivery**: The CRM outbox worker polls `crm_transactional_outbox` periodically. It retries publishing until Kafka acknowledges receipt.
  2. **Exactly-Once Processing**: SCM stores the incoming CRM event ID in its database (`scm_kafka_event_inbox`) before processing. Duplicate messages are ignored immediately without repeating inventory allocation.

### Scenario B: SCM $\rightarrow$ FM (via CRM) Link Failure
* **Risk**: SCM ships the goods out the door, but CRM or FM fails to process the shipment event. The accounting ledger remains unaware of the shipment, resulting in unregistered revenue (unbilled shipments).
* **Impact**: Inventory write-off, financial underreporting, compliance audit failures.
* **Mitigation**:
  1. **Decoupled Billing Accruals**: The billing process uses the `BillingTrigger` entity. If FM is offline or the network is partitioned, delivery details are staged locally in the range-partitioned database table and retried automatically.
  2. **Audit reconciliation scripts**: The background billing worker runs monthly range-partition checks to reconcile shipped sales order lines against registered ledger entries, raising alerts on any discrepancies.
