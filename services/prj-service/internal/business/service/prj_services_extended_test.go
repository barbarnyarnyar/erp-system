package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/erp-system/pm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

// ============================================================================
// Mocks wrapping Memory Repositories
// ============================================================================

type MockProjectRepository struct {
	domain.ProjectRepository
	CreateFunc  func(ctx context.Context, project *domain.Project) error
	GetByIDFunc func(ctx context.Context, id string) (*domain.Project, error)
	ListFunc    func(ctx context.Context) ([]domain.Project, error)
	UpdateFunc  func(ctx context.Context, project *domain.Project) error
	DeleteFunc  func(ctx context.Context, id string) error
}

func (m *MockProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, project)
	}
	return m.ProjectRepository.Create(ctx, project)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return m.ProjectRepository.GetByID(ctx, id)
}

func (m *MockProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return m.ProjectRepository.List(ctx)
}

func (m *MockProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, project)
	}
	return m.ProjectRepository.Update(ctx, project)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return m.ProjectRepository.Delete(ctx, id)
}

type MockWbsNodeRepository struct {
	domain.WbsNodeRepository
	CreateFunc          func(ctx context.Context, node *domain.WbsNode) error
	GetByIDFunc         func(ctx context.Context, id string) (*domain.WbsNode, error)
	ListByProjectIDFunc func(ctx context.Context, projectID string) ([]domain.WbsNode, error)
	UpdateFunc          func(ctx context.Context, node *domain.WbsNode) error
	DeleteFunc          func(ctx context.Context, id string) error
}

func (m *MockWbsNodeRepository) Create(ctx context.Context, node *domain.WbsNode) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, node)
	}
	return m.WbsNodeRepository.Create(ctx, node)
}

func (m *MockWbsNodeRepository) GetByID(ctx context.Context, id string) (*domain.WbsNode, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return m.WbsNodeRepository.GetByID(ctx, id)
}

func (m *MockWbsNodeRepository) ListByProjectID(ctx context.Context, projectID string) ([]domain.WbsNode, error) {
	if m.ListByProjectIDFunc != nil {
		return m.ListByProjectIDFunc(ctx, projectID)
	}
	return m.WbsNodeRepository.ListByProjectID(ctx, projectID)
}

func (m *MockWbsNodeRepository) Update(ctx context.Context, node *domain.WbsNode) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, node)
	}
	return m.WbsNodeRepository.Update(ctx, node)
}

func (m *MockWbsNodeRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return m.WbsNodeRepository.Delete(ctx, id)
}

type MockTimeLogRepository struct {
	domain.TimeLogRepository
	CreateFunc          func(ctx context.Context, log *domain.TimeLog) error
	GetByIDFunc         func(ctx context.Context, id string) (*domain.TimeLog, error)
	ListFunc            func(ctx context.Context) ([]domain.TimeLog, error)
	UpdateFunc          func(ctx context.Context, log *domain.TimeLog) error
	DeleteFunc          func(ctx context.Context, id string) error
	ApproveTimeLogsFunc func(ctx context.Context, ids []string, approverHrID string) error
}

func (m *MockTimeLogRepository) Create(ctx context.Context, log *domain.TimeLog) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, log)
	}
	return m.TimeLogRepository.Create(ctx, log)
}

func (m *MockTimeLogRepository) GetByID(ctx context.Context, id string) (*domain.TimeLog, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return m.TimeLogRepository.GetByID(ctx, id)
}

func (m *MockTimeLogRepository) List(ctx context.Context) ([]domain.TimeLog, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return m.TimeLogRepository.List(ctx)
}

func (m *MockTimeLogRepository) Update(ctx context.Context, log *domain.TimeLog) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, log)
	}
	return m.TimeLogRepository.Update(ctx, log)
}

func (m *MockTimeLogRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return m.TimeLogRepository.Delete(ctx, id)
}

func (m *MockTimeLogRepository) ApproveTimeLogs(ctx context.Context, ids []string, approverHrID string) error {
	if m.ApproveTimeLogsFunc != nil {
		return m.ApproveTimeLogsFunc(ctx, ids, approverHrID)
	}
	return m.TimeLogRepository.ApproveTimeLogs(ctx, ids, approverHrID)
}

type MockTransactionalOutboxRepository struct {
	domain.TransactionalOutboxRepository
	CreateFunc    func(ctx context.Context, msg *domain.TransactionalOutbox) error
	GetByIDFunc   func(ctx context.Context, id string) (*domain.TransactionalOutbox, error)
	GetUnsentFunc func(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error)
	UpdateFunc    func(ctx context.Context, msg *domain.TransactionalOutbox) error
}

func (m *MockTransactionalOutboxRepository) Create(ctx context.Context, msg *domain.TransactionalOutbox) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, msg)
	}
	return m.TransactionalOutboxRepository.Create(ctx, msg)
}

func (m *MockTransactionalOutboxRepository) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return m.TransactionalOutboxRepository.GetByID(ctx, id)
}

func (m *MockTransactionalOutboxRepository) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	if m.GetUnsentFunc != nil {
		return m.GetUnsentFunc(ctx, limit)
	}
	return m.TransactionalOutboxRepository.GetUnsent(ctx, limit)
}

func (m *MockTransactionalOutboxRepository) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, msg)
	}
	return m.TransactionalOutboxRepository.Update(ctx, msg)
}

type MockKafkaEventInboxRepository struct {
	domain.KafkaEventInboxRepository
	CreateFunc  func(ctx context.Context, msg *domain.KafkaEventInbox) error
	GetByIDFunc func(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error)
	UpdateFunc  func(ctx context.Context, msg *domain.KafkaEventInbox) error
}

func (m *MockKafkaEventInboxRepository) Create(ctx context.Context, msg *domain.KafkaEventInbox) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, msg)
	}
	return m.KafkaEventInboxRepository.Create(ctx, msg)
}

func (m *MockKafkaEventInboxRepository) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, eventID)
	}
	return m.KafkaEventInboxRepository.GetByID(ctx, eventID)
}

func (m *MockKafkaEventInboxRepository) Update(ctx context.Context, msg *domain.KafkaEventInbox) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, msg)
	}
	return m.KafkaEventInboxRepository.Update(ctx, msg)
}

// ============================================================================
// Extended Test Cases
// ============================================================================

func TestGetDB_Extended(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)

	// Context without txKey
	ctx := context.Background()
	res := service.GetDB(ctx, db)
	if res == nil {
		t.Error("expected non-nil db")
	}

	// Context with txKey (using string literal "gorm_tx" since txKey is package private)
	ctxTx := context.WithValue(context.Background(), "gorm_tx", db)
	resTx := service.GetDB(ctxTx, db)
	if resTx == nil {
		t.Error("expected non-nil db from context")
	}
}

func TestProjectTrackingService_InitializeProject_Error(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	mockRepo := &MockProjectRepository{
		ProjectRepository: memory.NewProjectRepository(),
		CreateFunc: func(ctx context.Context, project *domain.Project) error {
			return errors.New("mock create error")
		},
	}
	svc := service.NewProjectTrackingService(db, mockRepo)
	_, err := svc.InitializeProject(
		context.Background(),
		"tenant-1",
		"cust-1",
		"PRJ-001",
		"Implementation Project",
		domain.BillingMethodTIME_AND_MATERIALS,
		time.Now(),
	)
	if err == nil {
		t.Error("expected error on initialize project, got nil")
	}
}

func TestProjectTrackingService_TransitionProjectStatus_Errors(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)

	// Test GetByID error
	mockRepo := &MockProjectRepository{
		ProjectRepository: memory.NewProjectRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, errors.New("mock get error")
		},
	}
	svc := service.NewProjectTrackingService(db, mockRepo)
	_, err := svc.TransitionProjectStatus(context.Background(), "some-id", domain.ProjectStatusACTIVE)
	if err == nil {
		t.Error("expected error when GetByID fails, got nil")
	}

	// Test Update error
	memRepo := memory.NewProjectRepository()
	_ = memRepo.Create(context.Background(), &domain.Project{ID: "some-id"})
	mockRepo2 := &MockProjectRepository{
		ProjectRepository: memRepo,
		UpdateFunc: func(ctx context.Context, project *domain.Project) error {
			return errors.New("mock update error")
		},
	}
	svc2 := service.NewProjectTrackingService(db, mockRepo2)
	_, err = svc2.TransitionProjectStatus(context.Background(), "some-id", domain.ProjectStatusACTIVE)
	if err == nil {
		t.Error("expected error when Update fails, got nil")
	}
}

func TestWbsStructureService_AppendWbsNode_ProjectNotFound(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	mockProjRepo := &MockProjectRepository{
		ProjectRepository: memory.NewProjectRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, errors.New("project not found")
		},
	}
	svc := service.NewWbsStructureService(db, mockProjRepo, memory.NewWbsNodeRepository(), memory.NewTransactionalOutboxRepository())
	_, err := svc.AppendWbsNode(context.Background(), "non-existent", nil, "1.0", "Title", domain.WbsNodeTypeTASK, decimal.Zero)
	if err == nil {
		t.Error("expected error when project is not found, got nil")
	}
}

func TestWbsStructureService_AppendWbsNode_ParentNotFound(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	projRepo := memory.NewProjectRepository()
	_ = projRepo.Create(context.Background(), &domain.Project{ID: "prj-1"})

	mockWbsRepo := &MockWbsNodeRepository{
		WbsNodeRepository: memory.NewWbsNodeRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.WbsNode, error) {
			return nil, errors.New("parent node not found")
		},
	}
	svc := service.NewWbsStructureService(db, projRepo, mockWbsRepo, memory.NewTransactionalOutboxRepository())
	parentID := "parent-1"
	_, err := svc.AppendWbsNode(context.Background(), "prj-1", &parentID, "1.0", "Title", domain.WbsNodeTypeTASK, decimal.Zero)
	if err == nil {
		t.Error("expected error when parent node is not found, got nil")
	}
}

func TestWbsStructureService_AppendWbsNode_WithParentSuccess(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	projRepo := memory.NewProjectRepository()
	_ = projRepo.Create(context.Background(), &domain.Project{ID: "prj-1"})

	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{ID: "parent-1", WbsDepthLevel: 2})

	svc := service.NewWbsStructureService(db, projRepo, wbsRepo, memory.NewTransactionalOutboxRepository())
	parentID := "parent-1"
	node, err := svc.AppendWbsNode(context.Background(), "prj-1", &parentID, "1.0", "Title", domain.WbsNodeTypeTASK, decimal.Zero)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if node.WbsDepthLevel != 3 {
		t.Errorf("expected depth level 3, got: %d", node.WbsDepthLevel)
	}
}

func TestWbsStructureService_AppendWbsNode_CreateError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	projRepo := memory.NewProjectRepository()
	_ = projRepo.Create(context.Background(), &domain.Project{ID: "prj-1"})

	mockWbsRepo := &MockWbsNodeRepository{
		WbsNodeRepository: memory.NewWbsNodeRepository(),
		CreateFunc: func(ctx context.Context, node *domain.WbsNode) error {
			return errors.New("mock create node error")
		},
	}
	svc := service.NewWbsStructureService(db, projRepo, mockWbsRepo, memory.NewTransactionalOutboxRepository())
	_, err := svc.AppendWbsNode(context.Background(), "prj-1", nil, "1.0", "Title", domain.WbsNodeTypeTASK, decimal.Zero)
	if err == nil {
		t.Error("expected error when creating node fails, got nil")
	}
}

func TestWbsStructureService_DeclareNodeCompletion_GetNodeError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	mockWbsRepo := &MockWbsNodeRepository{
		WbsNodeRepository: memory.NewWbsNodeRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.WbsNode, error) {
			return nil, errors.New("node not found")
		},
	}
	svc := service.NewWbsStructureService(db, memory.NewProjectRepository(), mockWbsRepo, memory.NewTransactionalOutboxRepository())
	_, err := svc.DeclareNodeCompletion(context.Background(), "some-node", "emp-1")
	if err == nil {
		t.Error("expected error when GetByID fails, got nil")
	}
}

func TestWbsStructureService_DeclareNodeCompletion_AlreadyCompleted(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{ID: "node-1", IsCompleted: true})
	svc := service.NewWbsStructureService(db, memory.NewProjectRepository(), wbsRepo, memory.NewTransactionalOutboxRepository())
	node, err := svc.DeclareNodeCompletion(context.Background(), "node-1", "emp-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !node.IsCompleted {
		t.Error("expected node to remain completed")
	}
}

func TestWbsStructureService_DeclareNodeCompletion_UpdateError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{ID: "node-1", IsCompleted: false})
	mockWbsRepo := &MockWbsNodeRepository{
		WbsNodeRepository: wbsRepo,
		UpdateFunc: func(ctx context.Context, node *domain.WbsNode) error {
			return errors.New("update error")
		},
	}
	svc := service.NewWbsStructureService(db, memory.NewProjectRepository(), mockWbsRepo, memory.NewTransactionalOutboxRepository())
	_, err := svc.DeclareNodeCompletion(context.Background(), "node-1", "emp-1")
	if err == nil {
		t.Error("expected error when update fails, got nil")
	}
}

func TestWbsStructureService_DeclareNodeCompletion_ProjectGetError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{
		ID:        "node-1",
		ProjectID: "prj-1",
		NodeType:  domain.WbsNodeTypeMILESTONE,
	})
	mockProjRepo := &MockProjectRepository{
		ProjectRepository: memory.NewProjectRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, errors.New("project not found")
		},
	}
	svc := service.NewWbsStructureService(db, mockProjRepo, wbsRepo, memory.NewTransactionalOutboxRepository())
	_, err := svc.DeclareNodeCompletion(context.Background(), "node-1", "emp-1")
	if err == nil {
		t.Error("expected error when project GetByID fails, got nil")
	}
}

func TestWbsStructureService_DeclareNodeCompletion_OutboxError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{
		ID:        "node-1",
		ProjectID: "prj-1",
		NodeType:  domain.WbsNodeTypeMILESTONE,
	})
	projRepo := memory.NewProjectRepository()
	_ = projRepo.Create(context.Background(), &domain.Project{ID: "prj-1"})

	mockOutboxRepo := &MockTransactionalOutboxRepository{
		TransactionalOutboxRepository: memory.NewTransactionalOutboxRepository(),
		CreateFunc: func(ctx context.Context, msg *domain.TransactionalOutbox) error {
			return errors.New("outbox create error")
		},
	}
	svc := service.NewWbsStructureService(db, projRepo, wbsRepo, mockOutboxRepo)
	_, err := svc.DeclareNodeCompletion(context.Background(), "node-1", "emp-1")
	if err == nil {
		t.Error("expected error when outbox creation fails, got nil")
	}
}

func TestWbsStructureService_FetchProjectTree(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{ID: "n-1", ProjectID: "prj-1"})
	svc := service.NewWbsStructureService(db, memory.NewProjectRepository(), wbsRepo, memory.NewTransactionalOutboxRepository())
	list, err := svc.FetchProjectTree(context.Background(), "prj-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 node, got %d", len(list))
	}
}

func TestTimeTrackingService_LogOperationalHoursBulk_EmptyLogs(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	svc := service.NewTimeTrackingService(db, nil, nil, nil, nil)
	err := svc.LogOperationalHoursBulk(context.Background(), "tenant-1", "emp-1", nil)
	if err != nil {
		t.Errorf("expected no error when logs is nil, got: %v", err)
	}
}

func TestTimeTrackingService_LogOperationalHoursBulk_WbsNotFound(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	mockWbsRepo := &MockWbsNodeRepository{
		WbsNodeRepository: memory.NewWbsNodeRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.WbsNode, error) {
			return nil, errors.New("wbs node not found")
		},
	}
	svc := service.NewTimeTrackingService(db, nil, mockWbsRepo, nil, nil)
	logs := []domain.TimeLogSubmissionInput{{WbsNodeID: "invalid"}}
	err := svc.LogOperationalHoursBulk(context.Background(), "tenant-1", "emp-1", logs)
	if err == nil {
		t.Error("expected error when wbs node not found, got nil")
	}
}

func TestTimeTrackingService_LogOperationalHoursBulk_CreateError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{ID: "node-1"})

	mockTimeRepo := &MockTimeLogRepository{
		TimeLogRepository: memory.NewTimeLogRepository(),
		CreateFunc: func(ctx context.Context, log *domain.TimeLog) error {
			return errors.New("time log create error")
		},
	}
	svc := service.NewTimeTrackingService(db, nil, wbsRepo, mockTimeRepo, nil)
	logs := []domain.TimeLogSubmissionInput{{WbsNodeID: "node-1"}}
	err := svc.LogOperationalHoursBulk(context.Background(), "tenant-1", "emp-1", logs)
	if err == nil {
		t.Error("expected error when time log creation fails, got nil")
	}
}

func TestTimeTrackingService_ProcessTimesheetApproval_EmptyIDs(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	svc := service.NewTimeTrackingService(db, nil, nil, nil, nil)
	err := svc.ProcessTimesheetApproval(context.Background(), nil, "approver-1")
	if err != nil {
		t.Errorf("expected no error for empty logs list, got: %v", err)
	}
}

func TestTimeTrackingService_ProcessTimesheetApproval_ApproveError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	mockTimeRepo := &MockTimeLogRepository{
		TimeLogRepository: memory.NewTimeLogRepository(),
		ApproveTimeLogsFunc: func(ctx context.Context, ids []string, approverHrID string) error {
			return errors.New("approve error")
		},
	}
	svc := service.NewTimeTrackingService(db, nil, nil, mockTimeRepo, nil)
	err := svc.ProcessTimesheetApproval(context.Background(), []string{"log-1"}, "approver-1")
	if err == nil {
		t.Error("expected error when ApproveTimeLogs fails, got nil")
	}
}

func TestTimeTrackingService_ProcessTimesheetApproval_GetLogByIDError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	timeRepo := memory.NewTimeLogRepository()
	_ = timeRepo.Create(context.Background(), &domain.TimeLog{ID: "log-1"})

	mockTimeRepo := &MockTimeLogRepository{
		TimeLogRepository: timeRepo,
		GetByIDFunc: func(ctx context.Context, id string) (*domain.TimeLog, error) {
			return nil, errors.New("mock get log error")
		},
	}
	svc := service.NewTimeTrackingService(db, nil, nil, mockTimeRepo, nil)
	err := svc.ProcessTimesheetApproval(context.Background(), []string{"log-1"}, "approver-1")
	if err == nil {
		t.Error("expected error when GetByID fails, got nil")
	}
}

func TestTimeTrackingService_ProcessTimesheetApproval_GetWbsError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	timeRepo := memory.NewTimeLogRepository()
	_ = timeRepo.Create(context.Background(), &domain.TimeLog{ID: "log-1", WbsNodeID: "node-1"})

	mockWbsRepo := &MockWbsNodeRepository{
		WbsNodeRepository: memory.NewWbsNodeRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.WbsNode, error) {
			return nil, errors.New("mock get wbs node error")
		},
	}
	svc := service.NewTimeTrackingService(db, nil, mockWbsRepo, timeRepo, nil)
	err := svc.ProcessTimesheetApproval(context.Background(), []string{"log-1"}, "approver-1")
	if err == nil {
		t.Error("expected error when wbs GetByID fails, got nil")
	}
}

func TestTimeTrackingService_ProcessTimesheetApproval_GetProjectError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	timeRepo := memory.NewTimeLogRepository()
	_ = timeRepo.Create(context.Background(), &domain.TimeLog{ID: "log-1", WbsNodeID: "node-1"})
	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{ID: "node-1", ProjectID: "prj-1"})

	mockProjRepo := &MockProjectRepository{
		ProjectRepository: memory.NewProjectRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, errors.New("mock get project error")
		},
	}
	svc := service.NewTimeTrackingService(db, mockProjRepo, wbsRepo, timeRepo, nil)
	err := svc.ProcessTimesheetApproval(context.Background(), []string{"log-1"}, "approver-1")
	if err == nil {
		t.Error("expected error when project GetByID fails, got nil")
	}
}

func TestTimeTrackingService_ProcessTimesheetApproval_OutboxCreateError(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)
	timeRepo := memory.NewTimeLogRepository()
	_ = timeRepo.Create(context.Background(), &domain.TimeLog{ID: "log-1", WbsNodeID: "node-1"})
	wbsRepo := memory.NewWbsNodeRepository()
	_ = wbsRepo.Create(context.Background(), &domain.WbsNode{ID: "node-1", ProjectID: "prj-1"})
	projRepo := memory.NewProjectRepository()
	_ = projRepo.Create(context.Background(), &domain.Project{ID: "prj-1"})

	mockOutboxRepo := &MockTransactionalOutboxRepository{
		TransactionalOutboxRepository: memory.NewTransactionalOutboxRepository(),
		CreateFunc: func(ctx context.Context, msg *domain.TransactionalOutbox) error {
			return errors.New("mock outbox create error")
		},
	}
	svc := service.NewTimeTrackingService(db, projRepo, wbsRepo, timeRepo, mockOutboxRepo)
	err := svc.ProcessTimesheetApproval(context.Background(), []string{"log-1"}, "approver-1")
	if err == nil {
		t.Error("expected error when outbox creation fails, got nil")
	}
}

func TestOutboxRelayWorker(t *testing.T) {
	// GetByID error
	mockOutboxRepo1 := &MockTransactionalOutboxRepository{
		TransactionalOutboxRepository: memory.NewTransactionalOutboxRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
			return nil, errors.New("not found")
		},
	}
	worker1 := service.NewOutboxRelayWorker(mockOutboxRepo1)
	err := worker1.UpdateOutboxStatus(context.Background(), "msg-1", domain.OutboxStatusSENT)
	if err == nil {
		t.Error("expected error when outbox GetByID fails")
	}

	// Update error
	memOutboxRepo := memory.NewTransactionalOutboxRepository()
	_ = memOutboxRepo.Create(context.Background(), &domain.TransactionalOutbox{ID: "msg-1", Status: domain.OutboxStatusPENDING})
	mockOutboxRepo2 := &MockTransactionalOutboxRepository{
		TransactionalOutboxRepository: memOutboxRepo,
		UpdateFunc: func(ctx context.Context, msg *domain.TransactionalOutbox) error {
			return errors.New("update error")
		},
	}
	worker2 := service.NewOutboxRelayWorker(mockOutboxRepo2)
	err = worker2.UpdateOutboxStatus(context.Background(), "msg-1", domain.OutboxStatusSENT)
	if err == nil {
		t.Error("expected error when outbox Update fails")
	}

	// Happy path
	worker3 := service.NewOutboxRelayWorker(memOutboxRepo)
	unsent, err := worker3.GetUnsentMessages(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(unsent) != 1 {
		t.Errorf("expected 1 unsent msg, got %d", len(unsent))
	}
	err = worker3.UpdateOutboxStatus(context.Background(), "msg-1", domain.OutboxStatusSENT)
	if err != nil {
		t.Fatalf("expected no error updating status, got %v", err)
	}
	unsentAfter, _ := worker3.GetUnsentMessages(context.Background(), 10)
	if len(unsentAfter) != 0 {
		t.Errorf("expected 0 unsent msgs after status update, got %d", len(unsentAfter))
	}
}

func TestReliableMessagingService(t *testing.T) {
	db, _, _, _, _ := setupTestDB(t)

	// Test IsEventProcessed error
	mockInboxRepo := &MockKafkaEventInboxRepository{
		KafkaEventInboxRepository: memory.NewKafkaEventInboxRepository(),
		GetByIDFunc: func(ctx context.Context, id string) (*domain.KafkaEventInbox, error) {
			return nil, errors.New("db error")
		},
	}
	svc := service.NewReliableMessagingService(db, mockInboxRepo)
	processed, err := svc.IsEventProcessed(context.Background(), "evt-1")
	if err != nil {
		t.Errorf("IsEventProcessed should swallow error and return false, got: %v", err)
	}
	if processed {
		t.Error("expected processed to be false")
	}

	// Test IsEventProcessed returns success
	memInboxRepo := memory.NewKafkaEventInboxRepository()
	_ = memInboxRepo.Create(context.Background(), &domain.KafkaEventInbox{
		EventID:          "evt-2",
		ProcessingStatus: domain.EventProcessingStatusSUCCESS,
	})
	svc2 := service.NewReliableMessagingService(db, memInboxRepo)
	processed, err = svc2.IsEventProcessed(context.Background(), "evt-2")
	if err != nil || !processed {
		t.Errorf("expected processed to be true, got err: %v, processed: %t", err, processed)
	}

	// Test IsEventProcessed returns failure status
	_ = memInboxRepo.Create(context.Background(), &domain.KafkaEventInbox{
		EventID:          "evt-3",
		ProcessingStatus: domain.EventProcessingStatusFAILED,
	})
	processed, err = svc2.IsEventProcessed(context.Background(), "evt-3")
	if err != nil || processed {
		t.Errorf("expected processed to be false, got err: %v, processed: %t", err, processed)
	}

	// Test ExecuteIdempotentTransaction when already processed
	called := false
	err = svc2.ExecuteIdempotentTransaction(context.Background(), "evt-2", "type", nil, func(ctx context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if called {
		t.Error("expected business routine NOT to be called because event was already processed")
	}

	// Test ExecuteIdempotentTransaction when businessRoutine fails
	called = false
	err = svc2.ExecuteIdempotentTransaction(context.Background(), "evt-new-fail", "type", "payload", func(ctx context.Context) error {
		called = true
		return errors.New("business error")
	})
	if err == nil {
		t.Error("expected business error, got nil")
	}
	if !called {
		t.Error("expected business routine to be called")
	}
	// Verify failed inbox entry was created
	inboxEntry, err := memInboxRepo.GetByID(context.Background(), "evt-new-fail")
	if err != nil {
		t.Fatalf("expected inbox entry to exist, got: %v", err)
	}
	if inboxEntry.ProcessingStatus != domain.EventProcessingStatusFAILED {
		t.Errorf("expected FAILED status, got %v", inboxEntry.ProcessingStatus)
	}

	// Test ExecuteIdempotentTransaction when inboxRepo.Create fails on success path
	mockInboxRepoCreateFail := &MockKafkaEventInboxRepository{
		KafkaEventInboxRepository: memory.NewKafkaEventInboxRepository(),
		CreateFunc: func(ctx context.Context, msg *domain.KafkaEventInbox) error {
			return errors.New("create error")
		},
	}
	svc3 := service.NewReliableMessagingService(db, mockInboxRepoCreateFail)
	err = svc3.ExecuteIdempotentTransaction(context.Background(), "evt-success-create-fail", "type", "payload", func(ctx context.Context) error {
		return nil
	})
	if err == nil {
		t.Error("expected error when inbox creation fails, got nil")
	}

	// Test ExecuteIdempotentTransaction success
	err = svc2.ExecuteIdempotentTransaction(context.Background(), "evt-new-success", "type", "payload", func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	inboxEntrySuccess, err := memInboxRepo.GetByID(context.Background(), "evt-new-success")
	if err != nil {
		t.Fatalf("expected inbox entry to exist, got: %v", err)
	}
	if inboxEntrySuccess.ProcessingStatus != domain.EventProcessingStatusSUCCESS {
		t.Errorf("expected SUCCESS status, got %v", inboxEntrySuccess.ProcessingStatus)
	}
}
