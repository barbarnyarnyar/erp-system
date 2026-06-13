package sql

import (
	"context"
	"errors"
	"time"

	"github.com/erp-system/qms-service/internal/business/domain"
	"gorm.io/gorm"
)

// ==========================================
// InspectionPlan Repository
// ==========================================

type SQLInspectionPlanRepository struct {
	db *gorm.DB
}

func NewSQLInspectionPlanRepository(db *gorm.DB) domain.InspectionPlanRepository {
	return &SQLInspectionPlanRepository{db: db}
}

func (r *SQLInspectionPlanRepository) Create(ctx context.Context, ip *domain.InspectionPlan) error {
	db := GetDB(ctx, r.db)
	entity := FromInspectionPlanDomain(ip)
	return db.Create(entity).Error
}

func (r *SQLInspectionPlanRepository) GetByID(ctx context.Context, id string) (*domain.InspectionPlan, error) {
	db := GetDB(ctx, r.db)
	var entity InspectionPlan
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("inspection plan not found")
		}
		return nil, err
	}
	return ToInspectionPlanDomain(&entity), nil
}

func (r *SQLInspectionPlanRepository) GetByMaterial(ctx context.Context, legalEntityId string, materialId string) (*domain.InspectionPlan, error) {
	db := GetDB(ctx, r.db)
	var entity InspectionPlan
	err := db.First(&entity, "legal_entity_id = ? AND material_id = ?", legalEntityId, materialId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("inspection plan not found")
		}
		return nil, err
	}
	return ToInspectionPlanDomain(&entity), nil
}

func (r *SQLInspectionPlanRepository) List(ctx context.Context) ([]domain.InspectionPlan, error) {
	db := GetDB(ctx, r.db)
	var entities []InspectionPlan
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.InspectionPlan, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToInspectionPlanDomain(&e))
	}
	return list, nil
}

func (r *SQLInspectionPlanRepository) Update(ctx context.Context, ip *domain.InspectionPlan) error {
	db := GetDB(ctx, r.db)
	entity := FromInspectionPlanDomain(ip)
	return db.Save(entity).Error
}

// ==========================================
// InspectionMetricDefinition Repository
// ==========================================

type SQLInspectionMetricDefinitionRepository struct {
	db *gorm.DB
}

func NewSQLInspectionMetricDefinitionRepository(db *gorm.DB) domain.InspectionMetricDefinitionRepository {
	return &SQLInspectionMetricDefinitionRepository{db: db}
}

func (r *SQLInspectionMetricDefinitionRepository) Create(ctx context.Context, imd *domain.InspectionMetricDefinition) error {
	db := GetDB(ctx, r.db)
	entity := FromInspectionMetricDefinitionDomain(imd)
	return db.Create(entity).Error
}

func (r *SQLInspectionMetricDefinitionRepository) GetByID(ctx context.Context, id string) (*domain.InspectionMetricDefinition, error) {
	db := GetDB(ctx, r.db)
	var entity InspectionMetricDefinition
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("metric definition not found")
		}
		return nil, err
	}
	return ToInspectionMetricDefinitionDomain(&entity), nil
}

func (r *SQLInspectionMetricDefinitionRepository) ListByPlanID(ctx context.Context, planID string) ([]domain.InspectionMetricDefinition, error) {
	db := GetDB(ctx, r.db)
	var entities []InspectionMetricDefinition
	err := db.Where("inspection_plan_id = ?", planID).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.InspectionMetricDefinition, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToInspectionMetricDefinitionDomain(&e))
	}
	return list, nil
}

// ==========================================
// QualityInspection Repository
// ==========================================

type SQLQualityInspectionRepository struct {
	db *gorm.DB
}

func NewSQLQualityInspectionRepository(db *gorm.DB) domain.QualityInspectionRepository {
	return &SQLQualityInspectionRepository{db: db}
}

func (r *SQLQualityInspectionRepository) Create(ctx context.Context, qi *domain.QualityInspection) error {
	db := GetDB(ctx, r.db)
	entity := FromQualityInspectionDomain(qi)
	return db.Create(entity).Error
}

func (r *SQLQualityInspectionRepository) GetByID(ctx context.Context, id string) (*domain.QualityInspection, error) {
	db := GetDB(ctx, r.db)
	var entity QualityInspection
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("quality inspection not found")
		}
		return nil, err
	}
	return ToQualityInspectionDomain(&entity), nil
}

func (r *SQLQualityInspectionRepository) List(ctx context.Context) ([]domain.QualityInspection, error) {
	db := GetDB(ctx, r.db)
	var entities []QualityInspection
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.QualityInspection, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToQualityInspectionDomain(&e))
	}
	return list, nil
}

func (r *SQLQualityInspectionRepository) Update(ctx context.Context, qi *domain.QualityInspection) error {
	db := GetDB(ctx, r.db)
	entity := FromQualityInspectionDomain(qi)
	return db.Save(entity).Error
}

// ==========================================
// InspectionResultLine Repository
// ==========================================

type SQLInspectionResultLineRepository struct {
	db *gorm.DB
}

func NewSQLInspectionResultLineRepository(db *gorm.DB) domain.InspectionResultLineRepository {
	return &SQLInspectionResultLineRepository{db: db}
}

func (r *SQLInspectionResultLineRepository) Create(ctx context.Context, irl *domain.InspectionResultLine) error {
	db := GetDB(ctx, r.db)
	entity := FromInspectionResultLineDomain(irl)
	return db.Create(entity).Error
}

func (r *SQLInspectionResultLineRepository) ListByInspectionID(ctx context.Context, inspectionID string) ([]domain.InspectionResultLine, error) {
	db := GetDB(ctx, r.db)
	var entities []InspectionResultLine
	err := db.Where("inspection_id = ?", inspectionID).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.InspectionResultLine, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToInspectionResultLineDomain(&e))
	}
	return list, nil
}

func (r *SQLInspectionResultLineRepository) ListByMetricAndDateRange(ctx context.Context, metricDefID string, start, end time.Time) ([]domain.InspectionResultLine, error) {
	db := GetDB(ctx, r.db)
	var entities []InspectionResultLine
	// Forces partition pruning by matching SELECT filter on created_at coordinate range
	err := db.Where("metric_definition_id = ? AND created_at BETWEEN ? AND ?", metricDefID, start, end).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.InspectionResultLine, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToInspectionResultLineDomain(&e))
	}
	return list, nil
}

// ==========================================
// NonConformanceLog Repository
// ==========================================

type SQLNonConformanceLogRepository struct {
	db *gorm.DB
}

func NewSQLNonConformanceLogRepository(db *gorm.DB) domain.NonConformanceLogRepository {
	return &SQLNonConformanceLogRepository{db: db}
}

func (r *SQLNonConformanceLogRepository) Create(ctx context.Context, ncl *domain.NonConformanceLog) error {
	db := GetDB(ctx, r.db)
	entity := FromNonConformanceLogDomain(ncl)
	return db.Create(entity).Error
}

func (r *SQLNonConformanceLogRepository) GetByID(ctx context.Context, id string) (*domain.NonConformanceLog, error) {
	db := GetDB(ctx, r.db)
	var entity NonConformanceLog
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("non-conformance log not found")
		}
		return nil, err
	}
	return ToNonConformanceLogDomain(&entity), nil
}

func (r *SQLNonConformanceLogRepository) List(ctx context.Context) ([]domain.NonConformanceLog, error) {
	db := GetDB(ctx, r.db)
	var entities []NonConformanceLog
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.NonConformanceLog, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToNonConformanceLogDomain(&e))
	}
	return list, nil
}

func (r *SQLNonConformanceLogRepository) Update(ctx context.Context, ncl *domain.NonConformanceLog) error {
	db := GetDB(ctx, r.db)
	entity := FromNonConformanceLogDomain(ncl)
	return db.Save(entity).Error
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

func (r *SQLTransactionalOutboxRepository) UpdateStatus(ctx context.Context, id string, status domain.OutboxStatus) error {
	db := GetDB(ctx, r.db)
	return db.Model(&TransactionalOutbox{}).Where("id = ?", id).Update("status", string(status)).Error
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

func (r *SQLKafkaEventInboxRepository) Exists(ctx context.Context, eventID string) (bool, error) {
	db := GetDB(ctx, r.db)
	var count int64
	err := db.Model(&KafkaEventInbox{}).Where("event_id = ?", eventID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
