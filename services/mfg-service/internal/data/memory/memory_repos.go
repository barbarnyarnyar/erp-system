package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/erp-system/m-service/internal/business/domain"
)

// MemoryWorkCenterRepo implements domain.WorkCenterRepository
type MemoryWorkCenterRepo struct {
	mu  sync.RWMutex
	wcs map[string]domain.WorkCenter
}

func NewMemoryWorkCenterRepo() *MemoryWorkCenterRepo {
	return &MemoryWorkCenterRepo{wcs: make(map[string]domain.WorkCenter)}
}

func (r *MemoryWorkCenterRepo) Create(ctx context.Context, wc *domain.WorkCenter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wcs[wc.ID] = *wc
	return nil
}

func (r *MemoryWorkCenterRepo) GetByID(ctx context.Context, id string) (*domain.WorkCenter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wc, ok := r.wcs[id]
	if !ok {
		return nil, errors.New("work center not found")
	}
	return &wc, nil
}

func (r *MemoryWorkCenterRepo) GetByCode(ctx context.Context, legalEntityID, code string) (*domain.WorkCenter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, wc := range r.wcs {
		if wc.LegalEntityID == legalEntityID && wc.WorkCenterCode == code {
			return &wc, nil
		}
	}
	return nil, errors.New("work center not found by code")
}

func (r *MemoryWorkCenterRepo) List(ctx context.Context) ([]domain.WorkCenter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.WorkCenter, 0, len(r.wcs))
	for _, w := range r.wcs {
		list = append(list, w)
	}
	return list, nil
}

func (r *MemoryWorkCenterRepo) Update(ctx context.Context, wc *domain.WorkCenter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wcs[wc.ID] = *wc
	return nil
}

func (r *MemoryWorkCenterRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.wcs, id)
	return nil
}

// MemoryRoutingStationRepo implements domain.RoutingStationRepository
type MemoryRoutingStationRepo struct {
	mu       sync.RWMutex
	stations map[string]domain.RoutingStation
}

func NewMemoryRoutingStationRepo() *MemoryRoutingStationRepo {
	return &MemoryRoutingStationRepo{stations: make(map[string]domain.RoutingStation)}
}

func (r *MemoryRoutingStationRepo) Create(ctx context.Context, station *domain.RoutingStation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stations[station.ID] = *station
	return nil
}

func (r *MemoryRoutingStationRepo) GetByID(ctx context.Context, id string) (*domain.RoutingStation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	station, ok := r.stations[id]
	if !ok {
		return nil, errors.New("routing station not found")
	}
	return &station, nil
}

func (r *MemoryRoutingStationRepo) GetByCode(ctx context.Context, workCenterID, code string) (*domain.RoutingStation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, station := range r.stations {
		if station.WorkCenterID == workCenterID && station.RoutingCode == code {
			return &station, nil
		}
	}
	return nil, errors.New("routing station not found by code")
}

func (r *MemoryRoutingStationRepo) ListByWorkCenterID(ctx context.Context, workCenterID string) ([]domain.RoutingStation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.RoutingStation
	for _, station := range r.stations {
		if station.WorkCenterID == workCenterID {
			list = append(list, station)
		}
	}
	return list, nil
}

func (r *MemoryRoutingStationRepo) Update(ctx context.Context, station *domain.RoutingStation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stations[station.ID] = *station
	return nil
}

func (r *MemoryRoutingStationRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.stations, id)
	return nil
}

// MemoryWorkOrderRepo implements domain.WorkOrderRepository
type MemoryWorkOrderRepo struct {
	mu  sync.RWMutex
	wos map[string]domain.WorkOrder
}

func NewMemoryWorkOrderRepo() *MemoryWorkOrderRepo {
	return &MemoryWorkOrderRepo{wos: make(map[string]domain.WorkOrder)}
}

func (r *MemoryWorkOrderRepo) Create(ctx context.Context, wo *domain.WorkOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wos[wo.ID] = *wo
	return nil
}

func (r *MemoryWorkOrderRepo) GetByID(ctx context.Context, id string) (*domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wo, ok := r.wos[id]
	if !ok {
		return nil, errors.New("work order not found")
	}
	return &wo, nil
}

func (r *MemoryWorkOrderRepo) GetByNumber(ctx context.Context, legalEntityID, number string) (*domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, wo := range r.wos {
		if wo.LegalEntityID == legalEntityID && wo.WorkOrderNumber == number {
			return &wo, nil
		}
	}
	return nil, errors.New("work order not found by number")
}

func (r *MemoryWorkOrderRepo) List(ctx context.Context) ([]domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.WorkOrder, 0, len(r.wos))
	for _, w := range r.wos {
		list = append(list, w)
	}
	return list, nil
}

func (r *MemoryWorkOrderRepo) Update(ctx context.Context, wo *domain.WorkOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.wos[wo.ID]
	if !ok {
		return errors.New("work order not found for update")
	}
	if current.Version != wo.Version {
		return errors.New("concurrent modification error: work order version mismatch")
	}
	wo.Version++
	r.wos[wo.ID] = *wo
	return nil
}

func (r *MemoryWorkOrderRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.wos, id)
	return nil
}

// MemoryWorkOrderRoutingStateRepo implements domain.WorkOrderRoutingStateRepository
type MemoryWorkOrderRoutingStateRepo struct {
	mu     sync.RWMutex
	states map[string]domain.WorkOrderRoutingState
}

func NewMemoryWorkOrderRoutingStateRepo() *MemoryWorkOrderRoutingStateRepo {
	return &MemoryWorkOrderRoutingStateRepo{states: make(map[string]domain.WorkOrderRoutingState)}
}

func (r *MemoryWorkOrderRoutingStateRepo) Create(ctx context.Context, state *domain.WorkOrderRoutingState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[state.ID] = *state
	return nil
}

func (r *MemoryWorkOrderRoutingStateRepo) GetByID(ctx context.Context, id string) (*domain.WorkOrderRoutingState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	state, ok := r.states[id]
	if !ok {
		return nil, errors.New("routing state not found")
	}
	return &state, nil
}

func (r *MemoryWorkOrderRoutingStateRepo) GetActiveByWorkOrderID(ctx context.Context, workOrderID string) (*domain.WorkOrderRoutingState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.states {
		if s.WorkOrderID == workOrderID && s.ExitedAt == nil {
			return &s, nil
		}
	}
	return nil, errors.New("active routing state not found")
}

func (r *MemoryWorkOrderRoutingStateRepo) ListByWorkOrderID(ctx context.Context, workOrderID string) ([]domain.WorkOrderRoutingState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.WorkOrderRoutingState
	for _, s := range r.states {
		if s.WorkOrderID == workOrderID {
			list = append(list, s)
		}
	}
	return list, nil
}

func (r *MemoryWorkOrderRoutingStateRepo) Update(ctx context.Context, state *domain.WorkOrderRoutingState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[state.ID] = *state
	return nil
}

// MemoryMaterialConsumptionLogRepo implements domain.MaterialConsumptionLogRepository
type MemoryMaterialConsumptionLogRepo struct {
	mu   sync.RWMutex
	logs map[string]domain.MaterialConsumptionLog
}

func NewMemoryMaterialConsumptionLogRepo() *MemoryMaterialConsumptionLogRepo {
	return &MemoryMaterialConsumptionLogRepo{logs: make(map[string]domain.MaterialConsumptionLog)}
}

func (r *MemoryMaterialConsumptionLogRepo) Create(ctx context.Context, log *domain.MaterialConsumptionLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = *log
	return nil
}

func (r *MemoryMaterialConsumptionLogRepo) GetByID(ctx context.Context, id string) (*domain.MaterialConsumptionLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	log, ok := r.logs[id]
	if !ok {
		return nil, errors.New("consumption log not found")
	}
	return &log, nil
}

func (r *MemoryMaterialConsumptionLogRepo) ListByWorkOrderID(ctx context.Context, workOrderID string) ([]domain.MaterialConsumptionLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.MaterialConsumptionLog
	for _, log := range r.logs {
		if log.WorkOrderID == workOrderID {
			list = append(list, log)
		}
	}
	return list, nil
}

// MemoryProductionYieldLogRepo implements domain.ProductionYieldLogRepository
type MemoryProductionYieldLogRepo struct {
	mu   sync.RWMutex
	logs map[string]domain.ProductionYieldLog
}

func NewMemoryProductionYieldLogRepo() *MemoryProductionYieldLogRepo {
	return &MemoryProductionYieldLogRepo{logs: make(map[string]domain.ProductionYieldLog)}
}

func (r *MemoryProductionYieldLogRepo) Create(ctx context.Context, log *domain.ProductionYieldLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[log.ID] = *log
	return nil
}

func (r *MemoryProductionYieldLogRepo) GetByID(ctx context.Context, id string) (*domain.ProductionYieldLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	log, ok := r.logs[id]
	if !ok {
		return nil, errors.New("yield log not found")
	}
	return &log, nil
}

func (r *MemoryProductionYieldLogRepo) ListByWorkOrderID(ctx context.Context, workOrderID string) ([]domain.ProductionYieldLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ProductionYieldLog
	for _, log := range r.logs {
		if log.WorkOrderID == workOrderID {
			list = append(list, log)
		}
	}
	return list, nil
}

// MemoryTransactionalOutboxRepo implements domain.TransactionalOutboxRepository
type MemoryTransactionalOutboxRepo struct {
	mu   sync.RWMutex
	msgs map[string]domain.TransactionalOutbox
}

func NewMemoryTransactionalOutboxRepo() *MemoryTransactionalOutboxRepo {
	return &MemoryTransactionalOutboxRepo{msgs: make(map[string]domain.TransactionalOutbox)}
}

func (r *MemoryTransactionalOutboxRepo) Create(ctx context.Context, msg *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs[msg.ID] = *msg
	return nil
}

func (r *MemoryTransactionalOutboxRepo) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.msgs[id]
	if !ok {
		return nil, errors.New("outbox message not found")
	}
	return &msg, nil
}

func (r *MemoryTransactionalOutboxRepo) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.TransactionalOutbox
	for _, msg := range r.msgs {
		if msg.Status == domain.OutboxStatusPENDING {
			list = append(list, msg)
		}
		if len(list) >= limit {
			break
		}
	}
	return list, nil
}

func (r *MemoryTransactionalOutboxRepo) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs[msg.ID] = *msg
	return nil
}

// MemoryKafkaEventInboxRepo implements domain.KafkaEventInboxRepository
type MemoryKafkaEventInboxRepo struct {
	mu   sync.RWMutex
	msgs map[string]domain.KafkaEventInbox
}

func NewMemoryKafkaEventInboxRepo() *MemoryKafkaEventInboxRepo {
	return &MemoryKafkaEventInboxRepo{msgs: make(map[string]domain.KafkaEventInbox)}
}

func (r *MemoryKafkaEventInboxRepo) Create(ctx context.Context, msg *domain.KafkaEventInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs[msg.EventID] = *msg
	return nil
}

func (r *MemoryKafkaEventInboxRepo) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.msgs[eventID]
	if !ok {
		return nil, errors.New("inbox message not found")
	}
	return &msg, nil
}

func (r *MemoryKafkaEventInboxRepo) Update(ctx context.Context, msg *domain.KafkaEventInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs[msg.EventID] = *msg
	return nil
}
