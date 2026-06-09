package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/erp-system/m-service/internal/business/domain"
)

// MemoryBillOfMaterialsRepo implements domain.BillOfMaterialsRepository
type MemoryBillOfMaterialsRepo struct {
	mu   sync.RWMutex
	boms map[string]domain.BillOfMaterials
}

func NewMemoryBillOfMaterialsRepo() *MemoryBillOfMaterialsRepo {
	return &MemoryBillOfMaterialsRepo{boms: make(map[string]domain.BillOfMaterials)}
}

func (r *MemoryBillOfMaterialsRepo) Create(ctx context.Context, bom *domain.BillOfMaterials) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.boms[bom.ID] = *bom
	return nil
}

func (r *MemoryBillOfMaterialsRepo) GetByID(ctx context.Context, id string) (*domain.BillOfMaterials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bom, ok := r.boms[id]
	if !ok {
		return nil, errors.New("bill of materials not found")
	}
	return &bom, nil
}

func (r *MemoryBillOfMaterialsRepo) GetByProductID(ctx context.Context, productID string) (*domain.BillOfMaterials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, bom := range r.boms {
		if bom.ProductID == productID && bom.Status == "ACTIVE" {
			return &bom, nil
		}
	}
	// fallback to any version if active not found
	for _, bom := range r.boms {
		if bom.ProductID == productID {
			return &bom, nil
		}
	}
	return nil, errors.New("bill of materials not found for product")
}

func (r *MemoryBillOfMaterialsRepo) List(ctx context.Context) ([]domain.BillOfMaterials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.BillOfMaterials, 0, len(r.boms))
	for _, b := range r.boms {
		list = append(list, b)
	}
	return list, nil
}

func (r *MemoryBillOfMaterialsRepo) Update(ctx context.Context, bom *domain.BillOfMaterials) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.boms[bom.ID] = *bom
	return nil
}

func (r *MemoryBillOfMaterialsRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.boms, id)
	return nil
}

// MemoryBOMComponentRepo implements domain.BOMComponentRepository
type MemoryBOMComponentRepo struct {
	mu    sync.RWMutex
	comps map[string]domain.BOMComponent
}

func NewMemoryBOMComponentRepo() *MemoryBOMComponentRepo {
	return &MemoryBOMComponentRepo{comps: make(map[string]domain.BOMComponent)}
}

func (r *MemoryBOMComponentRepo) Create(ctx context.Context, comp *domain.BOMComponent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.comps[comp.ID] = *comp
	return nil
}

func (r *MemoryBOMComponentRepo) GetByID(ctx context.Context, id string) (*domain.BOMComponent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	comp, ok := r.comps[id]
	if !ok {
		return nil, errors.New("BOM component not found")
	}
	return &comp, nil
}

func (r *MemoryBOMComponentRepo) ListByBOMID(ctx context.Context, bomID string) ([]domain.BOMComponent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.BOMComponent
	for _, c := range r.comps {
		if c.BomID == bomID {
			list = append(list, c)
		}
	}
	return list, nil
}

func (r *MemoryBOMComponentRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.comps, id)
	return nil
}

// MemoryWorkCenterRepo implements domain.WorkCenterRepository
type MemoryWorkCenterRepo struct {
	mu  sync.RWMutex
	wcs map[string]domain.WorkCenter
}

func NewMemoryWorkCenterRepo() *MemoryWorkCenterRepo {
	return &MemoryWorkCenterRepo{wcs: make(map[string]domain.WorkCenter)}
}

func (r *MemoryWorkCenterRepo) Create(ctx context.Context, wc *domain.WorkCenter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wcs[wc.ID] = *wc
	return nil
}

func (r *MemoryWorkCenterRepo) GetByID(ctx context.Context, id string) (*domain.WorkCenter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wc, ok := r.wcs[id]
	if !ok {
		return nil, errors.New("work center not found")
	}
	return &wc, nil
}

func (r *MemoryWorkCenterRepo) GetByCode(ctx context.Context, code string) (*domain.WorkCenter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, wc := range r.wcs {
		if wc.Code == code {
			return &wc, nil
		}
	}
	return nil, errors.New("work center not found by code")
}

func (r *MemoryWorkCenterRepo) List(ctx context.Context) ([]domain.WorkCenter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.WorkCenter, 0, len(r.wcs))
	for _, w := range r.wcs {
		list = append(list, w)
	}
	return list, nil
}

func (r *MemoryWorkCenterRepo) Update(ctx context.Context, wc *domain.WorkCenter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wcs[wc.ID] = *wc
	return nil
}

func (r *MemoryWorkCenterRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.wcs, id)
	return nil
}

// MemoryRoutingOperationRepo implements domain.RoutingOperationRepository
type MemoryRoutingOperationRepo struct {
	mu  sync.RWMutex
	ops map[string]domain.RoutingOperation
}

func NewMemoryRoutingOperationRepo() *MemoryRoutingOperationRepo {
	return &MemoryRoutingOperationRepo{ops: make(map[string]domain.RoutingOperation)}
}

func (r *MemoryRoutingOperationRepo) Create(ctx context.Context, op *domain.RoutingOperation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ops[op.ID] = *op
	return nil
}

func (r *MemoryRoutingOperationRepo) GetByID(ctx context.Context, id string) (*domain.RoutingOperation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	op, ok := r.ops[id]
	if !ok {
		return nil, errors.New("routing operation not found")
	}
	return &op, nil
}

func (r *MemoryRoutingOperationRepo) ListByBOMID(ctx context.Context, bomID string) ([]domain.RoutingOperation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.RoutingOperation
	for _, op := range r.ops {
		if op.BomID == bomID {
			list = append(list, op)
		}
	}
	return list, nil
}

func (r *MemoryRoutingOperationRepo) List(ctx context.Context) ([]domain.RoutingOperation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.RoutingOperation, 0, len(r.ops))
	for _, op := range r.ops {
		list = append(list, op)
	}
	return list, nil
}

func (r *MemoryRoutingOperationRepo) Update(ctx context.Context, op *domain.RoutingOperation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ops[op.ID] = *op
	return nil
}

func (r *MemoryRoutingOperationRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.ops, id)
	return nil
}

// MemoryProductionOrderRepo implements domain.ProductionOrderRepository
type MemoryProductionOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]domain.ProductionOrder
}

func NewMemoryProductionOrderRepo() *MemoryProductionOrderRepo {
	return &MemoryProductionOrderRepo{orders: make(map[string]domain.ProductionOrder)}
}

func (r *MemoryProductionOrderRepo) Create(ctx context.Context, po *domain.ProductionOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[po.ID] = *po
	return nil
}

func (r *MemoryProductionOrderRepo) GetByID(ctx context.Context, id string) (*domain.ProductionOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	po, ok := r.orders[id]
	if !ok {
		return nil, errors.New("production order not found")
	}
	return &po, nil
}

func (r *MemoryProductionOrderRepo) List(ctx context.Context) ([]domain.ProductionOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ProductionOrder, 0, len(r.orders))
	for _, o := range r.orders {
		list = append(list, o)
	}
	return list, nil
}

func (r *MemoryProductionOrderRepo) Update(ctx context.Context, po *domain.ProductionOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[po.ID] = *po
	return nil
}

func (r *MemoryProductionOrderRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.orders, id)
	return nil
}

// MemoryWorkOrderRepo implements domain.WorkOrderRepository
type MemoryWorkOrderRepo struct {
	mu  sync.RWMutex
	wos map[string]domain.WorkOrder
}

func NewMemoryWorkOrderRepo() *MemoryWorkOrderRepo {
	return &MemoryWorkOrderRepo{wos: make(map[string]domain.WorkOrder)}
}

func (r *MemoryWorkOrderRepo) Create(ctx context.Context, wo *domain.WorkOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wos[wo.ID] = *wo
	return nil
}

func (r *MemoryWorkOrderRepo) GetByID(ctx context.Context, id string) (*domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wo, ok := r.wos[id]
	if !ok {
		return nil, errors.New("work order not found")
	}
	return &wo, nil
}

func (r *MemoryWorkOrderRepo) ListByProductionOrderID(ctx context.Context, poID string) ([]domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.WorkOrder
	for _, w := range r.wos {
		if w.ProductionOrderID == poID {
			list = append(list, w)
		}
	}
	return list, nil
}

func (r *MemoryWorkOrderRepo) List(ctx context.Context) ([]domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.WorkOrder, 0, len(r.wos))
	for _, w := range r.wos {
		list = append(list, w)
	}
	return list, nil
}

func (r *MemoryWorkOrderRepo) Update(ctx context.Context, wo *domain.WorkOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wos[wo.ID] = *wo
	return nil
}

func (r *MemoryWorkOrderRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.wos, id)
	return nil
}

// MemoryLaborReportRepo implements domain.LaborReportRepository
type MemoryLaborReportRepo struct {
	mu  sync.RWMutex
	lrs map[string]domain.LaborReport
}

func NewMemoryLaborReportRepo() *MemoryLaborReportRepo {
	return &MemoryLaborReportRepo{lrs: make(map[string]domain.LaborReport)}
}

func (r *MemoryLaborReportRepo) Create(ctx context.Context, lr *domain.LaborReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lrs[lr.ID] = *lr
	return nil
}

func (r *MemoryLaborReportRepo) ListByWorkOrderID(ctx context.Context, woID string) ([]domain.LaborReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.LaborReport
	for _, lr := range r.lrs {
		if lr.WorkOrderID == woID {
			list = append(list, lr)
		}
	}
	return list, nil
}

// MemoryMachineLogRepo implements domain.MachineLogRepository
type MemoryMachineLogRepo struct {
	mu  sync.RWMutex
	mls map[string]domain.MachineLog
}

func NewMemoryMachineLogRepo() *MemoryMachineLogRepo {
	return &MemoryMachineLogRepo{mls: make(map[string]domain.MachineLog)}
}

func (r *MemoryMachineLogRepo) Create(ctx context.Context, ml *domain.MachineLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mls[ml.ID] = *ml
	return nil
}

func (r *MemoryMachineLogRepo) ListByWorkCenterID(ctx context.Context, wcID string) ([]domain.MachineLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.MachineLog
	for _, ml := range r.mls {
		if ml.WorkCenterID == wcID {
			list = append(list, ml)
		}
	}
	return list, nil
}

// MemoryQualityInspectionRepo implements domain.QualityInspectionRepository
type MemoryQualityInspectionRepo struct {
	mu  sync.RWMutex
	qis map[string]domain.QualityInspection
}

func NewMemoryQualityInspectionRepo() *MemoryQualityInspectionRepo {
	return &MemoryQualityInspectionRepo{qis: make(map[string]domain.QualityInspection)}
}

func (r *MemoryQualityInspectionRepo) Create(ctx context.Context, qi *domain.QualityInspection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.qis[qi.ID] = *qi
	return nil
}

func (r *MemoryQualityInspectionRepo) GetByID(ctx context.Context, id string) (*domain.QualityInspection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	qi, ok := r.qis[id]
	if !ok {
		return nil, errors.New("quality inspection not found")
	}
	return &qi, nil
}

func (r *MemoryQualityInspectionRepo) GetByWorkOrderID(ctx context.Context, woID string) (*domain.QualityInspection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, qi := range r.qis {
		if qi.WorkOrderID == woID {
			return &qi, nil
		}
	}
	return nil, errors.New("quality inspection not found by work order")
}

func (r *MemoryQualityInspectionRepo) List(ctx context.Context) ([]domain.QualityInspection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.QualityInspection, 0, len(r.qis))
	for _, q := range r.qis {
		list = append(list, q)
	}
	return list, nil
}

func (r *MemoryQualityInspectionRepo) Update(ctx context.Context, qi *domain.QualityInspection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.qis[qi.ID] = *qi
	return nil
}

// MemoryNonConformanceRepo implements domain.NonConformanceRepository
type MemoryNonConformanceRepo struct {
	mu  sync.RWMutex
	ncs map[string]domain.NonConformance
}

func NewMemoryNonConformanceRepo() *MemoryNonConformanceRepo {
	return &MemoryNonConformanceRepo{ncs: make(map[string]domain.NonConformance)}
}

func (r *MemoryNonConformanceRepo) Create(ctx context.Context, nc *domain.NonConformance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ncs[nc.ID] = *nc
	return nil
}

func (r *MemoryNonConformanceRepo) GetByID(ctx context.Context, id string) (*domain.NonConformance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	nc, ok := r.ncs[id]
	if !ok {
		return nil, errors.New("non conformance not found")
	}
	return &nc, nil
}

func (r *MemoryNonConformanceRepo) ListByInspectionID(ctx context.Context, inspID string) ([]domain.NonConformance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.NonConformance
	for _, nc := range r.ncs {
		if nc.InspectionID == inspID {
			list = append(list, nc)
		}
	}
	return list, nil
}

func (r *MemoryNonConformanceRepo) Update(ctx context.Context, nc *domain.NonConformance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ncs[nc.ID] = *nc
	return nil
}

// MemoryEquipmentRepo implements domain.EquipmentRepository
type MemoryEquipmentRepo struct {
	mu  sync.RWMutex
	eqs map[string]domain.Equipment
}

func NewMemoryEquipmentRepo() *MemoryEquipmentRepo {
	return &MemoryEquipmentRepo{eqs: make(map[string]domain.Equipment)}
}

func (r *MemoryEquipmentRepo) Create(ctx context.Context, eq *domain.Equipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.eqs[eq.ID] = *eq
	return nil
}

func (r *MemoryEquipmentRepo) GetByID(ctx context.Context, id string) (*domain.Equipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	eq, ok := r.eqs[id]
	if !ok {
		return nil, errors.New("equipment not found")
	}
	return &eq, nil
}

func (r *MemoryEquipmentRepo) ListByWorkCenterID(ctx context.Context, wcID string) ([]domain.Equipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.Equipment
	for _, eq := range r.eqs {
		if eq.WorkCenterID == wcID {
			list = append(list, eq)
		}
	}
	return list, nil
}

func (r *MemoryEquipmentRepo) Update(ctx context.Context, eq *domain.Equipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.eqs[eq.ID] = *eq
	return nil
}

// MemoryMaintenanceOrderRepo implements domain.MaintenanceOrderRepository
type MemoryMaintenanceOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]domain.MaintenanceOrder
}

func NewMemoryMaintenanceOrderRepo() *MemoryMaintenanceOrderRepo {
	return &MemoryMaintenanceOrderRepo{orders: make(map[string]domain.MaintenanceOrder)}
}

func (r *MemoryMaintenanceOrderRepo) Create(ctx context.Context, mo *domain.MaintenanceOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[mo.ID] = *mo
	return nil
}

func (r *MemoryMaintenanceOrderRepo) GetByID(ctx context.Context, id string) (*domain.MaintenanceOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mo, ok := r.orders[id]
	if !ok {
		return nil, errors.New("maintenance order not found")
	}
	return &mo, nil
}

func (r *MemoryMaintenanceOrderRepo) List(ctx context.Context) ([]domain.MaintenanceOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.MaintenanceOrder, 0, len(r.orders))
	for _, m := range r.orders {
		list = append(list, m)
	}
	return list, nil
}

func (r *MemoryMaintenanceOrderRepo) Update(ctx context.Context, mo *domain.MaintenanceOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[mo.ID] = *mo
	return nil
}

// MemoryCostingRecordRepo implements domain.CostingRecordRepository
type MemoryCostingRecordRepo struct {
	mu      sync.RWMutex
	records map[string]domain.CostingRecord
}

func NewMemoryCostingRecordRepo() *MemoryCostingRecordRepo {
	return &MemoryCostingRecordRepo{records: make(map[string]domain.CostingRecord)}
}

func (r *MemoryCostingRecordRepo) Create(ctx context.Context, cr *domain.CostingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[cr.ID] = *cr
	return nil
}

func (r *MemoryCostingRecordRepo) GetByProductionOrderID(ctx context.Context, poID string) (*domain.CostingRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cr := range r.records {
		if cr.ProductionOrderID == poID {
			return &cr, nil
		}
	}
	return nil, errors.New("costing record not found by production order")
}

func (r *MemoryCostingRecordRepo) Update(ctx context.Context, cr *domain.CostingRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[cr.ID] = *cr
	return nil
}
