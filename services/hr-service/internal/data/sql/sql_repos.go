package sql

import (
	"context"
	"errors"

	"github.com/erp-system/hr-service/internal/business/domain"
	"gorm.io/gorm"
)

// ==========================================
// Department Repository
// ==========================================

type SQLDepartmentRepository struct {
	db *gorm.DB
}

func NewSQLDepartmentRepository(db *gorm.DB) domain.DepartmentRepository {
	return &SQLDepartmentRepository{db: db}
}

func (r *SQLDepartmentRepository) Create(ctx context.Context, dept *domain.Department) error {
	db := GetDB(ctx, r.db)
	entity := FromDepartmentDomain(dept)
	return db.Create(entity).Error
}

func (r *SQLDepartmentRepository) GetByID(ctx context.Context, id string) (*domain.Department, error) {
	db := GetDB(ctx, r.db)
	var entity Department
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("department not found")
		}
		return nil, err
	}
	return ToDepartmentDomain(&entity), nil
}

func (r *SQLDepartmentRepository) GetByCode(ctx context.Context, legalEntityID, code string) (*domain.Department, error) {
	db := GetDB(ctx, r.db)
	var entity Department
	err := db.First(&entity, "legal_entity_id = ? AND department_code = ?", legalEntityID, code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("department not found by code")
		}
		return nil, err
	}
	return ToDepartmentDomain(&entity), nil
}

func (r *SQLDepartmentRepository) List(ctx context.Context) ([]domain.Department, error) {
	db := GetDB(ctx, r.db)
	var entities []Department
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Department, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToDepartmentDomain(&e))
	}
	return list, nil
}

func (r *SQLDepartmentRepository) Update(ctx context.Context, dept *domain.Department) error {
	db := GetDB(ctx, r.db)
	entity := FromDepartmentDomain(dept)
	return db.Save(entity).Error
}

// ==========================================
// EmployeeMaster Repository
// ==========================================

type SQLEmployeeMasterRepository struct {
	db *gorm.DB
}

func NewSQLEmployeeMasterRepository(db *gorm.DB) domain.EmployeeMasterRepository {
	return &SQLEmployeeMasterRepository{db: db}
}

func (r *SQLEmployeeMasterRepository) Create(ctx context.Context, emp *domain.EmployeeMaster) error {
	db := GetDB(ctx, r.db)
	entity := FromEmployeeMasterDomain(emp)
	return db.Create(entity).Error
}

func (r *SQLEmployeeMasterRepository) GetByID(ctx context.Context, id string) (*domain.EmployeeMaster, error) {
	db := GetDB(ctx, r.db)
	var entity EmployeeMaster
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found")
		}
		return nil, err
	}
	return ToEmployeeMasterDomain(&entity), nil
}

func (r *SQLEmployeeMasterRepository) GetByNumber(ctx context.Context, legalEntityID, number string) (*domain.EmployeeMaster, error) {
	db := GetDB(ctx, r.db)
	var entity EmployeeMaster
	err := db.First(&entity, "legal_entity_id = ? AND employee_number = ?", legalEntityID, number).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found by number")
		}
		return nil, err
	}
	return ToEmployeeMasterDomain(&entity), nil
}

func (r *SQLEmployeeMasterRepository) GetByEmail(ctx context.Context, email string) (*domain.EmployeeMaster, error) {
	db := GetDB(ctx, r.db)
	var entity EmployeeMaster
	err := db.First(&entity, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found by email")
		}
		return nil, err
	}
	return ToEmployeeMasterDomain(&entity), nil
}

func (r *SQLEmployeeMasterRepository) List(ctx context.Context) ([]domain.EmployeeMaster, error) {
	db := GetDB(ctx, r.db)
	var entities []EmployeeMaster
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.EmployeeMaster, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToEmployeeMasterDomain(&e))
	}
	return list, nil
}

func (r *SQLEmployeeMasterRepository) Update(ctx context.Context, emp *domain.EmployeeMaster) error {
	db := GetDB(ctx, r.db)
	entity := FromEmployeeMasterDomain(emp)
	return db.Save(entity).Error
}

func (r *SQLEmployeeMasterRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&EmployeeMaster{}, "id = ?", id).Error
}

// ==========================================
// PayrollRun Repository
// ==========================================

type SQLPayrollRunRepository struct {
	db *gorm.DB
}

func NewSQLPayrollRunRepository(db *gorm.DB) domain.PayrollRunRepository {
	return &SQLPayrollRunRepository{db: db}
}

func (r *SQLPayrollRunRepository) Create(ctx context.Context, run *domain.PayrollRun) error {
	db := GetDB(ctx, r.db)
	entity := FromPayrollRunDomain(run)
	return db.Create(entity).Error
}

func (r *SQLPayrollRunRepository) GetByID(ctx context.Context, id string) (*domain.PayrollRun, error) {
	db := GetDB(ctx, r.db)
	var entity PayrollRun
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payroll run not found")
		}
		return nil, err
	}
	return ToPayrollRunDomain(&entity), nil
}

func (r *SQLPayrollRunRepository) GetByPeriod(ctx context.Context, legalEntityID string, year, period int) (*domain.PayrollRun, error) {
	db := GetDB(ctx, r.db)
	var entity PayrollRun
	err := db.First(&entity, "legal_entity_id = ? AND fiscal_year = ? AND period_number = ?", legalEntityID, year, period).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payroll run not found by period")
		}
		return nil, err
	}
	return ToPayrollRunDomain(&entity), nil
}

func (r *SQLPayrollRunRepository) List(ctx context.Context) ([]domain.PayrollRun, error) {
	db := GetDB(ctx, r.db)
	var entities []PayrollRun
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.PayrollRun, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToPayrollRunDomain(&e))
	}
	return list, nil
}

func (r *SQLPayrollRunRepository) Update(ctx context.Context, run *domain.PayrollRun) error {
	db := GetDB(ctx, r.db)
	entity := FromPayrollRunDomain(run)
	return db.Save(entity).Error
}

// ==========================================
// ExpenseClaim Repository
// ==========================================

type SQLExpenseClaimRepository struct {
	db *gorm.DB
}

func NewSQLExpenseClaimRepository(db *gorm.DB) domain.ExpenseClaimRepository {
	return &SQLExpenseClaimRepository{db: db}
}

func (r *SQLExpenseClaimRepository) Create(ctx context.Context, claim *domain.ExpenseClaim) error {
	db := GetDB(ctx, r.db)
	entity := FromExpenseClaimDomain(claim)
	return db.Create(entity).Error
}

func (r *SQLExpenseClaimRepository) GetByID(ctx context.Context, id string) (*domain.ExpenseClaim, error) {
	db := GetDB(ctx, r.db)
	var entity ExpenseClaim
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense claim not found")
		}
		return nil, err
	}
	return ToExpenseClaimDomain(&entity), nil
}

func (r *SQLExpenseClaimRepository) GetByNumber(ctx context.Context, legalEntityID, number string) (*domain.ExpenseClaim, error) {
	db := GetDB(ctx, r.db)
	var entity ExpenseClaim
	err := db.First(&entity, "legal_entity_id = ? AND claim_number = ?", legalEntityID, number).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense claim not found by number")
		}
		return nil, err
	}
	return ToExpenseClaimDomain(&entity), nil
}

func (r *SQLExpenseClaimRepository) List(ctx context.Context) ([]domain.ExpenseClaim, error) {
	db := GetDB(ctx, r.db)
	var entities []ExpenseClaim
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.ExpenseClaim, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToExpenseClaimDomain(&e))
	}
	return list, nil
}

func (r *SQLExpenseClaimRepository) Update(ctx context.Context, claim *domain.ExpenseClaim) error {
	db := GetDB(ctx, r.db)
	entity := FromExpenseClaimDomain(claim)
	return db.Save(entity).Error
}

// ==========================================
// ExpenseClaimLine Repository
// ==========================================

type SQLExpenseClaimLineRepository struct {
	db *gorm.DB
}

func NewSQLExpenseClaimLineRepository(db *gorm.DB) domain.ExpenseClaimLineRepository {
	return &SQLExpenseClaimLineRepository{db: db}
}

func (r *SQLExpenseClaimLineRepository) Create(ctx context.Context, line *domain.ExpenseClaimLine) error {
	db := GetDB(ctx, r.db)
	entity := FromExpenseClaimLineDomain(line)
	return db.Create(entity).Error
}

func (r *SQLExpenseClaimLineRepository) GetByID(ctx context.Context, id string) (*domain.ExpenseClaimLine, error) {
	db := GetDB(ctx, r.db)
	var entity ExpenseClaimLine
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense claim line not found")
		}
		return nil, err
	}
	return ToExpenseClaimLineDomain(&entity), nil
}

func (r *SQLExpenseClaimLineRepository) ListByClaimID(ctx context.Context, claimID string) ([]domain.ExpenseClaimLine, error) {
	db := GetDB(ctx, r.db)
	var entities []ExpenseClaimLine
	err := db.Find(&entities, "expense_claim_id = ?", claimID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.ExpenseClaimLine, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToExpenseClaimLineDomain(&e))
	}
	return list, nil
}

// ==========================================
// Transactional Outbox Repository
// ==========================================

type SQLTransactionalOutboxRepository struct {
	db *gorm.DB
}

func NewSQLTransactionalOutboxRepository(db *gorm.DB) domain.TransactionalOutboxRepository {
	return &SQLTransactionalOutboxRepository{db: db}
}

func (r *SQLTransactionalOutboxRepository) Create(ctx context.Context, msg *domain.TransactionalOutbox) error {
	db := GetDB(ctx, r.db)
	entity := FromTransactionalOutboxDomain(msg)
	return db.Create(entity).Error
}

func (r *SQLTransactionalOutboxRepository) GetByID(ctx context.Context, id string) (*domain.TransactionalOutbox, error) {
	db := GetDB(ctx, r.db)
	var entity TransactionalOutbox
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("outbox message not found")
		}
		return nil, err
	}
	return ToTransactionalOutboxDomain(&entity), nil
}

func (r *SQLTransactionalOutboxRepository) GetUnsent(ctx context.Context, limit int) ([]domain.TransactionalOutbox, error) {
	db := GetDB(ctx, r.db)
	var entities []TransactionalOutbox
	err := db.Where("status = ?", string(domain.OutboxStatusPENDING)).Order("created_at asc").Limit(limit).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.TransactionalOutbox, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToTransactionalOutboxDomain(&e))
	}
	return list, nil
}

func (r *SQLTransactionalOutboxRepository) Update(ctx context.Context, msg *domain.TransactionalOutbox) error {
	db := GetDB(ctx, r.db)
	entity := FromTransactionalOutboxDomain(msg)
	return db.Save(entity).Error
}

// ==========================================
// Kafka Event Inbox Repository
// ==========================================

type SQLKafkaEventInboxRepository struct {
	db *gorm.DB
}

func NewSQLKafkaEventInboxRepository(db *gorm.DB) domain.KafkaEventInboxRepository {
	return &SQLKafkaEventInboxRepository{db: db}
}

func (r *SQLKafkaEventInboxRepository) Create(ctx context.Context, msg *domain.KafkaEventInbox) error {
	db := GetDB(ctx, r.db)
	entity := FromKafkaEventInboxDomain(msg)
	return db.Create(entity).Error
}

func (r *SQLKafkaEventInboxRepository) GetByID(ctx context.Context, eventID string) (*domain.KafkaEventInbox, error) {
	db := GetDB(ctx, r.db)
	var entity KafkaEventInbox
	err := db.First(&entity, "event_id = ?", eventID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("inbox message not found")
		}
		return nil, err
	}
	return ToKafkaEventInboxDomain(&entity), nil
}

func (r *SQLKafkaEventInboxRepository) Update(ctx context.Context, msg *domain.KafkaEventInbox) error {
	db := GetDB(ctx, r.db)
	entity := FromKafkaEventInboxDomain(msg)
	return db.Save(entity).Error
}
