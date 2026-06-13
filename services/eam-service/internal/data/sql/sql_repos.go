package sql

import (
	"context"
	"errors"

	"github.com/erp-system/eam-service/internal/business/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ==========================================
// Facility Repository
// ==========================================

type SQLFacilityRepository struct {
	db *gorm.DB
}

func NewSQLFacilityRepository(db *gorm.DB) domain.FacilityRepository {
	return &SQLFacilityRepository{db: db}
}

func (r *SQLFacilityRepository) Create(ctx context.Context, f *domain.Facility) error {
	db := GetDB(ctx, r.db)
	entity := FromFacilityDomain(f)
	return db.Create(entity).Error
}

func (r *SQLFacilityRepository) GetByID(ctx context.Context, id string) (*domain.Facility, error) {
	db := GetDB(ctx, r.db)
	var entity Facility
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("facility not found")
		}
		return nil, err
	}
	return ToFacilityDomain(&entity), nil
}

func (r *SQLFacilityRepository) List(ctx context.Context) ([]domain.Facility, error) {
	db := GetDB(ctx, r.db)
	var entities []Facility
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Facility, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToFacilityDomain(&e))
	}
	return list, nil
}

func (r *SQLFacilityRepository) Update(ctx context.Context, f *domain.Facility) error {
	db := GetDB(ctx, r.db)
	entity := FromFacilityDomain(f)
	return db.Save(entity).Error
}

// ==========================================
// Equipment Repository
// ==========================================

type SQLEquipmentRepository struct {
	db *gorm.DB
}

func NewSQLEquipmentRepository(db *gorm.DB) domain.EquipmentRepository {
	return &SQLEquipmentRepository{db: db}
}

func (r *SQLEquipmentRepository) Create(ctx context.Context, eq *domain.Equipment) error {
	db := GetDB(ctx, r.db)
	entity := FromEquipmentDomain(eq)
	return db.Create(entity).Error
}

func (r *SQLEquipmentRepository) GetByID(ctx context.Context, id string) (*domain.Equipment, error) {
	db := GetDB(ctx, r.db)
	var entity Equipment
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("equipment not found")
		}
		return nil, err
	}
	return ToEquipmentDomain(&entity), nil
}

func (r *SQLEquipmentRepository) List(ctx context.Context) ([]domain.Equipment, error) {
	db := GetDB(ctx, r.db)
	var entities []Equipment
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Equipment, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToEquipmentDomain(&e))
	}
	return list, nil
}

func (r *SQLEquipmentRepository) Update(ctx context.Context, eq *domain.Equipment) error {
	db := GetDB(ctx, r.db)
	entity := FromEquipmentDomain(eq)
	return db.Save(entity).Error
}

func (r *SQLEquipmentRepository) ListByTenant(ctx context.Context, legalEntityId string) ([]domain.Equipment, error) {
	db := GetDB(ctx, r.db)
	var entities []Equipment
	err := db.Where("legal_entity_id = ?", legalEntityId).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Equipment, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToEquipmentDomain(&e))
	}
	return list, nil
}

func (r *SQLEquipmentRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&Equipment{}, "id = ?", id).Error
}

// ==========================================
// MaintenanceWorkOrder Repository
// ==========================================

type SQLMaintenanceWorkOrderRepository struct {
	db *gorm.DB
}

func NewSQLMaintenanceWorkOrderRepository(db *gorm.DB) domain.MaintenanceWorkOrderRepository {
	return &SQLMaintenanceWorkOrderRepository{db: db}
}

func (r *SQLMaintenanceWorkOrderRepository) Create(ctx context.Context, wo *domain.MaintenanceWorkOrder) error {
	db := GetDB(ctx, r.db)
	entity := FromMaintenanceWorkOrderDomain(wo)
	return db.Create(entity).Error
}

func (r *SQLMaintenanceWorkOrderRepository) GetByID(ctx context.Context, id string) (*domain.MaintenanceWorkOrder, error) {
	db := GetDB(ctx, r.db)
	var entity MaintenanceWorkOrder
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("work order not found")
		}
		return nil, err
	}
	return ToMaintenanceWorkOrderDomain(&entity), nil
}

func (r *SQLMaintenanceWorkOrderRepository) List(ctx context.Context) ([]domain.MaintenanceWorkOrder, error) {
	db := GetDB(ctx, r.db)
	var entities []MaintenanceWorkOrder
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.MaintenanceWorkOrder, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToMaintenanceWorkOrderDomain(&e))
	}
	return list, nil
}

func (r *SQLMaintenanceWorkOrderRepository) Update(ctx context.Context, wo *domain.MaintenanceWorkOrder) error {
	db := GetDB(ctx, r.db)
	entity := FromMaintenanceWorkOrderDomain(wo)
	return db.Save(entity).Error
}

// ==========================================
// PreventativeSchedule Repository
// ==========================================

type SQLPreventativeScheduleRepository struct {
	db *gorm.DB
}

func NewSQLPreventativeScheduleRepository(db *gorm.DB) domain.PreventativeScheduleRepository {
	return &SQLPreventativeScheduleRepository{db: db}
}

func (r *SQLPreventativeScheduleRepository) Create(ctx context.Context, ps *domain.PreventativeSchedule) error {
	db := GetDB(ctx, r.db)
	entity := FromPreventativeScheduleDomain(ps)
	return db.Create(entity).Error
}

func (r *SQLPreventativeScheduleRepository) GetByID(ctx context.Context, id string) (*domain.PreventativeSchedule, error) {
	db := GetDB(ctx, r.db)
	var entity PreventativeSchedule
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("schedule not found")
		}
		return nil, err
	}
	return ToPreventativeScheduleDomain(&entity), nil
}

func (r *SQLPreventativeScheduleRepository) List(ctx context.Context) ([]domain.PreventativeSchedule, error) {
	db := GetDB(ctx, r.db)
	var entities []PreventativeSchedule
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.PreventativeSchedule, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToPreventativeScheduleDomain(&e))
	}
	return list, nil
}

func (r *SQLPreventativeScheduleRepository) Update(ctx context.Context, ps *domain.PreventativeSchedule) error {
	db := GetDB(ctx, r.db)
	entity := FromPreventativeScheduleDomain(ps)
	return db.Save(entity).Error
}

// ==========================================
// TelemetryIngestBuffer Repository
// ==========================================

type SQLTelemetryIngestBufferRepository struct {
	db *gorm.DB
}

func NewSQLTelemetryIngestBufferRepository(db *gorm.DB) domain.TelemetryIngestBufferRepository {
	return &SQLTelemetryIngestBufferRepository{db: db}
}

func (r *SQLTelemetryIngestBufferRepository) Create(ctx context.Context, tb *domain.TelemetryIngestBuffer) error {
	db := GetDB(ctx, r.db)
	entity := FromTelemetryIngestBufferDomain(tb)
	return db.Create(entity).Error
}

func (r *SQLTelemetryIngestBufferRepository) List(ctx context.Context) ([]domain.TelemetryIngestBuffer, error) {
	db := GetDB(ctx, r.db)
	var entities []TelemetryIngestBuffer
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.TelemetryIngestBuffer, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToTelemetryIngestBufferDomain(&e))
	}
	return list, nil
}

func (r *SQLTelemetryIngestBufferRepository) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	db := GetDB(ctx, r.db)
	return db.Delete(&TelemetryIngestBuffer{}, "id IN ?", ids).Error
}

func (r *SQLTelemetryIngestBufferRepository) LockAndList(ctx context.Context, limit int) ([]domain.TelemetryIngestBuffer, error) {
	db := GetDB(ctx, r.db)
	var entities []TelemetryIngestBuffer
	err := db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Limit(limit).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.TelemetryIngestBuffer, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToTelemetryIngestBufferDomain(&e))
	}
	return list, nil
}

// ==========================================
// TransactionalOutbox Repository
// ==========================================

type SQLTransactionalOutboxRepository struct {
	db *gorm.DB
}

func NewSQLTransactionalOutboxRepository(db *gorm.DB) domain.TransactionalOutboxRepository {
	return &SQLTransactionalOutboxRepository{db: db}
}

func (r *SQLTransactionalOutboxRepository) Create(ctx context.Context, outbox *domain.TransactionalOutbox) error {
	db := GetDB(ctx, r.db)
	entity := FromTransactionalOutboxDomain(outbox)
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
	err := db.Where("status = ?", string(domain.OutboxStatusPENDING)).Order("created_at ASC").Limit(limit).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.TransactionalOutbox, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToTransactionalOutboxDomain(&e))
	}
	return list, nil
}

func (r *SQLTransactionalOutboxRepository) Update(ctx context.Context, outbox *domain.TransactionalOutbox) error {
	db := GetDB(ctx, r.db)
	entity := FromTransactionalOutboxDomain(outbox)
	return db.Save(entity).Error
}

// ==========================================
// KafkaEventInbox Repository
// ==========================================

type SQLKafkaEventInboxRepository struct {
	db *gorm.DB
}

func NewSQLKafkaEventInboxRepository(db *gorm.DB) domain.KafkaEventInboxRepository {
	return &SQLKafkaEventInboxRepository{db: db}
}

func (r *SQLKafkaEventInboxRepository) Create(ctx context.Context, inbox *domain.KafkaEventInbox) error {
	db := GetDB(ctx, r.db)
	entity := FromKafkaEventInboxDomain(inbox)
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

func (r *SQLKafkaEventInboxRepository) Update(ctx context.Context, inbox *domain.KafkaEventInbox) error {
	db := GetDB(ctx, r.db)
	entity := FromKafkaEventInboxDomain(inbox)
	return db.Save(entity).Error
}
