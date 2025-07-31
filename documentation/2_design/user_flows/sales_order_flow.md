# Sales Order Flow

This diagram shows the end-to-end flow of a sales order, from initial customer quote to fulfillment and invoicing. It demonstrates the interaction between the Sales, Supply Chain, and Finance modules.

```mermaid
flowchart TD
    subgraph Sales & CRM
        A[Start: Salesperson creates a new Quote] --> B{Send Quote to Customer};
        B --> C(Customer reviews and accepts Quote);
        C --> D[Salesperson converts Quote to Sales Order];
    end

    subgraph System & SCM
        D --> E{Check inventory for all line items};
        E -- All items in stock --> F[Reserve inventory for the order];
        E -- Some items out of stock --> G[Flag order for partial shipment / backorder];
        G --> H[Notify Operations Manager of stock shortage];
        H --> F;
    end

    subgraph Finance
        F --> I[Notify Finance: Create Invoice from Sales Order];
        I --> J[Generate and send Invoice to Customer];
    end

    subgraph Warehouse & Fulfillment
        F --> K[Notify Warehouse: New order ready for fulfillment];
        K --> L[Warehouse staff picks and packs items];
        L --> M[Integrate with shipping carrier to get tracking number];
        M --> N[Ship order to customer];
    end

    subgraph System & Customer
        N --> O[Update Order Status to 'Shipped'];
        O --> P[Send shipping confirmation email to customer with tracking info];
        P --> Q[End];
    end
```
