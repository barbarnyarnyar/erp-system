package sql

import (
	"context"
	"errors"

	"github.com/erp-system/m-service/internal/business/domain"
	"gorm.io/gorm"
)

// ==========================================
// Work Center Repository
// ==========================================

type SQLWorkCenterRepository struct {
	db *gorm.DB
}

func NewSQLWorkCenterRepository(db *gorm.DB) domain.WorkCenterRepository {
	return &SQLWorkCenterRepository{db: db}
}

func (r *SQLWorkCenterRepository) Create(ctx context.Context, wc *domain.WorkCenter) error {
	db := GetDB(ctx, r.db)
	entity := FromWorkCenterDomain(wc)
	return db.Create(entity).Error
}

func (r *SQLWorkCenterRepository) GetByID(ctx context.Context, id string) (*domain.WorkCenter, error) {
	db := GetDB(ctx, r.db)
	var entity WorkCenter
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("work center not found")
		}
		return nil, err
	}
	return ToWorkCenterDomain(&entity), nil
}

func (r *SQLWorkCenterRepository) GetByCode(ctx context.Context, legalEntityID, code string) (*domain.WorkCenter, error) {
	db := GetDB(ctx, r.db)
	var entity WorkCenter
	err := db.First(&entity, "legal_entity_id = ? AND work_center_code = ?", legalEntityID, code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("work center not found by code")
		}
		return nil, err
	}
	return ToWorkCenterDomain(&entity), nil
}

func (r *SQLWorkCenterRepository) List(ctx context.Context) ([]domain.WorkCenter, error) {
	db := GetDB(ctx, r.db)
	var entities []WorkCenter
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.WorkCenter, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToWorkCenterDomain(&e))
	}
	return list, nil
}

func (r *SQLWorkCenterRepository) Update(ctx context.Context, wc *domain.WorkCenter) error {
	db := GetDB(ctx, r.db)
	entity := FromWorkCenterDomain(wc)
	return db.Save(entity).Error
}

func (r *SQLWorkCenterRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&WorkCenter{}, "id = ?", id).Error
}

// ==========================================
// Routing Station Repository
// ==========================================

type SQLRoutingStationRepository struct {
	db *gorm.DB
}

func NewSQLRoutingStationRepository(db *gorm.DB) domain.RoutingStationRepository {
	return &SQLRoutingStationRepository{db: db}
}

func (r *SQLRoutingStationRepository) Create(ctx context.Context, station *domain.RoutingStation) error {
	db := GetDB(ctx, r.db)
	entity := FromRoutingStationDomain(station)
	return db.Create(entity).Error
}

func (r *SQLRoutingStationRepository) GetByID(ctx context.Context, id string) (*domain.RoutingStation, error) {
	db := GetDB(ctx, r.db)
	var entity RoutingStation
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("routing station not found")
		}
		return nil, err
	}
	return ToRoutingStationDomain(&entity), nil
}

func (r *SQLRoutingStationRepository) GetByCode(ctx context.Context, workCenterID, code string) (*domain.RoutingStation, error) {
	db := GetDB(ctx, r.db)
	var entity RoutingStation
	err := db.First(&entity, "work_center_id = ? AND routing_code = ?", workCenterID, code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("routing station not found by code")
		}
		return nil, err
	}
	return ToRoutingStationDomain(&entity), nil
}

func (r *SQLRoutingStationRepository) ListByWorkCenterID(ctx context.Context, workCenterID string) ([]domain.RoutingStation, error) {
	db := GetDB(ctx, r.db)
	var entities []RoutingStation
	err := db.Find(&entities, "work_center_id = ?", workCenterID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.RoutingStation, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToRoutingStationDomain(&e))
	}
	return list, nil
}

func (r *SQLRoutingStationRepository) Update(ctx context.Context, station *domain.RoutingStation) error {
	db := GetDB(ctx, r.db)
	entity := FromRoutingStationDomain(station)
	return db.Save(entity).Error
}

func (r *SQLRoutingStationRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&RoutingStation{}, "id = ?", id).Error
}

// ==========================================
// Work Order Repository
// ==========================================

type SQLWorkOrderRepository struct {
	db *gorm.DB
}

func NewSQLWorkOrderRepository(db *gorm.DB) domain.WorkOrderRepository {
	return &SQLWorkOrderRepository{db: db}
}

func (r *SQLWorkOrderRepository) Create(ctx context.Context, wo *domain.WorkOrder) error {
	db := GetDB(ctx, r.db)
	entity := FromWorkOrderDomain(wo)
	return db.Create(entity).Error
}

func (r *SQLWorkOrderRepository) GetByID(ctx context.Context, id string) (*domain.WorkOrder, error) {
	db := GetDB(ctx, r.db)
	var entity WorkOrder
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("work order not found")
		}
		return nil, err
	}
	return ToWorkOrderDomain(&entity), nil
}

func (r *SQLWorkOrderRepository) GetByNumber(ctx context.Context, legalEntityID, number string) (*domain.WorkOrder, error) {
	db := GetDB(ctx, r.db)
	var entity WorkOrder
	err := db.First(&entity, "legal_entity_id = ? AND work_order_number = ?", legalEntityID, number).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("work order not found by number")
		}
		return nil, err
	}
	return ToWorkOrderDomain(&entity), nil
}

func (r *SQLWorkOrderRepository) List(ctx context.Context) ([]domain.WorkOrder, error) {
	db := GetDB(ctx, r.db)
	var entities []WorkOrder
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.WorkOrder, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToWorkOrderDomain(&e))
	}
	return list, nil
}

func (r *SQLWorkOrderRepository) Update(ctx context.Context, wo *domain.WorkOrder) error {
	db := GetDB(ctx, r.db)
	entity := FromWorkOrderDomain(wo)
	res := db.Model(&WorkOrder{}).
		Where("id = ? AND version = ?", entity.ID, wo.Version).
		Updates(map[string]interface{}{
			"quantity_produced": entity.QuantityProduced,
			"status":            entity.Status,
			"version":           gorm.Expr("version + 1"),
			"updated_at":        entity.UpdatedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("concurrent modification error: work order version mismatch")
	}
	wo.Version++
	return nil
}

func (r *SQLWorkOrderRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&WorkOrder{}, "id = ?", id).Error
}

// ==========================================
// Work Order Routing State Repository
// ==========================================

type SQLWorkOrderRoutingStateRepository struct {
	db *gorm.DB
}

func NewSQLWorkOrderRoutingStateRepository(db *gorm.DB) domain.WorkOrderRoutingStateRepository {
	return &SQLWorkOrderRoutingStateRepository{db: db}
}

func (r *SQLWorkOrderRoutingStateRepository) Create(ctx context.Context, state *domain.WorkOrderRoutingState) error {
	db := GetDB(ctx, r.db)
	entity := FromWorkOrderRoutingStateDomain(state)
	return db.Create(entity).Error
}

func (r *SQLWorkOrderRoutingStateRepository) GetByID(ctx context.Context, id string) (*domain.WorkOrderRoutingState, error) {
	db := GetDB(ctx, r.db)
	var entity WorkOrderRoutingState
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("routing state not found")
		}
		return nil, err
	}
	return ToWorkOrderRoutingStateDomain(&entity), nil
}

func (r *SQLWorkOrderRoutingStateRepository) GetActiveByWorkOrderID(ctx context.Context, workOrderID string) (*domain.WorkOrderRoutingState, error) {
	db := GetDB(ctx, r.db)
	var entity WorkOrderRoutingState
	err := db.First(&entity, "work_order_id = ? AND exited_at IS NULL", workOrderID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("active routing state not found")
		}
		return nil, err
	}
	return ToWorkOrderRoutingStateDomain(&entity), nil
}

func (r *SQLWorkOrderRoutingStateRepository) ListByWorkOrderID(ctx context.Context, workOrderID string) ([]domain.WorkOrderRoutingState, error) {
	db := GetDB(ctx, r.db)
	var entities []WorkOrderRoutingState
	err := db.Find(&entities, "work_order_id = ?", workOrderID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.WorkOrderRoutingState, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToWorkOrderRoutingStateDomain(&e))
	}
	return list, nil
}

func (r *SQLWorkOrderRoutingStateRepository) Update(ctx context.Context, state *domain.WorkOrderRoutingState) error {
	db := GetDB(ctx, r.db)
	entity := FromWorkOrderRoutingStateDomain(state)
	return db.Save(entity).Error
}

// ==========================================
// Material Consumption Log Repository
// ==========================================

type SQLMaterialConsumptionLogRepository struct {
	db *gorm.DB
}

func NewSQLMaterialConsumptionLogRepository(db *gorm.DB) domain.MaterialConsumptionLogRepository {
	return &SQLMaterialConsumptionLogRepository{db: db}
}

func (r *SQLMaterialConsumptionLogRepository) Create(ctx context.Context, log *domain.MaterialConsumptionLog) error {
	db := GetDB(ctx, r.db)
	entity := FromMaterialConsumptionLogDomain(log)
	return db.Create(entity).Error
}

func (r *SQLMaterialConsumptionLogRepository) GetByID(ctx context.Context, id string) (*domain.MaterialConsumptionLog, error) {
	db := GetDB(ctx, r.db)
	var entity MaterialConsumptionLog
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("consumption log not found")
		}
		return nil, err
	}
	return ToMaterialConsumptionLogDomain(&entity), nil
}

func (r *SQLMaterialConsumptionLogRepository) ListByWorkOrderID(ctx context.Context, workOrderID string) ([]domain.MaterialConsumptionLog, error) {
	db := GetDB(ctx, r.db)
	var entities []MaterialConsumptionLog
	err := db.Find(&entities, "work_order_id = ?", workOrderID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.MaterialConsumptionLog, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToMaterialConsumptionLogDomain(&e))
	}
	return list, nil
}

// ==========================================
// Production Yield Log Repository
// ==========================================

type SQLProductionYieldLogRepository struct {
	db *gorm.DB
}

func NewSQLProductionYieldLogRepository(db *gorm.DB) domain.ProductionYieldLogRepository {
	return &SQLProductionYieldLogRepository{db: db}
}

func (r *SQLProductionYieldLogRepository) Create(ctx context.Context, log *domain.ProductionYieldLog) error {
	db := GetDB(ctx, r.db)
	entity := FromProductionYieldLogDomain(log)
	return db.Create(entity).Error
}

func (r *SQLProductionYieldLogRepository) GetByID(ctx context.Context, id string) (*domain.ProductionYieldLog, error) {
	db := GetDB(ctx, r.db)
	var entity ProductionYieldLog
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("yield log not found")
		}
		return nil, err
	}
	return ToProductionYieldLogDomain(&entity), nil
}

func (r *SQLProductionYieldLogRepository) ListByWorkOrderID(ctx context.Context, workOrderID string) ([]domain.ProductionYieldLog, error) {
	db := GetDB(ctx, r.db)
	var entities []ProductionYieldLog
	err := db.Find(&entities, "work_order_id = ?", workOrderID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.ProductionYieldLog, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToProductionYieldLogDomain(&e))
	}
	return list, nil
}

// ==========================================
// Transactional Outbox Repository
// ==========================================

type SQLTransactionalOutboxRepository struct {
	db *gorm.DB
}

func NewSQLTransactionalOutboxRepository(db *gorm.DB) domain.TransactionalOutboxRepository {
	return &SQLTransactionalOutboxRepository{db: db}
}

func (r *SQLTransactionalOutboxRepository) Create(ctx context.Context, msg *domain.TransactionalOutbox) error {
	db := GetDB(ctx, r.db)
	entity := FromTransactionalOutboxDomain(msg)
	return db.Create(entity).Error
}

func (r *SQLTransactionalOutboxRepository) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	db := GetDB(ctx, r.db)
	var entity TransactionalOutbox
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("outbox message not found")
		}
		return nil, err
	}
	return ToTransactionalOutboxDomain(&entity), nil
}

func (r *SQLTransactionalOutboxRepository) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	db := GetDB(ctx, r.db)
	var entities []TransactionalOutbox
	err := db.Where("status = ?", string(domain.OutboxStatusPENDING)).Order("created_at asc").Limit(limit).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.TransactionalOutbox, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToTransactionalOutboxDomain(&e))
	}
	return list, nil
}

func (r *SQLTransactionalOutboxRepository) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	db := GetDB(ctx, r.db)
	entity := FromTransactionalOutboxDomain(msg)
	return db.Save(entity).Error
}

// ==========================================
// Kafka Event Inbox Repository
// ==========================================

type SQLKafkaEventInboxRepository struct {
	db *gorm.DB
}

func NewSQLKafkaEventInboxRepository(db *gorm.DB) domain.KafkaEventInboxRepository {
	return &SQLKafkaEventInboxRepository{db: db}
}

func (r *SQLKafkaEventInboxRepository) Create(ctx context.Context, msg *domain.KafkaEventInbox) error {
	db := GetDB(ctx, r.db)
	entity := FromKafkaEventInboxDomain(msg)
	return db.Create(entity).Error
}

func (r *SQLKafkaEventInboxRepository) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	db := GetDB(ctx, r.db)
	var entity KafkaEventInbox
	err := db.First(&entity, "event_id = ?", eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("inbox message not found")
		}
		return nil, err
	}
	return ToKafkaEventInboxDomain(&entity), nil
}

func (r *SQLKafkaEventInboxRepository) Update(ctx context.Context, msg *domain.KafkaEventInbox) error {
	db := GetDB(ctx, r.db)
	entity := FromKafkaEventInboxDomain(msg)
	return db.Save(entity).Error
}
