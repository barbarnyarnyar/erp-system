package service

import (
	"log"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type EmployeeManagementService struct {
	repo        domain.EmployeeRepository
	claims      domain.ExpenseClaimRepository
	claimLines  domain.ExpenseClaimLineRepository
	historyRepo domain.EmployeeCompensationHistoryRepository
	depts       domain.DepartmentRepository
	positions   domain.PositionRepository
	publisher   domain.EventPublisher
}

func NewEmployeeManagementService(
	repo domain.EmployeeRepository,
	claims domain.ExpenseClaimRepository,
	claimLines domain.ExpenseClaimLineRepository,
	historyRepo domain.EmployeeCompensationHistoryRepository,
	depts domain.DepartmentRepository,
	positions domain.PositionRepository,
	publisher domain.EventPublisher,
) *EmployeeManagementService {
	return &EmployeeManagementService{
		repo:        repo,
		claims:      claims,
		claimLines:  claimLines,
		historyRepo: historyRepo,
		depts:       depts,
		positions:   positions,
		publisher:   publisher,
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

	// Record initial compensation history
	echID := fmt.Sprintf("ech_%d", time.Now().UnixNano())
	_ = s.historyRepo.Create(ctx, &domain.EmployeeCompensationHistory{
		ID:            echID,
		EmployeeID:    emp.ID,
		Salary:        salary,
		EffectiveDate: time.Now(),
		ChangedBy:     "system",
		CreatedAt:     time.Now(),
	})

	// Publish employee created event to Kafka
	if err := s.publisher.Publish(ctx, domain.TopicHrEmployeeCreated, emp.ID, domain.EmployeeCreatedEvent{
		EmployeeID:   emp.ID,
		FirstName:    emp.FirstName,
		LastName:     emp.LastName,
		DepartmentID: emp.DepartmentID,
		Salary:       emp.Salary,
		Timestamp:    time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrEmployeeCreated, err)
	}

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
	if err := s.publisher.Publish(ctx, domain.TopicHrEmployeeUpdated, emp.ID, domain.EmployeeUpdatedEvent{
		EmployeeID:   emp.ID,
		FirstName:    emp.FirstName,
		LastName:     emp.LastName,
		DepartmentID: emp.DepartmentID,
		PositionID:   emp.PositionID,
		Salary:       emp.Salary,
		Status:       emp.Status,
		Timestamp:    time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrEmployeeUpdated, err)
	}

	// Publish salary changed event and record compensation history if different
	if salaryChanged {
		echID := fmt.Sprintf("ech_%d", time.Now().UnixNano())
		_ = s.historyRepo.Create(ctx, &domain.EmployeeCompensationHistory{
			ID:            echID,
			EmployeeID:    emp.ID,
			Salary:        salary,
			EffectiveDate: time.Now(),
			ChangedBy:     "system",
			CreatedAt:     time.Now(),
		})

		if err := s.publisher.Publish(ctx, domain.TopicHrSalaryChanged, emp.ID, domain.SalaryChangedEvent{
			EmployeeID: emp.ID,
			OldSalary:  oldSalary,
			NewSalary:  emp.Salary,
			Timestamp:  time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrSalaryChanged, err)
		}
	}

	// Publish employee promoted event if position changed
	if positionChanged {
		if err := s.publisher.Publish(ctx, domain.TopicHrEmployeePromoted, emp.ID, domain.EmployeePromotedEvent{
			EmployeeID:    emp.ID,
			OldPositionID: oldPositionID,
			NewPositionID: emp.PositionID,
			NewSalary:     emp.Salary,
			Timestamp:     time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrEmployeePromoted, err)
		}
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
	if err := s.publisher.Publish(ctx, domain.TopicHrEmployeeTerminated, emp.ID, domain.EmployeeTerminatedEvent{
		EmployeeID: emp.ID,
		TermDate:   time.Now(),
		Reason:     "Deleted / Terminated from management system",
		Timestamp:  time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrEmployeeTerminated, err)
	}

	return nil
}

func (s *EmployeeManagementService) SubmitExpenseClaim(ctx context.Context, employeeID string, claimDate time.Time, lines []domain.ExpenseClaimLine) (*domain.ExpenseClaim, error) {
	claimID := fmt.Sprintf("exp_%d", time.Now().UnixNano())

	var total decimal.Decimal
	for i := range lines {
		lines[i].ID = fmt.Sprintf("expl_%d_%d", time.Now().UnixNano(), i)
		lines[i].ClaimID = claimID
		total = total.Add(lines[i].Amount)
	}

	claim := &domain.ExpenseClaim{
		ID:          claimID,
		EmployeeID:  employeeID,
		ClaimDate:   claimDate,
		Status:      "SUBMITTED",
		TotalAmount: total,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.claims.Create(ctx, claim)
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		err = s.claimLines.Create(ctx, &line)
		if err != nil {
			return nil, err
		}
	}

	// Publish hr.expense.submitted event
	desc := "Expense claim submission"
	if len(lines) > 0 {
		desc = lines[0].Description
	}

	if err := s.publisher.Publish(ctx, domain.TopicHrExpenseSubmitted, claim.ID, domain.ExpenseSubmittedEvent{
		ExpenseID:   claim.ID,
		EmployeeID:  claim.EmployeeID,
		Description: desc,
		Amount:      claim.TotalAmount,
		Timestamp:   time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrExpenseSubmitted, err)
	}

	return claim, nil
}

func (s *EmployeeManagementService) ListDepartments(ctx context.Context) ([]domain.Department, error) {
	return s.depts.List(ctx)
}

func (s *EmployeeManagementService) CreateDepartment(ctx context.Context, code, name, description, managerID string) (*domain.Department, error) {
	id := fmt.Sprintf("dept_%d", time.Now().UnixNano())
	dept := &domain.Department{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if managerID != "" {
		dept.ManagerID = &managerID
	}
	err := s.depts.Create(ctx, dept)
	return dept, err
}

func (s *EmployeeManagementService) ListPositions(ctx context.Context) ([]domain.Position, error) {
	return s.positions.List(ctx)
}

func (s *EmployeeManagementService) CreatePosition(ctx context.Context, code, title, description, departmentID string, minSalary, maxSalary decimal.Decimal) (*domain.Position, error) {
	id := fmt.Sprintf("pos_%d", time.Now().UnixNano())
	pos := &domain.Position{
		ID:           id,
		Code:         code,
		Title:        title,
		Description:  description,
		DepartmentID: departmentID,
		MinSalary:    minSalary,
		MaxSalary:    maxSalary,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := s.positions.Create(ctx, pos)
	return pos, err
}

func (s *EmployeeManagementService) UpdateEmployeeAvailability(ctx context.Context, employeeID string, status string) error {
	if err := s.publisher.Publish(ctx, domain.TopicHrEmployeeAvailable, employeeID, domain.EmployeeAvailableEvent{
		EmployeeID: employeeID,
		Status:     status,
		Timestamp:  time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrEmployeeAvailable, err)
		return err
	}
	return nil
}

func (s *EmployeeManagementService) UpdateEmployeeSkills(ctx context.Context, employeeID string, skills []string) error {
	if err := s.publisher.Publish(ctx, domain.TopicHrEmployeeSkillsUpdated, employeeID, domain.EmployeeSkillsUpdatedEvent{
		EmployeeID: employeeID,
		Skills:     skills,
		Timestamp:  time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrEmployeeSkillsUpdated, err)
		return err
	}
	return nil
}

