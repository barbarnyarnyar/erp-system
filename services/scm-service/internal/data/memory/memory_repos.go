package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/erp-system/scm-service/internal/business/domain"
)

// MemoryProductRepo implements domain.ProductRepository
type MemoryProductRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Product
}

func NewMemoryProductRepo() *MemoryProductRepo {
	return &MemoryProductRepo{data: make(map[string]domain.Product)}
}

func (r *MemoryProductRepo) Create(ctx context.Context, p *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[p.ID] = *p
	return nil
}

func (r *MemoryProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.data[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	return &p, nil
}

func (r *MemoryProductRepo) List(ctx context.Context) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Product, 0, len(r.data))
	for _, p := range r.data {
		list = append(list, p)
	}
	return list, nil
}

func (r *MemoryProductRepo) Update(ctx context.Context, p *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[p.ID] = *p
	return nil
}

func (r *MemoryProductRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

// MemoryLocationRepo implements domain.LocationRepository
type MemoryLocationRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Location
}

func NewMemoryLocationRepo() *MemoryLocationRepo {
	return &MemoryLocationRepo{data: make(map[string]domain.Location)}
}

func (r *MemoryLocationRepo) Create(ctx context.Context, loc *domain.Location) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[loc.ID] = *loc
	return nil
}

func (r *MemoryLocationRepo) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	loc, ok := r.data[id]
	if !ok {
		return nil, errors.New("location not found")
	}
	return &loc, nil
}

func (r *MemoryLocationRepo) List(ctx context.Context) ([]domain.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Location, 0, len(r.data))
	for _, l := range r.data {
		list = append(list, l)
	}
	return list, nil
}

func (r *MemoryLocationRepo) Update(ctx context.Context, loc *domain.Location) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[loc.ID] = *loc
	return nil
}

func (r *MemoryLocationRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

// MemorySupplierRepo implements domain.SupplierRepository
type MemorySupplierRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Supplier
}

func NewMemorySupplierRepo() *MemorySupplierRepo {
	return &MemorySupplierRepo{data: make(map[string]domain.Supplier)}
}

func (r *MemorySupplierRepo) Create(ctx context.Context, s *domain.Supplier) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.ID] = *s
	return nil
}

func (r *MemorySupplierRepo) GetByID(ctx context.Context, id string) (*domain.Supplier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.data[id]
	if !ok {
		return nil, errors.New("supplier not found")
	}
	return &s, nil
}

func (r *MemorySupplierRepo) List(ctx context.Context) ([]domain.Supplier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Supplier, 0, len(r.data))
	for _, s := range r.data {
		list = append(list, s)
	}
	return list, nil
}

func (r *MemorySupplierRepo) Update(ctx context.Context, s *domain.Supplier) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.ID] = *s
	return nil
}

func (r *MemorySupplierRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

// MemoryInventoryItemRepo implements domain.InventoryItemRepository
type MemoryInventoryItemRepo struct {
	mu   sync.RWMutex
	data map[string]domain.InventoryItem
}

func NewMemoryInventoryItemRepo() *MemoryInventoryItemRepo {
	return &MemoryInventoryItemRepo{data: make(map[string]domain.InventoryItem)}
}

func (r *MemoryInventoryItemRepo) Create(ctx context.Context, ii *domain.InventoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ii.ID] = *ii
	return nil
}

func (r *MemoryInventoryItemRepo) GetByID(ctx context.Context, id string) (*domain.InventoryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ii, ok := r.data[id]
	if !ok {
		return nil, errors.New("inventory item not found")
	}
	return &ii, nil
}

func (r *MemoryInventoryItemRepo) List(ctx context.Context) ([]domain.InventoryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.InventoryItem, 0, len(r.data))
	for _, ii := range r.data {
		list = append(list, ii)
	}
	return list, nil
}

func (r *MemoryInventoryItemRepo) Update(ctx context.Context, ii *domain.InventoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ii.ID] = *ii
	return nil
}

func (r *MemoryInventoryItemRepo) GetByProductAndLocation(ctx context.Context, productID string, locationID string) (*domain.InventoryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ii := range r.data {
		if ii.ProductID == productID && ii.LocationID == locationID {
			return &ii, nil
		}
	}
	return nil, errors.New("inventory item not found for product/location combination")
}

// MemoryInventoryMovementRepo implements domain.InventoryMovementRepository
type MemoryInventoryMovementRepo struct {
	mu   sync.RWMutex
	data map[string]domain.InventoryMovement
}

func NewMemoryInventoryMovementRepo() *MemoryInventoryMovementRepo {
	return &MemoryInventoryMovementRepo{data: make(map[string]domain.InventoryMovement)}
}

func (r *MemoryInventoryMovementRepo) Create(ctx context.Context, im *domain.InventoryMovement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[im.ID] = *im
	return nil
}

func (r *MemoryInventoryMovementRepo) GetByID(ctx context.Context, id string) (*domain.InventoryMovement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	im, ok := r.data[id]
	if !ok {
		return nil, errors.New("inventory movement not found")
	}
	return &im, nil
}

func (r *MemoryInventoryMovementRepo) List(ctx context.Context) ([]domain.InventoryMovement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.InventoryMovement, 0, len(r.data))
	for _, im := range r.data {
		list = append(list, im)
	}
	return list, nil
}

// MemoryPurchaseOrderRepo implements domain.PurchaseOrderRepository
type MemoryPurchaseOrderRepo struct {
	mu   sync.RWMutex
	data map[string]domain.PurchaseOrder
}

func NewMemoryPurchaseOrderRepo() *MemoryPurchaseOrderRepo {
	return &MemoryPurchaseOrderRepo{data: make(map[string]domain.PurchaseOrder)}
}

func (r *MemoryPurchaseOrderRepo) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[po.ID] = *po
	return nil
}

func (r *MemoryPurchaseOrderRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	po, ok := r.data[id]
	if !ok {
		return nil, errors.New("purchase order not found")
	}
	return &po, nil
}

func (r *MemoryPurchaseOrderRepo) List(ctx context.Context) ([]domain.PurchaseOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.PurchaseOrder, 0, len(r.data))
	for _, po := range r.data {
		list = append(list, po)
	}
	return list, nil
}

func (r *MemoryPurchaseOrderRepo) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[po.ID] = *po
	return nil
}

func (r *MemoryPurchaseOrderRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

// MemoryPurchaseOrderLineRepo implements domain.PurchaseOrderLineRepository
type MemoryPurchaseOrderLineRepo struct {
	mu   sync.RWMutex
	data map[string]domain.PurchaseOrderLine
}

func NewMemoryPurchaseOrderLineRepo() *MemoryPurchaseOrderLineRepo {
	return &MemoryPurchaseOrderLineRepo{data: make(map[string]domain.PurchaseOrderLine)}
}

func (r *MemoryPurchaseOrderLineRepo) Create(ctx context.Context, pol *domain.PurchaseOrderLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[pol.ID] = *pol
	return nil
}

func (r *MemoryPurchaseOrderLineRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseOrderLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pol, ok := r.data[id]
	if !ok {
		return nil, errors.New("purchase order line not found")
	}
	return &pol, nil
}

func (r *MemoryPurchaseOrderLineRepo) ListByPOID(ctx context.Context, poID string) ([]domain.PurchaseOrderLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.PurchaseOrderLine
	for _, pol := range r.data {
		if pol.PurchaseOrderID == poID {
			list = append(list, pol)
		}
	}
	return list, nil
}

func (r *MemoryPurchaseOrderLineRepo) DeleteByPOID(ctx context.Context, poID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, pol := range r.data {
		if pol.PurchaseOrderID == poID {
			delete(r.data, id)
		}
	}
	return nil
}

// MemoryReceiptRepo implements domain.ReceiptRepository
type MemoryReceiptRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Receipt
}

func NewMemoryReceiptRepo() *MemoryReceiptRepo {
	return &MemoryReceiptRepo{data: make(map[string]domain.Receipt)}
}

func (r *MemoryReceiptRepo) Create(ctx context.Context, rec *domain.Receipt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[rec.ID] = *rec
	return nil
}

func (r *MemoryReceiptRepo) GetByID(ctx context.Context, id string) (*domain.Receipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.data[id]
	if !ok {
		return nil, errors.New("receipt not found")
	}
	return &rec, nil
}

func (r *MemoryReceiptRepo) List(ctx context.Context) ([]domain.Receipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Receipt, 0, len(r.data))
	for _, rec := range r.data {
		list = append(list, rec)
	}
	return list, nil
}

func (r *MemoryReceiptRepo) Update(ctx context.Context, rec *domain.Receipt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[rec.ID] = *rec
	return nil
}

// MemoryReceiptLineRepo implements domain.ReceiptLineRepository
type MemoryReceiptLineRepo struct {
	mu   sync.RWMutex
	data map[string]domain.ReceiptLine
}

func NewMemoryReceiptLineRepo() *MemoryReceiptLineRepo {
	return &MemoryReceiptLineRepo{data: make(map[string]domain.ReceiptLine)}
}

func (r *MemoryReceiptLineRepo) Create(ctx context.Context, rl *domain.ReceiptLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[rl.ID] = *rl
	return nil
}

func (r *MemoryReceiptLineRepo) GetByID(ctx context.Context, id string) (*domain.ReceiptLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rl, ok := r.data[id]
	if !ok {
		return nil, errors.New("receipt line not found")
	}
	return &rl, nil
}

func (r *MemoryReceiptLineRepo) ListByReceiptID(ctx context.Context, receiptID string) ([]domain.ReceiptLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ReceiptLine
	for _, rl := range r.data {
		if rl.ReceiptID == receiptID {
			list = append(list, rl)
		}
	}
	return list, nil
}

// MemoryShipmentRepo implements domain.ShipmentRepository
type MemoryShipmentRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Shipment
}

func NewMemoryShipmentRepo() *MemoryShipmentRepo {
	return &MemoryShipmentRepo{data: make(map[string]domain.Shipment)}
}

func (r *MemoryShipmentRepo) Create(ctx context.Context, s *domain.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.ID] = *s
	return nil
}

func (r *MemoryShipmentRepo) GetByID(ctx context.Context, id string) (*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.data[id]
	if !ok {
		return nil, errors.New("shipment not found")
	}
	return &s, nil
}

func (r *MemoryShipmentRepo) List(ctx context.Context) ([]domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Shipment, 0, len(r.data))
	for _, s := range r.data {
		list = append(list, s)
	}
	return list, nil
}

func (r *MemoryShipmentRepo) Update(ctx context.Context, s *domain.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.ID] = *s
	return nil
}

// MemoryShipmentLineRepo implements domain.ShipmentLineRepository
type MemoryShipmentLineRepo struct {
	mu   sync.RWMutex
	data map[string]domain.ShipmentLine
}

func NewMemoryShipmentLineRepo() *MemoryShipmentLineRepo {
	return &MemoryShipmentLineRepo{data: make(map[string]domain.ShipmentLine)}
}

func (r *MemoryShipmentLineRepo) Create(ctx context.Context, sl *domain.ShipmentLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[sl.ID] = *sl
	return nil
}

func (r *MemoryShipmentLineRepo) GetByID(ctx context.Context, id string) (*domain.ShipmentLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sl, ok := r.data[id]
	if !ok {
		return nil, errors.New("shipment line not found")
	}
	return &sl, nil
}

func (r *MemoryShipmentLineRepo) ListByShipmentID(ctx context.Context, shipmentID string) ([]domain.ShipmentLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ShipmentLine
	for _, sl := range r.data {
		if sl.ShipmentID == shipmentID {
			list = append(list, sl)
		}
	}
	return list, nil
}

// MemoryDemandForecastRepo implements domain.DemandForecastRepository
type MemoryDemandForecastRepo struct {
	mu   sync.RWMutex
	data map[string]domain.DemandForecast
}

func NewMemoryDemandForecastRepo() *MemoryDemandForecastRepo {
	return &MemoryDemandForecastRepo{data: make(map[string]domain.DemandForecast)}
}

func (r *MemoryDemandForecastRepo) Create(ctx context.Context, df *domain.DemandForecast) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[df.ID] = *df
	return nil
}

func (r *MemoryDemandForecastRepo) GetByID(ctx context.Context, id string) (*domain.DemandForecast, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	df, ok := r.data[id]
	if !ok {
		return nil, errors.New("demand forecast not found")
	}
	return &df, nil
}

func (r *MemoryDemandForecastRepo) List(ctx context.Context) ([]domain.DemandForecast, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.DemandForecast, 0, len(r.data))
	for _, df := range r.data {
		list = append(list, df)
	}
	return list, nil
}

func (r *MemoryDemandForecastRepo) Update(ctx context.Context, df *domain.DemandForecast) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[df.ID] = *df
	return nil
}

func (r *MemoryDemandForecastRepo) ListByProductID(ctx context.Context, productID string) ([]domain.DemandForecast, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.DemandForecast
	for _, df := range r.data {
		if df.ProductID == productID {
			list = append(list, df)
		}
	}
	return list, nil
}

// MemoryProductCategoryRepo implements domain.ProductCategoryRepository
type MemoryProductCategoryRepo struct {
	mu   sync.RWMutex
	data map[string]domain.ProductCategory
}

func NewMemoryProductCategoryRepo() *MemoryProductCategoryRepo {
	return &MemoryProductCategoryRepo{data: make(map[string]domain.ProductCategory)}
}

func (r *MemoryProductCategoryRepo) Create(ctx context.Context, pc *domain.ProductCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[pc.ID] = *pc
	return nil
}

func (r *MemoryProductCategoryRepo) GetByID(ctx context.Context, id string) (*domain.ProductCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pc, ok := r.data[id]
	if !ok {
		return nil, errors.New("product category not found")
	}
	return &pc, nil
}

func (r *MemoryProductCategoryRepo) List(ctx context.Context) ([]domain.ProductCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ProductCategory, 0, len(r.data))
	for _, pc := range r.data {
		list = append(list, pc)
	}
	return list, nil
}

func (r *MemoryProductCategoryRepo) Update(ctx context.Context, pc *domain.ProductCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[pc.ID] = *pc
	return nil
}

func (r *MemoryProductCategoryRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

// MemoryVendorContractRepo implements domain.VendorContractRepository
type MemoryVendorContractRepo struct {
	mu   sync.RWMutex
	data map[string]domain.VendorContract
}

func NewMemoryVendorContractRepo() *MemoryVendorContractRepo {
	return &MemoryVendorContractRepo{data: make(map[string]domain.VendorContract)}
}

func (r *MemoryVendorContractRepo) Create(ctx context.Context, vc *domain.VendorContract) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[vc.ID] = *vc
	return nil
}

func (r *MemoryVendorContractRepo) GetByID(ctx context.Context, id string) (*domain.VendorContract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	vc, ok := r.data[id]
	if !ok {
		return nil, errors.New("vendor contract not found")
	}
	return &vc, nil
}

func (r *MemoryVendorContractRepo) List(ctx context.Context) ([]domain.VendorContract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.VendorContract, 0, len(r.data))
	for _, vc := range r.data {
		list = append(list, vc)
	}
	return list, nil
}

func (r *MemoryVendorContractRepo) Update(ctx context.Context, vc *domain.VendorContract) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[vc.ID] = *vc
	return nil
}

func (r *MemoryVendorContractRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

// MemoryPurchaseRequisitionRepo implements domain.PurchaseRequisitionRepository
type MemoryPurchaseRequisitionRepo struct {
	mu   sync.RWMutex
	data map[string]domain.PurchaseRequisition
}

func NewMemoryPurchaseRequisitionRepo() *MemoryPurchaseRequisitionRepo {
	return &MemoryPurchaseRequisitionRepo{data: make(map[string]domain.PurchaseRequisition)}
}

func (r *MemoryPurchaseRequisitionRepo) Create(ctx context.Context, pr *domain.PurchaseRequisition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[pr.ID] = *pr
	return nil
}

func (r *MemoryPurchaseRequisitionRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseRequisition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pr, ok := r.data[id]
	if !ok {
		return nil, errors.New("purchase requisition not found")
	}
	return &pr, nil
}

func (r *MemoryPurchaseRequisitionRepo) List(ctx context.Context) ([]domain.PurchaseRequisition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.PurchaseRequisition, 0, len(r.data))
	for _, pr := range r.data {
		list = append(list, pr)
	}
	return list, nil
}

func (r *MemoryPurchaseRequisitionRepo) Update(ctx context.Context, pr *domain.PurchaseRequisition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[pr.ID] = *pr
	return nil
}

func (r *MemoryPurchaseRequisitionRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

// MemoryPurchaseRequisitionLineRepo implements domain.PurchaseRequisitionLineRepository
type MemoryPurchaseRequisitionLineRepo struct {
	mu   sync.RWMutex
	data map[string]domain.PurchaseRequisitionLine
}

func NewMemoryPurchaseRequisitionLineRepo() *MemoryPurchaseRequisitionLineRepo {
	return &MemoryPurchaseRequisitionLineRepo{data: make(map[string]domain.PurchaseRequisitionLine)}
}

func (r *MemoryPurchaseRequisitionLineRepo) Create(ctx context.Context, prl *domain.PurchaseRequisitionLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[prl.ID] = *prl
	return nil
}

func (r *MemoryPurchaseRequisitionLineRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseRequisitionLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prl, ok := r.data[id]
	if !ok {
		return nil, errors.New("purchase requisition line not found")
	}
	return &prl, nil
}

func (r *MemoryPurchaseRequisitionLineRepo) ListByRequisitionID(ctx context.Context, reqID string) ([]domain.PurchaseRequisitionLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.PurchaseRequisitionLine
	for _, prl := range r.data {
		if prl.PurchaseRequisitionID == reqID {
			list = append(list, prl)
		}
	}
	return list, nil
}

func (r *MemoryPurchaseRequisitionLineRepo) DeleteByRequisitionID(ctx context.Context, reqID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, prl := range r.data {
		if prl.PurchaseRequisitionID == reqID {
			delete(r.data, id)
		}
	}
	return nil
}

// MemoryStockTransferRepo implements domain.StockTransferRepository
type MemoryStockTransferRepo struct {
	mu   sync.RWMutex
	data map[string]domain.StockTransfer
}

func NewMemoryStockTransferRepo() *MemoryStockTransferRepo {
	return &MemoryStockTransferRepo{data: make(map[string]domain.StockTransfer)}
}

func (r *MemoryStockTransferRepo) Create(ctx context.Context, st *domain.StockTransfer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[st.ID] = *st
	return nil
}

func (r *MemoryStockTransferRepo) GetByID(ctx context.Context, id string) (*domain.StockTransfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	st, ok := r.data[id]
	if !ok {
		return nil, errors.New("stock transfer not found")
	}
	return &st, nil
}

func (r *MemoryStockTransferRepo) List(ctx context.Context) ([]domain.StockTransfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.StockTransfer, 0, len(r.data))
	for _, st := range r.data {
		list = append(list, st)
	}
	return list, nil
}

func (r *MemoryStockTransferRepo) Update(ctx context.Context, st *domain.StockTransfer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[st.ID] = *st
	return nil
}

