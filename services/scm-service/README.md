# Supply Chain Management Service (scm-service)

The Supply Chain Management Service is a core microservice of the ERP system responsible for handling all logistics operations including:

- Product and category catalog
- Vendor (supplier) relationships and contracts
- Purchase requisitions, approvals, and purchase orders
- Inventory item tracking, stock reservations, and stock movements
- Multi-location stock transfers
- Goods receiving and outbound shipping logs
- Safety stock and forecasting metrics

## Architecture

This service is written in Go and implements clean architecture with the following layers:

- **Domain Layer**: Core business entities and rules
- **Business Layer**: Services and application logic
- **API Layer**: HTTP handlers and routes (Gin framework)
- **Data Layer**: In-memory mock repositories (simulating persistence)
- **Infrastructure**: Kafka integration for event-driven logic

## Getting Started

### Prerequisites

- Go 1.21+
- Kafka (for event publishing and consumption)

### Run the Service

1. Navigate to the scm-service directory:
```bash
cd services/scm-service
```

2. Run the service:
```bash
go run cmd/main.go
```

The service will start on port **8003** by default.

## API Endpoints

### Product Categories
- `GET /api/v1/product-categories` - List categories
- `POST /api/v1/product-categories` - Create category
- `GET /api/v1/product-categories/:id` - Get category
- `PUT /api/v1/product-categories/:id` - Update category
- `DELETE /api/v1/product-categories/:id` - Delete category

### Products
- `GET /api/v1/products` - List products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/:id` - Get product
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product

### Locations
- `GET /api/v1/locations` - List locations
- `POST /api/v1/locations` - Create location
- `GET /api/v1/locations/:id` - Get location
- `PUT /api/v1/locations/:id` - Update location
- `DELETE /api/v1/locations/:id` - Delete location

### Vendors & contracts
- `GET /api/v1/vendors` - List vendors
- `POST /api/v1/vendors` - Create vendor
- `GET /api/v1/vendors/:id` - Get vendor details
- `PUT /api/v1/vendors/:id` - Update vendor details
- `DELETE /api/v1/vendors/:id` - Delete vendor
- `GET /api/v1/vendor-contracts` - List contracts
- `POST /api/v1/vendor-contracts` - Create vendor contract
- `GET /api/v1/vendor-contracts/:id` - Get contract details
- `PUT /api/v1/vendor-contracts/:id` - Update contract details
- `DELETE /api/v1/vendor-contracts/:id` - Delete contract

### Purchase Requisitions
- `GET /api/v1/purchase-requisitions` - List requisitions
- `POST /api/v1/purchase-requisitions` - Create purchase requisition
- `GET /api/v1/purchase-requisitions/:id` - Get requisition details
- `PUT /api/v1/purchase-requisitions/:id` - Update requisition details
- `DELETE /api/v1/purchase-requisitions/:id` - Delete requisition
- `POST /api/v1/purchase-requisitions/:id/approve` - Approve requisition
- `POST /api/v1/purchase-requisitions/:id/reject` - Reject requisition
- `GET /api/v1/purchase-requisitions/:id/lines` - Get requisition line items

### Purchase Orders
- `GET /api/v1/purchase-orders` - List purchase orders
- `POST /api/v1/purchase-orders` - Create purchase order
- `GET /api/v1/purchase-orders/:id` - Get purchase order details
- `PUT /api/v1/purchase-orders/:id` - Update purchase order details
- `DELETE /api/v1/purchase-orders/:id` - Delete purchase order
- `POST /api/v1/purchase-orders/:id/send` - Send PO to supplier
- `GET /api/v1/purchase-orders/:id/lines` - Get PO line items

### Inventory & Transfers
- `GET /api/v1/inventory` - List inventory items
- `POST /api/v1/inventory` - Create inventory item
- `GET /api/v1/inventory/:id` - Get inventory item details
- `PUT /api/v1/inventory/:id` - Update inventory item details
- `DELETE /api/v1/inventory/:id` - Delete inventory item
- `POST /api/v1/inventory/reserve` - Reserve stock
- `POST /api/v1/inventory/release` - Release stock reservation
- `GET /api/v1/inventory/movements` - List inventory movements
- `GET /api/v1/stock-transfers` - List transfers
- `POST /api/v1/stock-transfers` - Create stock transfer
- `GET /api/v1/stock-transfers/:id` - Get transfer details
- `POST /api/v1/stock-transfers/:id/execute` - Execute stock transfer

### Warehouse Operations
- `GET /api/v1/receipts` - List goods receipts
- `POST /api/v1/receipts` - Create goods receipt
- `GET /api/v1/receipts/:id` - Get receipt details
- `PUT /api/v1/receipts/:id` - Update receipt details
- `GET /api/v1/receipts/:id/lines` - Get receipt line items
- `GET /api/v1/shipments` - List outbound shipments
- `POST /api/v1/shipments` - Create shipment
- `GET /api/v1/shipments/:id` - Get shipment details
- `PUT /api/v1/shipments/:id` - Update shipment details
- `GET /api/v1/shipments/:id/lines` - Get shipment line items

### Demand Forecasting & Reports
- `GET /api/v1/demand-forecasts` - List demand forecasts
- `POST /api/v1/demand-forecasts` - Create demand forecast
- `GET /api/v1/demand-forecasts/:id` - Get forecast details
- `PUT /api/v1/demand-forecasts/:id` - Update forecast details
- `GET /api/v1/reports/inventory-levels` - Current stock levels report
- `GET /api/v1/reports/vendor-performance` - Supplier performance metrics
- `GET /api/v1/reports/procurement-metrics` - Procurement stats
- `GET /api/v1/reports/safety-stock` - Safety stock warning report
