package sql

import (
	"encoding/json"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 1. Department
type Department struct {
	ID            string    `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_hr_dept_code_tenant"`
	DepartmentCode string   `gorm:"type:varchar(100);not null;uniqueIndex:idx_hr_dept_code_tenant"`
	Name          string    `gorm:"type:varchar(255);not null"`
	IsActive      bool      `gorm:"type:boolean;default:true"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (Department) TableName() string {
	return "hr_departments"
}

func ToDepartmentDomain(d *Department) *domain.Department {
	if d == nil {
		return nil
	}
	return &domain.Department{
		ID:            d.ID,
		LegalEntityID: d.LegalEntityID,
		DepartmentCode: d.DepartmentCode,
		Name:          d.Name,
		IsActive:      d.IsActive,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func FromDepartmentDomain(d *domain.Department) *Department {
	if d == nil {
		return nil
	}
	return &Department{
		ID:            d.ID,
		LegalEntityID: d.LegalEntityID,
		DepartmentCode: d.DepartmentCode,
		Name:          d.Name,
		IsActive:      d.IsActive,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

// 2. EmployeeMaster
type EmployeeMaster struct {
	ID             string         `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID  string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_hr_emp_num_tenant"`
	DepartmentID   string         `gorm:"type:varchar(255);not null"`
	ManagerHrID    *string        `gorm:"type:varchar(255);default:null"`
	OrgDepthLevel  int            `gorm:"type:integer;not null;default:0"`
	EmployeeNumber string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_hr_emp_num_tenant"`
	FirstName      string         `gorm:"type:varchar(255);not null"`
	LastName       string         `gorm:"type:varchar(255);not null"`
	Email          string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Status         string         `gorm:"type:varchar(50);not null"`
	Type           string         `gorm:"type:varchar(50);not null"`
	BaseSalary     decimal.Decimal `gorm:"type:numeric(18,4);not null"`
	Version        int            `gorm:"type:integer;not null;default:1"`
	CreatedAt      time.Time      `gorm:"not null"`
	UpdatedAt      time.Time      `gorm:"not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (EmployeeMaster) TableName() string {
	return "hr_employees"
}

func ToEmployeeMasterDomain(e *EmployeeMaster) *domain.EmployeeMaster {
	if e == nil {
		return nil
	}
	var deletedAt *time.Time
	if e.DeletedAt.Valid {
		deletedAt = &e.DeletedAt.Time
	}
	return &domain.EmployeeMaster{
		ID:             e.ID,
		LegalEntityID:  e.LegalEntityID,
		DepartmentID:   e.DepartmentID,
		ManagerHrID:    e.ManagerHrID,
		OrgDepthLevel:  e.OrgDepthLevel,
		EmployeeNumber: e.EmployeeNumber,
		FirstName:      e.FirstName,
		LastName:       e.LastName,
		Email:          e.Email,
		Status:         domain.EmployeeStatus(e.Status),
		Type:           domain.EmploymentType(e.Type),
		BaseSalary:     e.BaseSalary,
		Version:        e.Version,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

func FromEmployeeMasterDomain(e *domain.EmployeeMaster) *EmployeeMaster {
	if e == nil {
		return nil
	}
	var deletedAt gorm.DeletedAt
	if e.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *e.DeletedAt, Valid: true}
	}
	return &EmployeeMaster{
		ID:             e.ID,
		LegalEntityID:  e.LegalEntityID,
		DepartmentID:   e.DepartmentID,
		ManagerHrID:    e.ManagerHrID,
		OrgDepthLevel:  e.OrgDepthLevel,
		EmployeeNumber: e.EmployeeNumber,
		FirstName:      e.FirstName,
		LastName:       e.LastName,
		Email:          e.Email,
		Status:         string(e.Status),
		Type:           string(e.Type),
		BaseSalary:     e.BaseSalary,
		Version:        e.Version,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

// 3. PayrollRun
type PayrollRun struct {
	ID              string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID   string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_hr_payroll_period"`
	FiscalYear      int             `gorm:"type:integer;not null;uniqueIndex:idx_hr_payroll_period"`
	PeriodNumber    int             `gorm:"type:integer;not null;uniqueIndex:idx_hr_payroll_period"`
	Status          string          `gorm:"type:varchar(50);not null"`
	TotalGrossPay   decimal.Decimal `gorm:"type:numeric(18,4);not null;default:0"`
	TotalDeductions decimal.Decimal `gorm:"type:numeric(18,4);not null;default:0"`
	TotalNetPay     decimal.Decimal `gorm:"type:numeric(18,4);not null;default:0"`
	Version         int             `gorm:"type:integer;not null;default:1"`
	CreatedAt       time.Time       `gorm:"not null"`
	UpdatedAt       time.Time       `gorm:"not null"`
}

func (PayrollRun) TableName() string {
	return "hr_payroll_runs"
}

func ToPayrollRunDomain(p *PayrollRun) *domain.PayrollRun {
	if p == nil {
		return nil
	}
	return &domain.PayrollRun{
		ID:              p.ID,
		LegalEntityID:   p.LegalEntityID,
		FiscalYear:      p.FiscalYear,
		PeriodNumber:    p.PeriodNumber,
		Status:          domain.PayrollStatus(p.Status),
		TotalGrossPay:   p.TotalGrossPay,
		TotalDeductions: p.TotalDeductions,
		TotalNetPay:     p.TotalNetPay,
		Version:         p.Version,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func FromPayrollRunDomain(p *domain.PayrollRun) *PayrollRun {
	if p == nil {
		return nil
	}
	return &PayrollRun{
		ID:              p.ID,
		LegalEntityID:   p.LegalEntityID,
		FiscalYear:      p.FiscalYear,
		PeriodNumber:    p.PeriodNumber,
		Status:          string(p.Status),
		TotalGrossPay:   p.TotalGrossPay,
		TotalDeductions: p.TotalDeductions,
		TotalNetPay:     p.TotalNetPay,
		Version:         p.Version,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// 4. ExpenseClaim
type ExpenseClaim struct {
	ID            string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_hr_claim_num_tenant"`
	EmployeeID    string          `gorm:"type:varchar(255);not null"`
	ClaimNumber   string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_hr_claim_num_tenant"`
	Purpose       string          `gorm:"type:varchar(255);not null"`
	TotalAmount   decimal.Decimal `gorm:"type:numeric(18,4);not null;default:0"`
	Status        string          `gorm:"type:varchar(50);not null"`
	CostCenterTag string          `gorm:"type:varchar(100);not null"`
	Version       int             `gorm:"type:integer;not null;default:1"`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`
}

func (ExpenseClaim) TableName() string {
	return "hr_expense_claims"
}

func ToExpenseClaimDomain(e *ExpenseClaim) *domain.ExpenseClaim {
	if e == nil {
		return nil
	}
	return &domain.ExpenseClaim{
		ID:            e.ID,
		LegalEntityID: e.LegalEntityID,
		EmployeeID:    e.EmployeeID,
		ClaimNumber:   e.ClaimNumber,
		Purpose:       e.Purpose,
		TotalAmount:   e.TotalAmount,
		Status:        domain.ExpenseStatus(e.Status),
		CostCenterTag: e.CostCenterTag,
		Version:       e.Version,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func FromExpenseClaimDomain(e *domain.ExpenseClaim) *ExpenseClaim {
	if e == nil {
		return nil
	}
	return &ExpenseClaim{
		ID:            e.ID,
		LegalEntityID: e.LegalEntityID,
		EmployeeID:    e.EmployeeID,
		ClaimNumber:   e.ClaimNumber,
		Purpose:       e.Purpose,
		TotalAmount:   e.TotalAmount,
		Status:        string(e.Status),
		CostCenterTag: e.CostCenterTag,
		Version:       e.Version,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

// 5. ExpenseClaimLine
type ExpenseClaimLine struct {
	ID             string          `gorm:"primaryKey;type:varchar(255)"`
	ExpenseClaimID string          `gorm:"type:varchar(255);not null;index"`
	Description    string          `gorm:"type:varchar(255);not null"`
	LineAmount     decimal.Decimal `gorm:"type:numeric(18,4);not null"`
	CreatedAt      time.Time       `gorm:"not null"`
}

func (ExpenseClaimLine) TableName() string {
	return "hr_expense_claim_lines"
}

func ToExpenseClaimLineDomain(e *ExpenseClaimLine) *domain.ExpenseClaimLine {
	if e == nil {
		return nil
	}
	return &domain.ExpenseClaimLine{
		ID:             e.ID,
		ExpenseClaimID: e.ExpenseClaimID,
		Description:    e.Description,
		LineAmount:     e.LineAmount,
		CreatedAt:      e.CreatedAt,
	}
}

func FromExpenseClaimLineDomain(e *domain.ExpenseClaimLine) *ExpenseClaimLine {
	if e == nil {
		return nil
	}
	return &ExpenseClaimLine{
		ID:             e.ID,
		ExpenseClaimID: e.ExpenseClaimID,
		Description:    e.Description,
		LineAmount:     e.LineAmount,
		CreatedAt:      e.CreatedAt,
	}
}

// 6. TransactionalOutbox
type TransactionalOutbox struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType   string    `gorm:"type:varchar(255);not null"`
	AggregateID string    `gorm:"type:varchar(255);not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Status      string    `gorm:"type:varchar(50);not null;index:idx_hr_outbox_status_date"`
	CreatedAt   time.Time `gorm:"not null;index:idx_hr_outbox_status_date"`
}

func (TransactionalOutbox) TableName() string {
	return "hr_transactional_outbox"
}

func ToTransactionalOutboxDomain(o *TransactionalOutbox) *domain.TransactionalOutbox {
	if o == nil {
		return nil
	}
	var payload interface{}
	if len(o.Payload) > 0 {
		_ = json.Unmarshal(o.Payload, &payload)
	}
	return &domain.TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     payload,
		Status:      domain.OutboxStatus(o.Status),
		CreatedAt:   o.CreatedAt,
	}
}

func FromTransactionalOutboxDomain(o *domain.TransactionalOutbox) *TransactionalOutbox {
	if o == nil {
		return nil
	}
	payloadBytes, _ := json.Marshal(o.Payload)
	return &TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     payloadBytes,
		Status:      string(o.Status),
		CreatedAt:   o.CreatedAt,
	}
}

// 7. KafkaEventInbox
type KafkaEventInbox struct {
	EventID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType        string    `gorm:"type:varchar(255);not null"`
	ProcessedAt      time.Time `gorm:"not null"`
	ProcessingStatus string    `gorm:"type:varchar(50);not null"`
	Payload          []byte    `gorm:"type:jsonb;not null"`
}

func (KafkaEventInbox) TableName() string {
	return "hr_kafka_event_inbox"
}

func ToKafkaEventInboxDomain(i *KafkaEventInbox) *domain.KafkaEventInbox {
	if i == nil {
		return nil
	}
	var payload interface{}
	if len(i.Payload) > 0 {
		_ = json.Unmarshal(i.Payload, &payload)
	}
	return &domain.KafkaEventInbox{
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: domain.EventProcessingStatus(i.ProcessingStatus),
		Payload:          payload,
	}
}

func FromKafkaEventInboxDomain(i *domain.KafkaEventInbox) *KafkaEventInbox {
	if i == nil {
		return nil
	}
	payloadBytes, _ := json.Marshal(i.Payload)
	return &KafkaEventInbox{
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: string(i.ProcessingStatus),
		Payload:          payloadBytes,
	}
}
