package sql

import (
	"context"
	"errors"

	"github.com/erp-system/plm-service/internal/business/domain"
	"gorm.io/gorm"
)

type SQLMaterialMasterRepository struct {
	db *gorm.DB
}

func NewSQLMaterialMasterRepository(db *gorm.DB) domain.MaterialMasterRepository {
	return &SQLMaterialMasterRepository{db: db}
}

func (r *SQLMaterialMasterRepository) Create(ctx context.Context, m *domain.MaterialMaster) error {
	db := GetDB(ctx, r.db)
	entity := FromMaterialMasterDomain(m)
	return db.Create(entity).Error
}

func (r *SQLMaterialMasterRepository) GetByID(ctx context.Context, id string) (*domain.MaterialMaster, error) {
	db := GetDB(ctx, r.db)
	var entity MaterialMaster
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material not found")
		}
		return nil, err
	}
	return ToMaterialMasterDomain(&entity), nil
}

func (r *SQLMaterialMasterRepository) GetBySKU(ctx context.Context, legalEntityId string, sku string) (*domain.MaterialMaster, error) {
	db := GetDB(ctx, r.db)
	var entity MaterialMaster
	err := db.First(&entity, "legal_entity_id = ? AND sku = ?", legalEntityId, sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material with sku not found")
		}
		return nil, err
	}
	return ToMaterialMasterDomain(&entity), nil
}

func (r *SQLMaterialMasterRepository) List(ctx context.Context) ([]domain.MaterialMaster, error) {
	db := GetDB(ctx, r.db)
	var entities []MaterialMaster
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.MaterialMaster, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToMaterialMasterDomain(&e))
	}
	return list, nil
}

func (r *SQLMaterialMasterRepository) Update(ctx context.Context, m *domain.MaterialMaster) error {
	db := GetDB(ctx, r.db)
	entity := FromMaterialMasterDomain(m)
	return db.Save(entity).Error
}

type SQLBomHeaderRepository struct {
	db *gorm.DB
}

func NewSQLBomHeaderRepository(db *gorm.DB) domain.BomHeaderRepository {
	return &SQLBomHeaderRepository{db: db}
}

func (r *SQLBomHeaderRepository) Create(ctx context.Context, bh *domain.BomHeader) error {
	db := GetDB(ctx, r.db)
	entity := FromBomHeaderDomain(bh)
	return db.Create(entity).Error
}

func (r *SQLBomHeaderRepository) GetByID(ctx context.Context, id string) (*domain.BomHeader, error) {
	db := GetDB(ctx, r.db)
	var entity BomHeader
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bom header not found")
		}
		return nil, err
	}
	return ToBomHeaderDomain(&entity), nil
}

func (r *SQLBomHeaderRepository) List(ctx context.Context) ([]domain.BomHeader, error) {
	db := GetDB(ctx, r.db)
	var entities []BomHeader
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.BomHeader, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToBomHeaderDomain(&e))
	}
	return list, nil
}

func (r *SQLBomHeaderRepository) Update(ctx context.Context, bh *domain.BomHeader) error {
	db := GetDB(ctx, r.db)
	entity := FromBomHeaderDomain(bh)
	return db.Save(entity).Error
}

type SQLBomLineRepository struct {
	db *gorm.DB
}

func NewSQLBomLineRepository(db *gorm.DB) domain.BomLineRepository {
	return &SQLBomLineRepository{db: db}
}

func (r *SQLBomLineRepository) Create(ctx context.Context, bl *domain.BomLine) error {
	db := GetDB(ctx, r.db)
	entity := FromBomLineDomain(bl)
	return db.Create(entity).Error
}

func (r *SQLBomLineRepository) GetByID(ctx context.Context, id string) (*domain.BomLine, error) {
	db := GetDB(ctx, r.db)
	var entity BomLine
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bom line not found")
		}
		return nil, err
	}
	return ToBomLineDomain(&entity), nil
}

func (r *SQLBomLineRepository) ListByHeaderID(ctx context.Context, headerID string) ([]domain.BomLine, error) {
	db := GetDB(ctx, r.db)
	var entities []BomLine
	err := db.Find(&entities, "bom_header_id = ?", headerID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.BomLine, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToBomLineDomain(&e))
	}
	return list, nil
}

func (r *SQLBomLineRepository) DeleteByHeaderID(ctx context.Context, headerID string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&BomLine{}, "bom_header_id = ?", headerID).Error
}

type SQLEngineeringChangeOrderRepository struct {
	db *gorm.DB
}

func NewSQLEngineeringChangeOrderRepository(db *gorm.DB) domain.EngineeringChangeOrderRepository {
	return &SQLEngineeringChangeOrderRepository{db: db}
}

func (r *SQLEngineeringChangeOrderRepository) Create(ctx context.Context, eco *domain.EngineeringChangeOrder) error {
	db := GetDB(ctx, r.db)
	entity := FromEcoDomain(eco)
	return db.Create(entity).Error
}

func (r *SQLEngineeringChangeOrderRepository) GetByID(ctx context.Context, id string) (*domain.EngineeringChangeOrder, error) {
	db := GetDB(ctx, r.db)
	var entity EngineeringChangeOrder
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("eco not found")
		}
		return nil, err
	}
	return ToEcoDomain(&entity), nil
}

func (r *SQLEngineeringChangeOrderRepository) List(ctx context.Context) ([]domain.EngineeringChangeOrder, error) {
	db := GetDB(ctx, r.db)
	var entities []EngineeringChangeOrder
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.EngineeringChangeOrder, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToEcoDomain(&e))
	}
	return list, nil
}

func (r *SQLEngineeringChangeOrderRepository) Update(ctx context.Context, eco *domain.EngineeringChangeOrder) error {
	db := GetDB(ctx, r.db)
	entity := FromEcoDomain(eco)
	return db.Save(entity).Error
}

type SQLTransactionalOutboxRepository struct {
	db *gorm.DB
}

func NewSQLTransactionalOutboxRepository(db *gorm.DB) domain.TransactionalOutboxRepository {
	return &SQLTransactionalOutboxRepository{db: db}
}

func (r *SQLTransactionalOutboxRepository) Create(ctx context.Context, outbox *domain.TransactionalOutbox) error {
	db := GetDB(ctx, r.db)
	entity := FromOutboxDomain(outbox)
	return db.Create(entity).Error
}

func (r *SQLTransactionalOutboxRepository) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	db := GetDB(ctx, r.db)
	var entities []TransactionalOutbox
	err := db.Limit(limit).Find(&entities, "status = ?", string(domain.OutboxStatusPENDING)).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.TransactionalOutbox, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToOutboxDomain(&e))
	}
	return list, nil
}

func (r *SQLTransactionalOutboxRepository) UpdateStatus(ctx context.Context, id string, status domain.OutboxStatus) error {
	db := GetDB(ctx, r.db)
	return db.Model(&TransactionalOutbox{}).Where("id = ?", id).Update("status", string(status)).Error
}

type SQLKafkaEventInboxRepository struct {
	db *gorm.DB
}

func NewSQLKafkaEventInboxRepository(db *gorm.DB) domain.KafkaEventInboxRepository {
	return &SQLKafkaEventInboxRepository{db: db}
}

func (r *SQLKafkaEventInboxRepository) Create(ctx context.Context, inbox *domain.KafkaEventInbox) error {
	db := GetDB(ctx, r.db)
	entity := FromInboxDomain(inbox)
	return db.Create(entity).Error
}

func (r *SQLKafkaEventInboxRepository) Exists(ctx context.Context, eventID string) (bool, error) {
	db := GetDB(ctx, r.db)
	var count int64
	err := db.Model(&KafkaEventInbox{}).Where("event_id = ?", eventID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
