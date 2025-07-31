# Production Order Flow (Manufacturing)

This diagram shows the workflow for the optional Manufacturing module, from creating a production order to updating inventory with finished goods.

```mermaid
flowchart TD
    subgraph Planning
        A[Start: Demand forecast or low stock level triggers need for production] --> B[Production Planner creates a Production Order];
        B --> C{Select Bill of Materials (BOM) and specify quantity};
        C --> D[System checks for raw material availability];
        D -- Not enough materials --> E[Generate Purchase Requisitions for missing materials];
        E --> F[Wait for materials];
        F --> D;
        D -- Materials available --> G[System reserves raw materials from inventory];
    end

    subgraph Shop Floor
        G --> H[Schedule Production Order on the shop floor];
        H --> I[Release Production Order to the floor];
        I --> J[Shop Floor staff begin production work];
        J --> K[Record material consumption];
        K --> L[Record labor hours];
    end

    subgraph Quality & Inventory
        L --> M[Production is complete];
        M --> N{Perform Quality Control check};
        N -- Fails --> O[Quarantine batch and investigate];
        O --> P[End];
        N -- Passes --> Q[Record finished goods output];
        Q --> R[System updates inventory: decreases raw materials, increases finished goods];
        R --> S[Production Order is closed];
        S --> T[End];
    end
```
