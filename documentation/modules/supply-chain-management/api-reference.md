# Supply Chain Management API Reference

Complete REST API documentation for the Supply Chain Management module. Port **8003** (docker-compose: 8006).

## Base URL
```
http://localhost:8003/api/v1
```

## Response Format

All endpoints return JSON:
```json
{
  "data": { ... },
  "error": "..."  // only on errors
}
```

Error responses include HTTP status codes:
- `400 Bad Request` — validation error
- `404 Not Found` — resource not found
- `500 Internal Server Error` — server error

---

## Product Catalog

### List Product Categories
```http
GET /api/v1/product-categories
```

Response:
```json
{
  "data": [
    {
      "id": "cat_111",
      "code": "ELEC",
      "name": "Electronics",
      "description": "Electronic components and devices"
    }
  ]
}
```

### Create Product Category
```http
POST /api/v1/product-categories
Content-Type: application/json

{
  "code": "ELEC",
  "name": "Electronics",
  "description": "Electronic components and devices"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "cat_111",
    "code": "ELEC",
    "name": "Electronics",
    "description": "Electronic components and devices"
  }
}
```

### Get Product Category
```http
GET /api/v1/product-categories/:id
```

### Update Product Category
```http
PUT /api/v1/product-categories/:id
Content-Type: application/json

{
  "code": "ELEC-UPDATED",
  "name": "Consumer Electronics",
  "description": "Finished electronic goods"
}
```

### Delete Product Category
```http
DELETE /api/v1/product-categories/:id
```

---

## Products

### List Products
```http
GET /api/v1/products
```

Response:
```json
{
  "data": [
    {
      "id": "prod_server001",
      "product_code": "EQ-SERVER-001",
      "product_name": "ProLiant Gen10 Server",
      "description": "Enterprise rack server",
      "product_type": "FINISHED_GOOD",
      "category_id": "cat_111",
      "unit_of_measure": "PCS",
      "standard_cost": "10000.0000",
      "list_price": "12000.0000",
      "is_active": true,
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Create Product
```http
POST /api/v1/products
Content-Type: application/json

{
  "product_code": "EQ-SERVER-001",
  "product_name": "ProLiant Gen10 Server",
  "description": "Enterprise rack server",
  "product_type": "FINISHED_GOOD",
  "category_id": "cat_111",
  "unit_of_measure": "PCS",
  "standard_cost": "10000.00",
  "list_price": "12000.00"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "prod_server001",
    "product_code": "EQ-SERVER-001",
    "product_name": "ProLiant Gen10 Server",
    "description": "Enterprise rack server",
    "product_type": "FINISHED_GOOD",
    "category_id": "cat_111",
    "unit_of_measure": "PCS",
    "standard_cost": "10000.0000",
    "list_price": "12000.0000",
    "is_active": true,
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Get Product
```http
GET /api/v1/products/:id
```

### Update Product
```http
PUT /api/v1/products/:id
Content-Type: application/json

{
  "product_code": "EQ-SERVER-001",
  "product_name": "ProLiant Gen10 Rack Server",
  "description": "Enterprise rack server",
  "product_type": "FINISHED_GOOD",
  "category_id": "cat_111",
  "unit_of_measure": "PCS",
  "standard_cost": "10000.00",
  "list_price": "12500.00",
  "is_active": true
}
```

### Delete Product
```http
DELETE /api/v1/products/:id
```

---

## Locations

Warehouses, retail stores, or in-transit hubs.

### List Locations
```http
GET /api/v1/locations
```

Response:
```json
{
  "data": [
    {
      "id": "loc_main_wh",
      "location_code": "MAIN-WH",
      "location_name": "Main Warehouse",
      "location_type": "WAREHOUSE",
      "is_active": true,
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Create Location
```http
POST /api/v1/locations
Content-Type: application/json

{
  "location_code": "MAIN-WH",
  "location_name": "Main Warehouse",
  "location_type": "WAREHOUSE"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "loc_main_wh",
    "location_code": "MAIN-WH",
    "location_name": "Main Warehouse",
    "location_type": "WAREHOUSE",
    "is_active": true,
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

---

## Vendor Management

### List Vendors (Suppliers)
```http
GET /api/v1/vendors
```

Response:
```json
{
  "data": [
    {
      "id": "vend_hp",
      "supplier_code": "VND-HP",
      "supplier_name": "HP Enterprise",
      "contact_name": "John Doe",
      "email": "john.doe@hpe.com",
      "phone": "+1-555-0199",
      "is_active": true,
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Create Vendor
```http
POST /api/v1/vendors
Content-Type: application/json

{
  "supplier_code": "VND-HP",
  "supplier_name": "HP Enterprise",
  "contact_name": "John Doe",
  "email": "john.doe@hpe.com",
  "phone": "+1-555-0199"
}
```

### Vendor Contracts
```http
GET /api/v1/vendor-contracts
```

```http
POST /api/v1/vendor-contracts
Content-Type: application/json

{
  "contract_number": "CON-2026-001",
  "supplier_id": "vend_hp",
  "start_date": "2026-01-01",
  "end_date": "2026-12-31",
  "terms": "Net 30 Payment Terms"
}
```

---

## Purchase Requisitions

Employee-initiated purchase request workflow.

### List Requisitions
```http
GET /api/v1/purchase-requisitions
```

### Create Requisition
```http
POST /api/v1/purchase-requisitions
Content-Type: application/json

{
  "requester_id": "emp_001",
  "request_date": "2026-06-13",
  "notes": "Requesting server replacement components",
  "lines": [
    {
      "product_id": "prod_server001",
      "quantity_requested": 2,
      "estimated_unit_price": "12000.00"
    }
  ]
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "req_123",
    "req_number": "REQ-10023",
    "requester_id": "emp_001",
    "request_date": "2026-06-13T00:00:00Z",
    "status": "DRAFT",
    "total_amount": "24000.0000",
    "notes": "Requesting server replacement components",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Requisition Approvals
```http
POST /api/v1/purchase-requisitions/:id/approve
```

```http
POST /api/v1/purchase-requisitions/:id/reject
```

### Get Requisition Lines
```http
GET /api/v1/purchase-requisitions/:id/lines
```

---

## Purchase Orders

Supplier-facing purchase order processing.

### Create Purchase Order
```http
POST /api/v1/purchase-orders
Content-Type: application/json

{
  "supplier_id": "vend_hp",
  "expected_delivery": "2026-06-20T00:00:00Z",
  "notes": "PO for rack servers",
  "lines": [
    {
      "product_id": "prod_server001",
      "quantity_ordered": 2,
      "unit_price": "10000.00",
      "description": "Enterprise server unit"
    }
  ]
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "po_999",
    "po_number": "PO-2026-1004",
    "supplier_id": "vend_hp",
    "order_date": "2026-06-13T02:00:00Z",
    "expected_delivery": "2026-06-20T00:00:00Z",
    "status": "DRAFT",
    "total_amount": "20000.0000",
    "notes": "PO for rack servers",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Send Purchase Order
```http
POST /api/v1/purchase-orders/:id/send
```

Sends PO to vendor, publishing `scm.purchase.order.sent` and reserving inventory stock.

### Get Purchase Order Lines
```http
GET /api/v1/purchase-orders/:id/lines
```

---

## Inventory

### List Inventory Items
```http
GET /api/v1/inventory
```

Response:
```json
{
  "data": [
    {
      "id": "inv_item_888",
      "product_id": "prod_server001",
      "location_id": "loc_main_wh",
      "quantity_on_hand": 10,
      "quantity_reserved": 2,
      "quantity_available": 8,
      "reorder_point": 5,
      "maximum_stock": 50,
      "unit_cost": "10000.0000",
      "created_at": "2026-06-13T02:00:00Z",
      "updated_at": "2026-06-13T02:00:00Z"
    }
  ]
}
```

### Create Inventory Item
```http
POST /api/v1/inventory
Content-Type: application/json

{
  "product_id": "prod_server001",
  "location_id": "loc_main_wh",
  "quantity_on_hand": 10,
  "reorder_point": 5,
  "maximum_stock": 50,
  "unit_cost": "10000.00"
}
```

### Reserve Stock
```http
POST /api/v1/inventory/reserve
Content-Type: application/json

{
  "product_id": "prod_server001",
  "location_id": "loc_main_wh",
  "quantity": 2,
  "reference_id": "ref_so_1002"
}
```

### Release Stock
```http
POST /api/v1/inventory/release
Content-Type: application/json

{
  "reference_id": "ref_so_1002"
}
```

### List Inventory Movements
```http
GET /api/v1/inventory/movements
```

---

## Stock Transfers

### Create Stock Transfer
```http
POST /api/v1/stock-transfers
Content-Type: application/json

{
  "from_location_id": "loc_main_wh",
  "to_location_id": "loc_retail_store",
  "product_id": "prod_server001",
  "quantity": 1
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "st_555",
    "from_location_id": "loc_main_wh",
    "to_location_id": "loc_retail_store",
    "product_id": "prod_server001",
    "quantity": 1,
    "status": "REQUESTED",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Execute Stock Transfer
```http
POST /api/v1/stock-transfers/:id/execute
```

Executes stock transfer, moving physical quantities from the source location to the target location.

---

## Warehouse Operations

### Create Receipt (Goods Receipt)
```http
POST /api/v1/receipts
Content-Type: application/json

{
  "purchase_order_id": "po_999",
  "notes": "Server rack arrival",
  "lines": [
    {
      "product_id": "prod_server001",
      "quantity_received": 2,
      "location_id": "loc_main_wh"
    }
  ]
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "rec_777",
    "purchase_order_id": "po_999",
    "received_date": "2026-06-13T02:00:00Z",
    "status": "RECEIVED",
    "notes": "Server rack arrival",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Create Shipment (Outbound Shipments)
```http
POST /api/v1/shipments
Content-Type: application/json

{
  "carrier": "FedEx",
  "tracking_number": "TRK-1294819",
  "estimated_delivery": "2026-06-16T12:00:00Z",
  "notes": "Deliver to customer site",
  "lines": [
    {
      "product_id": "prod_server001",
      "quantity_shipped": 1,
      "location_id": "loc_main_wh"
    }
  ]
}
```

---

## Demand Forecasting

### Create Forecast
```http
POST /api/v1/demand-forecasts
Content-Type: application/json

{
  "product_id": "prod_server001",
  "forecast_date": "2026-07-01T00:00:00Z",
  "forecast_quantity": 15,
  "confidence_level": "0.85",
  "notes": "Expected high demand for Q3"
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "fore_222",
    "product_id": "prod_server001",
    "forecast_date": "2026-07-01T00:00:00Z",
    "forecast_quantity": 15,
    "confidence_level": "0.8500",
    "notes": "Expected high demand for Q3",
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

---

## Reports

### Inventory Levels
```http
GET /api/v1/reports/inventory-levels
```

### Vendor Performance
```http
GET /api/v1/reports/vendor-performance
```

### Procurement Metrics
```http
GET /api/v1/reports/procurement-metrics
```

### Safety Stock Report
```http
GET /api/v1/reports/safety-stock
```
