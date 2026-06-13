# PRD: SCM Service Documentation Alignment & API Reference Update

**PRD ID**: PRD-2026-06-13-0223  
**Date**: 2026-06-13  
**Status**: Implemented  
**Parent Initiative**: Technical Documentation Sync & API Standard  
**Target Alignment**: 100% parity between SCM service implementation, CDD contract definitions, and user-facing documentation  

---

## 1. Objective & Problem Statement

A detailed comparison between the Supply Chain Management (SCM) module's documentation and its actual codebase implementation reveals several key variations:
1. **Model Field Mismatches**: 
   - `Product` has fields like `ProductCode` and `ProductName` instead of `SKU` and `Name`.
   - `ProductCategory` is flat (no parent category ID).
   - `Location` uses `LocationCode`, `LocationName`, and `LocationType` instead of `Code`, `Name`, `Type`, and `Address`.
   - `Supplier` uses `SupplierCode` and `SupplierName` instead of `Code` and `Name`.
2. **Missing Endpoint Listings**: Several active REST API routes are not documented in the current SCM `README.md`, including:
   - All `Locations` endpoints (`/api/v1/locations`).
   - Line subroutes (`/api/v1/purchase-requisitions/:id/lines`, `/api/v1/purchase-orders/:id/lines`, `/api/v1/receipts/:id/lines`, `/api/v1/shipments/:id/lines`).
   - Inventory movements endpoint (`/api/v1/inventory/movements`).
3. **Missing API Reference File**: The planned `api-reference.md` document for SCM is completely missing, leaving REST endpoints undocumented with respect to request payloads and response formats.

This PRD defines the scope to synchronize all SCM documentation under `documentation/modules/supply-chain-management/` with the current Go codebase.

---

## 2. Alignment Matrix (Variations to Resolve)

| SCM Model/Endpoint | Current Documentation Specification | Actual Code Schema |
| :--- | :--- | :--- |
| **Product** | `ID, SKU, Name, Description, CategoryID, UnitPrice, UnitCost, ReorderPoint` | `ID, ProductCode, ProductName, Description, ProductType, CategoryID, UnitOfMeasure, StandardCost, ListPrice, IsActive` |
| **ProductCategory** | `ID, Name, Description, ParentCategoryID` | `ID, Code, Name, Description` |
| **Supplier** | `ID, Code, Name, ContactPerson, Email, Phone, PaymentTerms, Status` | `ID, SupplierCode, SupplierName, ContactName, Email, Phone, IsActive` |
| **Location** | `ID, Name, Code, Type, Address` | `ID, LocationCode, LocationName, LocationType, IsActive` |
| **DemandForecast** | `ID, ProductID, PeriodStart, PeriodEnd, ForecastQuantity, ActualQuantity` | `ID, ProductID, ForecastDate, ForecastQuantity, ConfidenceLevel, Notes` |
| **Locations API** | Completely missing. | `GET/POST/PUT/DELETE /api/v1/locations` |

---

## 3. Scope & Checklist

### Phase 1: Create SCM API Reference
- [x] Create [api-reference.md](file:///Users/sithuhlaing/Projects/erp-system/documentation/modules/supply-chain-management/api-reference.md) containing:
  - Base URL configuration (Port **8003**).
  - Clear JSON schemas and request/response examples for Product Categories, Products, Locations, Vendors, Purchase Requisitions, Purchase Orders, Inventory, Stock Transfers, Receipts, Shipments, Demand Forecasts, and Reports.

### Phase 2: SCM README & Concept Update
- [x] Update SCM module [README.md](file:///Users/sithuhlaing/Projects/erp-system/documentation/modules/supply-chain-management/README.md) to reflect the correct model fields, registered services, and the full list of 47 API endpoints.
- [x] Align master [README.md](file:///Users/sithuhlaing/Projects/erp-system/documentation/modules/README.md) SCM module section with the new model schema definitions.

---

## 4. Definition of Done
- [x] Zero outdated properties (e.g. `SKU` or `ParentCategoryID`) remain in SCM documentation.
- [x] All 47 routes registered in `routes.go` have matching API definitions and documentation.
