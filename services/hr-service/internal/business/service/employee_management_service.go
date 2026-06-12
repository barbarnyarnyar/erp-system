package service

import (
	"context"
	"erp-system/shared/utils"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type EmployeeManagementService struct {
	repo            domain.EmployeeRepository
	claims          domain.ExpenseClaimRepository
	claimLines      domain.ExpenseClaimLineRepository
	historyRepo     domain.EmployeeCompensationHistoryRepository
	posHistoryRepo  domain.PositionHistoryRepository
	deptHistoryRepo domain.DepartmentHistoryRepository
	depts           domain.DepartmentRepository
	positions       domain.PositionRepository
	publisher       domain.EventPublisher
}

func NewEmployeeManagementService(
	repo domain.EmployeeRepository,
	claims domain.ExpenseClaimRepository,
	claimLines domain.ExpenseClaimLineRepository,
	historyRepo domain.EmployeeCompensationHistoryRepository,
	posHistoryRepo domain.PositionHistoryRepository,
	deptHistoryRepo domain.DepartmentHistoryRepository,
	depts domain.DepartmentRepository,
	positions domain.PositionRepository,
	publisher domain.EventPublisher,
) *EmployeeManagementService {
	return &EmployeeManagementService{
		repo:            repo,
		claims:          claims,
		claimLines:      claimLines,
		historyRepo:     historyRepo,
		posHistoryRepo:  posHistoryRepo,
		deptHistoryRepo: deptHistoryRepo,
		depts:           depts,
		positions:       positions,
		publisher:       publisher,
	}
}

func (s *EmployeeManagementService) ListEmployees(ctx context.Context) ([]domain.Employee, error) {
	return s.repo.List(ctx)
}

func (s *EmployeeManagementService) CreateEmployee(ctx context.Context, firstName, lastName, email, deptID, posID string, salary decimal.Decimal) (*domain.Employee, error) {
	if firstName == "" || lastName == "" || email == "" {
		return nil, errors.New("first name, last name, and email are required")
	}

	id := utils.NewID("emp")
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
	echID := utils.NewID("ech")
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
		utils.LogPublishErr("hr-service", domain.TopicHrEmployeeCreated, err)
	}

	return emp, nil
}

func (s *EmployeeManagementService) GetEmployee(ctx context.Context, id string) (*domain.Employee, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EmployeeManagementService) UpdateEmployee(ctx context.Context, id, firstName, lastName, email, deptID, posID string, status string) (*domain.Employee, error) {
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldPositionID := emp.PositionID
	positionChanged := oldPositionID != posID

	oldDepartmentID := emp.DepartmentID
	departmentChanged := oldDepartmentID != deptID

	emp.FirstName = firstName
	emp.LastName = lastName
	emp.Email = email
	emp.DepartmentID = deptID
	emp.PositionID = posID
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
		utils.LogPublishErr("hr-service", domain.TopicHrEmployeeUpdated, err)
	}

	// Record Position History and Publish Promoted Event
	if positionChanged {
		phID := utils.NewID("ph")
		_ = s.posHistoryRepo.Create(ctx, &domain.PositionHistory{
			ID:            phID,
			EmployeeID:    emp.ID,
			PositionID:    posID,
			EffectiveDate: time.Now(),
			ChangedBy:     "system",
			CreatedAt:     time.Now(),
		})

		if err := s.publisher.Publish(ctx, domain.TopicHrEmployeePromoted, emp.ID, domain.EmployeePromotedEvent{
			EmployeeID:    emp.ID,
			OldPositionID: oldPositionID,
			NewPositionID: emp.PositionID,
			NewSalary:     emp.Salary,
			Timestamp:     time.Now(),
		}); err != nil {
			utils.LogPublishErr("hr-service", domain.TopicHrEmployeePromoted, err)
		}
	}

	// Record Department History and Publish Transferred Event
	if departmentChanged {
		dhID := utils.NewID("dh")
		_ = s.deptHistoryRepo.Create(ctx, &domain.DepartmentHistory{
			ID:            dhID,
			EmployeeID:    emp.ID,
			DepartmentID:  deptID,
			EffectiveDate: time.Now(),
			ChangedBy:     "system",
			CreatedAt:     time.Now(),
		})

		if err := s.publisher.Publish(ctx, domain.TopicHrEmployeeTransferred, emp.ID, domain.EmployeeTransferredEvent{
			EmployeeID:      emp.ID,
			OldDepartmentID: oldDepartmentID,
			NewDepartmentID: emp.DepartmentID,
			Timestamp:       time.Now(),
		}); err != nil {
			utils.LogPublishErr("hr-service", domain.TopicHrEmployeeTransferred, err)
		}
	}

	return emp, nil
}

func (s *EmployeeManagementService) UpdateCompensation(ctx context.Context, employeeID string, salary decimal.Decimal, effectiveDate time.Time, changedBy string) (*domain.Employee, error) {
	emp, err := s.repo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	oldSalary := emp.Salary

	// Record compensation history
	echID := utils.NewID("ech")
	err = s.historyRepo.Create(ctx, &domain.EmployeeCompensationHistory{
		ID:            echID,
		EmployeeID:    emp.ID,
		Salary:        salary,
		EffectiveDate: effectiveDate,
		ChangedBy:     changedBy,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		return nil, err
	}

	// Update base employee salary
	emp.Salary = salary
	emp.UpdatedAt = time.Now()
	err = s.repo.Update(ctx, emp)
	if err != nil {
		return nil, err
	}

	// Publish salary changed event
	if !oldSalary.Equal(salary) {
		if err := s.publisher.Publish(ctx, domain.TopicHrSalaryChanged, emp.ID, domain.SalaryChangedEvent{
			EmployeeID: emp.ID,
			OldSalary:  oldSalary,
			NewSalary:  salary,
			Timestamp:  time.Now(),
		}); err != nil {
			utils.LogPublishErr("hr-service", domain.TopicHrSalaryChanged, err)
		}
	}

	return emp, nil
}

func (s *EmployeeManagementService) ListPositionHistory(ctx context.Context, employeeID string) ([]domain.PositionHistory, error) {
	return s.posHistoryRepo.ListByEmployeeID(ctx, employeeID)
}

func (s *EmployeeManagementService) ListDepartmentHistory(ctx context.Context, employeeID string) ([]domain.DepartmentHistory, error) {
	return s.deptHistoryRepo.ListByEmployeeID(ctx, employeeID)
}

func (s *EmployeeManagementService) ListCompensationHistory(ctx context.Context, employeeID string) ([]domain.EmployeeCompensationHistory, error) {
	return s.historyRepo.ListByEmployeeID(ctx, employeeID)
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
		utils.LogPublishErr("hr-service", domain.TopicHrEmployeeTerminated, err)
	}

	return nil
}

func (s *EmployeeManagementService) SubmitExpenseClaim(ctx context.Context, employeeID string, claimDate time.Time, lines []domain.ExpenseClaimLine) (*domain.ExpenseClaim, error) {
	claimID := utils.NewID("exp")

	var total decimal.Decimal
	for i := range lines {
		lines[i].ID = utils.NewID("expl")
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
		utils.LogPublishErr("hr-service", domain.TopicHrExpenseSubmitted, err)
	}

	return claim, nil
}

func (s *EmployeeManagementService) ListDepartments(ctx context.Context) ([]domain.Department, error) {
	return s.depts.List(ctx)
}

func (s *EmployeeManagementService) CreateDepartment(ctx context.Context, code, name, description, managerID string) (*domain.Department, error) {
	id := utils.NewID("dept")
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
	id := utils.NewID("pos")
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
		utils.LogPublishErr("hr-service", domain.TopicHrEmployeeAvailable, err)
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
		utils.LogPublishErr("hr-service", domain.TopicHrEmployeeSkillsUpdated, err)
		return err
	}
	return nil
}
