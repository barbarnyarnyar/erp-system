# PRD: Version 1.0 Survival Scorecard & Release Scope

This document evaluates the 9 software modules of the ERP system based strictly on their necessity for **Version 1.0 survival**. To prevent over-engineering and infrastructure tax, all theoretical elegance is ignored. Modules scoring below **8 / 10** are immediately cut from the V1.0 release.

---

## 1. The Core Decision Framework

Each module is graded using three strict binary metrics:

1. **Revenue Dependency (RD):** If this module is deleted, does the company immediately lose the ability to capture money?  
   * *Yes = 5 points, No = 0 points*
2. **Operational Substitute (OS):** Can this module be temporarily replaced by a spreadsheet, a human, or a cheap SaaS tool?  
   * *No = 3 points, Yes = 0 points*
3. **Infrastructure Tax (IT):** Can this be built in the core monolith database without distributed tracing or message brokers?  
   * *Yes = 2 points, No = -5 points*

> [!IMPORTANT]
> **Survival Threshold:** Any module scoring **< 8 points** is CUT. The target is to run a lean, monolithic database core for V1.0.

---

## 2. 9-Module Scorecard Matrix

| Module | Revenue Dependency (RD) | Operational Substitute (OS) | Infrastructure Tax (IT) | Total Score | Status (V1.0) |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Financials (FM)** | 5 (Yes) | 3 (No) | 2 (Yes) | **10 / 10** | **RETAIN (Core)** |
| **CRM Operations** | 5 (Yes) | 3 (No) | 2 (Yes) | **10 / 10** | **RETAIN (Core)** |
| **Supply Chain (SCM)** | 5 (Yes) | 3 (No) | 2 (Yes) | **10 / 10** | **RETAIN (Core)** |
| **Manufacturing (MFG)** | 0 (No) | 0 (Yes) | 2 (Yes) | **2 / 10** | ❌ **CUT** |
| **Product Lifecycle (PLM)** | 0 (No) | 0 (Yes) | 2 (Yes) | **2 / 10** | ❌ **CUT** |
| **Human Resources (HR)** | 0 (No) | 0 (Yes) | 2 (Yes) | **2 / 10** | ❌ **CUT** |
| **Project Management (PRJ)**| 0 (No) | 0 (Yes) | 2 (Yes) | **2 / 10** | ❌ **CUT** |
| **Quality Management (QMS)**| 0 (No) | 0 (Yes) | 2 (Yes) | **2 / 10** | ❌ **CUT** |
| **Asset Management (EAM)** | 0 (No) | 0 (Yes) | 2 (Yes) | **2 / 10** | ❌ **CUT** |

---

## 3. Scope Definition: The Order-to-Cash (O2C) Trinity

Only three modules survive the cut, representing the **O2C Trinity**. These three systems are tightly coupled to the flow of cash through the business and cannot be substituted without operational paralysis.

```mermaid
graph TD
    subgraph "V1.0 Survival Scope (The O2C Trinity)"
        CRM[CRM: Customer & Sales Orders] -->|Order Fulfill| SCM[SCM: Inventory & Fulfillment]
        SCM -->|Ship/Invoice| FM[FM: Invoices, AR/AP, Ledger]
    end

    subgraph "Cut Scope (Postponed / Spreadsheet Substitutes)"
        MFG[MFG: Shop Floor Execution] -.- x CRM
        PLM[PLM: BOM & CAD] -.- x SCM
        HR[HR: Payroll & Schedules] -.- x FM
        PRJ[PRJ: Projects & Tasks] -.- x CRM
        QMS[QMS: Inspection Logs] -.- x SCM
        EAM[EAM: Maintenance Schedules] -.- x SCM
    end

    style CRM fill:#c8e6c9,stroke:#388e3c
    style SCM fill:#c8e6c9,stroke:#388e3c
    style FM fill:#c8e6c9,stroke:#388e3c
    style MFG fill:#ffcdd2,stroke:#d32f2f
    style PLM fill:#ffcdd2,stroke:#d32f2f
    style HR fill:#ffcdd2,stroke:#d32f2f
    style PRJ fill:#ffcdd2,stroke:#d32f2f
    style QMS fill:#ffcdd2,stroke:#d32f2f
    style EAM fill:#ffcdd2,stroke:#d32f2f
```

### Why They Survive:
* **CRM:** Captures customer profiles and issues Sales Orders. If deleted, we cannot record who owes us money or what they bought. RD = 5, OS = 3 (Manual order matching fails under transaction load).
* **SCM:** Allocates physical warehouse inventory and manages shipping. If deleted, we sell products we don't have, leading to instant refund claims and operational collapse. RD = 5, OS = 3.
* **Financials (FM):** Generates customer invoices, manages General Ledgers, and enforces tax compliance. If deleted, we ship products but never recognize revenue or record accounts receivable. RD = 5, OS = 3.

---

## 4. Rationale for Cut Modules (Spreadsheet & SaaS Alternatives)

For the cut modules, business operations will temporarily fall back to manual substitutes for the V1.0 launch:

1. **Manufacturing (MFG):**
   * *V1.0 Substitute:* The factory floor will log production runs, routings, and station transitions using clipboard checklists and a shared Google Sheet.
2. **Product Lifecycle (PLM):**
   * *V1.0 Substitute:* BOMs will be maintained inside Excel sheets, and CAD drawing files will be organized inside a shared Google Drive folder structure.
3. **Human Resources (HR):**
   * *V1.0 Substitute:* Shift scheduling will be planned via spreadsheets, and payroll processing will be outsourced to a cheap SaaS tool (e.g. Gusto or ADP).
4. **Project Management (PRJ):**
   * *V1.0 Substitute:* Milestone tracking and tasks will be managed using a free Trello board or Jira Cloud instance.
5. **Quality Management (QMS):**
   * *V1.0 Substitute:* Inspectors will manually fill out paper checklists and log non-conformance issues in a shared spreadsheet tracker.
6. **Enterprise Asset Management (EAM):**
   * *V1.0 Substitute:* Machine logs and preventive maintenance intervals will be scheduled on a calendar app with recurring reminders.
