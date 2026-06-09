package domain

import "context"

type BillOfMaterialsRepository interface {
	Create(ctx context.Context, bom *BillOfMaterials) error
	GetByID(ctx context.Context, id string) (*BillOfMaterials, error)
	GetByProductID(ctx context.Context, productID string) (*BillOfMaterials, error)
	List(ctx context.Context) ([]BillOfMaterials, error)
	Update(ctx context.Context, bom *BillOfMaterials) error
	Delete(ctx context.Context, id string) error
}

type BOMComponentRepository interface {
	Create(ctx context.Context, comp *BOMComponent) error
	GetByID(ctx context.Context, id string) (*BOMComponent, error)
	ListByBOMID(ctx context.Context, bomID string) ([]BOMComponent, error)
	Delete(ctx context.Context, id string) error
}

type WorkCenterRepository interface {
	Create(ctx context.Context, wc *WorkCenter) error
	GetByID(ctx context.Context, id string) (*WorkCenter, error)
	GetByCode(ctx context.Context, code string) (*WorkCenter, error)
	List(ctx context.Context) ([]WorkCenter, error)
	Update(ctx context.Context, wc *WorkCenter) error
	Delete(ctx context.Context, id string) error
}

type RoutingOperationRepository interface {
	Create(ctx context.Context, op *RoutingOperation) error
	GetByID(ctx context.Context, id string) (*RoutingOperation, error)
	ListByBOMID(ctx context.Context, bomID string) ([]RoutingOperation, error)
	List(ctx context.Context) ([]RoutingOperation, error)
	Update(ctx context.Context, op *RoutingOperation) error
	Delete(ctx context.Context, id string) error
}

type ProductionOrderRepository interface {
	Create(ctx context.Context, po *ProductionOrder) error
	GetByID(ctx context.Context, id string) (*ProductionOrder, error)
	List(ctx context.Context) ([]ProductionOrder, error)
	Update(ctx context.Context, po *ProductionOrder) error
	Delete(ctx context.Context, id string) error
}

type WorkOrderRepository interface {
	Create(ctx context.Context, wo *WorkOrder) error
	GetByID(ctx context.Context, id string) (*WorkOrder, error)
	ListByProductionOrderID(ctx context.Context, poID string) ([]WorkOrder, error)
	List(ctx context.Context) ([]WorkOrder, error)
	Update(ctx context.Context, wo *WorkOrder) error
	Delete(ctx context.Context, id string) error
}

type LaborReportRepository interface {
	Create(ctx context.Context, lr *LaborReport) error
	ListByWorkOrderID(ctx context.Context, woID string) ([]LaborReport, error)
}

type MachineLogRepository interface {
	Create(ctx context.Context, ml *MachineLog) error
	ListByWorkCenterID(ctx context.Context, wcID string) ([]MachineLog, error)
}

type QualityInspectionRepository interface {
	Create(ctx context.Context, qi *QualityInspection) error
	GetByID(ctx context.Context, id string) (*QualityInspection, error)
	GetByWorkOrderID(ctx context.Context, woID string) (*QualityInspection, error)
	List(ctx context.Context) ([]QualityInspection, error)
	Update(ctx context.Context, qi *QualityInspection) error
}

type NonConformanceRepository interface {
	Create(ctx context.Context, nc *NonConformance) error
	GetByID(ctx context.Context, id string) (*NonConformance, error)
	ListByInspectionID(ctx context.Context, inspID string) ([]NonConformance, error)
	Update(ctx context.Context, nc *NonConformance) error
}

type EquipmentRepository interface {
	Create(ctx context.Context, eq *Equipment) error
	GetByID(ctx context.Context, id string) (*Equipment, error)
	ListByWorkCenterID(ctx context.Context, wcID string) ([]Equipment, error)
	Update(ctx context.Context, eq *Equipment) error
}

type MaintenanceOrderRepository interface {
	Create(ctx context.Context, mo *MaintenanceOrder) error
	GetByID(ctx context.Context, id string) (*MaintenanceOrder, error)
	List(ctx context.Context) ([]MaintenanceOrder, error)
	Update(ctx context.Context, mo *MaintenanceOrder) error
}

type CostingRecordRepository interface {
	Create(ctx context.Context, cr *CostingRecord) error
	GetByProductionOrderID(ctx context.Context, poID string) (*CostingRecord, error)
	Update(ctx context.Context, cr *CostingRecord) error
}
