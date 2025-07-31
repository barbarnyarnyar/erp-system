# Project Time Tracking & Billing Flow (Project Management)

This diagram illustrates the workflow for the optional Project Management module, showing how time is tracked against projects and subsequently billed to clients.

```mermaid
flowchart TD
    subgraph Project Setup
        A[Start: New client project is approved] --> B[Project Manager creates a new Project in the system];
        B --> C[Define project tasks, milestones, and budget];
        C --> D[Assign employees to the project team];
    end

    subgraph Time & Expense Tracking
        D --> E[Team members are notified of assignment];
        E --> F(Employee performs work on assigned tasks);
        F --> G[Employee logs hours against project tasks in their timesheet];
        G --> H(Employee submits timesheet weekly);
        H --> I{Project Manager reviews and approves timesheet};
        I -- Rejected --> J[Employee corrects and resubmits];
        J --> H;
        I -- Approved --> K[Approved hours are logged to the project];
    end

    subgraph Billing & Finance
        K --> L[End of billing cycle (e.g., monthly)];
        L --> M[Finance Manager generates a draft invoice for the project];
        M --> N{System pulls all unbilled approved hours and expenses};
        N --> O[Project Manager reviews the draft invoice for accuracy];
        O --> P[Send final invoice to the client];
        P --> Q[Record invoice in Accounts Receivable];
        Q --> R[Track project profitability by comparing billed amounts to labor costs];
        R --> S[End];
    end
```
