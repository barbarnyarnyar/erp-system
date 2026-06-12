package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/erp-system/plm-service/internal/business/domain"
)

type MemoryMaterialMasterRepo struct {
	mu   sync.RWMutex
	data map[string]domain.MaterialMaster
}

func NewMemoryMaterialMasterRepo() *MemoryMaterialMasterRepo {
	return &MemoryMaterialMasterRepo{data: make(map[string]domain.MaterialMaster)}
}

func (r *MemoryMaterialMasterRepo) Create(ctx context.Context, m *domain.MaterialMaster) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[m.ID] = *m
	return nil
}

func (r *MemoryMaterialMasterRepo) GetByID(ctx context.Context, id string) (*domain.MaterialMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.data[id]
	if !ok {
		return nil, errors.New("material not found")
	}
	return &m, nil
}

func (r *MemoryMaterialMasterRepo) GetBySKU(ctx context.Context, legalEntityId string, sku string) (*domain.MaterialMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.data {
		if m.LegalEntityID == legalEntityId && m.Sku == sku {
			return &m, nil
		}
	}
	return nil, errors.New("material with sku not found")
}

func (r *MemoryMaterialMasterRepo) List(ctx context.Context) ([]domain.MaterialMaster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.MaterialMaster, 0, len(r.data))
	for _, m := range r.data {
		list = append(list, m)
	}
	return list, nil
}

func (r *MemoryMaterialMasterRepo) Update(ctx context.Context, m *domain.MaterialMaster) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[m.ID] = *m
	return nil
}

type MemoryBomHeaderRepo struct {
	mu   sync.RWMutex
	data map[string]domain.BomHeader
}

func NewMemoryBomHeaderRepo() *MemoryBomHeaderRepo {
	return &MemoryBomHeaderRepo{data: make(map[string]domain.BomHeader)}
}

func (r *MemoryBomHeaderRepo) Create(ctx context.Context, bh *domain.BomHeader) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[bh.ID] = *bh
	return nil
}

func (r *MemoryBomHeaderRepo) GetByID(ctx context.Context, id string) (*domain.BomHeader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bh, ok := r.data[id]
	if !ok {
		return nil, errors.New("bom header not found")
	}
	return &bh, nil
}

func (r *MemoryBomHeaderRepo) List(ctx context.Context) ([]domain.BomHeader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.BomHeader, 0, len(r.data))
	for _, bh := range r.data {
		list = append(list, bh)
	}
	return list, nil
}

func (r *MemoryBomHeaderRepo) Update(ctx context.Context, bh *domain.BomHeader) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[bh.ID] = *bh
	return nil
}

type MemoryBomLineRepo struct {
	mu   sync.RWMutex
	data map[string]domain.BomLine
}

func NewMemoryBomLineRepo() *MemoryBomLineRepo {
	return &MemoryBomLineRepo{data: make(map[string]domain.BomLine)}
}

func (r *MemoryBomLineRepo) Create(ctx context.Context, bl *domain.BomLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[bl.ID] = *bl
	return nil
}

func (r *MemoryBomLineRepo) GetByID(ctx context.Context, id string) (*domain.BomLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bl, ok := r.data[id]
	if !ok {
		return nil, errors.New("bom line not found")
	}
	return &bl, nil
}

func (r *MemoryBomLineRepo) ListByHeaderID(ctx context.Context, headerID string) ([]domain.BomLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.BomLine, 0)
	for _, bl := range r.data {
		if bl.BomHeaderID == headerID {
			list = append(list, bl)
		}
	}
	return list, nil
}

func (r *MemoryBomLineRepo) DeleteByHeaderID(ctx context.Context, headerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, bl := range r.data {
		if bl.BomHeaderID == headerID {
			delete(r.data, id)
		}
	}
	return nil
}

type MemoryEngineeringChangeOrderRepo struct {
	mu   sync.RWMutex
	data map[string]domain.EngineeringChangeOrder
}

func NewMemoryEngineeringChangeOrderRepo() *MemoryEngineeringChangeOrderRepo {
	return &MemoryEngineeringChangeOrderRepo{data: make(map[string]domain.EngineeringChangeOrder)}
}

func (r *MemoryEngineeringChangeOrderRepo) Create(ctx context.Context, eco *domain.EngineeringChangeOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[eco.ID] = *eco
	return nil
}

func (r *MemoryEngineeringChangeOrderRepo) GetByID(ctx context.Context, id string) (*domain.EngineeringChangeOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	eco, ok := r.data[id]
	if !ok {
		return nil, errors.New("eco not found")
	}
	return &eco, nil
}

func (r *MemoryEngineeringChangeOrderRepo) List(ctx context.Context) ([]domain.EngineeringChangeOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.EngineeringChangeOrder, 0, len(r.data))
	for _, eco := range r.data {
		list = append(list, eco)
	}
	return list, nil
}

func (r *MemoryEngineeringChangeOrderRepo) Update(ctx context.Context, eco *domain.EngineeringChangeOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[eco.ID] = *eco
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
