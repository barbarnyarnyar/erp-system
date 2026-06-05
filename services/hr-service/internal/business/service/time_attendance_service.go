package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type TimeAttendanceService struct {
	repo      domain.TimeEntryRepository
	publisher domain.EventPublisher
}

func NewTimeAttendanceService(repo domain.TimeEntryRepository, publisher domain.EventPublisher) *TimeAttendanceService {
	return &TimeAttendanceService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *TimeAttendanceService) ListTimesheets(ctx context.Context) ([]domain.TimeEntry, error) {
	return s.repo.List(ctx)
}

func (s *TimeAttendanceService) CreateTimesheet(ctx context.Context, employeeID string, entryDate time.Time, clockIn, clockOut time.Time, notes string) (*domain.TimeEntry, error) {
	id := fmt.Sprintf("te_%d", time.Now().UnixNano())

	// Calculate total hours
	diff := clockOut.Sub(clockIn)
	hours := decimal.NewFromFloat(diff.Hours())

	te := &domain.TimeEntry{
		ID:          id,
		EmployeeID:  employeeID,
		EntryDate:   entryDate,
		ClockIn:     clockIn,
		ClockOut:    clockOut,
		TotalHours:  hours,
		Notes:       notes,
		Status:      "SUBMITTED",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.repo.Create(ctx, te)
	if err != nil {
		return nil, err
	}

	// Calculate and publish overtime if hours > 8
	if hours.GreaterThan(decimal.NewFromInt(8)) {
		otHours := hours.Sub(decimal.NewFromInt(8))
		_ = s.publisher.Publish(ctx, domain.TopicHrOvertimeRecorded, te.EmployeeID, domain.OvertimeRecordedEvent{
			EmployeeID:    te.EmployeeID,
			EntryDate:     te.EntryDate,
			OvertimeHours: otHours,
			Timestamp:     time.Now(),
		})
	}

	return te, nil
}

func (s *TimeAttendanceService) GetTimesheet(ctx context.Context, id string) (*domain.TimeEntry, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TimeAttendanceService) UpdateTimesheet(ctx context.Context, id string, clockIn, clockOut time.Time, notes string) (*domain.TimeEntry, error) {
	te, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	diff := clockOut.Sub(clockIn)
	hours := decimal.NewFromFloat(diff.Hours())

	te.ClockIn = clockIn
	te.ClockOut = clockOut
	te.TotalHours = hours
	te.Notes = notes
	te.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, te)
	if err != nil {
		return nil, err
	}

	return te, nil
}

func (s *TimeAttendanceService) SubmitTimesheet(ctx context.Context, id string) (*domain.TimeEntry, error) {
	te, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	te.Status = "SUBMITTED"
	te.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, te)
	if err != nil {
		return nil, err
	}

	// Publish timesheet submitted event
	_ = s.publisher.Publish(ctx, domain.TopicHrTimesheetSubmitted, te.ID, domain.TimesheetSubmittedEvent{
		TimesheetID: te.ID,
		EmployeeID:  te.EmployeeID,
		EntryDate:   te.EntryDate,
		TotalHours:  te.TotalHours,
		Timestamp:   time.Now(),
	})

	return te, nil
}

func (s *TimeAttendanceService) ApproveTimesheet(ctx context.Context, id string, approvedBy string) (*domain.TimeEntry, error) {
	te, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	te.Status = "APPROVED"
	te.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, te)
	if err != nil {
		return nil, err
	}

	// Publish timesheet approved event
	_ = s.publisher.Publish(ctx, domain.TopicHrTimesheetApproved, te.ID, domain.TimesheetApprovedEvent{
		TimesheetID: te.ID,
		EmployeeID:  te.EmployeeID,
		ApprovedBy:  approvedBy,
		Timestamp:   time.Now(),
	})

	return te, nil
}
