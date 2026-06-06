package service

import (
	"context"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type PortfolioAnalyticsService struct {
	projectRepo domain.ProjectRepository
	taskRepo    domain.TaskRepository
	timeRepo    domain.ProjectTimeEntryRepository
	expenseRepo domain.ProjectExpenseRepository
}

func NewPortfolioAnalyticsService(
	projectRepo domain.ProjectRepository,
	taskRepo domain.TaskRepository,
	timeRepo domain.ProjectTimeEntryRepository,
	expenseRepo domain.ProjectExpenseRepository,
) *PortfolioAnalyticsService {
	return &PortfolioAnalyticsService{
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
		timeRepo:    timeRepo,
		expenseRepo: expenseRepo,
	}
}

type ProjectSummary struct {
	ProjectID      string          `json:"project_id"`
	ProjectName    string          `json:"project_name"`
	Status         string          `json:"status"`
	TaskCount      int             `json:"task_count"`
	CompletedTasks int             `json:"completed_tasks"`
	TotalHours     decimal.Decimal `json:"total_hours"`
	TotalExpenses  decimal.Decimal `json:"total_expenses"`
}

func (s *PortfolioAnalyticsService) GetPortfolioSummary(ctx context.Context, portfolioID string) ([]ProjectSummary, error) {
	projects, err := s.projectRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	var summaries []ProjectSummary

	for _, p := range projects {
		if p.PortfolioID != nil && *p.PortfolioID == portfolioID {
			tasks, _ := s.taskRepo.ListByProject(ctx, p.ID)
			completed := 0
			for _, t := range tasks {
				if t.Status == "DONE" {
					completed++
				}
			}

			// Aggregate time
			hours := decimal.Zero
			entries, _ := s.timeRepo.ListByProject(ctx, p.ID)
			for _, entry := range entries {
				if entry.Status == "APPROVED" {
					hours = hours.Add(entry.Hours)
				}
			}

			// Aggregate expenses
			expenses := decimal.Zero
			exps, _ := s.expenseRepo.ListByProject(ctx, p.ID)
			for _, exp := range exps {
				if exp.Status == "APPROVED" {
					expenses = expenses.Add(exp.Amount)
				}
			}

			summaries = append(summaries, ProjectSummary{
				ProjectID:      p.ID,
				ProjectName:    p.Name,
				Status:         p.Status,
				TaskCount:      len(tasks),
				CompletedTasks: completed,
				TotalHours:     hours,
				TotalExpenses:  expenses,
			})
		}
	}

	return summaries, nil
}
