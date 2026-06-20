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
		ContactName:   "",
		Email:         "",
		Phone:         "",
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
		ID:            dbModel.ID,
		LegalEntityID: dbModel.LegalEntityID,
		SupplierCode:  dbModel.SupplierCode,
		SupplierName:  dbModel.SupplierName,
		IsActive:      dbModel.IsActive,
		CreatedAt:     dbModel.CreatedAt,
		UpdatedAt:     dbModel.UpdatedAt,
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
		Terms:          "",
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
		LegalEntityID:  DefaultLegalEntityID,
		ContractNumber: dbModel.ContractNumber,
		SupplierID:     dbModel.SupplierID,
		StartDate:      dbModel.StartDate,
		EndDate:        dbModel.EndDate,
		Status:         dbModel.Status,
		CreatedAt:      dbModel.CreatedAt,
		UpdatedAt:      dbModel.UpdatedAt,
	}
}

// StockBalance GORM struct
type StockBalance struct {
	ID                string          `gorm:"primaryKey"`
	LegalEntityID     string          `gorm:"type:uuid;not null;index:idx_tenant_sb_loc_mat,unique;default:'00000000-0000-0000-0000-000000000000'"`
	MaterialID        string          `gorm:"index:idx_tenant_sb_loc_mat,unique"`
	LocationID        string          `gorm:"index:idx_tenant_sb_loc_mat,unique"`
	QuantityOnHand    decimal.Decimal `gorm:"type:numeric(18,4)"`
	QuantityReserved  decimal.Decimal `gorm:"type:numeric(18,4)"`
	QuantityAvailable decimal.Decimal `gorm:"type:numeric(18,4)"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Version           int             `gorm:"type:integer;not null;default:0"` // OCC concurrency shield

	Location Location `gorm:"foreignKey:LocationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func (StockBalance) TableName() string {
	return "scm_stock_balances"
}

func FromDomainStockBalance(d *domain.StockBalance) *StockBalance {
	if d == nil {
		return nil
	}
	return &StockBalance{
		ID:                d.ID,
		LegalEntityID:     d.LegalEntityID,
		MaterialID:        d.MaterialID,
		LocationID:        d.LocationID,
		QuantityOnHand:    d.QuantityOnHand,
		QuantityReserved:  d.QuantityReserved,
		QuantityAvailable: d.QuantityAvailable,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
		Version:           d.Version,
	}
}

func ToDomainStockBalance(dbModel *StockBalance) *domain.StockBalance {
	if dbModel == nil {
		return nil
	}
	return &domain.StockBalance{
		ID:                dbModel.ID,
		LegalEntityID:     dbModel.LegalEntityID,
		LocationID:        dbModel.LocationID,
		MaterialID:        dbModel.MaterialID,
		QuantityOnHand:    dbModel.QuantityOnHand,
		QuantityReserved:  dbModel.QuantityReserved,
		QuantityAvailable: dbModel.QuantityAvailable,
		Version:           dbModel.Version,
		CreatedAt:         dbModel.CreatedAt,
		UpdatedAt:         dbModel.UpdatedAt,
	}
}

// InventoryMovement GORM struct
type InventoryMovement struct {
	ID            string          `gorm:"primaryKey"`
	LegalEntityID string          `gorm:"type:uuid;not null;index;default:'00000000-0000-0000-0000-000000000000'"`
	MaterialID    string          `gorm:"index"`
	LocationID    string          `gorm:"index"`
	MovementType  string          // e.g. RECEIPT, ISSUE, TRANSFER, ADJUSTMENT
	Quantity      decimal.Decimal `gorm:"type:numeric(18,4)"`
	UnitCost      decimal.Decimal `gorm:"type:numeric(18,4)"`
	ReferenceType string          // e.g. MANUAL_ADJUSTMENT, PO_RECEIPT, STOCK_TRANSFER, SHIPMENT
	ReferenceID   string
	Notes         string
	CreatedAt     time.Time

	Location Location `gorm:"foreignKey:LocationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainInventoryMovement(d *domain.InventoryMovement) *InventoryMovement {
	if d == nil {
		return nil
	}
	return &InventoryMovement{
		ID:            d.ID,
		LegalEntityID: d.LegalEntityID,
		MaterialID:    d.MaterialID,
		LocationID:    d.LocationID,
		MovementType:  d.MovementType,
		Quantity:      d.Quantity,
		UnitCost:      decimal.Zero, // no longer in domain
		ReferenceType: d.ReferenceType,
		ReferenceID:   d.ReferenceID,
		Notes:         "", // no longer in domain
		CreatedAt:     d.CreatedAt,
	}
}

func ToDomainInventoryMovement(dbModel *InventoryMovement) *domain.InventoryMovement {
	if dbModel == nil {
		return nil
	}
	return &domain.InventoryMovement{
		ID:            dbModel.ID,
		LegalEntityID: dbModel.LegalEntityID,
		LocationID:    dbModel.LocationID,
		MaterialID:    dbModel.MaterialID,
		MovementType:  dbModel.MovementType,
		Quantity:      dbModel.Quantity,
		ReferenceType: dbModel.ReferenceType,
		ReferenceID:   dbModel.ReferenceID,
		CreatedAt:     dbModel.CreatedAt,
	}
}

// StockTransfer GORM struct
type StockTransfer struct {
	ID             string     `gorm:"primaryKey"`
	FromLocationID string     `gorm:"index"`
	ToLocationID   string     `gorm:"index"`
	MaterialID     string     `gorm:"index"`
	Quantity       decimal.Decimal `gorm:"type:numeric(18,4)"`
	Status         string // e.g. PENDING, TRANSFERRED, CANCELLED
	Version        int             `gorm:"type:integer;not null;default:0"` // OCC concurrency shield
	TransferredAt  *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	FromLocation Location `gorm:"foreignKey:FromLocationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ToLocation   Location `gorm:"foreignKey:ToLocationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Product      Product  `gorm:"foreignKey:MaterialID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainStockTransfer(d *domain.StockTransfer) *StockTransfer {
	if d == nil {
		return nil
	}
	return &StockTransfer{
		ID:             d.ID,
		FromLocationID: d.FromLocationID,
		ToLocationID:   d.ToLocationID,
		MaterialID:     d.MaterialID,
		Quantity:       d.Quantity,
		Status:         d.Status,
		Version:        d.Version,
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
		MaterialID:     dbModel.MaterialID,
		Quantity:       dbModel.Quantity,
		Status:         dbModel.Status,
		Version:        dbModel.Version,
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
	MaterialID            string `gorm:"index"`
	QuantityRequested     decimal.Decimal `gorm:"type:numeric(18,4)"`
	EstimatedUnitPrice    decimal.Decimal `gorm:"type:numeric(18,4)"`
	LineTotal             decimal.Decimal `gorm:"type:numeric(18,4)"`

	PurchaseRequisition PurchaseRequisition `gorm:"foreignKey:PurchaseRequisitionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Product             Product             `gorm:"foreignKey:MaterialID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainPurchaseRequisitionLine(d *domain.PurchaseRequisitionLine) *PurchaseRequisitionLine {
	if d == nil {
		return nil
	}
	return &PurchaseRequisitionLine{
		ID:                    d.ID,
		PurchaseRequisitionID: d.PurchaseRequisitionID,
		MaterialID:            d.MaterialID,
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
		MaterialID:            dbModel.MaterialID,
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
		Status:           string(d.Status),
		TotalAmount:      d.TotalAmount,
		Notes:            "",
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
		Status:           domain.PurchaseOrderStatus(dbModel.Status),
		TotalAmount:      dbModel.TotalAmount,
		CreatedAt:        dbModel.CreatedAt,
		UpdatedAt:        dbModel.UpdatedAt,
	}
}

// PurchaseOrderLine GORM struct
type PurchaseOrderLine struct {
	ID               string `gorm:"primaryKey"`
	PurchaseOrderID  string `gorm:"index"`
	MaterialID       string `gorm:"index"`
	QuantityOrdered  decimal.Decimal `gorm:"type:numeric(18,4)"`
	QuantityReceived decimal.Decimal `gorm:"type:numeric(18,4)"`
	UnitPrice        decimal.Decimal `gorm:"type:numeric(18,4)"`
	LineTotal        decimal.Decimal `gorm:"type:numeric(18,4)"`
	Description      string
	CreatedAt        time.Time

	PurchaseOrder PurchaseOrder `gorm:"foreignKey:PurchaseOrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Product       Product       `gorm:"foreignKey:MaterialID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainPurchaseOrderLine(d *domain.PurchaseOrderLine) *PurchaseOrderLine {
	if d == nil {
		return nil
	}
	return &PurchaseOrderLine{
		ID:               d.ID,
		PurchaseOrderID:  d.PurchaseOrderID,
		MaterialID:       d.MaterialID,
		QuantityOrdered:  d.QuantityOrdered,
		QuantityReceived: d.QuantityReceived,
		UnitPrice:        d.UnitPrice,
		LineTotal:        d.LineTotal,
		Description:      "",
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
		MaterialID:       dbModel.MaterialID,
		QuantityOrdered:  dbModel.QuantityOrdered,
		QuantityReceived: dbModel.QuantityReceived,
		UnitPrice:        dbModel.UnitPrice,
		LineTotal:        dbModel.LineTotal,
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
	var poID *string
	if d.PurchaseOrderID != "" {
		poID = &d.PurchaseOrderID
	}
	return &Receipt{
		ID:              d.ID,
		ReceiptNumber:   d.ReceiptNumber,
		PurchaseOrderID: poID,
		ReceivedDate:    d.ReceivedDate,
		Status:          d.Status,
		Notes:           "",
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func ToDomainReceipt(dbModel *Receipt) *domain.Receipt {
	if dbModel == nil {
		return nil
	}
	var poID string
	if dbModel.PurchaseOrderID != nil {
		poID = *dbModel.PurchaseOrderID
	}
	return &domain.Receipt{
		ID:              dbModel.ID,
		ReceiptNumber:   dbModel.ReceiptNumber,
		PurchaseOrderID: poID,
		ReceivedDate:    dbModel.ReceivedDate,
		Status:          dbModel.Status,
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
	var soID *string
	if d.SalesOrderID != "" {
		soID = &d.SalesOrderID
	}
	return &Shipment{
		ID:                d.ID,
		ShipmentNumber:    d.ShipmentNumber,
		SalesOrderID:      soID,
		Carrier:           d.Carrier,
		TrackingNumber:    d.TrackingNumber,
		ShippedDate:       d.ShippedDate,
		EstimatedDelivery: time.Time{},
		Status:            d.Status,
		Notes:             "",
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

func ToDomainShipment(dbModel *Shipment) *domain.Shipment {
	if dbModel == nil {
		return nil
	}
	var soID string
	if dbModel.SalesOrderID != nil {
		soID = *dbModel.SalesOrderID
	}
	return &domain.Shipment{
		ID:             dbModel.ID,
		ShipmentNumber: dbModel.ShipmentNumber,
		SalesOrderID:   soID,
		Carrier:        dbModel.Carrier,
		TrackingNumber: dbModel.TrackingNumber,
		ShippedDate:    dbModel.ShippedDate,
		Status:         dbModel.Status,
		CreatedAt:      dbModel.CreatedAt,
		UpdatedAt:      dbModel.UpdatedAt,
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
	ID               string          `gorm:"primaryKey"`
	MaterialID       string          `gorm:"index"`
	ForecastDate     time.Time
	ForecastQuantity decimal.Decimal `gorm:"type:numeric(18,4)"`
	ConfidenceLevel  decimal.Decimal `gorm:"type:numeric(18,4)"`
	Notes            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func FromDomainDemandForecast(d *domain.DemandForecast) *DemandForecast {
	if d == nil {
		return nil
	}
	return &DemandForecast{
		ID:               d.ID,
		MaterialID:       d.MaterialID,
		ForecastDate:     d.ForecastDate,
		ForecastQuantity: d.ForecastQuantity,
		ConfidenceLevel:  d.ConfidenceLevel,
		Notes:            "",
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
		LegalEntityID:    DefaultLegalEntityID,
		MaterialID:       dbModel.MaterialID,
		ForecastDate:     dbModel.ForecastDate,
		ForecastQuantity: dbModel.ForecastQuantity,
		ConfidenceLevel:  dbModel.ConfidenceLevel,
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
	var payloadStr string
	if d.Payload != nil {
		if s, ok := d.Payload.(string); ok {
			payloadStr = s
		} else {
			bytes, _ := json.Marshal(d.Payload)
			payloadStr = string(bytes)
		}
	}
	return &KafkaEventInbox{
		AttemptCount:     0,
		EventID:          d.EventID,
		EventType:        d.EventType,
		ProcessedAt:      d.ProcessedAt,
		ProcessingStatus: d.ProcessingStatus,
		Payload:          payloadStr,
	}
}

func ToDomainKafkaEventInbox(dbModel *KafkaEventInbox) *domain.KafkaEventInbox {
	if dbModel == nil {
		return nil
	}
	return &domain.KafkaEventInbox{
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
