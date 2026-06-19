package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erp-system/m-service/internal/api/handlers"
	"github.com/erp-system/m-service/internal/api/routes"
	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/erp-system/m-service/internal/business/service"
	"github.com/erp-system/m-service/internal/data/sql"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type testEnv struct {
	router *gin.Engine
	db     *gorm.DB
}

func setupTestEnv(t *testing.T) *testEnv {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&sql.WorkCenter{},
		&sql.RoutingStation{},
		&sql.WorkOrder{},
		&sql.WorkOrderRoutingState{},
		&sql.MaterialConsumptionLog{},
		&sql.ProductionYieldLog{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	wcRepo := sql.NewSQLWorkCenterRepository(db)
	stationRepo := sql.NewSQLRoutingStationRepository(db)
	woRepo := sql.NewSQLWorkOrderRepository(db)
	stateRepo := sql.NewSQLWorkOrderRoutingStateRepository(db)
	consumeRepo := sql.NewSQLMaterialConsumptionLogRepository(db)
	yieldRepo := sql.NewSQLProductionYieldLogRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)

	floorSvc := service.NewFloorConfigurationService(wcRepo, stationRepo)
	execSvc := service.NewWorkOrderExecutionService(db, woRepo, stateRepo, stationRepo, outboxRepo)
	teleSvc := service.NewShopFloorTelemetryService(db, woRepo, stationRepo, consumeRepo, yieldRepo, outboxRepo)

	mfgHandler := handlers.NewMfgHandler(floorSvc, execSvc, teleSvc)

	router := gin.New()
	routes.RegisterRoutes(router, mfgHandler)

	return &testEnv{
		router: router,
		db:     db,
	}
}

func TestWorkCenterAndStation(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Establish Work Center
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"code":            "WC-001",
		"name":            "Assembly Line 1",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/mfg/work-centers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var wc domain.WorkCenter
	_ = json.Unmarshal(w.Body.Bytes(), &wc)

	// 2. Append Station to Work Center
	body, _ = json.Marshal(map[string]interface{}{
		"routing_code":             "ST-01",
		"station_type":             "MANUAL",
		"standard_setup_time_mins": 10,
		"standard_run_time_mins":   30,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/mfg/work-centers/"+wc.ID+"/stations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestWorkOrderFlow(t *testing.T) {
	env := setupTestEnv(t)

	// Seed WorkCenter and Station
	wc := &sql.WorkCenter{
		ID:             "wc-123",
		LegalEntityID:  "tenant-1",
		WorkCenterCode: "WC-MAIN",
		Name:           "Main Assembly",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_ = env.db.Create(wc).Error

	station := &sql.RoutingStation{
		ID:                    "station-123",
		WorkCenterID:          "wc-123",
		RoutingCode:           "R-01",
		StationType:           "ASSEMBLY",
		StandardSetupTimeMins: 5,
		StandardRunTimeMins:   15,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	_ = env.db.Create(station).Error

	// 1. Instantiate Work Order
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"material_id":     "mat-456",
		"bom_header_id":    "bom-789",
		"quantity_target":  decimal.NewFromInt(100),
		"scheduled_start":  time.Now(),
		"scheduled_end":    time.Now().Add(24 * time.Hour),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/mfg/work-orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var wo domain.WorkOrder
	_ = json.Unmarshal(w.Body.Bytes(), &wo)

	// Seed a second station to reroute to
	station2 := &sql.RoutingStation{
		ID:                    "station-456",
		WorkCenterID:          "wc-123",
		RoutingCode:           "R-02",
		StationType:           "ASSEMBLY",
		StandardSetupTimeMins: 5,
		StandardRunTimeMins:   15,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	_ = env.db.Create(station2).Error

	// Seed a state so that we can transition/reroute/consume/yield
	woState := &sql.WorkOrderRoutingState{
		ID:               "state-123",
		WorkOrderID:      wo.ID,
		CurrentStationID: "station-123",
		EnteredAt:        time.Now(),
	}
	_ = env.db.Create(woState).Error

	// 2. Transition Work Order State
	body, _ = json.Marshal(map[string]interface{}{
		"current_state": "STAGED",
		"target_state":  "IN_PROGRESS",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/mfg/work-orders/"+wo.ID+"/transition", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 3. Reroute Work Order Station
	body, _ = json.Marshal(map[string]interface{}{
		"current_station_id": "station-123",
		"target_station_id":  "station-456",
		"is_rework":          false,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/mfg/work-orders/"+wo.ID+"/reroute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 4. Record Bulk Material Consumption
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"lines": []map[string]interface{}{
			{
				"material_id":        "component-1",
				"quantity_consumed":  decimal.NewFromInt(5),
				"routing_station_id": "station-123",
				"warehouse_id":       "wh-123",
			},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/mfg/work-orders/"+wo.ID+"/consumption", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 5. Commit Production Yield
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"station_id":      "station-123",
		"quantity_good":   decimal.NewFromInt(10),
		"quantity_scrap":  decimal.NewFromInt(1),
		"operator_hr_id":  "operator-01",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/mfg/work-orders/"+wo.ID+"/yield", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestMfgErrorPaths(t *testing.T) {
	env := setupTestEnv(t)

	// Validation errors (missing fields)
	badRequests := []struct {
		url    string
		method string
		body   string
	}{
		{"/api/v1/mfg/work-centers", http.MethodPost, `{"legal_entity_id":""}`},
		{"/api/v1/mfg/work-centers/wc-123/stations", http.MethodPost, `{"routing_code":""}`},
		{"/api/v1/mfg/work-orders", http.MethodPost, `{"legal_entity_id":""}`},
		{"/api/v1/mfg/work-orders/wo-123/transition", http.MethodPost, `{"current_state":""}`},
		{"/api/v1/mfg/work-orders/wo-123/reroute", http.MethodPost, `{"current_station_id":""}`},
		{"/api/v1/mfg/work-orders/wo-123/consumption", http.MethodPost, `{"legal_entity_id":""}`},
		{"/api/v1/mfg/work-orders/wo-123/yield", http.MethodPost, `{"legal_entity_id":""}`},
	}

	for _, item := range badRequests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(item.method, item.url, bytes.NewBufferString(item.body))
		req.Header.Set("Content-Type", "application/json")
		env.router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for %s, got %d. Body: %s", item.url, w.Code, w.Body.String())
		}
	}

	// Service/Database error branch (non-existent work order or station)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/mfg/work-centers/non-existent/stations", bytes.NewBufferString(`{"routing_code":"ST-01","station_type":"MANUAL"}`))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// Mock services for handler error tests
type mockFloorService struct {
	service.FloorConfigurationService
	establishWorkCenterFunc func(ctx context.Context, legalEntityID, code, name string) (*domain.WorkCenter, error)
	appendStationToCenterFunc func(ctx context.Context, workCenterID, routingCode string, stationType domain.StationType, equipmentID *string, setupTime, runTime int) (*domain.RoutingStation, error)
}

func (m *mockFloorService) EstablishWorkCenter(ctx context.Context, legalEntityID, code, name string) (*domain.WorkCenter, error) {
	if m.establishWorkCenterFunc != nil {
		return m.establishWorkCenterFunc(ctx, legalEntityID, code, name)
	}
	return nil, nil
}

func (m *mockFloorService) AppendStationToCenter(ctx context.Context, workCenterID, routingCode string, stationType domain.StationType, equipmentID *string, setupTime, runTime int) (*domain.RoutingStation, error) {
	if m.appendStationToCenterFunc != nil {
		return m.appendStationToCenterFunc(ctx, workCenterID, routingCode, stationType, equipmentID, setupTime, runTime)
	}
	return nil, nil
}

type mockExecService struct {
	service.WorkOrderExecutionService
	instantiateWorkOrderFunc func(ctx context.Context, legalEntityID, materialID, bomHeaderID string, qtyTarget decimal.Decimal, start, end time.Time) (*domain.WorkOrder, error)
	transitionWorkOrderStateFunc func(ctx context.Context, workOrderID string, currentState, targetState domain.WorkOrderState) (*domain.WorkOrder, error)
	rerouteWorkOrderStationFunc func(ctx context.Context, workOrderID, currentStationID, targetStationID string, isRework bool) error
}

func (m *mockExecService) InstantiateWorkOrder(ctx context.Context, legalEntityID, materialID, bomHeaderID string, qtyTarget decimal.Decimal, start, end time.Time) (*domain.WorkOrder, error) {
	if m.instantiateWorkOrderFunc != nil {
		return m.instantiateWorkOrderFunc(ctx, legalEntityID, materialID, bomHeaderID, qtyTarget, start, end)
	}
	return nil, nil
}

func (m *mockExecService) TransitionWorkOrderState(ctx context.Context, workOrderID string, currentState, targetState domain.WorkOrderState) (*domain.WorkOrder, error) {
	if m.transitionWorkOrderStateFunc != nil {
		return m.transitionWorkOrderStateFunc(ctx, workOrderID, currentState, targetState)
	}
	return nil, nil
}

func (m *mockExecService) RerouteWorkOrderStation(ctx context.Context, workOrderID, currentStationID, targetStationID string, isRework bool) error {
	if m.rerouteWorkOrderStationFunc != nil {
		return m.rerouteWorkOrderStationFunc(ctx, workOrderID, currentStationID, targetStationID, isRework)
	}
	return nil
}

type mockTeleService struct {
	service.ShopFloorTelemetryService
	recordBulkMaterialConsumptionFunc func(ctx context.Context, legalEntityID, workOrderID string, lines []domain.ConsumptionSubmissionInput) error
	commitProductionYieldFunc func(ctx context.Context, legalEntityID, workOrderID, stationID string, qtyGood, qtyScrap decimal.Decimal, operatorHrID string) error
}

func (m *mockTeleService) RecordBulkMaterialConsumption(ctx context.Context, legalEntityID, workOrderID string, lines []domain.ConsumptionSubmissionInput) error {
	if m.recordBulkMaterialConsumptionFunc != nil {
		return m.recordBulkMaterialConsumptionFunc(ctx, legalEntityID, workOrderID, lines)
	}
	return nil
}

func (m *mockTeleService) CommitProductionYield(ctx context.Context, legalEntityID, workOrderID, stationID string, qtyGood, qtyScrap decimal.Decimal, operatorHrID string) error {
	if m.commitProductionYieldFunc != nil {
		return m.commitProductionYieldFunc(ctx, legalEntityID, workOrderID, stationID, qtyGood, qtyScrap, operatorHrID)
	}
	return nil
}

func TestMfgHandlerServiceErrors(t *testing.T) {
	floorSvc := &mockFloorService{
		establishWorkCenterFunc: func(ctx context.Context, legalEntityID, code, name string) (*domain.WorkCenter, error) {
			return nil, errors.New("establish error")
		},
		appendStationToCenterFunc: func(ctx context.Context, workCenterID, routingCode string, stationType domain.StationType, equipmentID *string, setupTime, runTime int) (*domain.RoutingStation, error) {
			return nil, errors.New("append error")
		},
	}

	execSvc := &mockExecService{
		instantiateWorkOrderFunc: func(ctx context.Context, legalEntityID, materialID, bomHeaderID string, qtyTarget decimal.Decimal, start, end time.Time) (*domain.WorkOrder, error) {
			return nil, errors.New("instantiate error")
		},
		transitionWorkOrderStateFunc: func(ctx context.Context, workOrderID string, currentState, targetState domain.WorkOrderState) (*domain.WorkOrder, error) {
			return nil, errors.New("transition error")
		},
		rerouteWorkOrderStationFunc: func(ctx context.Context, workOrderID, currentStationID, targetStationID string, isRework bool) error {
			return errors.New("reroute error")
		},
	}

	teleSvc := &mockTeleService{
		recordBulkMaterialConsumptionFunc: func(ctx context.Context, legalEntityID, workOrderID string, lines []domain.ConsumptionSubmissionInput) error {
			return errors.New("consumption error")
		},
		commitProductionYieldFunc: func(ctx context.Context, legalEntityID, workOrderID, stationID string, qtyGood, qtyScrap decimal.Decimal, operatorHrID string) error {
			return errors.New("yield error")
		},
	}

	mfgHandler := handlers.NewMfgHandler(floorSvc, execSvc, teleSvc)
	router := gin.New()
	routes.RegisterRoutes(router, mfgHandler)

	tests := []struct {
		url    string
		method string
		body   string
	}{
		{"/api/v1/mfg/work-centers", http.MethodPost, `{"legal_entity_id":"t-1","code":"c-1","name":"n-1"}`},
		{"/api/v1/mfg/work-centers/wc-123/stations", http.MethodPost, `{"routing_code":"rc-1","station_type":"ASSEMBLY"}`},
		{"/api/v1/mfg/work-orders", http.MethodPost, `{"legal_entity_id":"t-1","material_id":"m-1","bom_header_id":"b-1","quantity_target":"10","scheduled_start":"2026-06-19T00:00:00Z","scheduled_end":"2026-06-20T00:00:00Z"}`},
		{"/api/v1/mfg/work-orders/wo-123/transition", http.MethodPost, `{"current_state":"STAGED","target_state":"RELEASED"}`},
		{"/api/v1/mfg/work-orders/wo-123/reroute", http.MethodPost, `{"current_station_id":"s-1","target_station_id":"s-2"}`},
		{"/api/v1/mfg/work-orders/wo-123/consumption", http.MethodPost, `{"legal_entity_id":"t-1","lines":[{"material_id":"m-1","quantity_consumed":"5","routing_station_id":"s-1","warehouse_id":"w-1"}]}`},
		{"/api/v1/mfg/work-orders/wo-123/yield", http.MethodPost, `{"legal_entity_id":"t-1","station_id":"s-1","quantity_good":"10","operator_hr_id":"op-1"}`},
	}

	for _, item := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(item.method, item.url, bytes.NewBufferString(item.body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500 for %s, got %d. Body: %s", item.url, w.Code, w.Body.String())
		}
	}
}

