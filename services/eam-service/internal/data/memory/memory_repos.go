package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/erp-system/eam-service/internal/business/domain"
)

type MemoryFacilityRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Facility
}

func NewMemoryFacilityRepo() *MemoryFacilityRepo {
	return &MemoryFacilityRepo{data: make(map[string]domain.Facility)}
}

func (r *MemoryFacilityRepo) Create(ctx context.Context, f *domain.Facility) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[f.ID] = *f
	return nil
}

func (r *MemoryFacilityRepo) GetByID(ctx context.Context, id string) (*domain.Facility, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.data[id]
	if !ok {
		return nil, errors.New("facility not found")
	}
	return &f, nil
}

func (r *MemoryFacilityRepo) List(ctx context.Context) ([]domain.Facility, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Facility, 0, len(r.data))
	for _, f := range r.data {
		list = append(list, f)
	}
	return list, nil
}

func (r *MemoryFacilityRepo) Update(ctx context.Context, f *domain.Facility) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[f.ID] = *f
	return nil
}

type MemoryEquipmentRepo struct {
	mu   sync.RWMutex
	data map[string]domain.Equipment
}

func NewMemoryEquipmentRepo() *MemoryEquipmentRepo {
	return &MemoryEquipmentRepo{data: make(map[string]domain.Equipment)}
}

func (r *MemoryEquipmentRepo) Create(ctx context.Context, eq *domain.Equipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[eq.ID] = *eq
	return nil
}

func (r *MemoryEquipmentRepo) GetByID(ctx context.Context, id string) (*domain.Equipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	eq, ok := r.data[id]
	if !ok || eq.DeletedAt != nil {
		return nil, errors.New("equipment not found")
	}
	return &eq, nil
}

func (r *MemoryEquipmentRepo) List(ctx context.Context) ([]domain.Equipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Equipment, 0, len(r.data))
	for _, eq := range r.data {
		if eq.DeletedAt == nil {
			list = append(list, eq)
		}
	}
	return list, nil
}

func (r *MemoryEquipmentRepo) Update(ctx context.Context, eq *domain.Equipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[eq.ID] = *eq
	return nil
}

func (r *MemoryEquipmentRepo) ListByTenant(ctx context.Context, legalEntityId string) ([]domain.Equipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Equipment, 0)
	for _, eq := range r.data {
		if eq.LegalEntityID == legalEntityId && eq.DeletedAt == nil {
			list = append(list, eq)
		}
	}
	return list, nil
}

func (r *MemoryEquipmentRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if eq, ok := r.data[id]; ok {
		now := time.Now()
		eq.DeletedAt = &now
		r.data[id] = eq
		return nil
	}
	return errors.New("equipment not found")
}

type MemoryMaintenanceWorkOrderRepo struct {
	mu   sync.RWMutex
	data map[string]domain.MaintenanceWorkOrder
}

func NewMemoryMaintenanceWorkOrderRepo() *MemoryMaintenanceWorkOrderRepo {
	return &MemoryMaintenanceWorkOrderRepo{data: make(map[string]domain.MaintenanceWorkOrder)}
}

func (r *MemoryMaintenanceWorkOrderRepo) Create(ctx context.Context, wo *domain.MaintenanceWorkOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[wo.ID] = *wo
	return nil
}

func (r *MemoryMaintenanceWorkOrderRepo) GetByID(ctx context.Context, id string) (*domain.MaintenanceWorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wo, ok := r.data[id]
	if !ok {
		return nil, errors.New("work order not found")
	}
	return &wo, nil
}

func (r *MemoryMaintenanceWorkOrderRepo) List(ctx context.Context) ([]domain.MaintenanceWorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.MaintenanceWorkOrder, 0, len(r.data))
	for _, wo := range r.data {
		list = append(list, wo)
	}
	return list, nil
}

func (r *MemoryMaintenanceWorkOrderRepo) Update(ctx context.Context, wo *domain.MaintenanceWorkOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[wo.ID] = *wo
	return nil
}

type MemoryPreventativeScheduleRepo struct {
	mu   sync.RWMutex
	data map[string]domain.PreventativeSchedule
}

func NewMemoryPreventativeScheduleRepo() *MemoryPreventativeScheduleRepo {
	return &MemoryPreventativeScheduleRepo{data: make(map[string]domain.PreventativeSchedule)}
}

func (r *MemoryPreventativeScheduleRepo) Create(ctx context.Context, ps *domain.PreventativeSchedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ps.ID] = *ps
	return nil
}

func (r *MemoryPreventativeScheduleRepo) GetByID(ctx context.Context, id string) (*domain.PreventativeSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ps, ok := r.data[id]
	if !ok {
		return nil, errors.New("schedule not found")
	}
	return &ps, nil
}

func (r *MemoryPreventativeScheduleRepo) List(ctx context.Context) ([]domain.PreventativeSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.PreventativeSchedule, 0, len(r.data))
	for _, ps := range r.data {
		list = append(list, ps)
	}
	return list, nil
}

func (r *MemoryPreventativeScheduleRepo) Update(ctx context.Context, ps *domain.PreventativeSchedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[ps.ID] = *ps
	return nil
}

type MemoryTelemetryIngestBufferRepo struct {
	mu   sync.RWMutex
	data map[string]domain.TelemetryIngestBuffer
}

func NewMemoryTelemetryIngestBufferRepo() *MemoryTelemetryIngestBufferRepo {
	return &MemoryTelemetryIngestBufferRepo{data: make(map[string]domain.TelemetryIngestBuffer)}
}

func (r *MemoryTelemetryIngestBufferRepo) Create(ctx context.Context, tb *domain.TelemetryIngestBuffer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[tb.ID] = *tb
	return nil
}

func (r *MemoryTelemetryIngestBufferRepo) List(ctx context.Context) ([]domain.TelemetryIngestBuffer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.TelemetryIngestBuffer, 0, len(r.data))
	for _, tb := range r.data {
		list = append(list, tb)
	}
	return list, nil
}

func (r *MemoryTelemetryIngestBufferRepo) DeleteBatch(ctx context.Context, ids []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range ids {
		delete(r.data, id)
	}
	return nil
}

func (r *MemoryTelemetryIngestBufferRepo) LockAndList(ctx context.Context, limit int) ([]domain.TelemetryIngestBuffer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	list := make([]domain.TelemetryIngestBuffer, 0)
	for _, tb := range r.data {
		list = append(list, tb)
		if len(list) >= limit {
			break
		}
	}
	return list, nil
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

func (r *MemoryTransactionalOutboxRepo) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.data[id]
	if !ok {
		return nil, errors.New("outbox message not found")
	}
	return &o, nil
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

func (r *MemoryTransactionalOutboxRepo) Update(ctx context.Context, outbox *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[outbox.ID] = *outbox
	return nil
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

func (r *MemoryKafkaEventInboxRepo) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, ok := r.data[eventID]
	if !ok {
		return nil, errors.New("inbox message not found")
	}
	return &i, nil
}

func (r *MemoryKafkaEventInboxRepo) Update(ctx context.Context, inbox *domain.KafkaEventInbox) error {
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
