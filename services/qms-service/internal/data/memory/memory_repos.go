package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/erp-system/qms-service/internal/business/domain"
)

type MemoryInspectionPlanRepo struct {
	mu   sync.RWMutex
	data map[string]domain.InspectionPlan
}

func NewMemoryInspectionPlanRepo() *MemoryInspectionPlanRepo {
	return &MemoryInspectionPlanRepo{data: make(map[string]domain.InspectionPlan)}
}

func (r *MemoryInspectionPlanRepo) Create(ctx context.Context, ip *domain.InspectionPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ip.ID] = *ip
	return nil
}

func (r *MemoryInspectionPlanRepo) GetByID(ctx context.Context, id string) (*domain.InspectionPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ip, ok := r.data[id]
	if !ok {
		return nil, errors.New("inspection plan not found")
	}
	return &ip, nil
}

func (r *MemoryInspectionPlanRepo) GetByMaterial(ctx context.Context, legalEntityId string, materialId string) (*domain.InspectionPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ip := range r.data {
		if ip.LegalEntityID == legalEntityId && ip.MaterialID == materialId {
			return &ip, nil
		}
	}
	return nil, errors.New("inspection plan for material not found")
}

func (r *MemoryInspectionPlanRepo) List(ctx context.Context) ([]domain.InspectionPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.InspectionPlan, 0, len(r.data))
	for _, ip := range r.data {
		list = append(list, ip)
	}
	return list, nil
}

func (r *MemoryInspectionPlanRepo) Update(ctx context.Context, ip *domain.InspectionPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ip.ID] = *ip
	return nil
}

type MemoryInspectionMetricDefinitionRepo struct {
	mu   sync.RWMutex
	data map[string]domain.InspectionMetricDefinition
}

func NewMemoryInspectionMetricDefinitionRepo() *MemoryInspectionMetricDefinitionRepo {
	return &MemoryInspectionMetricDefinitionRepo{data: make(map[string]domain.InspectionMetricDefinition)}
}

func (r *MemoryInspectionMetricDefinitionRepo) Create(ctx context.Context, imd *domain.InspectionMetricDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[imd.ID] = *imd
	return nil
}

func (r *MemoryInspectionMetricDefinitionRepo) GetByID(ctx context.Context, id string) (*domain.InspectionMetricDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	imd, ok := r.data[id]
	if !ok {
		return nil, errors.New("metric definition not found")
	}
	return &imd, nil
}

func (r *MemoryInspectionMetricDefinitionRepo) ListByPlanID(ctx context.Context, planID string) ([]domain.InspectionMetricDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.InspectionMetricDefinition, 0)
	for _, imd := range r.data {
		if imd.InspectionPlanID == planID {
			list = append(list, imd)
		}
	}
	return list, nil
}

type MemoryQualityInspectionRepo struct {
	mu   sync.RWMutex
	data map[string]domain.QualityInspection
}

func NewMemoryQualityInspectionRepo() *MemoryQualityInspectionRepo {
	return &MemoryQualityInspectionRepo{data: make(map[string]domain.QualityInspection)}
}

func (r *MemoryQualityInspectionRepo) Create(ctx context.Context, qi *domain.QualityInspection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[qi.ID] = *qi
	return nil
}

func (r *MemoryQualityInspectionRepo) GetByID(ctx context.Context, id string) (*domain.QualityInspection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	qi, ok := r.data[id]
	if !ok {
		return nil, errors.New("inspection not found")
	}
	return &qi, nil
}

func (r *MemoryQualityInspectionRepo) List(ctx context.Context) ([]domain.QualityInspection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.QualityInspection, 0, len(r.data))
	for _, qi := range r.data {
		list = append(list, qi)
	}
	return list, nil
}

func (r *MemoryQualityInspectionRepo) Update(ctx context.Context, qi *domain.QualityInspection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[qi.ID] = *qi
	return nil
}

type MemoryInspectionResultLineRepo struct {
	mu   sync.RWMutex
	data map[string]domain.InspectionResultLine
}

func NewMemoryInspectionResultLineRepo() *MemoryInspectionResultLineRepo {
	return &MemoryInspectionResultLineRepo{data: make(map[string]domain.InspectionResultLine)}
}

func (r *MemoryInspectionResultLineRepo) Create(ctx context.Context, irl *domain.InspectionResultLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[irl.ID] = *irl
	return nil
}

func (r *MemoryInspectionResultLineRepo) ListByInspectionID(ctx context.Context, inspectionID string) ([]domain.InspectionResultLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.InspectionResultLine, 0)
	for _, irl := range r.data {
		if irl.InspectionID == inspectionID {
			list = append(list, irl)
		}
	}
	return list, nil
}

func (r *MemoryInspectionResultLineRepo) ListByMetricAndDateRange(ctx context.Context, metricDefID string, start, end time.Time) ([]domain.InspectionResultLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.InspectionResultLine, 0)
	for _, irl := range r.data {
		if irl.MetricDefinitionID == metricDefID && (irl.CreatedAt.After(start) || irl.CreatedAt.Equal(start)) && (irl.CreatedAt.Before(end) || irl.CreatedAt.Equal(end)) {
			list = append(list, irl)
		}
	}
	return list, nil
}

type MemoryNonConformanceLogRepo struct {
	mu   sync.RWMutex
	data map[string]domain.NonConformanceLog
}

func NewMemoryNonConformanceLogRepo() *MemoryNonConformanceLogRepo {
	return &MemoryNonConformanceLogRepo{data: make(map[string]domain.NonConformanceLog)}
}

func (r *MemoryNonConformanceLogRepo) Create(ctx context.Context, ncl *domain.NonConformanceLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ncl.ID] = *ncl
	return nil
}

func (r *MemoryNonConformanceLogRepo) GetByID(ctx context.Context, id string) (*domain.NonConformanceLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ncl, ok := r.data[id]
	if !ok {
		return nil, errors.New("non-conformance log not found")
	}
	return &ncl, nil
}

func (r *MemoryNonConformanceLogRepo) List(ctx context.Context) ([]domain.NonConformanceLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.NonConformanceLog, 0, len(r.data))
	for _, ncl := range r.data {
		list = append(list, ncl)
	}
	return list, nil
}

func (r *MemoryNonConformanceLogRepo) Update(ctx context.Context, ncl *domain.NonConformanceLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ncl.ID] = *ncl
	return nil
}

type MemoryTransactionalOutboxRepo struct {
	mu   sync.RWMutex
	data map[string]domain.TransactionalOutbox
}

func NewMemoryTransactionalOutboxRepo() *MemoryTransactionalOutboxRepo {
	return &MemoryTransactionalOutboxRepo{data: make(map[string]domain.TransactionalOutbox)}
}

func (r *MemoryTransactionalOutboxRepo) Create(ctx context.Context, outbox *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[outbox.ID] = *outbox
	return nil
}

func (r *MemoryTransactionalOutboxRepo) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.TransactionalOutbox, 0)
	for _, o := range r.data {
		if o.Status == domain.OutboxStatusPENDING {
			list = append(list, o)
			if len(list) >= limit {
				break
			}
		}
	}
	return list, nil
}

func (r *MemoryTransactionalOutboxRepo) UpdateStatus(ctx context.Context, id string, status domain.OutboxStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.data[id]; ok {
		o.Status = status
		r.data[id] = o
		return nil
	}
	return errors.New("outbox message not found")
}

type MemoryKafkaEventInboxRepo struct {
	mu   sync.RWMutex
	data map[string]domain.KafkaEventInbox
}

func NewMemoryKafkaEventInboxRepo() *MemoryKafkaEventInboxRepo {
	return &MemoryKafkaEventInboxRepo{data: make(map[string]domain.KafkaEventInbox)}
}

func (r *MemoryKafkaEventInboxRepo) Create(ctx context.Context, inbox *domain.KafkaEventInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[inbox.EventID] = *inbox
	return nil
}

func (r *MemoryKafkaEventInboxRepo) Exists(ctx context.Context, eventID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.data[eventID]
	return ok, nil
}
