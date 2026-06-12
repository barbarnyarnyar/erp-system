package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type PayrollService struct {
	repo       domain.PayrollRecordRepository
	deductions domain.PayrollDeductionRepository
	employees  domain.EmployeeRepository
	publisher  domain.EventPublisher
}

func NewPayrollService(repo domain.PayrollRecordRepository, deductions domain.PayrollDeductionRepository, employees domain.EmployeeRepository, publisher domain.EventPublisher) *PayrollService {
	return &PayrollService{
		repo:       repo,
		deductions: deductions,
		employees:  employees,
		publisher:  publisher,
	}
}

func (s *PayrollService) ListPayrollRecords(ctx context.Context) ([]domain.PayrollRecord, error) {
	return s.repo.List(ctx)
}

func (s *PayrollService) ProcessPayroll(ctx context.Context, employeeID string, start, end time.Time, regularHours, overtimeHours decimal.Decimal) (*domain.PayrollRecord, error) {
	emp, err := s.employees.GetByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// Calculate gross pay (Hourly rate = salary / 160 hours per month. Overtime = Hourly * 1.5)
	hourlyRate := emp.Salary.Div(decimal.NewFromInt(160))
	regularPay := hourlyRate.Mul(regularHours)
	overtimeRate := hourlyRate.Mul(decimal.NewFromFloat(1.5))
	overtimePay := overtimeRate.Mul(overtimeHours)
	grossPay := regularPay.Add(overtimePay)

	// Deductions (15% Income Tax, 5% Social Security)
	incomeTaxAmt := grossPay.Mul(decimal.NewFromFloat(0.15))
	socialSecurityAmt := grossPay.Mul(decimal.NewFromFloat(0.05))
	totalDeductions := incomeTaxAmt.Add(socialSecurityAmt)
	netPay := grossPay.Sub(totalDeductions)

	id := utils.NewID("pay")

	pr := &domain.PayrollRecord{
		ID:             id,
		EmployeeID:     employeeID,
		PayPeriodStart: start,
		PayPeriodEnd:   end,
		RegularHours:   regularHours,
		OvertimeHours:  overtimeHours,
		GrossPay:       grossPay,
		NetPay:         netPay,
		Status:         "PAID",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = s.repo.Create(ctx, pr)
	if err != nil {
		return nil, err
	}

	// Create deductions breakdown
	ded1 := &domain.PayrollDeduction{
		ID:        utils.NewID("ded-tax"),
		PayrollID: pr.ID,
		Type:      "Income Tax",
		Amount:    incomeTaxAmt,
	}
	_ = s.deductions.Create(ctx, ded1)

	ded2 := &domain.PayrollDeduction{
		ID:        utils.NewID("ded-ss"),
		PayrollID: pr.ID,
		Type:      "Social Security",
		Amount:    socialSecurityAmt,
	}
	_ = s.deductions.Create(ctx, ded2)

	// Publish payroll processed event to Kafka
	if err := s.publisher.Publish(ctx, domain.TopicHrPayrollProcessed, pr.ID, domain.PayrollProcessedEvent{
		PayrollID:   pr.ID,
		PeriodStart: pr.PayPeriodStart,
		PeriodEnd:   pr.PayPeriodEnd,
		TotalGross:  pr.GrossPay,
		TotalNet:    pr.NetPay,
		Timestamp:   time.Now(),
	}); err != nil {
		utils.LogPublishErr("hr-service", domain.TopicHrPayrollProcessed, err)
	}

	return pr, nil
}

func (s *PayrollService) GetPayrollRecord(ctx context.Context, id string) (*domain.PayrollRecord, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PayrollService) UpdatePayrollRecord(ctx context.Context, id, status string) (*domain.PayrollRecord, error) {
	pr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	pr.Status = status
	pr.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, pr)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *PayrollService) GetEmployeePayroll(ctx context.Context, employeeID string) ([]domain.PayrollRecord, error) {
	return s.repo.GetByEmployeeID(ctx, employeeID)
}

func (s *PayrollService) FailPayroll(ctx context.Context, employeeID string, start, end time.Time, reason string) error {
	if err := s.publisher.Publish(ctx, domain.TopicHrPayrollFailed, employeeID, domain.PayrollFailedEvent{
		EmployeeID:  employeeID,
		PeriodStart: start,
		PeriodEnd:   end,
		Reason:      reason,
		Timestamp:   time.Now(),
	}); err != nil {
		utils.LogPublishErr("hr-service", domain.TopicHrPayrollFailed, err)
		return err
	}
	return nil
}
