package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type TaskManagementService struct {
	taskRepo  domain.TaskRepository
	depRepo   domain.TaskDependencyRepository
	publisher domain.EventPublisher
}

func NewTaskManagementService(
	taskRepo domain.TaskRepository,
	depRepo domain.TaskDependencyRepository,
	publisher domain.EventPublisher,
) *TaskManagementService {
	return &TaskManagementService{
		taskRepo:  taskRepo,
		depRepo:   depRepo,
		publisher: publisher,
	}
}

func (s *TaskManagementService) CreateTask(ctx context.Context, projectID, parentID, title, description, assignedTo string, startDate, endDate *time.Time) (*domain.Task, error) {
	id := fmt.Sprintf("task_%d", time.Now().UnixNano())
	task := &domain.Task{
		ID:          id,
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      "TODO",
		Progress:    0,
		ActualHours: decimal.NewFromInt(0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if parentID != "" {
		task.ParentID = &parentID
	}
	if assignedTo != "" {
		task.AssignedTo = &assignedTo
	}
	if startDate != nil {
		task.StartDate = startDate
	}
	if endDate != nil {
		task.EndDate = endDate
	}

	err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	// Publish Task Created Event
	_ = s.publisher.Publish(ctx, domain.TopicPrjTaskCreated, id, domain.TaskCreatedEvent{
		TaskID:    id,
		ProjectID: projectID,
		Title:     title,
		Timestamp: time.Now(),
	})

	if assignedTo != "" {
		_ = s.publisher.Publish(ctx, domain.TopicPrjTaskAssigned, id, domain.TaskAssignedEvent{
			TaskID:     id,
			ProjectID:  projectID,
			EmployeeID: assignedTo,
			Workload:   8,
			Timestamp:  time.Now(),
		})
	}

	return task, nil
}

func (s *TaskManagementService) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}

func (s *TaskManagementService) ListTasksByProject(ctx context.Context, projectID string) ([]domain.Task, error) {
	return s.taskRepo.ListByProject(ctx, projectID)
}

func (s *TaskManagementService) UpdateTaskProgress(ctx context.Context, taskID string, progress int, actualHours decimal.Decimal, status string) (*domain.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	oldStatus := task.Status
	task.Progress = progress
	task.ActualHours = actualHours
	task.Status = status
	task.UpdatedAt = time.Now()

	err = s.taskRepo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	if oldStatus != status {
		if status == "IN_PROGRESS" {
			_ = s.publisher.Publish(ctx, domain.TopicPrjTaskStarted, taskID, domain.TaskStartedEvent{
				TaskID:    taskID,
				ProjectID: task.ProjectID,
				Timestamp: time.Now(),
			})
		} else if status == "DONE" {
			_ = s.publisher.Publish(ctx, domain.TopicPrjTaskCompleted, taskID, domain.TaskCompletedEvent{
				TaskID:    taskID,
				ProjectID: task.ProjectID,
				Timestamp: time.Now(),
			})
		}
	}

	return task, nil
}

func (s *TaskManagementService) AssignTask(ctx context.Context, taskID string, employeeID string) (*domain.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.AssignedTo = &employeeID
	task.UpdatedAt = time.Now()

	err = s.taskRepo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	_ = s.publisher.Publish(ctx, domain.TopicPrjTaskAssigned, taskID, domain.TaskAssignedEvent{
		TaskID:     taskID,
		ProjectID:  task.ProjectID,
		EmployeeID: employeeID,
		Workload:   8,
		Timestamp:  time.Now(),
	})

	return task, nil
}

func (s *TaskManagementService) AddTaskDependency(ctx context.Context, taskID, dependsOnTaskID, depType string) (*domain.TaskDependency, error) {
	id := fmt.Sprintf("dep_%d", time.Now().UnixNano())
	dep := &domain.TaskDependency{
		ID:              id,
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
		DependencyType:  depType,
		CreatedAt:       time.Now(),
	}

	err := s.depRepo.Create(ctx, dep)
	if err != nil {
		return nil, err
	}
	return dep, nil
}

func (s *TaskManagementService) ListDependencies(ctx context.Context, taskID string) ([]domain.TaskDependency, error) {
	return s.depRepo.ListByTask(ctx, taskID)
}

func (s *TaskManagementService) RequestMaterial(ctx context.Context, projectID, taskID, productID string, qty int) error {
	return s.publisher.Publish(ctx, domain.TopicPrjMaterialRequested, projectID, domain.MaterialRequestedEvent{
		ProjectID:   projectID,
		TaskID:      taskID,
		ProductID:   productID,
		QtyRequired: qty,
		Timestamp:   time.Now(),
	})
}
