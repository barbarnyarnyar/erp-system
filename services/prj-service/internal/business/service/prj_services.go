package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"erp-system/shared/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const txKey = "gorm_tx"

func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}

// ============================================================================
// ProjectTrackingService
// ============================================================================

type ProjectTrackingService interface {
	InitializeProject(ctx context.Context, legalEntityID string, customerID string, code string, name string, method domain.BillingMethod, start time.Time) (*domain.Project, error)
	TransitionProjectStatus(ctx context.Context, projectID string, newStatus domain.ProjectStatus) (*domain.Project, error)
}

type ProjectTrackingServiceImpl struct {
	db        *gorm.DB
	projRepo  domain.ProjectRepository
}

func NewProjectTrackingService(db *gorm.DB, projRepo domain.ProjectRepository) ProjectTrackingService {
	return &ProjectTrackingServiceImpl{
		db:       db,
		projRepo: projRepo,
	}
}

func (s *ProjectTrackingServiceImpl) InitializeProject(
	ctx context.Context,
	legalEntityID string,
	customerID string,
	code string,
	name string,
	method domain.BillingMethod,
	start time.Time,
) (*domain.Project, error) {
	proj := &domain.Project{
		ID:            utils.NewID("prj"),
		LegalEntityID: legalEntityID,
		CustomerID:    customerID,
		ProjectCode:   code,
		Name:          name,
		Status:        domain.ProjectStatusDRAFT,
		BillingMethod: method,
		StartDate:     start,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.projRepo.Create(ctx, proj)
	if err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *ProjectTrackingServiceImpl) TransitionProjectStatus(
	ctx context.Context,
	projectID string,
	newStatus domain.ProjectStatus,
) (*domain.Project, error) {
	proj, err := s.projRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	proj.Status = newStatus
	proj.UpdatedAt = time.Now()

	err = s.projRepo.Update(ctx, proj)
	if err != nil {
		return nil, err
	}
	return proj, nil
}

// ============================================================================
// WbsStructureService
// ============================================================================

type WbsStructureService interface {
	AppendWbsNode(ctx context.Context, projectID string, parentNodeID *string, code string, title string, nodeType domain.WbsNodeType, hours decimal.Decimal) (*domain.WbsNode, error)
	DeclareNodeCompletion(ctx context.Context, nodeID string, completionHrID string) (*domain.WbsNode, error)
	FetchProjectTree(ctx context.Context, projectID string) ([]domain.WbsNode, error)
}

type WbsStructureServiceImpl struct {
	db         *gorm.DB
	projRepo   domain.ProjectRepository
	wbsRepo    domain.WbsNodeRepository
	outboxRepo domain.TransactionalOutboxRepository
}

func NewWbsStructureService(
	db *gorm.DB,
	projRepo domain.ProjectRepository,
	wbsRepo domain.WbsNodeRepository,
	outboxRepo domain.TransactionalOutboxRepository,
) WbsStructureService {
	return &WbsStructureServiceImpl{
		db:         db,
		projRepo:   projRepo,
		wbsRepo:    wbsRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *WbsStructureServiceImpl) AppendWbsNode(
	ctx context.Context,
	projectID string,
	parentNodeID *string,
	code string,
	title string,
	nodeType domain.WbsNodeType,
	hours decimal.Decimal,
) (*domain.WbsNode, error) {
	// Verify project exists
	_, err := s.projRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	depth := 0
	if parentNodeID != nil && *parentNodeID != "" {
		parent, err := s.wbsRepo.GetByID(ctx, *parentNodeID)
		if err != nil {
			return nil, fmt.Errorf("parent WBS node not found: %w", err)
		}
		depth = parent.WbsDepthLevel + 1
	}

	node := &domain.WbsNode{
		ID:            utils.NewID("wbs"),
		ProjectID:     projectID,
		ParentNodeID:  parentNodeID,
		WbsDepthLevel: depth,
		NodeCode:      code,
		Title:         title,
		NodeType:      nodeType,
		EstimatedHours: hours,
		IsCompleted:   false,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.wbsRepo.Create(ctx, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (s *WbsStructureServiceImpl) DeclareNodeCompletion(
	ctx context.Context,
	nodeID string,
	completionHrID string,
) (*domain.WbsNode, error) {
	var node *domain.WbsNode
	var proj *domain.Project

	err := s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		var err error
		node, err = s.wbsRepo.GetByID(txCtx, nodeID)
		if err != nil {
			return err
		}

		if node.IsCompleted {
			return nil // Idempotent success
		}

		node.IsCompleted = true
		node.UpdatedAt = time.Now()
		if err := s.wbsRepo.Update(txCtx, node); err != nil {
			return err
		}

		if node.NodeType == domain.WbsNodeTypeMILESTONE {
			proj, err = s.projRepo.GetByID(txCtx, node.ProjectID)
			if err != nil {
				return err
			}

			var rev decimal.Decimal
			if node.BudgetRevenueFunctional != nil {
				rev = *node.BudgetRevenueFunctional
			}

			evtPayload := domain.PrjMilestoneAchievedEvent{
				EventID:       utils.NewID("evt"),
				LegalEntityID: proj.LegalEntityID,
				ProjectID:     proj.ID,
				CustomerID:    proj.CustomerID,
				WbsNodeID:     node.ID,
				RevenueAmount: rev,
				Timestamp:     time.Now(),
			}

			outbox := &domain.TransactionalOutbox{
				ID:          utils.NewID("outbox"),
				EventType:   domain.TopicPrjMilestoneAchieved,
				AggregateID: node.ID,
				Payload:     evtPayload,
				Status:      domain.OutboxStatusPENDING,
				CreatedAt:   time.Now(),
			}

			if err := s.outboxRepo.Create(txCtx, outbox); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return node, nil
}

func (s *WbsStructureServiceImpl) FetchProjectTree(ctx context.Context, projectID string) ([]domain.WbsNode, error) {
	return s.wbsRepo.ListByProjectID(ctx, projectID)
}

// ============================================================================
// TimeTrackingService
// ============================================================================

type TimeTrackingService interface {
	LogOperationalHoursBulk(ctx context.Context, legalEntityID string, employeeID string, logs []domain.TimeLogSubmissionInput) error
	ProcessTimesheetApproval(ctx context.Context, timeLogIDs []string, approverHrID string) error
}

type TimeTrackingServiceImpl struct {
	db         *gorm.DB
	projRepo   domain.ProjectRepository
	wbsRepo    domain.WbsNodeRepository
	timeRepo   domain.TimeLogRepository
	outboxRepo domain.TransactionalOutboxRepository
}

func NewTimeTrackingService(
	db *gorm.DB,
	projRepo domain.ProjectRepository,
	wbsRepo domain.WbsNodeRepository,
	timeRepo domain.TimeLogRepository,
	outboxRepo domain.TransactionalOutboxRepository,
) TimeTrackingService {
	return &TimeTrackingServiceImpl{
		db:         db,
		projRepo:   projRepo,
		wbsRepo:    wbsRepo,
		timeRepo:   timeRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *TimeTrackingServiceImpl) LogOperationalHoursBulk(
	ctx context.Context,
	legalEntityID string,
	employeeID string,
	logs []domain.TimeLogSubmissionInput,
) error {
	if len(logs) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		for _, logInput := range logs {
			// Verify WBS node exists
			_, err := s.wbsRepo.GetByID(txCtx, logInput.WbsNodeID)
			if err != nil {
				return fmt.Errorf("wbs node %s not found: %w", logInput.WbsNodeID, err)
			}

			timeLog := &domain.TimeLog{
				ID:               utils.NewID("tim"),
				LegalEntityID:    legalEntityID,
				WbsNodeID:        logInput.WbsNodeID,
				EmployeeID:       employeeID,
				WorkDate:         logInput.WorkDate,
				HoursSpent:       logInput.HoursSpent,
				InternalCostRate: logInput.InternalCostRate,
				BillingRate:      logInput.BillingRate,
				IsBillable:       logInput.IsBillable,
				IsApproved:       false,
				CreatedAt:        time.Now(),
			}

			if err := s.timeRepo.Create(txCtx, timeLog); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *TimeTrackingServiceImpl) ProcessTimesheetApproval(
	ctx context.Context,
	timeLogIDs []string,
	approverHrID string,
) error {
	if len(timeLogIDs) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		// 1. Approve time logs
		if err := s.timeRepo.ApproveTimeLogs(txCtx, timeLogIDs, approverHrID); err != nil {
			return err
		}

		// 2. Fetch all approved time logs
		var logs []domain.TimeLog
		for _, id := range timeLogIDs {
			log, err := s.timeRepo.GetByID(txCtx, id)
			if err != nil {
				return err
			}
			logs = append(logs, *log)
		}

		// Group logs by Project ID
		projectLogs := make(map[string][]domain.TimeLog)
		for _, log := range logs {
			node, err := s.wbsRepo.GetByID(txCtx, log.WbsNodeID)
			if err != nil {
				return err
			}
			projectLogs[node.ProjectID] = append(projectLogs[node.ProjectID], log)
		}

		// Grouped outbox emission per project
		for projID, pLogs := range projectLogs {
			proj, err := s.projRepo.GetByID(txCtx, projID)
			if err != nil {
				return err
			}

			var totalHours decimal.Decimal
			var details []domain.TimeLogPayload
			for _, log := range pLogs {
				totalHours = totalHours.Add(log.HoursSpent)
				details = append(details, domain.TimeLogPayload{
					TimeLogID:   log.ID,
					WbsNodeID:   log.WbsNodeID,
					EmployeeID:  log.EmployeeID,
					HoursSpent:  log.HoursSpent,
					BillingRate: log.BillingRate,
				})
			}

			evtPayload := domain.PrjTimeLoggedEvent{
				EventID:               utils.NewID("evt"),
				LegalEntityID:         proj.LegalEntityID,
				ProjectID:             proj.ID,
				CustomerID:            proj.CustomerID,
				TotalAccumulatedHours: totalHours,
				Details:               details,
				Timestamp:             time.Now(),
			}

			outbox := &domain.TransactionalOutbox{
				ID:          utils.NewID("outbox"),
				EventType:   domain.TopicPrjTimeLogged,
				AggregateID: proj.ID,
				Payload:     evtPayload,
				Status:      domain.OutboxStatusPENDING,
				CreatedAt:   time.Now(),
			}

			if err := s.outboxRepo.Create(txCtx, outbox); err != nil {
				return err
			}
		}

		return nil
	})
}

// ============================================================================
// OutboxRelayWorker
// ============================================================================

type OutboxRelayWorker interface {
	GetUnsentMessages(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error)
	UpdateOutboxStatus(ctx context.Context, outboxID string, status domain.OutboxStatus) error
}

type OutboxRelayWorkerImpl struct {
	outboxRepo domain.TransactionalOutboxRepository
}

func NewOutboxRelayWorker(outboxRepo domain.TransactionalOutboxRepository) OutboxRelayWorker {
	return &OutboxRelayWorkerImpl{outboxRepo: outboxRepo}
}

func (s *OutboxRelayWorkerImpl) GetUnsentMessages(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	return s.outboxRepo.GetUnsent(ctx, limit)
}

func (s *OutboxRelayWorkerImpl) UpdateOutboxStatus(ctx context.Context, outboxID string, status domain.OutboxStatus) error {
	msg, err := s.outboxRepo.GetByID(ctx, outboxID)
	if err != nil {
		return err
	}
	msg.Status = status
	return s.outboxRepo.Update(ctx, msg)
}

// ============================================================================
// ReliableMessagingService
// ============================================================================

type ReliableMessagingService interface {
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
	ExecuteIdempotentTransaction(ctx context.Context, eventID string, eventType string, payload interface{}, businessRoutine func(ctx context.Context) error) error
}

type ReliableMessagingServiceImpl struct {
	db        *gorm.DB
	inboxRepo domain.KafkaEventInboxRepository
}

func NewReliableMessagingService(db *gorm.DB, inboxRepo domain.KafkaEventInboxRepository) ReliableMessagingService {
	return &ReliableMessagingServiceImpl{
		db:        db,
		inboxRepo: inboxRepo,
	}
}

func (s *ReliableMessagingServiceImpl) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	msg, err := s.inboxRepo.GetByID(ctx, eventID)
	if err == nil && msg != nil {
		return msg.ProcessingStatus == domain.EventProcessingStatusSUCCESS, nil
	}
	return false, nil
}

func (s *ReliableMessagingServiceImpl) ExecuteIdempotentTransaction(
	ctx context.Context,
	eventID string,
	eventType string,
	payload interface{},
	businessRoutine func(ctx context.Context) error,
) error {
	// First check if the event was already successfully processed or sent to DLQ
	msg, err := s.inboxRepo.GetByID(ctx, eventID)
	if err == nil && msg != nil {
		if msg.ProcessingStatus == domain.EventProcessingStatusSUCCESS || msg.ProcessingStatus == "FAILED_DLQ" {
			return nil
		}
	}

	attempts := 0
	if msg != nil {
		attempts = msg.AttemptCount
	}
	attempts++

	// If attempts >= 5, route to DLQ
	if attempts >= 5 {
		inboxEntry := &domain.KafkaEventInbox{
			EventID:          eventID,
			EventType:        eventType,
			ProcessedAt:      time.Now(),
			ProcessingStatus: "FAILED_DLQ",
			Payload:          payload,
			AttemptCount:     attempts,
		}

		if msg != nil {
			_ = s.inboxRepo.Update(ctx, inboxEntry)
		} else {
			_ = s.inboxRepo.Create(ctx, inboxEntry)
		}

		// Central DLQ topic
		dlqTopic := "erp.system.dlq"
		dlqMsg := map[string]interface{}{
			"event_id":       eventID,
			"original_topic": eventType,
			"payload":        payload,
			"error":          "Max retries reached (5 attempts)",
			"failed_at":      time.Now(),
		}

		if pub, ok := ctx.Value("publisher").(interface {
			Publish(ctx context.Context, topic string, key string, payload interface{}) error
		}); ok && pub != nil {
			_ = pub.Publish(ctx, dlqTopic, eventID, dlqMsg)
		}

		return nil // Return nil so consumer commits offset and doesn't loop
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)

		if err := businessRoutine(txCtx); err != nil {
			inboxEntry := &domain.KafkaEventInbox{
				EventID:          eventID,
				EventType:        eventType,
				ProcessedAt:      time.Now(),
				ProcessingStatus: domain.EventProcessingStatusFAILED,
				Payload:          payload,
				AttemptCount:     attempts,
			}
			if msg != nil {
				_ = s.inboxRepo.Update(txCtx, inboxEntry)
			} else {
				_ = s.inboxRepo.Create(txCtx, inboxEntry)
			}
			return err
		}

		inboxEntry := &domain.KafkaEventInbox{
			EventID:          eventID,
			EventType:        eventType,
			ProcessedAt:      time.Now(),
			ProcessingStatus: domain.EventProcessingStatusSUCCESS,
			Payload:          payload,
			AttemptCount:     attempts,
		}
		if msg != nil {
			return s.inboxRepo.Update(txCtx, inboxEntry)
		} else {
			return s.inboxRepo.Create(txCtx, inboxEntry)
		}
	})
}
