package domain

import "context"

type PortfolioRepository interface {
	Create(ctx context.Context, portfolio *Portfolio) error
	GetByID(ctx context.Context, id string) (*Portfolio, error)
	List(ctx context.Context) ([]Portfolio, error)
	Update(ctx context.Context, portfolio *Portfolio) error
	Delete(ctx context.Context, id string) error
}

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context) ([]Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id string) (*Task, error)
	ListByProject(ctx context.Context, projectID string) ([]Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
}

type TaskDependencyRepository interface {
	Create(ctx context.Context, dep *TaskDependency) error
	ListByTask(ctx context.Context, taskID string) ([]TaskDependency, error)
	Delete(ctx context.Context, id string) error
}

type ResourceAllocationRepository interface {
	Create(ctx context.Context, alloc *ResourceAllocation) error
	GetByID(ctx context.Context, id string) (*ResourceAllocation, error)
	ListByProject(ctx context.Context, projectID string) ([]ResourceAllocation, error)
	Update(ctx context.Context, alloc *ResourceAllocation) error
	Delete(ctx context.Context, id string) error
}

type ProjectTimeEntryRepository interface {
	Create(ctx context.Context, entry *ProjectTimeEntry) error
	GetByID(ctx context.Context, id string) (*ProjectTimeEntry, error)
	ListByProject(ctx context.Context, projectID string) ([]ProjectTimeEntry, error)
	Update(ctx context.Context, entry *ProjectTimeEntry) error
	Delete(ctx context.Context, id string) error
}

type ProjectExpenseRepository interface {
	Create(ctx context.Context, expense *ProjectExpense) error
	GetByID(ctx context.Context, id string) (*ProjectExpense, error)
	ListByProject(ctx context.Context, projectID string) ([]ProjectExpense, error)
	Update(ctx context.Context, expense *ProjectExpense) error
	Delete(ctx context.Context, id string) error
}

type ProjectDocumentRepository interface {
	Create(ctx context.Context, doc *ProjectDocument) error
	GetByID(ctx context.Context, id string) (*ProjectDocument, error)
	ListByProject(ctx context.Context, projectID string) ([]ProjectDocument, error)
	Delete(ctx context.Context, id string) error
}

type ProjectIssueRepository interface {
	Create(ctx context.Context, issue *ProjectIssue) error
	GetByID(ctx context.Context, id string) (*ProjectIssue, error)
	ListByProject(ctx context.Context, projectID string) ([]ProjectIssue, error)
	Update(ctx context.Context, issue *ProjectIssue) error
	Delete(ctx context.Context, id string) error
}

type ChangeRequestRepository interface {
	Create(ctx context.Context, req *ChangeRequest) error
	GetByID(ctx context.Context, id string) (*ChangeRequest, error)
	ListByProject(ctx context.Context, projectID string) ([]ChangeRequest, error)
	Update(ctx context.Context, req *ChangeRequest) error
	Delete(ctx context.Context, id string) error
}

type MilestoneRepository interface {
	Create(ctx context.Context, m *Milestone) error
	GetByID(ctx context.Context, id string) (*Milestone, error)
	ListByProject(ctx context.Context, projectID string) ([]Milestone, error)
	Update(ctx context.Context, m *Milestone) error
	Delete(ctx context.Context, id string) error
}
