package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/erp-system/pm-service/internal/business/domain"
)

// ==========================================
// Portfolio Memory Repository
// ==========================================

type PortfolioRepository struct {
	mu         sync.RWMutex
	portfolios map[string]domain.Portfolio
}

func NewPortfolioRepository() *PortfolioRepository {
	return &PortfolioRepository{
		portfolios: make(map[string]domain.Portfolio),
	}
}

func (r *PortfolioRepository) Create(ctx context.Context, portfolio *domain.Portfolio) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.portfolios[portfolio.ID] = *portfolio
	return nil
}

func (r *PortfolioRepository) GetByID(ctx context.Context, id string) (*domain.Portfolio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.portfolios[id]
	if !ok {
		return nil, fmt.Errorf("portfolio not found: %s", id)
	}
	return &p, nil
}

func (r *PortfolioRepository) List(ctx context.Context) ([]domain.Portfolio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Portfolio, 0, len(r.portfolios))
	for _, p := range r.portfolios {
		list = append(list, p)
	}
	return list, nil
}

func (r *PortfolioRepository) Update(ctx context.Context, portfolio *domain.Portfolio) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.portfolios[portfolio.ID]; !ok {
		return fmt.Errorf("portfolio not found: %s", portfolio.ID)
	}
	r.portfolios[portfolio.ID] = *portfolio
	return nil
}

func (r *PortfolioRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.portfolios, id)
	return nil
}

// ==========================================
// Project Memory Repository
// ==========================================

type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]domain.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects: make(map[string]domain.Project),
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.projects[project.ID] = *project
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.projects[id]
	if !ok {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	return &p, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Project, 0, len(r.projects))
	for _, p := range r.projects {
		list = append(list, p)
	}
	return list, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[project.ID]; !ok {
		return fmt.Errorf("project not found: %s", project.ID)
	}
	r.projects[project.ID] = *project
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.projects, id)
	return nil
}

// ==========================================
// Task Memory Repository
// ==========================================

type TaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]domain.Task
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make(map[string]domain.Task),
	}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = *task
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return &t, nil
}

func (r *TaskRepository) ListByProject(ctx context.Context, projectID string) ([]domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.Task
	for _, t := range r.tasks {
		if t.ProjectID == projectID {
			list = append(list, t)
		}
	}
	return list, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[task.ID]; !ok {
		return fmt.Errorf("task not found: %s", task.ID)
	}
	r.tasks[task.ID] = *task
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, id)
	return nil
}

// ==========================================
// TaskDependency Memory Repository
// ==========================================

type TaskDependencyRepository struct {
	mu   sync.RWMutex
	deps map[string]domain.TaskDependency
}

func NewTaskDependencyRepository() *TaskDependencyRepository {
	return &TaskDependencyRepository{
		deps: make(map[string]domain.TaskDependency),
	}
}

func (r *TaskDependencyRepository) Create(ctx context.Context, dep *domain.TaskDependency) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deps[dep.ID] = *dep
	return nil
}

func (r *TaskDependencyRepository) ListByTask(ctx context.Context, taskID string) ([]domain.TaskDependency, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.TaskDependency
	for _, d := range r.deps {
		if d.TaskID == taskID {
			list = append(list, d)
		}
	}
	return list, nil
}

func (r *TaskDependencyRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.deps, id)
	return nil
}

// ==========================================
// ResourceAllocation Memory Repository
// ==========================================

type ResourceAllocationRepository struct {
	mu     sync.RWMutex
	allocs map[string]domain.ResourceAllocation
}

func NewResourceAllocationRepository() *ResourceAllocationRepository {
	return &ResourceAllocationRepository{
		allocs: make(map[string]domain.ResourceAllocation),
	}
}

func (r *ResourceAllocationRepository) Create(ctx context.Context, alloc *domain.ResourceAllocation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.allocs[alloc.ID] = *alloc
	return nil
}

func (r *ResourceAllocationRepository) GetByID(ctx context.Context, id string) (*domain.ResourceAllocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.allocs[id]
	if !ok {
		return nil, fmt.Errorf("allocation not found: %s", id)
	}
	return &a, nil
}

func (r *ResourceAllocationRepository) ListByProject(ctx context.Context, projectID string) ([]domain.ResourceAllocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ResourceAllocation
	for _, a := range r.allocs {
		if a.ProjectID == projectID {
			list = append(list, a)
		}
	}
	return list, nil
}

func (r *ResourceAllocationRepository) Update(ctx context.Context, alloc *domain.ResourceAllocation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.allocs[alloc.ID]; !ok {
		return fmt.Errorf("allocation not found: %s", alloc.ID)
	}
	r.allocs[alloc.ID] = *alloc
	return nil
}

func (r *ResourceAllocationRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.allocs, id)
	return nil
}

// ==========================================
// ProjectTimeEntry Memory Repository
// ==========================================

type ProjectTimeEntryRepository struct {
	mu      sync.RWMutex
	entries map[string]domain.ProjectTimeEntry
}

func NewProjectTimeEntryRepository() *ProjectTimeEntryRepository {
	return &ProjectTimeEntryRepository{
		entries: make(map[string]domain.ProjectTimeEntry),
	}
}

func (r *ProjectTimeEntryRepository) Create(ctx context.Context, entry *domain.ProjectTimeEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[entry.ID] = *entry
	return nil
}

func (r *ProjectTimeEntryRepository) GetByID(ctx context.Context, id string) (*domain.ProjectTimeEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[id]
	if !ok {
		return nil, fmt.Errorf("project time entry not found: %s", id)
	}
	return &e, nil
}

func (r *ProjectTimeEntryRepository) ListByProject(ctx context.Context, projectID string) ([]domain.ProjectTimeEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ProjectTimeEntry
	for _, e := range r.entries {
		if e.ProjectID == projectID {
			list = append(list, e)
		}
	}
	return list, nil
}

func (r *ProjectTimeEntryRepository) Update(ctx context.Context, entry *domain.ProjectTimeEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.entries[entry.ID]; !ok {
		return fmt.Errorf("project time entry not found: %s", entry.ID)
	}
	r.entries[entry.ID] = *entry
	return nil
}

func (r *ProjectTimeEntryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, id)
	return nil
}

// ==========================================
// ProjectExpense Memory Repository
// ==========================================

type ProjectExpenseRepository struct {
	mu       sync.RWMutex
	expenses map[string]domain.ProjectExpense
}

func NewProjectExpenseRepository() *ProjectExpenseRepository {
	return &ProjectExpenseRepository{
		expenses: make(map[string]domain.ProjectExpense),
	}
}

func (r *ProjectExpenseRepository) Create(ctx context.Context, expense *domain.ProjectExpense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expenses[expense.ID] = *expense
	return nil
}

func (r *ProjectExpenseRepository) GetByID(ctx context.Context, id string) (*domain.ProjectExpense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.expenses[id]
	if !ok {
		return nil, fmt.Errorf("expense not found: %s", id)
	}
	return &e, nil
}

func (r *ProjectExpenseRepository) ListByProject(ctx context.Context, projectID string) ([]domain.ProjectExpense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ProjectExpense
	for _, e := range r.expenses {
		if e.ProjectID == projectID {
			list = append(list, e)
		}
	}
	return list, nil
}

func (r *ProjectExpenseRepository) Update(ctx context.Context, expense *domain.ProjectExpense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.expenses[expense.ID]; !ok {
		return fmt.Errorf("expense not found: %s", expense.ID)
	}
	r.expenses[expense.ID] = *expense
	return nil
}

func (r *ProjectExpenseRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.expenses, id)
	return nil
}

// ==========================================
// ProjectDocument Memory Repository
// ==========================================

type ProjectDocumentRepository struct {
	mu   sync.RWMutex
	docs map[string]domain.ProjectDocument
}

func NewProjectDocumentRepository() *ProjectDocumentRepository {
	return &ProjectDocumentRepository{
		docs: make(map[string]domain.ProjectDocument),
	}
}

func (r *ProjectDocumentRepository) Create(ctx context.Context, doc *domain.ProjectDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docs[doc.ID] = *doc
	return nil
}

func (r *ProjectDocumentRepository) GetByID(ctx context.Context, id string) (*domain.ProjectDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.docs[id]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", id)
	}
	return &d, nil
}

func (r *ProjectDocumentRepository) ListByProject(ctx context.Context, projectID string) ([]domain.ProjectDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ProjectDocument
	for _, d := range r.docs {
		if d.ProjectID == projectID {
			list = append(list, d)
		}
	}
	return list, nil
}

func (r *ProjectDocumentRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.docs, id)
	return nil
}

// ==========================================
// ProjectIssue Memory Repository
// ==========================================

type ProjectIssueRepository struct {
	mu     sync.RWMutex
	issues map[string]domain.ProjectIssue
}

func NewProjectIssueRepository() *ProjectIssueRepository {
	return &ProjectIssueRepository{
		issues: make(map[string]domain.ProjectIssue),
	}
}

func (r *ProjectIssueRepository) Create(ctx context.Context, issue *domain.ProjectIssue) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issues[issue.ID] = *issue
	return nil
}

func (r *ProjectIssueRepository) GetByID(ctx context.Context, id string) (*domain.ProjectIssue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, ok := r.issues[id]
	if !ok {
		return nil, fmt.Errorf("issue not found: %s", id)
	}
	return &i, nil
}

func (r *ProjectIssueRepository) ListByProject(ctx context.Context, projectID string) ([]domain.ProjectIssue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ProjectIssue
	for _, i := range r.issues {
		if i.ProjectID == projectID {
			list = append(list, i)
		}
	}
	return list, nil
}

func (r *ProjectIssueRepository) Update(ctx context.Context, issue *domain.ProjectIssue) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.issues[issue.ID]; !ok {
		return fmt.Errorf("issue not found: %s", issue.ID)
	}
	r.issues[issue.ID] = *issue
	return nil
}

func (r *ProjectIssueRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.issues, id)
	return nil
}

// ==========================================
// ChangeRequest Memory Repository
// ==========================================

type ChangeRequestRepository struct {
	mu   sync.RWMutex
	reqs map[string]domain.ChangeRequest
}

func NewChangeRequestRepository() *ChangeRequestRepository {
	return &ChangeRequestRepository{
		reqs: make(map[string]domain.ChangeRequest),
	}
}

func (r *ChangeRequestRepository) Create(ctx context.Context, req *domain.ChangeRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reqs[req.ID] = *req
	return nil
}

func (r *ChangeRequestRepository) GetByID(ctx context.Context, id string) (*domain.ChangeRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rq, ok := r.reqs[id]
	if !ok {
		return nil, fmt.Errorf("change request not found: %s", id)
	}
	return &rq, nil
}

func (r *ChangeRequestRepository) ListByProject(ctx context.Context, projectID string) ([]domain.ChangeRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.ChangeRequest
	for _, rq := range r.reqs {
		if rq.ProjectID == projectID {
			list = append(list, rq)
		}
	}
	return list, nil
}

func (r *ChangeRequestRepository) Update(ctx context.Context, req *domain.ChangeRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reqs[req.ID]; !ok {
		return fmt.Errorf("change request not found: %s", req.ID)
	}
	r.reqs[req.ID] = *req
	return nil
}

func (r *ChangeRequestRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.reqs, id)
	return nil
}
