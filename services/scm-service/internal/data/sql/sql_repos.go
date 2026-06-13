package sql

import (
	"context"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"gorm.io/gorm"
)

// SQLProductCategoryRepo implements domain.ProductCategoryRepository
type SQLProductCategoryRepo struct {
	db *gorm.DB
}

func NewSQLProductCategoryRepo(db *gorm.DB) *SQLProductCategoryRepo {
	return &SQLProductCategoryRepo{db: db}
}

func (r *SQLProductCategoryRepo) Create(ctx context.Context, pc *domain.ProductCategory) error {
	dbModel := FromDomainProductCategory(pc)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLProductCategoryRepo) GetByID(ctx context.Context, id string) (*domain.ProductCategory, error) {
	var dbModel ProductCategory
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainProductCategory(&dbModel), nil
}

func (r *SQLProductCategoryRepo) List(ctx context.Context) ([]domain.ProductCategory, error) {
	var dbModels []ProductCategory
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.ProductCategory, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainProductCategory(&m)
	}
	return res, nil
}

func (r *SQLProductCategoryRepo) Update(ctx context.Context, pc *domain.ProductCategory) error {
	dbModel := FromDomainProductCategory(pc)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLProductCategoryRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&ProductCategory{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// SQLProductRepo implements domain.ProductRepository
type SQLProductRepo struct {
	db *gorm.DB
}

func NewSQLProductRepo(db *gorm.DB) *SQLProductRepo {
	return &SQLProductRepo{db: db}
}

func (r *SQLProductRepo) Create(ctx context.Context, p *domain.Product) error {
	dbModel := FromDomainProduct(p)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	p.CreatedAt = dbModel.CreatedAt
	p.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var dbModel Product
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainProduct(&dbModel), nil
}

func (r *SQLProductRepo) List(ctx context.Context) ([]domain.Product, error) {
	var dbModels []Product
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.Product, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainProduct(&m)
	}
	return res, nil
}

func (r *SQLProductRepo) Update(ctx context.Context, p *domain.Product) error {
	dbModel := FromDomainProduct(p)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLProductRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&Product{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// SQLLocationRepo implements domain.LocationRepository
type SQLLocationRepo struct {
	db *gorm.DB
}

func NewSQLLocationRepo(db *gorm.DB) *SQLLocationRepo {
	return &SQLLocationRepo{db: db}
}

func (r *SQLLocationRepo) Create(ctx context.Context, loc *domain.Location) error {
	dbModel := FromDomainLocation(loc)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	loc.CreatedAt = dbModel.CreatedAt
	loc.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLLocationRepo) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	var dbModel Location
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainLocation(&dbModel), nil
}

func (r *SQLLocationRepo) List(ctx context.Context) ([]domain.Location, error) {
	var dbModels []Location
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.Location, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainLocation(&m)
	}
	return res, nil
}

func (r *SQLLocationRepo) Update(ctx context.Context, loc *domain.Location) error {
	dbModel := FromDomainLocation(loc)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLLocationRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&Location{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// SQLSupplierRepo implements domain.SupplierRepository
type SQLSupplierRepo struct {
	db *gorm.DB
}

func NewSQLSupplierRepo(db *gorm.DB) *SQLSupplierRepo {
	return &SQLSupplierRepo{db: db}
}

func (r *SQLSupplierRepo) Create(ctx context.Context, s *domain.Supplier) error {
	dbModel := FromDomainSupplier(s)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	s.CreatedAt = dbModel.CreatedAt
	s.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLSupplierRepo) GetByID(ctx context.Context, id string) (*domain.Supplier, error) {
	var dbModel Supplier
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainSupplier(&dbModel), nil
}

func (r *SQLSupplierRepo) List(ctx context.Context) ([]domain.Supplier, error) {
	var dbModels []Supplier
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.Supplier, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainSupplier(&m)
	}
	return res, nil
}

func (r *SQLSupplierRepo) Update(ctx context.Context, s *domain.Supplier) error {
	dbModel := FromDomainSupplier(s)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLSupplierRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&Supplier{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// SQLVendorContractRepo implements domain.VendorContractRepository
type SQLVendorContractRepo struct {
	db *gorm.DB
}

func NewSQLVendorContractRepo(db *gorm.DB) *SQLVendorContractRepo {
	return &SQLVendorContractRepo{db: db}
}

func (r *SQLVendorContractRepo) Create(ctx context.Context, vc *domain.VendorContract) error {
	dbModel := FromDomainVendorContract(vc)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	vc.CreatedAt = dbModel.CreatedAt
	vc.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLVendorContractRepo) GetByID(ctx context.Context, id string) (*domain.VendorContract, error) {
	var dbModel VendorContract
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainVendorContract(&dbModel), nil
}

func (r *SQLVendorContractRepo) List(ctx context.Context) ([]domain.VendorContract, error) {
	var dbModels []VendorContract
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.VendorContract, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainVendorContract(&m)
	}
	return res, nil
}

func (r *SQLVendorContractRepo) Update(ctx context.Context, vc *domain.VendorContract) error {
	dbModel := FromDomainVendorContract(vc)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLVendorContractRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&VendorContract{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// SQLInventoryItemRepo implements domain.InventoryItemRepository with OCC version check
type SQLInventoryItemRepo struct {
	db *gorm.DB
}

func NewSQLInventoryItemRepo(db *gorm.DB) *SQLInventoryItemRepo {
	return &SQLInventoryItemRepo{db: db}
}

func (r *SQLInventoryItemRepo) Create(ctx context.Context, ii *domain.InventoryItem) error {
	dbModel := FromDomainInventoryItem(ii)
	dbModel.Version = 0
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	ii.CreatedAt = dbModel.CreatedAt
	ii.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLInventoryItemRepo) GetByID(ctx context.Context, id string) (*domain.InventoryItem, error) {
	var dbModel InventoryItem
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainInventoryItem(&dbModel), nil
}

func (r *SQLInventoryItemRepo) List(ctx context.Context) ([]domain.InventoryItem, error) {
	var dbModels []InventoryItem
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.InventoryItem, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainInventoryItem(&m)
	}
	return res, nil
}

func (r *SQLInventoryItemRepo) Update(ctx context.Context, ii *domain.InventoryItem) error {
	tx := GetDB(ctx, r.db)

	var dbModel InventoryItem
	if err := tx.First(&dbModel, "id = ?", ii.ID).Error; err != nil {
		return err
	}

	expectedVersion := dbModel.Version
	newVersion := expectedVersion + 1

	res := tx.Model(&InventoryItem{}).
		Where("id = ? AND version = ?", ii.ID, expectedVersion).
		Updates(map[string]interface{}{
			"quantity_on_hand":   ii.QuantityOnHand,
			"quantity_reserved":  ii.QuantityReserved,
			"quantity_available": ii.QuantityAvailable,
			"reorder_point":      ii.ReorderPoint,
			"maximum_stock":      ii.MaximumStock,
			"unit_cost":          ii.UnitCost,
			"updated_at":         time.Now(),
			"version":            newVersion,
		})

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return domain.ErrOptimisticLock
	}

	return nil
}

func (r *SQLInventoryItemRepo) GetByProductAndLocation(ctx context.Context, productID string, locationID string) (*domain.InventoryItem, error) {
	var dbModel InventoryItem
	if err := GetDB(ctx, r.db).First(&dbModel, "product_id = ? AND location_id = ?", productID, locationID).Error; err != nil {
		return nil, err
	}
	return ToDomainInventoryItem(&dbModel), nil
}

// SQLInventoryMovementRepo implements domain.InventoryMovementRepository
type SQLInventoryMovementRepo struct {
	db *gorm.DB
}

func NewSQLInventoryMovementRepo(db *gorm.DB) *SQLInventoryMovementRepo {
	return &SQLInventoryMovementRepo{db: db}
}

func (r *SQLInventoryMovementRepo) Create(ctx context.Context, im *domain.InventoryMovement) error {
	dbModel := FromDomainInventoryMovement(im)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	im.CreatedAt = dbModel.CreatedAt
	return nil
}

func (r *SQLInventoryMovementRepo) GetByID(ctx context.Context, id string) (*domain.InventoryMovement, error) {
	var dbModel InventoryMovement
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainInventoryMovement(&dbModel), nil
}

func (r *SQLInventoryMovementRepo) List(ctx context.Context) ([]domain.InventoryMovement, error) {
	var dbModels []InventoryMovement
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.InventoryMovement, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainInventoryMovement(&m)
	}
	return res, nil
}

// SQLStockTransferRepo implements domain.StockTransferRepository
type SQLStockTransferRepo struct {
	db *gorm.DB
}

func NewSQLStockTransferRepo(db *gorm.DB) *SQLStockTransferRepo {
	return &SQLStockTransferRepo{db: db}
}

func (r *SQLStockTransferRepo) Create(ctx context.Context, st *domain.StockTransfer) error {
	dbModel := FromDomainStockTransfer(st)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	st.CreatedAt = dbModel.CreatedAt
	st.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLStockTransferRepo) GetByID(ctx context.Context, id string) (*domain.StockTransfer, error) {
	var dbModel StockTransfer
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainStockTransfer(&dbModel), nil
}

func (r *SQLStockTransferRepo) List(ctx context.Context) ([]domain.StockTransfer, error) {
	var dbModels []StockTransfer
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.StockTransfer, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainStockTransfer(&m)
	}
	return res, nil
}

func (r *SQLStockTransferRepo) Update(ctx context.Context, st *domain.StockTransfer) error {
	dbModel := FromDomainStockTransfer(st)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

// SQLPurchaseRequisitionRepo implements domain.PurchaseRequisitionRepository
type SQLPurchaseRequisitionRepo struct {
	db *gorm.DB
}

func NewSQLPurchaseRequisitionRepo(db *gorm.DB) *SQLPurchaseRequisitionRepo {
	return &SQLPurchaseRequisitionRepo{db: db}
}

func (r *SQLPurchaseRequisitionRepo) Create(ctx context.Context, pr *domain.PurchaseRequisition) error {
	dbModel := FromDomainPurchaseRequisition(pr)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	pr.CreatedAt = dbModel.CreatedAt
	pr.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLPurchaseRequisitionRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseRequisition, error) {
	var dbModel PurchaseRequisition
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainPurchaseRequisition(&dbModel), nil
}

func (r *SQLPurchaseRequisitionRepo) List(ctx context.Context) ([]domain.PurchaseRequisition, error) {
	var dbModels []PurchaseRequisition
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.PurchaseRequisition, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainPurchaseRequisition(&m)
	}
	return res, nil
}

func (r *SQLPurchaseRequisitionRepo) Update(ctx context.Context, pr *domain.PurchaseRequisition) error {
	dbModel := FromDomainPurchaseRequisition(pr)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLPurchaseRequisitionRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&PurchaseRequisition{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// SQLPurchaseRequisitionLineRepo implements domain.PurchaseRequisitionLineRepository
type SQLPurchaseRequisitionLineRepo struct {
	db *gorm.DB
}

func NewSQLPurchaseRequisitionLineRepo(db *gorm.DB) *SQLPurchaseRequisitionLineRepo {
	return &SQLPurchaseRequisitionLineRepo{db: db}
}

func (r *SQLPurchaseRequisitionLineRepo) Create(ctx context.Context, prl *domain.PurchaseRequisitionLine) error {
	dbModel := FromDomainPurchaseRequisitionLine(prl)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLPurchaseRequisitionLineRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseRequisitionLine, error) {
	var dbModel PurchaseRequisitionLine
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainPurchaseRequisitionLine(&dbModel), nil
}

func (r *SQLPurchaseRequisitionLineRepo) ListByRequisitionID(ctx context.Context, reqID string) ([]domain.PurchaseRequisitionLine, error) {
	var dbModels []PurchaseRequisitionLine
	if err := GetDB(ctx, r.db).Where("purchase_requisition_id = ?", reqID).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.PurchaseRequisitionLine, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainPurchaseRequisitionLine(&m)
	}
	return res, nil
}

func (r *SQLPurchaseRequisitionLineRepo) DeleteByRequisitionID(ctx context.Context, reqID string) error {
	if err := GetDB(ctx, r.db).Delete(&PurchaseRequisitionLine{}, "purchase_requisition_id = ?", reqID).Error; err != nil {
		return err
	}
	return nil
}

// SQLPurchaseOrderRepo implements domain.PurchaseOrderRepository
type SQLPurchaseOrderRepo struct {
	db *gorm.DB
}

func NewSQLPurchaseOrderRepo(db *gorm.DB) *SQLPurchaseOrderRepo {
	return &SQLPurchaseOrderRepo{db: db}
}

func (r *SQLPurchaseOrderRepo) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	dbModel := FromDomainPurchaseOrder(po)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	po.CreatedAt = dbModel.CreatedAt
	po.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLPurchaseOrderRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	var dbModel PurchaseOrder
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainPurchaseOrder(&dbModel), nil
}

func (r *SQLPurchaseOrderRepo) List(ctx context.Context) ([]domain.PurchaseOrder, error) {
	var dbModels []PurchaseOrder
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.PurchaseOrder, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainPurchaseOrder(&m)
	}
	return res, nil
}

func (r *SQLPurchaseOrderRepo) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	dbModel := FromDomainPurchaseOrder(po)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLPurchaseOrderRepo) Delete(ctx context.Context, id string) error {
	if err := GetDB(ctx, r.db).Delete(&PurchaseOrder{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// SQLPurchaseOrderLineRepo implements domain.PurchaseOrderLineRepository
type SQLPurchaseOrderLineRepo struct {
	db *gorm.DB
}

func NewSQLPurchaseOrderLineRepo(db *gorm.DB) *SQLPurchaseOrderLineRepo {
	return &SQLPurchaseOrderLineRepo{db: db}
}

func (r *SQLPurchaseOrderLineRepo) Create(ctx context.Context, pol *domain.PurchaseOrderLine) error {
	dbModel := FromDomainPurchaseOrderLine(pol)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	pol.CreatedAt = dbModel.CreatedAt
	return nil
}

func (r *SQLPurchaseOrderLineRepo) GetByID(ctx context.Context, id string) (*domain.PurchaseOrderLine, error) {
	var dbModel PurchaseOrderLine
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainPurchaseOrderLine(&dbModel), nil
}

func (r *SQLPurchaseOrderLineRepo) ListByPOID(ctx context.Context, poID string) ([]domain.PurchaseOrderLine, error) {
	var dbModels []PurchaseOrderLine
	if err := GetDB(ctx, r.db).Where("purchase_order_id = ?", poID).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.PurchaseOrderLine, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainPurchaseOrderLine(&m)
	}
	return res, nil
}

func (r *SQLPurchaseOrderLineRepo) DeleteByPOID(ctx context.Context, poID string) error {
	if err := GetDB(ctx, r.db).Delete(&PurchaseOrderLine{}, "purchase_order_id = ?", poID).Error; err != nil {
		return err
	}
	return nil
}

// SQLReceiptRepo implements domain.ReceiptRepository
type SQLReceiptRepo struct {
	db *gorm.DB
}

func NewSQLReceiptRepo(db *gorm.DB) *SQLReceiptRepo {
	return &SQLReceiptRepo{db: db}
}

func (r *SQLReceiptRepo) Create(ctx context.Context, rec *domain.Receipt) error {
	dbModel := FromDomainReceipt(rec)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	rec.CreatedAt = dbModel.CreatedAt
	rec.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLReceiptRepo) GetByID(ctx context.Context, id string) (*domain.Receipt, error) {
	var dbModel Receipt
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainReceipt(&dbModel), nil
}

func (r *SQLReceiptRepo) List(ctx context.Context) ([]domain.Receipt, error) {
	var dbModels []Receipt
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.Receipt, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainReceipt(&m)
	}
	return res, nil
}

func (r *SQLReceiptRepo) Update(ctx context.Context, rec *domain.Receipt) error {
	dbModel := FromDomainReceipt(rec)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

// SQLReceiptLineRepo implements domain.ReceiptLineRepository
type SQLReceiptLineRepo struct {
	db *gorm.DB
}

func NewSQLReceiptLineRepo(db *gorm.DB) *SQLReceiptLineRepo {
	return &SQLReceiptLineRepo{db: db}
}

func (r *SQLReceiptLineRepo) Create(ctx context.Context, rl *domain.ReceiptLine) error {
	dbModel := FromDomainReceiptLine(rl)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	rl.CreatedAt = dbModel.CreatedAt
	return nil
}

func (r *SQLReceiptLineRepo) GetByID(ctx context.Context, id string) (*domain.ReceiptLine, error) {
	var dbModel ReceiptLine
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainReceiptLine(&dbModel), nil
}

func (r *SQLReceiptLineRepo) ListByReceiptID(ctx context.Context, receiptID string) ([]domain.ReceiptLine, error) {
	var dbModels []ReceiptLine
	if err := GetDB(ctx, r.db).Where("receipt_id = ?", receiptID).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.ReceiptLine, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainReceiptLine(&m)
	}
	return res, nil
}

// SQLShipmentRepo implements domain.ShipmentRepository
type SQLShipmentRepo struct {
	db *gorm.DB
}

func NewSQLShipmentRepo(db *gorm.DB) *SQLShipmentRepo {
	return &SQLShipmentRepo{db: db}
}

func (r *SQLShipmentRepo) Create(ctx context.Context, s *domain.Shipment) error {
	dbModel := FromDomainShipment(s)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	s.CreatedAt = dbModel.CreatedAt
	s.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLShipmentRepo) GetByID(ctx context.Context, id string) (*domain.Shipment, error) {
	var dbModel Shipment
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainShipment(&dbModel), nil
}

func (r *SQLShipmentRepo) List(ctx context.Context) ([]domain.Shipment, error) {
	var dbModels []Shipment
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.Shipment, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainShipment(&m)
	}
	return res, nil
}

func (r *SQLShipmentRepo) Update(ctx context.Context, s *domain.Shipment) error {
	dbModel := FromDomainShipment(s)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

// SQLShipmentLineRepo implements domain.ShipmentLineRepository
type SQLShipmentLineRepo struct {
	db *gorm.DB
}

func NewSQLShipmentLineRepo(db *gorm.DB) *SQLShipmentLineRepo {
	return &SQLShipmentLineRepo{db: db}
}

func (r *SQLShipmentLineRepo) Create(ctx context.Context, sl *domain.ShipmentLine) error {
	dbModel := FromDomainShipmentLine(sl)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	sl.CreatedAt = dbModel.CreatedAt
	return nil
}

func (r *SQLShipmentLineRepo) GetByID(ctx context.Context, id string) (*domain.ShipmentLine, error) {
	var dbModel ShipmentLine
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainShipmentLine(&dbModel), nil
}

func (r *SQLShipmentLineRepo) ListByShipmentID(ctx context.Context, shipmentID string) ([]domain.ShipmentLine, error) {
	var dbModels []ShipmentLine
	if err := GetDB(ctx, r.db).Where("shipment_id = ?", shipmentID).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.ShipmentLine, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainShipmentLine(&m)
	}
	return res, nil
}

// SQLDemandForecastRepo implements domain.DemandForecastRepository
type SQLDemandForecastRepo struct {
	db *gorm.DB
}

func NewSQLDemandForecastRepo(db *gorm.DB) *SQLDemandForecastRepo {
	return &SQLDemandForecastRepo{db: db}
}

func (r *SQLDemandForecastRepo) Create(ctx context.Context, df *domain.DemandForecast) error {
	dbModel := FromDomainDemandForecast(df)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	df.CreatedAt = dbModel.CreatedAt
	df.UpdatedAt = dbModel.UpdatedAt
	return nil
}

func (r *SQLDemandForecastRepo) GetByID(ctx context.Context, id string) (*domain.DemandForecast, error) {
	var dbModel DemandForecast
	if err := GetDB(ctx, r.db).First(&dbModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainDemandForecast(&dbModel), nil
}

func (r *SQLDemandForecastRepo) List(ctx context.Context) ([]domain.DemandForecast, error) {
	var dbModels []DemandForecast
	if err := GetDB(ctx, r.db).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.DemandForecast, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainDemandForecast(&m)
	}
	return res, nil
}

func (r *SQLDemandForecastRepo) Update(ctx context.Context, df *domain.DemandForecast) error {
	dbModel := FromDomainDemandForecast(df)
	if err := GetDB(ctx, r.db).Save(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLDemandForecastRepo) ListByProductID(ctx context.Context, productID string) ([]domain.DemandForecast, error) {
	var dbModels []DemandForecast
	if err := GetDB(ctx, r.db).Where("product_id = ?", productID).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.DemandForecast, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainDemandForecast(&m)
	}
	return res, nil
}

// SQLKafkaEventInboxRepo implements domain.KafkaEventInboxRepository
type SQLKafkaEventInboxRepo struct {
	db *gorm.DB
}

func NewSQLKafkaEventInboxRepo(db *gorm.DB) *SQLKafkaEventInboxRepo {
	return &SQLKafkaEventInboxRepo{db: db}
}

func (r *SQLKafkaEventInboxRepo) Create(ctx context.Context, e *domain.KafkaEventInbox) error {
	dbModel := FromDomainKafkaEventInbox(e)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLKafkaEventInboxRepo) GetByID(ctx context.Context, id string) (*domain.KafkaEventInbox, error) {
	var dbModel KafkaEventInbox
	if err := GetDB(ctx, r.db).First(&dbModel, "event_id = ?", id).Error; err != nil {
		return nil, err
	}
	return ToDomainKafkaEventInbox(&dbModel), nil
}

// SQLTransactionalOutboxRepo implements domain.TransactionalOutboxRepository
type SQLTransactionalOutboxRepo struct {
	db *gorm.DB
}

func NewSQLTransactionalOutboxRepo(db *gorm.DB) *SQLTransactionalOutboxRepo {
	return &SQLTransactionalOutboxRepo{db: db}
}

func (r *SQLTransactionalOutboxRepo) Create(ctx context.Context, o *domain.TransactionalOutbox) error {
	dbModel := FromDomainTransactionalOutbox(o)
	if err := GetDB(ctx, r.db).Create(dbModel).Error; err != nil {
		return err
	}
	return nil
}

func (r *SQLTransactionalOutboxRepo) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	var dbModels []TransactionalOutbox
	if err := GetDB(ctx, r.db).Where("status IN ?", []domain.OutboxStatus{domain.OutboxStatusPENDING, domain.OutboxStatusFAILED}).Limit(limit).Find(&dbModels).Error; err != nil {
		return nil, err
	}
	res := make([]domain.TransactionalOutbox, len(dbModels))
	for i, m := range dbModels {
		res[i] = *ToDomainTransactionalOutbox(&m)
	}
	return res, nil
}

func (r *SQLTransactionalOutboxRepo) UpdateStatus(ctx context.Context, id string, status domain.OutboxStatus, retryCount int) error {
	if err := GetDB(ctx, r.db).Model(&TransactionalOutbox{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"retry_count": retryCount,
	}).Error; err != nil {
		return err
	}
	return nil
}
