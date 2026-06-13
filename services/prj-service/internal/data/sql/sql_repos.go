package sql

import (
	"context"
	"errors"

	"github.com/erp-system/pm-service/internal/business/domain"
	"gorm.io/gorm"
)

// ==========================================
// Project Repository
// ==========================================

type SQLProjectRepository struct {
	db *gorm.DB
}

func NewSQLProjectRepository(db *gorm.DB) domain.ProjectRepository {
	return &SQLProjectRepository{db: db}
}

func (r *SQLProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	db := GetDB(ctx, r.db)
	entity := FromProjectDomain(project)
	return db.Create(entity).Error
}

func (r *SQLProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	db := GetDB(ctx, r.db)
	var entity Project
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return ToProjectDomain(&entity), nil
}

func (r *SQLProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	db := GetDB(ctx, r.db)
	var entities []Project
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Project, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToProjectDomain(&e))
	}
	return list, nil
}

func (r *SQLProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	db := GetDB(ctx, r.db)
	entity := FromProjectDomain(project)
	var currentVersion = project.Version
	entity.Version = currentVersion + 1
	tx := db.Model(&Project{}).Where("id = ? AND version = ?", project.ID, currentVersion).Select("*").Updates(entity)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("optimistic concurrency lock failure")
	}
	project.Version = entity.Version
	return nil
}

func (r *SQLProjectRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&Project{}, "id = ?", id).Error
}

// ==========================================
// WbsNode Repository
// ==========================================

type SQLWbsNodeRepository struct {
	db *gorm.DB
}

func NewSQLWbsNodeRepository(db *gorm.DB) domain.WbsNodeRepository {
	return &SQLWbsNodeRepository{db: db}
}

func (r *SQLWbsNodeRepository) Create(ctx context.Context, node *domain.WbsNode) error {
	db := GetDB(ctx, r.db)
	entity := FromWbsNodeDomain(node)
	return db.Create(entity).Error
}

func (r *SQLWbsNodeRepository) GetByID(ctx context.Context, id string) (*domain.WbsNode, error) {
	db := GetDB(ctx, r.db)
	var entity WbsNode
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wbs node not found")
		}
		return nil, err
	}
	return ToWbsNodeDomain(&entity), nil
}

func (r *SQLWbsNodeRepository) ListByProjectID(ctx context.Context, projectID string) ([]domain.WbsNode, error) {
	db := GetDB(ctx, r.db)
	var entities []WbsNode
	err := db.Where("project_id = ?", projectID).Order("node_code ASC").Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.WbsNode, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToWbsNodeDomain(&e))
	}
	return list, nil
}

func (r *SQLWbsNodeRepository) Update(ctx context.Context, node *domain.WbsNode) error {
	db := GetDB(ctx, r.db)
	entity := FromWbsNodeDomain(node)
	var currentVersion = node.Version
	entity.Version = currentVersion + 1
	tx := db.Model(&WbsNode{}).Where("id = ? AND version = ?", node.ID, currentVersion).Select("*").Updates(entity)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("optimistic concurrency lock failure")
	}
	node.Version = entity.Version
	return nil
}

func (r *SQLWbsNodeRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&WbsNode{}, "id = ?", id).Error
}

// ==========================================
// TimeLog Repository
// ==========================================

type SQLTimeLogRepository struct {
	db *gorm.DB
}

func NewSQLTimeLogRepository(db *gorm.DB) domain.TimeLogRepository {
	return &SQLTimeLogRepository{db: db}
}

func (r *SQLTimeLogRepository) Create(ctx context.Context, log *domain.TimeLog) error {
	db := GetDB(ctx, r.db)
	entity := FromTimeLogDomain(log)
	return db.Create(entity).Error
}

func (r *SQLTimeLogRepository) GetByID(ctx context.Context, id string) (*domain.TimeLog, error) {
	db := GetDB(ctx, r.db)
	var entity TimeLog
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("time log not found")
		}
		return nil, err
	}
	return ToTimeLogDomain(&entity), nil
}

func (r *SQLTimeLogRepository) List(ctx context.Context) ([]domain.TimeLog, error) {
	db := GetDB(ctx, r.db)
	var entities []TimeLog
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.TimeLog, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToTimeLogDomain(&e))
	}
	return list, nil
}

func (r *SQLTimeLogRepository) Update(ctx context.Context, log *domain.TimeLog) error {
	db := GetDB(ctx, r.db)
	entity := FromTimeLogDomain(log)
	return db.Save(entity).Error
}

func (r *SQLTimeLogRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&TimeLog{}, "id = ?", id).Error
}

func (r *SQLTimeLogRepository) ApproveTimeLogs(ctx context.Context, ids []string, approverHrID string) error {
	db := GetDB(ctx, r.db)
	return db.Model(&TimeLog{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"is_approved":       true,
		"approved_by_hr_id": approverHrID,
	}).Error
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

func (r *SQLTransactionalOutboxRepository) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	db := GetDB(ctx, r.db)
	entity := FromTransactionalOutboxDomain(msg)
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
