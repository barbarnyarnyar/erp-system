package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type EmployeeManagementService struct {
	repo      domain.EmployeeRepository
	publisher domain.EventPublisher
}

func NewEmployeeManagementService(repo domain.EmployeeRepository, publisher domain.EventPublisher) *EmployeeManagementService {
	return &EmployeeManagementService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *EmployeeManagementService) ListEmployees(ctx context.Context) ([]domain.Employee, error) {
	return s.repo.List(ctx)
}

func (s *EmployeeManagementService) CreateEmployee(ctx context.Context, firstName, lastName, email, deptID, posID string, salary decimal.Decimal) (*domain.Employee, error) {
	if firstName == "" || lastName == "" || email == "" {
		return nil, errors.New("first name, last name, and email are required")
	}

	id := fmt.Sprintf("emp_%d", time.Now().UnixNano())
	empCode := fmt.Sprintf("EMP-%d", time.Now().Unix())

	emp := &domain.Employee{
		ID:           id,
		EmployeeID:   empCode,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		HireDate:     time.Now(),
		DepartmentID: deptID,
		PositionID:   posID,
		Status:       "ACTIVE",
		Salary:       salary,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.Create(ctx, emp)
	if err != nil {
		return nil, err
	}

	// Publish employee created event to Kafka
	_ = s.publisher.Publish(ctx, domain.TopicHrEmployeeCreated, emp.ID, domain.EmployeeCreatedEvent{
		EmployeeID:   emp.ID,
		FirstName:    emp.FirstName,
		LastName:     emp.LastName,
		DepartmentID: emp.DepartmentID,
		Salary:       emp.Salary,
		Timestamp:    time.Now(),
	})

	return emp, nil
}

func (s *EmployeeManagementService) GetEmployee(ctx context.Context, id string) (*domain.Employee, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EmployeeManagementService) UpdateEmployee(ctx context.Context, id, firstName, lastName, email, deptID, posID string, salary decimal.Decimal, status string) (*domain.Employee, error) {
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldSalary := emp.Salary
	salaryChanged := !oldSalary.Equal(salary)
	oldPositionID := emp.PositionID
	positionChanged := oldPositionID != posID

	emp.FirstName = firstName
	emp.LastName = lastName
	emp.Email = email
	emp.DepartmentID = deptID
	emp.PositionID = posID
	emp.Salary = salary
	emp.Status = status
	emp.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, emp)
	if err != nil {
		return nil, err
	}

	// Publish employee updated event
	_ = s.publisher.Publish(ctx, domain.TopicHrEmployeeUpdated, emp.ID, domain.EmployeeUpdatedEvent{
		EmployeeID:   emp.ID,
		FirstName:    emp.FirstName,
		LastName:     emp.LastName,
		DepartmentID: emp.DepartmentID,
		PositionID:   emp.PositionID,
		Salary:       emp.Salary,
		Status:       emp.Status,
		Timestamp:    time.Now(),
	})

	// Publish salary changed event if different
	if salaryChanged {
		_ = s.publisher.Publish(ctx, domain.TopicHrSalaryChanged, emp.ID, domain.SalaryChangedEvent{
			EmployeeID: emp.ID,
			OldSalary:  oldSalary,
			NewSalary:  emp.Salary,
			Timestamp:  time.Now(),
		})
	}

	// Publish employee promoted event if position changed
	if positionChanged {
		_ = s.publisher.Publish(ctx, domain.TopicHrEmployeePromoted, emp.ID, domain.EmployeePromotedEvent{
			EmployeeID:    emp.ID,
			OldPositionID: oldPositionID,
			NewPositionID: emp.PositionID,
			NewSalary:     emp.Salary,
			Timestamp:     time.Now(),
		})
	}

	return emp, nil
}

func (s *EmployeeManagementService) DeleteEmployee(ctx context.Context, id string) error {
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Publish employee terminated event
	_ = s.publisher.Publish(ctx, domain.TopicHrEmployeeTerminated, emp.ID, domain.EmployeeTerminatedEvent{
		EmployeeID: emp.ID,
		TermDate:   time.Now(),
		Reason:     "Deleted / Terminated from management system",
		Timestamp:  time.Now(),
	})

	return nil
}
