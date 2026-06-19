package sql

import (
	"encoding/json"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

// DefaultLegalEntityID is used as a fallback tenant ID when not provided in domain models
const DefaultLegalEntityID = "00000000-0000-0000-0000-000000000000"

// ProductCategory GORM struct
type ProductCategory struct {
	ID          string `gorm:"primaryKey"`
	Code        string `gorm:"uniqueIndex"`
	Name        string
	Description string
}

func FromDomainProductCategory(d *domain.ProductCategory) *ProductCategory {
	if d == nil {
		return nil
	}
	return &ProductCategory{
		ID:          d.ID,
		Code:        d.Code,
		Name:        d.Name,
		Description: d.Description,
	}
}

func ToDomainProductCategory(dbModel *ProductCategory) *domain.ProductCategory {
	if dbModel == nil {
		return nil
	}
	return &domain.ProductCategory{
		ID:          dbModel.ID,
		Code:        dbModel.Code,
		Name:        dbModel.Name,
		Description: dbModel.Description,
	}
}

// Product GORM struct
type Product struct {
	ID            string `gorm:"primaryKey"`
	ProductCode   string `gorm:"uniqueIndex"`
	ProductName   string
	Description   string
	ProductType   string
	CategoryID    *string `gorm:"index"`
	UnitOfMeasure string
	StandardCost  decimal.Decimal `gorm:"type:numeric(18,4)"`
	ListPrice     decimal.Decimal `gorm:"type:numeric(18,4)"`
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Category *ProductCategory `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func FromDomainProduct(d *domain.Product) *Product {
	if d == nil {
		return nil
	}
	return &Product{
		ID:            d.ID,
		ProductCode:   d.ProductCode,
		ProductName:   d.ProductName,
		Description:   d.Description,
		ProductType:   d.ProductType,
		CategoryID:    d.CategoryID,
		UnitOfMeasure: d.UnitOfMeasure,
		StandardCost:  d.StandardCost,
		ListPrice:     d.ListPrice,
		IsActive:      d.IsActive,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func ToDomainProduct(dbModel *Product) *domain.Product {
	if dbModel == nil {
		return nil
	}
	return &domain.Product{
		ID:            dbModel.ID,
		ProductCode:   dbModel.ProductCode,
		ProductName:   dbModel.ProductName,
		Description:   dbModel.Description,
		ProductType:   dbModel.ProductType,
		CategoryID:    dbModel.CategoryID,
		UnitOfMeasure: dbModel.UnitOfMeasure,
		StandardCost:  dbModel.StandardCost,
		ListPrice:     dbModel.ListPrice,
		IsActive:      dbModel.IsActive,
		CreatedAt:     dbModel.CreatedAt,
		UpdatedAt:     dbModel.UpdatedAt,
	}
}

// Location GORM struct
type Location struct {
	ID            string    `gorm:"primaryKey"`
	LegalEntityID string    `gorm:"type:uuid;not null;index:idx_tenant_wh_code,unique;default:'00000000-0000-0000-0000-000000000000'"`
	LocationCode  string    `gorm:"index:idx_tenant_wh_code,unique"`
	LocationName  string
	LocationType  string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func FromDomainLocation(d *domain.Location) *Location {
	if d == nil {
		return nil
	}
	return &Location{
		ID:            d.ID,
		LegalEntityID: DefaultLegalEntityID,
		LocationCode:  d.LocationCode,
		LocationName:  d.LocationName,
		LocationType:  d.LocationType,
		IsActive:      d.IsActive,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func ToDomainLocation(dbModel *Location) *domain.Location {
	if dbModel == nil {
		return nil
	}
	return &domain.Location{
		ID:           dbModel.ID,
		LocationCode: dbModel.LocationCode,
		LocationName: dbModel.LocationName,
		LocationType: dbModel.LocationType,
		IsActive:     dbModel.IsActive,
		CreatedAt:    dbModel.CreatedAt,
		UpdatedAt:    dbModel.UpdatedAt,
	}
}

// Supplier GORM struct
type Supplier struct {
	ID            string    `gorm:"primaryKey"`
	LegalEntityID string    `gorm:"type:uuid;not null;index:idx_tenant_sup_code,unique;default:'00000000-0000-0000-0000-000000000000'"`
	SupplierCode  string    `gorm:"index:idx_tenant_sup_code,unique"`
	SupplierName  string
	ContactName   string
	Email         string
	Phone         string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func FromDomainSupplier(d *domain.Supplier) *Supplier {
	if d == nil {
		return nil
	}
	return &Supplier{
		ID:            d.ID,
		LegalEntityID: DefaultLegalEntityID,
		SupplierCode:  d.SupplierCode,
		SupplierName:  d.SupplierName,
		ContactName:   d.ContactName,
		Email:         d.Email,
		Phone:         d.Phone,
		IsActive:      d.IsActive,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func ToDomainSupplier(dbModel *Supplier) *domain.Supplier {
	if dbModel == nil {
		return nil
	}
	return &domain.Supplier{
		ID:           dbModel.ID,
		SupplierCode: dbModel.SupplierCode,
		SupplierName: dbModel.SupplierName,
		ContactName:  dbModel.ContactName,
		Email:        dbModel.Email,
		Phone:        dbModel.Phone,
		IsActive:     dbModel.IsActive,
		CreatedAt:    dbModel.CreatedAt,
		UpdatedAt:    dbModel.UpdatedAt,
	}
}

// VendorContract GORM struct
type VendorContract struct {
	ID             string    `gorm:"primaryKey"`
	ContractNumber string    `gorm:"uniqueIndex"`
	SupplierID     string    `gorm:"index"`
	StartDate      time.Time
	EndDate        time.Time
	Terms          string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Supplier Supplier `gorm:"foreignKey:SupplierID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainVendorContract(d *domain.VendorContract) *VendorContract {
	if d == nil {
		return nil
	}
	return &VendorContract{
		ID:             d.ID,
		ContractNumber: d.ContractNumber,
		SupplierID:     d.SupplierID,
		StartDate:      d.StartDate,
		EndDate:        d.EndDate,
		Terms:          d.Terms,
		Status:         d.Status,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func ToDomainVendorContract(dbModel *VendorContract) *domain.VendorContract {
	if dbModel == nil {
		return nil
	}
	return &domain.VendorContract{
		ID:             dbModel.ID,
		ContractNumber: dbModel.ContractNumber,
		SupplierID:     dbModel.SupplierID,
		StartDate:      dbModel.StartDate,
		EndDate:        dbModel.EndDate,
		Terms:          dbModel.Terms,
		Status:         dbModel.Status,
		CreatedAt:      dbModel.CreatedAt,
		UpdatedAt:      dbModel.UpdatedAt,
	}
}

// InventoryItem GORM struct
type InventoryItem struct {
	ID                string          `gorm:"primaryKey"`
	LegalEntityID     string          `gorm:"type:uuid;not null;index:idx_tenant_wh_mat,unique;default:'00000000-0000-0000-0000-000000000000'"`
	ProductID         string          `gorm:"index:idx_tenant_wh_mat,unique"`
	LocationID        string          `gorm:"index:idx_tenant_wh_mat,unique"`
	QuantityOnHand    int
	QuantityReserved  int
	QuantityAvailable int
	ReorderPoint      int
	MaximumStock      int
	UnitCost          decimal.Decimal `gorm:"type:numeric(18,4)"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Version           int32           `gorm:"type:integer;not null;default:0"` // OCC concurrency shield

	Product  Product  `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Location Location `gorm:"foreignKey:LocationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainInventoryItem(d *domain.InventoryItem) *InventoryItem {
	if d == nil {
		return nil
	}
	return &InventoryItem{
		ID:                d.ID,
		LegalEntityID:     DefaultLegalEntityID,
		ProductID:         d.ProductID,
		LocationID:        d.LocationID,
		QuantityOnHand:    d.QuantityOnHand,
		QuantityReserved:  d.QuantityReserved,
		QuantityAvailable: d.QuantityAvailable,
		ReorderPoint:      d.ReorderPoint,
		MaximumStock:      d.MaximumStock,
		UnitCost:          d.UnitCost,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
		Version:           0,
	}
}

func ToDomainInventoryItem(dbModel *InventoryItem) *domain.InventoryItem {
	if dbModel == nil {
		return nil
	}
	return &domain.InventoryItem{
		ID:                dbModel.ID,
		ProductID:         dbModel.ProductID,
		LocationID:        dbModel.LocationID,
		QuantityOnHand:    dbModel.QuantityOnHand,
		QuantityReserved:  dbModel.QuantityReserved,
		QuantityAvailable: dbModel.QuantityAvailable,
		ReorderPoint:      dbModel.ReorderPoint,
		MaximumStock:      dbModel.MaximumStock,
		UnitCost:          dbModel.UnitCost,
		CreatedAt:         dbModel.CreatedAt,
		UpdatedAt:         dbModel.UpdatedAt,
	}
}

// InventoryMovement GORM struct
type InventoryMovement struct {
	ID            string          `gorm:"primaryKey"`
	LegalEntityID string          `gorm:"type:uuid;not null;index;default:'00000000-0000-0000-0000-000000000000'"`
	ProductID     string          `gorm:"index"`
	LocationID    string          `gorm:"index"`
	MovementType  string          // e.g. RECEIPT, ISSUE, TRANSFER, ADJUSTMENT
	Quantity      int
	UnitCost      decimal.Decimal `gorm:"type:numeric(18,4)"`
	ReferenceType string          // e.g. MANUAL_ADJUSTMENT, PO_RECEIPT, STOCK_TRANSFER, SHIPMENT
	ReferenceID   string
	Notes         string
	CreatedAt     time.Time

	Product  Product  `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Location Location `gorm:"foreignKey:LocationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainInventoryMovement(d *domain.InventoryMovement) *InventoryMovement {
	if d == nil {
		return nil
	}
	return &InventoryMovement{
		ID:            d.ID,
		LegalEntityID: DefaultLegalEntityID,
		ProductID:     d.ProductID,
		LocationID:    d.LocationID,
		MovementType:  d.MovementType,
		Quantity:      d.Quantity,
		UnitCost:      d.UnitCost,
		ReferenceType: d.ReferenceType,
		ReferenceID:   d.ReferenceID,
		Notes:         d.Notes,
		CreatedAt:     d.CreatedAt,
	}
}

func ToDomainInventoryMovement(dbModel *InventoryMovement) *domain.InventoryMovement {
	if dbModel == nil {
		return nil
	}
	return &domain.InventoryMovement{
		ID:            dbModel.ID,
		ProductID:     dbModel.ProductID,
		LocationID:    dbModel.LocationID,
		MovementType:  dbModel.MovementType,
		Quantity:      dbModel.Quantity,
		UnitCost:      dbModel.UnitCost,
		ReferenceType: dbModel.ReferenceType,
		ReferenceID:   dbModel.ReferenceID,
		Notes:         dbModel.Notes,
		CreatedAt:     dbModel.CreatedAt,
	}
}

// StockTransfer GORM struct
type StockTransfer struct {
	ID             string     `gorm:"primaryKey"`
	FromLocationID string     `gorm:"index"`
	ToLocationID   string     `gorm:"index"`
	ProductID      string     `gorm:"index"`
	Quantity       int
	Status         string // e.g. PENDING, TRANSFERRED, CANCELLED
	TransferredAt  *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	FromLocation Location `gorm:"foreignKey:FromLocationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ToLocation   Location `gorm:"foreignKey:ToLocationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Product      Product  `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainStockTransfer(d *domain.StockTransfer) *StockTransfer {
	if d == nil {
		return nil
	}
	return &StockTransfer{
		ID:             d.ID,
		FromLocationID: d.FromLocationID,
		ToLocationID:   d.ToLocationID,
		ProductID:      d.ProductID,
		Quantity:       d.Quantity,
		Status:         d.Status,
		TransferredAt:  d.TransferredAt,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func ToDomainStockTransfer(dbModel *StockTransfer) *domain.StockTransfer {
	if dbModel == nil {
		return nil
	}
	return &domain.StockTransfer{
		ID:             dbModel.ID,
		FromLocationID: dbModel.FromLocationID,
		ToLocationID:   dbModel.ToLocationID,
		ProductID:      dbModel.ProductID,
		Quantity:       dbModel.Quantity,
		Status:         dbModel.Status,
		TransferredAt:  dbModel.TransferredAt,
		CreatedAt:      dbModel.CreatedAt,
		UpdatedAt:      dbModel.UpdatedAt,
	}
}

// PurchaseRequisition GORM struct
type PurchaseRequisition struct {
	ID          string `gorm:"primaryKey"`
	ReqNumber   string `gorm:"uniqueIndex"`
	RequesterID string
	RequestDate time.Time
	Status      string
	TotalAmount decimal.Decimal `gorm:"type:numeric(18,4)"`
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func FromDomainPurchaseRequisition(d *domain.PurchaseRequisition) *PurchaseRequisition {
	if d == nil {
		return nil
	}
	return &PurchaseRequisition{
		ID:          d.ID,
		ReqNumber:   d.ReqNumber,
		RequesterID: d.RequesterID,
		RequestDate: d.RequestDate,
		Status:      d.Status,
		TotalAmount: d.TotalAmount,
		Notes:       d.Notes,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func ToDomainPurchaseRequisition(dbModel *PurchaseRequisition) *domain.PurchaseRequisition {
	if dbModel == nil {
		return nil
	}
	return &domain.PurchaseRequisition{
		ID:          dbModel.ID,
		ReqNumber:   dbModel.ReqNumber,
		RequesterID: dbModel.RequesterID,
		RequestDate: dbModel.RequestDate,
		Status:      dbModel.Status,
		TotalAmount: dbModel.TotalAmount,
		Notes:       dbModel.Notes,
		CreatedAt:   dbModel.CreatedAt,
		UpdatedAt:   dbModel.UpdatedAt,
	}
}

// PurchaseRequisitionLine GORM struct
type PurchaseRequisitionLine struct {
	ID                    string `gorm:"primaryKey"`
	PurchaseRequisitionID string `gorm:"index"`
	ProductID             string `gorm:"index"`
	QuantityRequested     int
	EstimatedUnitPrice    decimal.Decimal `gorm:"type:numeric(18,4)"`
	LineTotal             decimal.Decimal `gorm:"type:numeric(18,4)"`

	PurchaseRequisition PurchaseRequisition `gorm:"foreignKey:PurchaseRequisitionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Product             Product             `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainPurchaseRequisitionLine(d *domain.PurchaseRequisitionLine) *PurchaseRequisitionLine {
	if d == nil {
		return nil
	}
	return &PurchaseRequisitionLine{
		ID:                    d.ID,
		PurchaseRequisitionID: d.PurchaseRequisitionID,
		ProductID:             d.ProductID,
		QuantityRequested:     d.QuantityRequested,
		EstimatedUnitPrice:    d.EstimatedUnitPrice,
		LineTotal:             d.LineTotal,
	}
}

func ToDomainPurchaseRequisitionLine(dbModel *PurchaseRequisitionLine) *domain.PurchaseRequisitionLine {
	if dbModel == nil {
		return nil
	}
	return &domain.PurchaseRequisitionLine{
		ID:                    dbModel.ID,
		PurchaseRequisitionID: dbModel.PurchaseRequisitionID,
		ProductID:             dbModel.ProductID,
		QuantityRequested:     dbModel.QuantityRequested,
		EstimatedUnitPrice:    dbModel.EstimatedUnitPrice,
		LineTotal:             dbModel.LineTotal,
	}
}

// PurchaseOrder GORM struct
type PurchaseOrder struct {
	ID               string    `gorm:"primaryKey"`
	LegalEntityID    string    `gorm:"type:uuid;not null;index;default:'00000000-0000-0000-0000-000000000000'"`
	PoNumber         string    `gorm:"uniqueIndex"`
	SupplierID       string    `gorm:"index"`
	OrderDate        time.Time
	ExpectedDelivery time.Time
	Status           string
	TotalAmount      decimal.Decimal `gorm:"type:numeric(18,4)"`
	Notes            string
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Supplier Supplier `gorm:"foreignKey:SupplierID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainPurchaseOrder(d *domain.PurchaseOrder) *PurchaseOrder {
	if d == nil {
		return nil
	}
	return &PurchaseOrder{
		ID:               d.ID,
		LegalEntityID:    DefaultLegalEntityID,
		PoNumber:         d.PoNumber,
		SupplierID:       d.SupplierID,
		OrderDate:        d.OrderDate,
		ExpectedDelivery: d.ExpectedDelivery,
		Status:           d.Status,
		TotalAmount:      d.TotalAmount,
		Notes:            d.Notes,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

func ToDomainPurchaseOrder(dbModel *PurchaseOrder) *domain.PurchaseOrder {
	if dbModel == nil {
		return nil
	}
	return &domain.PurchaseOrder{
		ID:               dbModel.ID,
		PoNumber:         dbModel.PoNumber,
		SupplierID:       dbModel.SupplierID,
		OrderDate:        dbModel.OrderDate,
		ExpectedDelivery: dbModel.ExpectedDelivery,
		Status:           dbModel.Status,
		TotalAmount:      dbModel.TotalAmount,
		Notes:            dbModel.Notes,
		CreatedAt:        dbModel.CreatedAt,
		UpdatedAt:        dbModel.UpdatedAt,
	}
}

// PurchaseOrderLine GORM struct
type PurchaseOrderLine struct {
	ID               string `gorm:"primaryKey"`
	PurchaseOrderID  string `gorm:"index"`
	ProductID        string `gorm:"index"`
	QuantityOrdered  int
	QuantityReceived int
	UnitPrice        decimal.Decimal `gorm:"type:numeric(18,4)"`
	LineTotal        decimal.Decimal `gorm:"type:numeric(18,4)"`
	Description      string
	CreatedAt        time.Time

	PurchaseOrder PurchaseOrder `gorm:"foreignKey:PurchaseOrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Product       Product       `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainPurchaseOrderLine(d *domain.PurchaseOrderLine) *PurchaseOrderLine {
	if d == nil {
		return nil
	}
	return &PurchaseOrderLine{
		ID:               d.ID,
		PurchaseOrderID:  d.PurchaseOrderID,
		ProductID:        d.ProductID,
		QuantityOrdered:  d.QuantityOrdered,
		QuantityReceived: d.QuantityReceived,
		UnitPrice:        d.UnitPrice,
		LineTotal:        d.LineTotal,
		Description:      d.Description,
		CreatedAt:        d.CreatedAt,
	}
}

func ToDomainPurchaseOrderLine(dbModel *PurchaseOrderLine) *domain.PurchaseOrderLine {
	if dbModel == nil {
		return nil
	}
	return &domain.PurchaseOrderLine{
		ID:               dbModel.ID,
		PurchaseOrderID:  dbModel.PurchaseOrderID,
		ProductID:        dbModel.ProductID,
		QuantityOrdered:  dbModel.QuantityOrdered,
		QuantityReceived: dbModel.QuantityReceived,
		UnitPrice:        dbModel.UnitPrice,
		LineTotal:        dbModel.LineTotal,
		Description:      dbModel.Description,
		CreatedAt:        dbModel.CreatedAt,
	}
}

// Receipt GORM struct
type Receipt struct {
	ID              string    `gorm:"primaryKey"`
	ReceiptNumber   string    `gorm:"uniqueIndex"`
	PurchaseOrderID *string   `gorm:"index"`
	ReceivedDate    time.Time
	Status          string
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time

	PurchaseOrder *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func FromDomainReceipt(d *domain.Receipt) *Receipt {
	if d == nil {
		return nil
	}
	return &Receipt{
		ID:              d.ID,
		ReceiptNumber:   d.ReceiptNumber,
		PurchaseOrderID: d.PurchaseOrderID,
		ReceivedDate:    d.ReceivedDate,
		Status:          d.Status,
		Notes:           d.Notes,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func ToDomainReceipt(dbModel *Receipt) *domain.Receipt {
	if dbModel == nil {
		return nil
	}
	return &domain.Receipt{
		ID:              dbModel.ID,
		ReceiptNumber:   dbModel.ReceiptNumber,
		PurchaseOrderID: dbModel.PurchaseOrderID,
		ReceivedDate:    dbModel.ReceivedDate,
		Status:          dbModel.Status,
		Notes:           dbModel.Notes,
		CreatedAt:       dbModel.CreatedAt,
		UpdatedAt:       dbModel.UpdatedAt,
	}
}

// ReceiptLine GORM struct
type ReceiptLine struct {
	ID               string `gorm:"primaryKey"`
	ReceiptID        string `gorm:"index"`
	ProductID        string `gorm:"index"`
	QuantityReceived int
	UnitCost         decimal.Decimal `gorm:"type:numeric(18,4)"`
	CreatedAt        time.Time

	Receipt Receipt `gorm:"foreignKey:ReceiptID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Product Product `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainReceiptLine(d *domain.ReceiptLine) *ReceiptLine {
	if d == nil {
		return nil
	}
	return &ReceiptLine{
		ID:               d.ID,
		ReceiptID:        d.ReceiptID,
		ProductID:        d.ProductID,
		QuantityReceived: d.QuantityReceived,
		UnitCost:         d.UnitCost,
		CreatedAt:        d.CreatedAt,
	}
}

func ToDomainReceiptLine(dbModel *ReceiptLine) *domain.ReceiptLine {
	if dbModel == nil {
		return nil
	}
	return &domain.ReceiptLine{
		ID:               dbModel.ID,
		ReceiptID:        dbModel.ReceiptID,
		ProductID:        dbModel.ProductID,
		QuantityReceived: dbModel.QuantityReceived,
		UnitCost:         dbModel.UnitCost,
		CreatedAt:        dbModel.CreatedAt,
	}
}

// Shipment GORM struct
type Shipment struct {
	ID                string    `gorm:"primaryKey"`
	ShipmentNumber    string    `gorm:"uniqueIndex"`
	SalesOrderID      *string   `gorm:"index"`
	Carrier           string
	TrackingNumber    string
	ShippedDate       time.Time
	EstimatedDelivery time.Time
	Status            string
	Notes             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func FromDomainShipment(d *domain.Shipment) *Shipment {
	if d == nil {
		return nil
	}
	return &Shipment{
		ID:                d.ID,
		ShipmentNumber:    d.ShipmentNumber,
		SalesOrderID:      d.SalesOrderID,
		Carrier:           d.Carrier,
		TrackingNumber:    d.TrackingNumber,
		ShippedDate:       d.ShippedDate,
		EstimatedDelivery: d.EstimatedDelivery,
		Status:            d.Status,
		Notes:             d.Notes,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

func ToDomainShipment(dbModel *Shipment) *domain.Shipment {
	if dbModel == nil {
		return nil
	}
	return &domain.Shipment{
		ID:                dbModel.ID,
		ShipmentNumber:    dbModel.ShipmentNumber,
		SalesOrderID:      dbModel.SalesOrderID,
		Carrier:           dbModel.Carrier,
		TrackingNumber:    dbModel.TrackingNumber,
		ShippedDate:       dbModel.ShippedDate,
		EstimatedDelivery: dbModel.EstimatedDelivery,
		Status:            dbModel.Status,
		Notes:             dbModel.Notes,
		CreatedAt:         dbModel.CreatedAt,
		UpdatedAt:         dbModel.UpdatedAt,
	}
}

// ShipmentLine GORM struct
type ShipmentLine struct {
	ID              string `gorm:"primaryKey"`
	ShipmentID      string `gorm:"index"`
	ProductID       string `gorm:"index"`
	QuantityShipped int
	CreatedAt       time.Time

	Shipment Shipment `gorm:"foreignKey:ShipmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Product  Product  `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainShipmentLine(d *domain.ShipmentLine) *ShipmentLine {
	if d == nil {
		return nil
	}
	return &ShipmentLine{
		ID:              d.ID,
		ShipmentID:      d.ShipmentID,
		ProductID:       d.ProductID,
		QuantityShipped: d.QuantityShipped,
		CreatedAt:       d.CreatedAt,
	}
}

func ToDomainShipmentLine(dbModel *ShipmentLine) *domain.ShipmentLine {
	if dbModel == nil {
		return nil
	}
	return &domain.ShipmentLine{
		ID:              dbModel.ID,
		ShipmentID:      dbModel.ShipmentID,
		ProductID:       dbModel.ProductID,
		QuantityShipped: dbModel.QuantityShipped,
		CreatedAt:       dbModel.CreatedAt,
	}
}

// DemandForecast GORM struct
type DemandForecast struct {
	ID               string `gorm:"primaryKey"`
	ProductID        string `gorm:"index"`
	ForecastDate     time.Time
	ForecastQuantity int
	ConfidenceLevel  decimal.Decimal `gorm:"type:numeric(18,4)"`
	Notes            string
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Product Product `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainDemandForecast(d *domain.DemandForecast) *DemandForecast {
	if d == nil {
		return nil
	}
	return &DemandForecast{
		ID:               d.ID,
		ProductID:        d.ProductID,
		ForecastDate:     d.ForecastDate,
		ForecastQuantity: d.ForecastQuantity,
		ConfidenceLevel:  d.ConfidenceLevel,
		Notes:            d.Notes,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

func ToDomainDemandForecast(dbModel *DemandForecast) *domain.DemandForecast {
	if dbModel == nil {
		return nil
	}
	return &domain.DemandForecast{
		ID:               dbModel.ID,
		ProductID:        dbModel.ProductID,
		ForecastDate:     dbModel.ForecastDate,
		ForecastQuantity: dbModel.ForecastQuantity,
		ConfidenceLevel:  dbModel.ConfidenceLevel,
		Notes:            dbModel.Notes,
		CreatedAt:        dbModel.CreatedAt,
		UpdatedAt:        dbModel.UpdatedAt,
	}
}

// KafkaEventInbox GORM struct
type KafkaEventInbox struct {
	AttemptCount     int       `gorm:"type:integer;default:0;not null"`
	EventID          string `gorm:"primaryKey"`
	EventType        string
	ProcessedAt      time.Time
	ProcessingStatus domain.EventProcessingStatus `gorm:"type:varchar(50)"`
	Payload          string                       `gorm:"type:text"`
}

func FromDomainKafkaEventInbox(d *domain.KafkaEventInbox) *KafkaEventInbox {
	if d == nil {
		return nil
	}
	return &KafkaEventInbox{
		AttemptCount:     d.AttemptCount,
		EventID:          d.EventID,
		EventType:        d.EventType,
		ProcessedAt:      d.ProcessedAt,
		ProcessingStatus: d.ProcessingStatus,
		Payload:          d.Payload,
	}
}

func ToDomainKafkaEventInbox(dbModel *KafkaEventInbox) *domain.KafkaEventInbox {
	if dbModel == nil {
		return nil
	}
	return &domain.KafkaEventInbox{
		AttemptCount:     dbModel.AttemptCount,
		EventID:          dbModel.EventID,
		EventType:        dbModel.EventType,
		ProcessedAt:      dbModel.ProcessedAt,
		ProcessingStatus: dbModel.ProcessingStatus,
		Payload:          dbModel.Payload,
	}
}

// TransactionalOutbox GORM struct
type TransactionalOutbox struct {
	ID          string `gorm:"primaryKey"`
	EventType   string
	AggregateID string
	Payload     []byte              `gorm:"type:jsonb"`
	Status      domain.OutboxStatus `gorm:"type:varchar(50)"`
	RetryCount  int
	CreatedAt   time.Time
}

func FromDomainTransactionalOutbox(d *domain.TransactionalOutbox) *TransactionalOutbox {
	if d == nil {
		return nil
	}
	pBytes, _ := json.Marshal(d.Payload)
	return &TransactionalOutbox{
		ID:          d.ID,
		EventType:   d.EventType,
		AggregateID: d.AggregateID,
		Payload:     pBytes,
		Status:      d.Status,
		RetryCount:  d.RetryCount,
		CreatedAt:   d.CreatedAt,
	}
}

func ToDomainTransactionalOutbox(dbModel *TransactionalOutbox) *domain.TransactionalOutbox {
	if dbModel == nil {
		return nil
	}
	var p interface{}
	if len(dbModel.Payload) > 0 {
		_ = json.Unmarshal(dbModel.Payload, &p)
	}
	return &domain.TransactionalOutbox{
		ID:          dbModel.ID,
		EventType:   dbModel.EventType,
		AggregateID: dbModel.AggregateID,
		Payload:     p,
		Status:      dbModel.Status,
		RetryCount:  dbModel.RetryCount,
		CreatedAt:   dbModel.CreatedAt,
	}
}
