# Procure-to-Pay Flow

This diagram illustrates the end-to-end procurement process, from creating a purchase order to paying the vendor. It involves the Supply Chain and Finance modules.

```mermaid
flowchart TD
    subgraph SCM / Operations
        A[Start: Operations Manager identifies need for a product] --> B[Create Purchase Requisition];
        B --> C{Approval Workflow};
        C -- Approved --> D[Convert Requisition to Purchase Order (PO)];
        C -- Rejected --> E[End];
        D --> F[Send PO to Vendor];
    end

    subgraph Vendor
        F --> G(Vendor receives PO and ships goods);
        G --> H(Vendor sends invoice to company);
    end

    subgraph SCM / Warehouse
        G --> I[Warehouse receives shipment];
        I --> J{Match goods against PO};
        J -- Discrepancy --> K[Resolve discrepancy with Vendor];
        K --> J;
        J -- Matched --> L[Update inventory stock levels];
    end

    subgraph Finance / Accounts Payable
        H --> M[AP Accountant receives Vendor Invoice];
        M --> N{3-Way Match: Invoice vs. PO vs. Goods Receipt};
        N -- Discrepancy --> O[Resolve matching issue with Operations/Vendor];
        O --> N;
        N -- Matched --> P[Approve Invoice for Payment];
        P --> Q[Schedule payment according to vendor terms];
        Q --> R[Process payment to Vendor];
        R --> S[Record transaction in General Ledger];
        S --> T[End];
    end
```
