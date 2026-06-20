package domain

import (
	"context"
	"errors"
)

var ErrOptimisticLock = errors.New("optimistic lock conflict: inventory item updated by another transaction")

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context) ([]Product, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
}

type LocationRepository interface {
	Create(ctx context.Context, loc *Location) error
	GetByID(ctx context.Context, id string) (*Location, error)
	List(ctx context.Context) ([]Location, error)
	Update(ctx context.Context, loc *Location) error
	Delete(ctx context.Context, id string) error
}

type SupplierRepository interface {
	Create(ctx context.Context, s *Supplier) error
	GetByID(ctx context.Context, id string) (*Supplier, error)
	List(ctx context.Context) ([]Supplier, error)
	Update(ctx context.Context, s *Supplier) error
	Delete(ctx context.Context, id string) error
}

type StockBalanceRepository interface {
	Create(ctx context.Context, sb *StockBalance) error
	GetByID(ctx context.Context, id string) (*StockBalance, error)
	List(ctx context.Context) ([]StockBalance, error)
	Update(ctx context.Context, sb *StockBalance) error
	GetByMaterialAndLocation(ctx context.Context, materialID string, locationID string) (*StockBalance, error)
}

type InventoryMovementRepository interface {
	Create(ctx context.Context, im *InventoryMovement) error
	GetByID(ctx context.Context, id string) (*InventoryMovement, error)
	List(ctx context.Context) ([]InventoryMovement, error)
}

type PurchaseOrderRepository interface {
	Create(ctx context.Context, po *PurchaseOrder) error
	GetByID(ctx context.Context, id string) (*PurchaseOrder, error)
	List(ctx context.Context) ([]PurchaseOrder, error)
	Update(ctx context.Context, po *PurchaseOrder) error
	Delete(ctx context.Context, id string) error
}

type PurchaseOrderLineRepository interface {
	Create(ctx context.Context, pol *PurchaseOrderLine) error
	GetByID(ctx context.Context, id string) (*PurchaseOrderLine, error)
	ListByPOID(ctx context.Context, poID string) ([]PurchaseOrderLine, error)
	DeleteByPOID(ctx context.Context, poID string) error
}

type ReceiptRepository interface {
	Create(ctx context.Context, r *Receipt) error
	GetByID(ctx context.Context, id string) (*Receipt, error)
	List(ctx context.Context) ([]Receipt, error)
	Update(ctx context.Context, r *Receipt) error
}

type ReceiptLineRepository interface {
	Create(ctx context.Context, rl *ReceiptLine) error
	GetByID(ctx context.Context, id string) (*ReceiptLine, error)
	ListByReceiptID(ctx context.Context, receiptID string) ([]ReceiptLine, error)
}

type ShipmentRepository interface {
	Create(ctx context.Context, s *Shipment) error
	GetByID(ctx context.Context, id string) (*Shipment, error)
	List(ctx context.Context) ([]Shipment, error)
	Update(ctx context.Context, s *Shipment) error
}

type ShipmentLineRepository interface {
	Create(ctx context.Context, sl *ShipmentLine) error
	GetByID(ctx context.Context, id string) (*ShipmentLine, error)
	ListByShipmentID(ctx context.Context, shipmentID string) ([]ShipmentLine, error)
}

type DemandForecastRepository interface {
	Create(ctx context.Context, df *DemandForecast) error
	GetByID(ctx context.Context, id string) (*DemandForecast, error)
	List(ctx context.Context) ([]DemandForecast, error)
	Update(ctx context.Context, df *DemandForecast) error
	ListByMaterialID(ctx context.Context, materialID string) ([]DemandForecast, error)
}

type ProductCategoryRepository interface {
	Create(ctx context.Context, pc *ProductCategory) error
	GetByID(ctx context.Context, id string) (*ProductCategory, error)
	List(ctx context.Context) ([]ProductCategory, error)
	Update(ctx context.Context, pc *ProductCategory) error
	Delete(ctx context.Context, id string) error
}

type VendorContractRepository interface {
	Create(ctx context.Context, vc *VendorContract) error
	GetByID(ctx context.Context, id string) (*VendorContract, error)
	List(ctx context.Context) ([]VendorContract, error)
	Update(ctx context.Context, vc *VendorContract) error
	Delete(ctx context.Context, id string) error
}

type PurchaseRequisitionRepository interface {
	Create(ctx context.Context, pr *PurchaseRequisition) error
	GetByID(ctx context.Context, id string) (*PurchaseRequisition, error)
	List(ctx context.Context) ([]PurchaseRequisition, error)
	Update(ctx context.Context, pr *PurchaseRequisition) error
	Delete(ctx context.Context, id string) error
}

type PurchaseRequisitionLineRepository interface {
	Create(ctx context.Context, prl *PurchaseRequisitionLine) error
	GetByID(ctx context.Context, id string) (*PurchaseRequisitionLine, error)
	ListByRequisitionID(ctx context.Context, reqID string) ([]PurchaseRequisitionLine, error)
	DeleteByRequisitionID(ctx context.Context, reqID string) error
}

type StockTransferRepository interface {
	Create(ctx context.Context, st *StockTransfer) error
	GetByID(ctx context.Context, id string) (*StockTransfer, error)
	List(ctx context.Context) ([]StockTransfer, error)
	Update(ctx context.Context, st *StockTransfer) error
}

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type KafkaEventInboxRepository interface {
	Create(ctx context.Context, e *KafkaEventInbox) error
	GetByID(ctx context.Context, id string) (*KafkaEventInbox, error)
}

type TransactionalOutboxRepository interface {
	Create(ctx context.Context, o *TransactionalOutbox) error
	GetUnsent(ctx context.Context, limit int) ([]TransactionalOutbox, error)
	UpdateStatus(ctx context.Context, id string, status OutboxStatus, retryCount int) error
}
