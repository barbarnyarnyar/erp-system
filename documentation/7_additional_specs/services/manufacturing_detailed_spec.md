# Manufacturing

### Core Features

- **Bill of Materials (BOM)**
    - Product Structure
    - Component Relationships
    - BOM Versions
    - Engineering Changes
- **Routing Management**
    - Operation Sequences
    - Work Centers
    - Setup Times
    - Labor Standards
- **Production Planning**
    - Master Production Schedule
    - Material Requirements Planning (MRP)
    - Capacity Planning
    - Production Scheduling
- **Shop Floor Control**
    - Work Order Management
    - Production Tracking
    - Labor Reporting
    - Machine Monitoring
- **Quality Management**
    - Quality Control Plans
    - Inspection Results
    - Non-Conformance Tracking
    - Corrective Actions
- **Maintenance Management**
    - Preventive Maintenance
    - Work Order Management
    - Equipment History
    - Spare Parts Management
- **Costing**
    - Standard Costing
    - Actual Costing
    - Variance Analysis
    - Cost Roll-up

### REST APIs

```go
go
// BOM Management
GET    /api/v1/boms// List BOMs
POST   /api/v1/boms// Create BOM
GET    /api/v1/boms/{id}// Get BOM details
PUT    /api/v1/boms/{id}// Update BOM
DELETE /api/v1/boms/{id}// Delete BOM// Routing Management
GET    /api/v1/routings// List routings
POST   /api/v1/routings// Create routing
GET    /api/v1/routings/{id}// Get routing details
PUT    /api/v1/routings/{id}// Update routing
DELETE /api/v1/routings/{id}// Delete routing// Work Orders
GET    /api/v1/work-orders// List work orders
POST   /api/v1/work-orders// Create work order
GET    /api/v1/work-orders/{id}// Get work order details
PUT    /api/v1/work-orders/{id}// Update work order
DELETE /api/v1/work-orders/{id}// Delete work order
POST   /api/v1/work-orders/{id}/start// Start work order
POST   /api/v1/work-orders/{id}/complete// Complete work order// Production Planning
GET    /api/v1/production-plans// List production plans
POST   /api/v1/production-plans// Create production plan
GET    /api/v1/production-plans/{id}// Get production plan
PUT    /api/v1/production-plans/{id}// Update production plan
POST   /api/v1/mrp/run// Run MRP// Quality Control
GET    /api/v1/quality-inspections// List inspections
POST   /api/v1/quality-inspections// Create inspection
GET    /api/v1/quality-inspections/{id}// Get inspection details
PUT    /api/v1/quality-inspections/{id}// Update inspection// Work Centers
GET    /api/v1/work-centers// List work centers
POST   /api/v1/work-centers// Create work center
GET    /api/v1/work-centers/{id}// Get work center details
PUT    /api/v1/work-centers/{id}// Update work center
DELETE /api/v1/work-centers/{id}// Delete work center// Maintenance
GET    /api/v1/maintenance-schedules// List maintenance schedules
POST   /api/v1/maintenance-schedules// Create maintenance schedule
GET    /api/v1/maintenance-schedules/{id}// Get maintenance schedule
PUT    /api/v1/maintenance-schedules/{id}// Update maintenance schedule
```

### Message Queue Events

### Published Events

```go
go
// Production Events
mfg.production.scheduled
mfg.production.started
mfg.production.completed
mfg.production.delayed

// Work Order Events
mfg.work.order.created
mfg.work.order.started
mfg.work.order.completed
mfg.work.order.cancelled

// Material Events
mfg.material.consumed
mfg.material.wasted
mfg.material.required

// Quality Events
mfg.quality.inspection.passed
mfg.quality.inspection.failed
mfg.quality.non.conformance.detected

// Maintenance Events
mfg.maintenance.scheduled
mfg.maintenance.completed
mfg.equipment.down
mfg.equipment.up
```

### Consumed Events

```go
go
// From SCM Module
scm.material.received// Update material availability
scm.inventory.updated// Update material status// From CRM Module
crm.sales.order.created// Create production demand// From Financial Module
fin.cost.budget.allocated// Update cost budgets// From HR Module
hr.employee.scheduled// Update labor capacity// From Project Module
prj.custom.order.created// Create custom production order
```