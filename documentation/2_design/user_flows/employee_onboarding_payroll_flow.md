# Employee Onboarding & Payroll Flow

This diagram outlines the process of onboarding a new employee and running their first payroll, involving the HR module and integrating with the Finance module.

```mermaid
flowchart TD
    subgraph HR Onboarding
        A[Start: Candidate accepts job offer] --> B[HR Manager creates new Employee Profile in system];
        B --> C[Assign position, department, and salary];
        C --> D[Send onboarding documents to new employee];
        D --> E(Employee completes forms and uploads documents via Self-Service Portal);
        E --> F[HR Manager verifies documents and activates employee profile];
    end

    subgraph Employee & Manager
        F --> G[Employee is assigned a manager];
        G --> H[Manager assigns initial tasks/projects];
        H --> I(Employee logs time and attendance);
    end

    subgraph HR Payroll
        I --> J[End of pay period: HR Manager initiates payroll run];
        J --> K{System calculates gross pay from salary and approved time entries};
        K --> L[System calculates taxes and benefit deductions];
        L --> M[HR Manager reviews the draft payroll register];
        M --> N{Is payroll correct?};
        N -- No --> O[Make corrections and recalculate];
        O --> M;
        N -- Yes --> P[Approve payroll run];
    end

    subgraph Finance & Banking
        P --> Q[System generates payment instructions (e.g., ACH file)];
        Q --> R[Finance Manager submits payment file to bank];
        R --> S[System posts payroll expenses to the General Ledger];
        S --> T(Bank processes payments to employees);
        T --> U[End];
    end
```
