package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/erp-system/pm-service/internal/business/domain"
)

// ==========================================
// Project Memory Repository
// ==========================================

type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]domain.Project
}

func NewProjectRepository() domain.ProjectRepository {
	return &ProjectRepository{
		projects: make(map[string]domain.Project),
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.projects[project.ID] = *project
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.projects[id]
	if !ok {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	return &p, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Project, 0, len(r.projects))
	for _, p := range r.projects {
		list = append(list, p)
	}
	return list, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.projects[project.ID]
	if !ok {
		return fmt.Errorf("project not found: %s", project.ID)
	}
	if existing.Version != project.Version {
		return errors.New("optimistic concurrency lock failure")
	}
	project.Version++
	r.projects[project.ID] = *project
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.projects, id)
	return nil
}

// ==========================================
// WbsNode Memory Repository
// ==========================================

type WbsNodeRepository struct {
	mu    sync.RWMutex
	nodes map[string]domain.WbsNode
}

func NewWbsNodeRepository() domain.WbsNodeRepository {
	return &WbsNodeRepository{
		nodes: make(map[string]domain.WbsNode),
	}
}

func (r *WbsNodeRepository) Create(ctx context.Context, node *domain.WbsNode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[node.ID] = *node
	return nil
}

func (r *WbsNodeRepository) GetByID(ctx context.Context, id string) (*domain.WbsNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	node, ok := r.nodes[id]
	if !ok {
		return nil, fmt.Errorf("wbs node not found: %s", id)
	}
	return &node, nil
}

func (r *WbsNodeRepository) ListByProjectID(ctx context.Context, projectID string) ([]domain.WbsNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.WbsNode
	for _, node := range r.nodes {
		if node.ProjectID == projectID {
			list = append(list, node)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].NodeCode < list[j].NodeCode
	})
	return list, nil
}

func (r *WbsNodeRepository) Update(ctx context.Context, node *domain.WbsNode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.nodes[node.ID]
	if !ok {
		return fmt.Errorf("wbs node not found: %s", node.ID)
	}
	if existing.Version != node.Version {
		return errors.New("optimistic concurrency lock failure")
	}
	node.Version++
	r.nodes[node.ID] = *node
	return nil
}

func (r *WbsNodeRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, id)
	return nil
}

// ==========================================
// TimeLog Memory Repository
// ==========================================

type TimeLogRepository struct {
	mu   sync.RWMutex
	logs map[string]domain.TimeLog
}

func NewTimeLogRepository() domain.TimeLogRepository {
	return &TimeLogRepository{
		logs: make(map[string]domain.TimeLog),
	}
}

func (r *TimeLogRepository) Create(ctx context.Context, log *domain.TimeLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = *log
	return nil
}

func (r *TimeLogRepository) GetByID(ctx context.Context, id string) (*domain.TimeLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	log, ok := r.logs[id]
	if !ok {
		return nil, fmt.Errorf("time log not found: %s", id)
	}
	return &log, nil
}

func (r *TimeLogRepository) List(ctx context.Context) ([]domain.TimeLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.TimeLog, 0, len(r.logs))
	for _, l := range r.logs {
		list = append(list, l)
	}
	return list, nil
}

func (r *TimeLogRepository) Update(ctx context.Context, log *domain.TimeLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.logs[log.ID]; !ok {
		return fmt.Errorf("time log not found: %s", log.ID)
	}
	r.logs[log.ID] = *log
	return nil
}

func (r *TimeLogRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.logs, id)
	return nil
}

func (r *TimeLogRepository) ApproveTimeLogs(ctx context.Context, ids []string, approverHrID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range ids {
		if log, ok := r.logs[id]; ok {
			log.IsApproved = true
			log.ApprovedByHrID = &approverHrID
			r.logs[id] = log
		}
	}
	return nil
}

// ==========================================
// TransactionalOutbox Memory Repository
// ==========================================

type TransactionalOutboxRepository struct {
	mu   sync.RWMutex
	msgs map[string]domain.TransactionalOutbox
}

func NewTransactionalOutboxRepository() domain.TransactionalOutboxRepository {
	return &TransactionalOutboxRepository{
		msgs: make(map[string]domain.TransactionalOutbox),
	}
}

func (r *TransactionalOutboxRepository) Create(ctx context.Context, msg *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs[msg.ID] = *msg
	return nil
}

func (r *TransactionalOutboxRepository) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.msgs[id]
	if !ok {
		return nil, fmt.Errorf("outbox message not found: %s", id)
	}
	return &msg, nil
}

func (r *TransactionalOutboxRepository) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.TransactionalOutbox
	for _, msg := range r.msgs {
		if msg.Status == domain.OutboxStatusPENDING {
			list = append(list, msg)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (r *TransactionalOutboxRepository) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.msgs[msg.ID]; !ok {
		return fmt.Errorf("outbox message not found: %s", msg.ID)
	}
	r.msgs[msg.ID] = *msg
	return nil
}

// ==========================================
// KafkaEventInbox Memory Repository
// ==========================================

type KafkaEventInboxRepository struct {
	mu   sync.RWMutex
	msgs map[string]domain.KafkaEventInbox
}

func NewKafkaEventInboxRepository() domain.KafkaEventInboxRepository {
	return &KafkaEventInboxRepository{
		msgs: make(map[string]domain.KafkaEventInbox),
	}
}

func (r *KafkaEventInboxRepository) Create(ctx context.Context, msg *domain.KafkaEventInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs[msg.EventID] = *msg
	return nil
}

func (r *KafkaEventInboxRepository) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.msgs[eventID]
	if !ok {
		return nil, fmt.Errorf("inbox message not found: %s", eventID)
	}
	return &msg, nil
}

func (r *KafkaEventInboxRepository) Update(ctx context.Context, msg *domain.KafkaEventInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.msgs[msg.EventID]; !ok {
		return fmt.Errorf("inbox message not found: %s", msg.EventID)
	}
	r.msgs[msg.EventID] = *msg
	return nil
}
