package service

import (
	"context"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type ReportService struct {
	empRepo       domain.EmployeeRepository
	payrollRepo   domain.PayrollRecordRepository
	timesheetRepo domain.AttendanceEntryRepository
}

func NewReportService(
	empRepo domain.EmployeeRepository,
	payrollRepo domain.PayrollRecordRepository,
	timesheetRepo domain.AttendanceEntryRepository,
) *ReportService {
	return &ReportService{
		empRepo:       empRepo,
		payrollRepo:   payrollRepo,
		timesheetRepo: timesheetRepo,
	}
}

type HeadcountReport struct {
	TotalEmployees    int            `json:"total_employees"`
	ActiveEmployees   int            `json:"active_employees"`
	TerminatedCount   int            `json:"terminated_employees"`
	ByDepartmentCount map[string]int `json:"by_department"`
}

func (s *ReportService) GetHeadcountReport(ctx context.Context) (*HeadcountReport, error) {
	list, err := s.empRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	report := &HeadcountReport{
		ByDepartmentCount: make(map[string]int),
	}

	for _, emp := range list {
		report.TotalEmployees++
		if emp.Status == "ACTIVE" {
			report.ActiveEmployees++
		} else if emp.Status == "TERMINATED" {
			report.TerminatedCount++
		}

		if emp.DepartmentID != "" {
			report.ByDepartmentCount[emp.DepartmentID]++
		}
	}

	return report, nil
}

type PayrollReport struct {
	TotalGrossPay      decimal.Decimal `json:"total_gross_pay"`
	TotalNetPay        decimal.Decimal `json:"total_net_pay"`
	TotalDeductions    decimal.Decimal `json:"total_deductions"`
	TotalRegularHours  decimal.Decimal `json:"total_regular_hours"`
	TotalOvertimeHours decimal.Decimal `json:"total_overtime_hours"`
	PayrollRunCount    int             `json:"payroll_run_count"`
}

func (s *ReportService) GetPayrollReport(ctx context.Context) (*PayrollReport, error) {
	list, err := s.payrollRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	report := &PayrollReport{
		TotalGrossPay:      decimal.Zero,
		TotalNetPay:        decimal.Zero,
		TotalDeductions:    decimal.Zero,
		TotalRegularHours:  decimal.Zero,
		TotalOvertimeHours: decimal.Zero,
	}

	for _, pr := range list {
		report.PayrollRunCount++
		report.TotalGrossPay = report.TotalGrossPay.Add(pr.GrossPay)
		report.TotalNetPay = report.TotalNetPay.Add(pr.NetPay)
		report.TotalRegularHours = report.TotalRegularHours.Add(pr.RegularHours)
		report.TotalOvertimeHours = report.TotalOvertimeHours.Add(pr.OvertimeHours)
	}

	report.TotalDeductions = report.TotalGrossPay.Sub(report.TotalNetPay)

	return report, nil
}

type AttendanceReport struct {
	TotalHoursLogged decimal.Decimal `json:"total_hours_logged"`
	TotalShiftsCount int             `json:"total_shifts_count"`
	AverageShiftHours decimal.Decimal `json:"average_shift_hours"`
	LateCheckinsCount int             `json:"late_checkins_count"`
}

func (s *ReportService) GetAttendanceReport(ctx context.Context) (*AttendanceReport, error) {
	list, err := s.timesheetRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	report := &AttendanceReport{
		TotalHoursLogged: decimal.Zero,
		AverageShiftHours: decimal.Zero,
	}

	for _, te := range list {
		report.TotalShiftsCount++
		report.TotalHoursLogged = report.TotalHoursLogged.Add(te.TotalHours)

		// Check if check-in was late (e.g. check in after 9:00 AM local time of punch date)
		if te.ClockIn.Hour() > 9 || (te.ClockIn.Hour() == 9 && te.ClockIn.Minute() > 0) {
			report.LateCheckinsCount++
		}
	}

	if report.TotalShiftsCount > 0 {
		report.AverageShiftHours = report.TotalHoursLogged.Div(decimal.NewFromInt(int64(report.TotalShiftsCount)))
	}

	return report, nil
}
