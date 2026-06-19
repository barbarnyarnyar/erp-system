package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erp-system/pm-service/internal/api/handlers"
	"github.com/erp-system/pm-service/internal/api/routes"
	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/erp-system/pm-service/internal/data/sql"
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
		&sql.Project{},
		&sql.WbsNode{},
		&sql.TimeLog{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	projRepo := sql.NewSQLProjectRepository(db)
	wbsRepo := sql.NewSQLWbsNodeRepository(db)
	timeRepo := sql.NewSQLTimeLogRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)

	projTrackingSvc := service.NewProjectTrackingService(db, projRepo)
	wbsSvc := service.NewWbsStructureService(db, projRepo, wbsRepo, outboxRepo)
	timeSvc := service.NewTimeTrackingService(db, projRepo, wbsRepo, timeRepo, outboxRepo)

	prjHandler := handlers.NewPrjHandler(projTrackingSvc, wbsSvc, timeSvc, projRepo, wbsRepo, timeRepo)

	router := gin.New()
	routes.SetupPMRoutes(router, prjHandler)

	return &testEnv{
		router: router,
		db:     db,
	}
}

func TestProjectTrackingEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Initialize Project
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"customer_id":     "customer-1",
		"project_code":    "PRJ-001",
		"name":            "ERP Implementation",
		"billing_method":  "TIME_AND_MATERIALS",
		"start_date":      "2026-06-01",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var proj domain.Project
	_ = json.Unmarshal(w.Body.Bytes(), &proj)

	// 2. Get Project
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/projects/"+proj.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 3. List Projects
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 4. Transition Status
	body, _ = json.Marshal(map[string]interface{}{
		"status": "IN_PROGRESS",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/projects/"+proj.ID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestWbsStructureEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Seed Project
	startDate, _ := time.Parse("2006-01-02", "2026-06-01")
	proj := &sql.Project{
		ID:            "proj-123",
		LegalEntityID: "tenant-1",
		CustomerID:    "customer-1",
		ProjectCode:   "PRJ-002",
		Name:          "Seed Project",
		Status:        "DRAFT",
		BillingMethod: "FIXED_PRICE",
		StartDate:     startDate,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_ = env.db.Create(proj).Error

	// 1. Append Wbs Node
	body, _ := json.Marshal(map[string]interface{}{
		"node_code":       "WBS-01",
		"title":           "Requirements Gathering",
		"node_type":       "MILESTONE",
		"estimated_hours": decimal.NewFromInt(40),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/projects/proj-123/wbs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var node domain.WbsNode
	_ = json.Unmarshal(w.Body.Bytes(), &node)

	// 2. Fetch Project Tree
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/projects/proj-123/wbs", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 3. Declare Node Completion
	body, _ = json.Marshal(map[string]interface{}{
		"completion_hr_id": "emp-001",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/wbs/"+node.ID+"/complete", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestTimeTrackingEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Seed Project and WBS Node
	startDate, _ := time.Parse("2006-01-02", "2026-06-01")
	proj := &sql.Project{
		ID:            "proj-123",
		LegalEntityID: "tenant-1",
		CustomerID:    "customer-1",
		ProjectCode:   "PRJ-002",
		Name:          "Seed Project",
		Status:        "IN_PROGRESS",
		BillingMethod: "FIXED_PRICE",
		StartDate:     startDate,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_ = env.db.Create(proj).Error

	node := &sql.WbsNode{
		ID:             "node-123",
		ProjectID:      "proj-123",
		NodeCode:       "WBS-01",
		Title:          "Phase 1",
		NodeType:       "TASK",
		EstimatedHours: decimal.NewFromInt(100),
		IsCompleted:    false,
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_ = env.db.Create(node).Error

	// 1. Log Operational Hours Bulk
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"employee_id":     "emp-001",
		"logs": []map[string]interface{}{
			{
				"wbs_node_id":        "node-123",
				"work_date":          "2026-06-14",
				"hours_spent":        decimal.NewFromInt(8),
				"internal_cost_rate": decimal.NewFromInt(50),
				"billing_rate":       decimal.NewFromInt(80),
				"is_billable":        true,
			},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/time-logs/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Retrieve seeded TimeLog
	var log sql.TimeLog
	_ = env.db.First(&log).Error

	// 2. Process Timesheet Approval
	body, _ = json.Marshal(map[string]interface{}{
		"time_log_ids":    []string{log.ID},
		"approver_hr_id": "approver-01",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/time-logs/approve", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestPrjErrorPaths(t *testing.T) {
	env := setupTestEnv(t)

	// Validation Bad Requests (Missing required fields)
	badRequests := []struct {
		url    string
		method string
		body   string
	}{
		{"/api/v1/projects", http.MethodPost, `{"legal_entity_id":""}`},
		{"/api/v1/projects/proj-123/status", http.MethodPut, `{"status":""}`},
		{"/api/v1/projects/proj-123/wbs", http.MethodPost, `{"node_code":""}`},
		{"/api/v1/wbs/node-123/complete", http.MethodPut, `{"completion_hr_id":""}`},
		{"/api/v1/time-logs/bulk", http.MethodPost, `{"legal_entity_id":""}`},
		{"/api/v1/time-logs/approve", http.MethodPost, `{"time_log_ids":null}`},
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

	// 404 Project or WBS Node
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/projects/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 1. InitializeProject with invalid date
	w = httptest.NewRecorder()
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"customer_id":     "customer-1",
		"project_code":    "PRJ-ERR",
		"name":            "Err Project",
		"billing_method":  "TIME_AND_MATERIALS",
		"start_date":      "bad-date",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad date, got %d", w.Code)
	}

	// 2. InitializeProject with RFC3339 date (to cover RFC3339 branch in parseDate)
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"customer_id":     "customer-1",
		"project_code":    "PRJ-RFC3339",
		"name":            "RFC Project",
		"billing_method":  "TIME_AND_MATERIALS",
		"start_date":      "2026-06-14T08:53:00Z",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 for RFC3339 date, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Database Error path (via canceled context) for ListProjects
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for canceled context on ListProjects, got %d", w.Code)
	}

	// Canceled context for InitializeProject
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"customer_id":     "customer-1",
		"project_code":    "PRJ-CANCEL",
		"name":            "Cancel Project",
		"billing_method":  "TIME_AND_MATERIALS",
		"start_date":      "2026-06-01",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for canceled context on InitializeProject, got %d", w.Code)
	}

	// Canceled context for TransitionProjectStatus
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]interface{}{
		"status": "IN_PROGRESS",
	})
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/projects/proj-123/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for canceled context on TransitionProjectStatus, got %d", w.Code)
	}

	// Canceled context for AppendWbsNode
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]interface{}{
		"node_code":       "WBS-CANCEL",
		"title":           "Cancel Node",
		"node_type":       "TASK",
		"estimated_hours": decimal.NewFromInt(10),
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/projects/proj-123/wbs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for canceled context on AppendWbsNode, got %d", w.Code)
	}

	// Canceled context for DeclareNodeCompletion
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]interface{}{
		"completion_hr_id": "emp-001",
	})
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/wbs/node-123/complete", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for canceled context on DeclareNodeCompletion, got %d", w.Code)
	}

	// Canceled context for FetchProjectTree
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/projects/proj-123/wbs", nil)
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for canceled context on FetchProjectTree, got %d", w.Code)
	}

	// LogOperationalHoursBulk with invalid work_date format
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"employee_id":     "emp-001",
		"logs": []map[string]interface{}{
			{
				"wbs_node_id":        "node-123",
				"work_date":          "invalid-date-format",
				"hours_spent":        decimal.NewFromInt(8),
				"internal_cost_rate": decimal.NewFromInt(50),
				"billing_rate":       decimal.NewFromInt(80),
				"is_billable":        true,
			},
		},
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/time-logs/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid work_date, got %d", w.Code)
	}

	// Canceled context for LogOperationalHoursBulk
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"employee_id":     "emp-001",
		"logs": []map[string]interface{}{
			{
				"wbs_node_id":        "node-123",
				"work_date":          "2026-06-14",
				"hours_spent":        decimal.NewFromInt(8),
				"internal_cost_rate": decimal.NewFromInt(50),
				"billing_rate":       decimal.NewFromInt(80),
				"is_billable":        true,
			},
		},
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/time-logs/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for canceled context on LogOperationalHoursBulk, got %d", w.Code)
	}

	// Canceled context for ProcessTimesheetApproval
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]interface{}{
		"time_log_ids":    []string{"some-id"},
		"approver_hr_id": "approver-01",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/time-logs/approve", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(canceledCtx)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for canceled context on ProcessTimesheetApproval, got %d", w.Code)
	}
}
