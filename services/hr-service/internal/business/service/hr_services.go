package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
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

// ==========================================
// EmployeeService Interface & Implementation
// ==========================================

type EmployeeService interface {
	HireEmployee(ctx context.Context, legalEntityId string, departmentId string, managerHrId *string, empNum string, firstName string, lastName string, email string, salary decimal.Decimal, empType domain.EmploymentType) (*domain.EmployeeMaster, error)
	TerminateEmployee(ctx context.Context, employeeId string, terminationDate time.Time) (*domain.EmployeeMaster, error)
	AdjustCompensation(ctx context.Context, employeeId string, targetSalary decimal.Decimal) (*domain.EmployeeMaster, error)
	FetchManagementChain(ctx context.Context, employeeId string) ([]domain.EmployeeMaster, error)
}

type EmployeeServiceImpl struct {
	db         *gorm.DB
	empRepo    domain.EmployeeMasterRepository
	deptRepo   domain.DepartmentRepository
	outboxRepo domain.TransactionalOutboxRepository
}

func NewEmployeeService(
	db *gorm.DB,
	empRepo domain.EmployeeMasterRepository,
	deptRepo domain.DepartmentRepository,
	outboxRepo domain.TransactionalOutboxRepository,
) EmployeeService {
	return &EmployeeServiceImpl{
		db:         db,
		empRepo:    empRepo,
		deptRepo:   deptRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *EmployeeServiceImpl) HireEmployee(
	ctx context.Context,
	legalEntityId string,
	departmentId string,
	managerHrId *string,
	empNum string,
	firstName string,
	lastName string,
	email string,
	salary decimal.Decimal,
	empType domain.EmploymentType,
) (*domain.EmployeeMaster, error) {
	// Verify department exists
	_, err := s.deptRepo.GetByID(ctx, departmentId)
	if err != nil {
		return nil, fmt.Errorf("department not found: %w", err)
	}

	orgDepth := 0
	if managerHrId != nil && *managerHrId != "" {
		mgr, err := s.empRepo.GetByID(ctx, *managerHrId)
		if err != nil {
			return nil, fmt.Errorf("manager not found: %w", err)
		}
		orgDepth = mgr.OrgDepthLevel + 1
	}

	emp := &domain.EmployeeMaster{
		ID:             utils.NewID("emp"),
		LegalEntityID:  legalEntityId,
		DepartmentID:   departmentId,
		ManagerHrID:    managerHrId,
		OrgDepthLevel:  orgDepth,
		EmployeeNumber: empNum,
		FirstName:      firstName,
		LastName:       lastName,
		Email:          email,
		Status:         domain.EmployeeStatusACTIVE,
		Type:           empType,
		BaseSalary:     salary,
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		if err := s.empRepo.Create(txCtx, emp); err != nil {
			return err
		}

		evtPayload := domain.EmployeeCreatedEvent{
			EventID:        utils.NewID("evt"),
			LegalEntityID:  legalEntityId,
			EmployeeID:     emp.ID,
			ManagerHrID:    managerHrId,
			EmployeeNumber: empNum,
			Email:          email,
			BaseSalary:     salary,
			Type:           string(empType),
			Timestamp:      time.Now(),
		}

		outbox := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   "hr.employee.created",
			AggregateID: emp.ID,
			Payload:     evtPayload,
			Status:      domain.OutboxStatusPENDING,
			CreatedAt:   time.Now(),
		}

		return s.outboxRepo.Create(txCtx, outbox)
	})

	if err != nil {
		return nil, err
	}

	return emp, nil
}

func (s *EmployeeServiceImpl) TerminateEmployee(
	ctx context.Context,
	employeeId string,
	terminationDate time.Time,
) (*domain.EmployeeMaster, error) {
	emp, err := s.empRepo.GetByID(ctx, employeeId)
	if err != nil {
		return nil, err
	}

	emp.Status = domain.EmployeeStatusTERMINATED
	now := time.Now()
	emp.DeletedAt = &now
	emp.UpdatedAt = now

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		if err := s.empRepo.Update(txCtx, emp); err != nil {
			return err
		}

		evtPayload := domain.EmployeeTerminatedEvent{
			EventID:        utils.NewID("evt"),
			LegalEntityID:  emp.LegalEntityID,
			EmployeeID:     emp.ID,
			EmployeeNumber: emp.EmployeeNumber,
			Email:          emp.Email,
			Timestamp:      time.Now(),
		}

		outbox := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   "hr.employee.terminated",
			AggregateID: emp.ID,
			Payload:     evtPayload,
			Status:      domain.OutboxStatusPENDING,
			CreatedAt:   time.Now(),
		}

		return s.outboxRepo.Create(txCtx, outbox)
	})

	if err != nil {
		return nil, err
	}

	return emp, nil
}

func (s *EmployeeServiceImpl) AdjustCompensation(
	ctx context.Context,
	employeeId string,
	targetSalary decimal.Decimal,
) (*domain.EmployeeMaster, error) {
	emp, err := s.empRepo.GetByID(ctx, employeeId)
	if err != nil {
		return nil, err
	}

	emp.BaseSalary = targetSalary
	emp.UpdatedAt = time.Now()

	err = s.empRepo.Update(ctx, emp)
	if err != nil {
		return nil, err
	}

	return emp, nil
}

func (s *EmployeeServiceImpl) FetchManagementChain(
	ctx context.Context,
	employeeId string,
) ([]domain.EmployeeMaster, error) {
	var chain []domain.EmployeeMaster
	currID := employeeId

	for {
		emp, err := s.empRepo.GetByID(ctx, currID)
		if err != nil {
			break
		}
		chain = append(chain, *emp)
		if emp.ManagerHrID == nil || *emp.ManagerHrID == "" || *emp.ManagerHrID == currID {
			break
		}
		currID = *emp.ManagerHrID
	}

	return chain, nil
}

// ==========================================
// PayrollService Interface & Implementation
// ==========================================

type PayrollService interface {
	InitiatePeriodRun(ctx context.Context, legalEntityId string, fiscalYear int, periodNumber int) (*domain.PayrollRun, error)
	ExecuteCalculations(ctx context.Context, payrollRunId string) (*domain.PayrollRun, error)
	CloseAndApprovePayroll(ctx context.Context, payrollRunId string) (*domain.PayrollRun, error)
}

type PayrollServiceImpl struct {
	db          *gorm.DB
	payrollRepo domain.PayrollRunRepository
	empRepo     domain.EmployeeMasterRepository
	outboxRepo  domain.TransactionalOutboxRepository
}

func NewPayrollService(
	db *gorm.DB,
	payrollRepo domain.PayrollRunRepository,
	empRepo domain.EmployeeMasterRepository,
	outboxRepo domain.TransactionalOutboxRepository,
) PayrollService {
	return &PayrollServiceImpl{
		db:          db,
		payrollRepo: payrollRepo,
		empRepo:     empRepo,
		outboxRepo:  outboxRepo,
	}
}

func (s *PayrollServiceImpl) InitiatePeriodRun(
	ctx context.Context,
	legalEntityId string,
	fiscalYear int,
	periodNumber int,
) (*domain.PayrollRun, error) {
	// Check if already exists for this period
	existing, err := s.payrollRepo.GetByPeriod(ctx, legalEntityId, fiscalYear, periodNumber)
	if err == nil && existing != nil {
		return nil, errors.New("payroll run for this period already exists")
	}

	run := &domain.PayrollRun{
		ID:              utils.NewID("pay"),
		LegalEntityID:   legalEntityId,
		FiscalYear:      fiscalYear,
		PeriodNumber:    periodNumber,
		Status:          domain.PayrollStatusDRAFT,
		TotalGrossPay:   decimal.Zero,
		TotalDeductions: decimal.Zero,
		TotalNetPay:     decimal.Zero,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.payrollRepo.Create(ctx, run); err != nil {
		return nil, err
	}

	return run, nil
}

func (s *PayrollServiceImpl) ExecuteCalculations(
	ctx context.Context,
	payrollRunId string,
) (*domain.PayrollRun, error) {
	run, err := s.payrollRepo.GetByID(ctx, payrollRunId)
	if err != nil {
		return nil, err
	}

	if run.Status != domain.PayrollStatusDRAFT {
		return nil, errors.New("calculations can only be executed on DRAFT payroll runs")
	}

	employees, err := s.empRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	totalGross := decimal.Zero
	for _, e := range employees {
		if e.LegalEntityID == run.LegalEntityID && e.Status == domain.EmployeeStatusACTIVE {
			totalGross = totalGross.Add(e.BaseSalary)
		}
	}

	// Deductions are 10% of gross
	totalDeductions := totalGross.Mul(decimal.NewFromFloat(0.1))
	totalNet := totalGross.Sub(totalDeductions)

	run.TotalGrossPay = totalGross
	run.TotalDeductions = totalDeductions
	run.TotalNetPay = totalNet
	run.UpdatedAt = time.Now()

	if err := s.payrollRepo.Update(ctx, run); err != nil {
		return nil, err
	}

	return run, nil
}

func (s *PayrollServiceImpl) CloseAndApprovePayroll(
	ctx context.Context,
	payrollRunId string,
) (*domain.PayrollRun, error) {
	run, err := s.payrollRepo.GetByID(ctx, payrollRunId)
	if err != nil {
		return nil, err
	}

	if run.Status != domain.PayrollStatusDRAFT {
		return nil, errors.New("only DRAFT payroll runs can be approved")
	}

	run.Status = domain.PayrollStatusAPPROVED
	run.UpdatedAt = time.Now()

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		if err := s.payrollRepo.Update(txCtx, run); err != nil {
			return err
		}

		evtPayload := domain.PayrollProcessedEvent{
			EventID:       utils.NewID("evt"),
			LegalEntityID: run.LegalEntityID,
			PayrollRunID:  run.ID,
			FiscalYear:    run.FiscalYear,
			PeriodNumber:  run.PeriodNumber,
			TotalNetPay:   run.TotalNetPay,
			TotalGrossPay: run.TotalGrossPay,
			Timestamp:     time.Now(),
		}

		outbox := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   "hr.payroll.processed",
			AggregateID: run.ID,
			Payload:     evtPayload,
			Status:      domain.OutboxStatusPENDING,
			CreatedAt:   time.Now(),
		}

		return s.outboxRepo.Create(txCtx, outbox)
	})

	if err != nil {
		return nil, err
	}

	return run, nil
}

// ==========================================
// ExpenseService Interface & Implementation
// ==========================================

type ExpenseService interface {
	SubmitClaim(ctx context.Context, legalEntityId string, employeeId string, claimNumber string, purpose string, costCenter string, lines []domain.ExpenseClaimLineInput) (*domain.ExpenseClaim, error)
	VerifyAndApproveClaim(ctx context.Context, claimId string, reviewerHrId string) (*domain.ExpenseClaim, error)
	ClearClaimForPayment(ctx context.Context, claimId string) error
}

type ExpenseServiceImpl struct {
	db         *gorm.DB
	claimRepo  domain.ExpenseClaimRepository
	lineRepo   domain.ExpenseClaimLineRepository
	empRepo    domain.EmployeeMasterRepository
	outboxRepo domain.TransactionalOutboxRepository
}

func NewExpenseService(
	db *gorm.DB,
	claimRepo domain.ExpenseClaimRepository,
	lineRepo domain.ExpenseClaimLineRepository,
	empRepo domain.EmployeeMasterRepository,
	outboxRepo domain.TransactionalOutboxRepository,
) ExpenseService {
	return &ExpenseServiceImpl{
		db:         db,
		claimRepo:  claimRepo,
		lineRepo:   lineRepo,
		empRepo:    empRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *ExpenseServiceImpl) SubmitClaim(
	ctx context.Context,
	legalEntityId string,
	employeeId string,
	claimNumber string,
	purpose string,
	costCenter string,
	lines []domain.ExpenseClaimLineInput,
) (*domain.ExpenseClaim, error) {
	// Verify employee
	_, err := s.empRepo.GetByID(ctx, employeeId)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	totalAmount := decimal.Zero
	for _, l := range lines {
		totalAmount = totalAmount.Add(l.Amount)
	}

	claim := &domain.ExpenseClaim{
		ID:            utils.NewID("exp"),
		LegalEntityID: legalEntityId,
		EmployeeID:    employeeId,
		ClaimNumber:   claimNumber,
		Purpose:       purpose,
		TotalAmount:   totalAmount,
		Status:        domain.ExpenseStatusSUBMITTED,
		CostCenterTag: costCenter,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		if err := s.claimRepo.Create(txCtx, claim); err != nil {
			return err
		}

		for _, l := range lines {
			line := &domain.ExpenseClaimLine{
				ID:             utils.NewID("expl"),
				ExpenseClaimID: claim.ID,
				Description:    l.Description,
				LineAmount:     l.Amount,
				CreatedAt:      time.Now(),
			}
			if err := s.lineRepo.Create(txCtx, line); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *ExpenseServiceImpl) VerifyAndApproveClaim(
	ctx context.Context,
	claimId string,
	reviewerHrId string,
) (*domain.ExpenseClaim, error) {
	claim, err := s.claimRepo.GetByID(ctx, claimId)
	if err != nil {
		return nil, err
	}

	if claim.Status != domain.ExpenseStatusSUBMITTED {
		return nil, errors.New("only SUBMITTED claims can be approved")
	}

	// Verify reviewer
	_, err = s.empRepo.GetByID(ctx, reviewerHrId)
	if err != nil {
		return nil, fmt.Errorf("reviewer employee not found: %w", err)
	}

	claim.Status = domain.ExpenseStatusAPPROVED
	claim.UpdatedAt = time.Now()

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		if err := s.claimRepo.Update(txCtx, claim); err != nil {
			return err
		}

		evtPayload := domain.ExpenseApprovedEvent{
			EventID:       utils.NewID("evt"),
			LegalEntityID: claim.LegalEntityID,
			ClaimID:       claim.ID,
			EmployeeID:    claim.EmployeeID,
			TotalAmount:   claim.TotalAmount,
			CostCenter:    claim.CostCenterTag,
			Timestamp:     time.Now(),
		}

		outbox := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   "hr.expense.approved",
			AggregateID: claim.ID,
			Payload:     evtPayload,
			Status:      domain.OutboxStatusPENDING,
			CreatedAt:   time.Now(),
		}

		return s.outboxRepo.Create(txCtx, outbox)
	})

	if err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *ExpenseServiceImpl) ClearClaimForPayment(
	ctx context.Context,
	claimId string,
) error {
	claim, err := s.claimRepo.GetByID(ctx, claimId)
	if err != nil {
		return err
	}

	claim.Status = domain.ExpenseStatusPAID
	claim.UpdatedAt = time.Now()

	return s.claimRepo.Update(ctx, claim)
}

// ==========================================
// OutboxRelayWorker Interface & Implementation
// ==========================================

type OutboxRelayWorker interface {
	GetUnsentMessages(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error)
	LogProcessingAttempt(ctx context.Context, outboxID string, currentRetries int, errorNotes string) error
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

func (s *OutboxRelayWorkerImpl) LogProcessingAttempt(ctx context.Context, outboxID string, currentRetries int, errorNotes string) error {
	msg, err := s.outboxRepo.GetByID(ctx, outboxID)
	if err != nil {
		return err
	}
	if currentRetries >= 5 {
		msg.Status = domain.OutboxStatusFAILED
	}
	return s.outboxRepo.Update(ctx, msg)
}

func (s *OutboxRelayWorkerImpl) UpdateOutboxStatus(ctx context.Context, outboxID string, status domain.OutboxStatus) error {
	msg, err := s.outboxRepo.GetByID(ctx, outboxID)
	if err != nil {
		return err
	}
	msg.Status = status
	return s.outboxRepo.Update(ctx, msg)
}

// ==========================================
// ReliableMessagingService Interface & Implementation
// ==========================================

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
