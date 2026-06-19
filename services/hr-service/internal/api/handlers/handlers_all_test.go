package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/hr-service/internal/api/handlers"
	"github.com/erp-system/hr-service/internal/api/routes"
	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/erp-system/hr-service/internal/data/sql"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
	utils.InitLogger("hr-service-test")
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
		&sql.Department{},
		&sql.EmployeeMaster{},
		&sql.PayrollRun{},
		&sql.ExpenseClaim{},
		&sql.ExpenseClaimLine{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	deptRepo := sql.NewSQLDepartmentRepository(db)
	empRepo := sql.NewSQLEmployeeMasterRepository(db)
	payrollRepo := sql.NewSQLPayrollRunRepository(db)
	expenseRepo := sql.NewSQLExpenseClaimRepository(db)
	lineRepo := sql.NewSQLExpenseClaimLineRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)

	empService := service.NewEmployeeService(db, empRepo, deptRepo, outboxRepo)
	payrollSvc := service.NewPayrollService(db, payrollRepo, empRepo, outboxRepo)
	expenseSvc := service.NewExpenseService(db, expenseRepo, lineRepo, empRepo, outboxRepo)

	hrHandler := handlers.NewHrHandler(empService, payrollSvc, expenseSvc, deptRepo, empRepo, payrollRepo, expenseRepo, lineRepo)

	router := gin.New()
	routes.RegisterRoutes(router, hrHandler)

	return &testEnv{
		router: router,
		db:     db,
	}
}

func TestDepartmentEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Create Department validation error
	body, _ := json.Marshal(map[string]interface{}{
		"manager_id": "manager-1",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/departments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad dept req, got %d", w.Code)
	}

	// 2. Create Department success
	body, _ = json.Marshal(map[string]interface{}{
		"id":              "dept-1",
		"legal_entity_id": "tenant-1",
		"name":            "Engineering",
		"department_code": "ENG",
		"manager_id":      "manager-1",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/departments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var dept domain.Department
	_ = json.Unmarshal(w.Body.Bytes(), &dept)

	// 3. Get Department
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/departments/"+dept.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 4. Update Department
	body, _ = json.Marshal(map[string]interface{}{
		"name":       "Engineering Updated",
		"dept_code":  "ENG-UPD",
		"manager_id": "manager-2",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/departments/"+dept.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 5. Get list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/departments", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestEmployeeEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Seed department
	dept := &sql.Department{ID: "dept-1", LegalEntityID: "tenant-1", Name: "Sales", DepartmentCode: "SLS"}
	_ = env.db.Create(dept).Error

	// 1. Hire Employee
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"department_id":   "dept-1",
		"employee_number": "EMP-100",
		"first_name":      "John",
		"last_name":       "Doe",
		"email":           "john@doe.com",
		"base_salary":     decimal.NewFromInt(4000),
		"type":            "FULL_TIME",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var emp domain.EmployeeMaster
	_ = json.Unmarshal(w.Body.Bytes(), &emp)

	// 2. Get Employee
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/employees/"+emp.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 3. Update Employee
	body, _ = json.Marshal(map[string]interface{}{
		"first_name": "JohnUpdated",
		"last_name":  "DoeUpdated",
		"status":     "ACTIVE",
		"type":       "FULL_TIME",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/employees/"+emp.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 4. Update Compensation
	body, _ = json.Marshal(map[string]interface{}{
		"target_salary": decimal.NewFromInt(5000),
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/employees/"+emp.ID+"/compensation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 5. Get management chain
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/employees/"+emp.ID+"/management-chain", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 5b. Get Employees list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/employees", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 6. Terminate Employee
	body, _ = json.Marshal(map[string]interface{}{
		"termination_date": time.Now(),
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/employees/"+emp.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestPayrollEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Initiate Payroll
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"fiscal_year":     2026,
		"period_number":   6,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/payroll/initiate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var run domain.PayrollRun
	_ = json.Unmarshal(w.Body.Bytes(), &run)

	// Execute Calculations
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/payroll/calculate/"+run.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Approve Payroll
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/payroll/approve/"+run.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Payroll Runs list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/payroll/runs", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Payroll Run by ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/payroll/runs/"+run.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestExpenseClaimEndpoints(t *testing.T) {
	env := setupTestEnv(t)

	// Seed employee
	emp := &sql.EmployeeMaster{
		ID:            "emp-123",
		LegalEntityID: "tenant-1",
		DepartmentID:   "dept-1",
		FirstName:     "Bob",
		LastName:      "Jones",
		Email:         "bob@jones.com",
		Status:        "ACTIVE",
		Type:          "FULL_TIME",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_ = env.db.Create(emp).Error

	body, _ := json.Marshal(map[string]interface{}{
		"employee_id":     "emp-123",
		"legal_entity_id": "tenant-1",
		"claim_number":    "CLM-1001",
		"purpose":         "Business travel",
		"cost_center_tag": "CC-ENG",
		"lines": []map[string]interface{}{
			{
				"expense_date": time.Now(),
				"category":     "TRAVEL",
				"amount":       decimal.NewFromInt(150),
				"currency":     "USD",
				"description":  "Flights",
			},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var claim domain.ExpenseClaim
	_ = json.Unmarshal(w.Body.Bytes(), &claim)

	// Get lines
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/expenses/"+claim.ID+"/lines", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Approve claim
	body, _ = json.Marshal(map[string]interface{}{
		"reviewer_hr_id": "emp-123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/expenses/"+claim.ID+"/approve", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Pay claim
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/expenses/"+claim.ID+"/pay", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Get Expense Claims list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/expenses", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Get Expense Claim by ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/expenses/"+claim.ID, nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestErrorPaths(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Create Department missing ID
	body, _ := json.Marshal(map[string]interface{}{
		"legal_entity_id": "tenant-1",
		"name":            "Engineering",
		"department_code": "ENG",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/departments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 2. Get Department 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/departments/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 3. Update Department 404
	body, _ = json.Marshal(map[string]interface{}{
		"name": "Engineering Updated",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/departments/non-existent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 4. Update Department bad request (validation error)
	// Seed a department first
	dept := &sql.Department{ID: "dept-1", LegalEntityID: "tenant-1", Name: "Sales", DepartmentCode: "SLS"}
	_ = env.db.Create(dept).Error
	body, _ = json.Marshal(map[string]interface{}{
		"is_active": true, // missing required 'name'
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/departments/dept-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 5. Get Employee 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/employees/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 6. Update Employee 404
	body, _ = json.Marshal(map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
		"status":     "ACTIVE",
		"type":       "FULL_TIME",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/employees/non-existent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 7. Update Employee validation error
	emp := &sql.EmployeeMaster{
		ID:            "emp-1",
		LegalEntityID: "tenant-1",
		DepartmentID:   "dept-1",
		FirstName:     "Bob",
		LastName:      "Jones",
		Email:         "bob@jones.com",
		Status:        "ACTIVE",
		Type:          "FULL_TIME",
	}
	_ = env.db.Create(emp).Error
	body, _ = json.Marshal(map[string]interface{}{
		"first_name": "", // missing first_name binding
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/employees/emp-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 8. Update Compensation validation error
	body, _ = json.Marshal(map[string]interface{}{
		"target_salary": "abc",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/employees/emp-1/compensation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 9. Initiate Payroll Run validation error
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "", // missing required fields
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/payroll/initiate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 10. Get Payroll Run 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/payroll/runs/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 11. Submit Expense Claim validation error
	body, _ = json.Marshal(map[string]interface{}{
		"employee_id": "emp-1",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 12. Verify & Approve Claim validation error (missing reviewer_hr_id)
	body, _ = json.Marshal(map[string]interface{}{})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/expenses/claim-1/approve", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 13. Get Expense Claim 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/expenses/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 14. Hire Employee validation error
	body, _ = json.Marshal(map[string]interface{}{
		"legal_entity_id": "", // missing required fields
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 15. Terminate Employee error (non-existent)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/employees/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 16. Get Management Chain error (non-existent)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/employees/non-existent/management-chain", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// 17. Execute calculations error (non-existent payroll run)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/payroll/calculate/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// 18. Close/Approve error (non-existent payroll run)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/payroll/approve/non-existent", nil)
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
