package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

// MockDemandForecastRepository is used to inject custom errors for error path testing
type MockDemandForecastRepository struct {
	domain.DemandForecastRepository
	createErr error
	updateErr error
}

func (m *MockDemandForecastRepository) Create(ctx context.Context, df *domain.DemandForecast) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.DemandForecastRepository.Create(ctx, df)
}

func (m *MockDemandForecastRepository) Update(ctx context.Context, df *domain.DemandForecast) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.DemandForecastRepository.Update(ctx, df)
}

func TestDemandPlanningService_ListForecasts(t *testing.T) {
	repo := memory.NewMemoryDemandForecastRepo()
	svc := NewDemandPlanningService(repo)
	ctx := context.Background()

	list, err := svc.ListForecasts(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 forecasts, got %d", len(list))
	}

	// Create a forecast
	confidence := decimal.NewFromFloat(0.85)
	notes := "Test forecast"
	created, err := svc.CreateForecast(ctx, "prod_1", time.Now().Add(24*time.Hour), 100, confidence, notes)
	if err != nil {
		t.Fatalf("create forecast: %v", err)
	}

	list, err = svc.ListForecasts(ctx)
	if err != nil {
		t.Fatalf("list forecasts: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 forecast, got %d", len(list))
	}
	if list[0].ID != created.ID {
		t.Errorf("expected forecast ID %s, got %s", created.ID, list[0].ID)
	}
}

func TestDemandPlanningService_CreateForecast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := memory.NewMemoryDemandForecastRepo()
		svc := NewDemandPlanningService(repo)
		ctx := context.Background()

		forecastDate := time.Now().Add(48 * time.Hour)
		qty := 250
		confidence := decimal.NewFromFloat(0.95)
		notes := "Holiday Season Demand"

		df, err := svc.CreateForecast(ctx, "prod_1", forecastDate, qty, confidence, notes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if df.ID == "" {
			t.Error("expected generated ID, got empty string")
		}
		if df.ProductID != "prod_1" {
			t.Errorf("expected ProductID prod_1, got %s", df.ProductID)
		}
		if df.ForecastQuantity != qty {
			t.Errorf("expected qty %d, got %d", qty, df.ForecastQuantity)
		}
		if !df.ConfidenceLevel.Equal(confidence) {
			t.Errorf("expected confidence %s, got %s", confidence, df.ConfidenceLevel)
		}
		if df.Notes != notes {
			t.Errorf("expected notes %s, got %s", notes, df.Notes)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &MockDemandForecastRepository{
			DemandForecastRepository: memory.NewMemoryDemandForecastRepo(),
			createErr:                errors.New("db insert failed"),
		}
		svc := NewDemandPlanningService(repo)
		ctx := context.Background()

		_, err := svc.CreateForecast(ctx, "prod_1", time.Now(), 10, decimal.NewFromFloat(0.5), "")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "db insert failed" {
			t.Errorf("expected db insert failed error, got: %v", err)
		}
	})
}

func TestDemandPlanningService_GetForecast(t *testing.T) {
	repo := memory.NewMemoryDemandForecastRepo()
	svc := NewDemandPlanningService(repo)
	ctx := context.Background()

	confidence := decimal.NewFromFloat(0.85)
	created, err := svc.CreateForecast(ctx, "prod_1", time.Now().Add(24*time.Hour), 100, confidence, "notes")
	if err != nil {
		t.Fatalf("create forecast: %v", err)
	}

	t.Run("found", func(t *testing.T) {
		got, err := svc.GetForecast(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("expected ID %s, got %s", created.ID, got.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetForecast(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestDemandPlanningService_UpdateForecast(t *testing.T) {
	repo := memory.NewMemoryDemandForecastRepo()
	svc := NewDemandPlanningService(repo)
	ctx := context.Background()

	created, err := svc.CreateForecast(ctx, "prod_1", time.Now().Add(24*time.Hour), 100, decimal.NewFromFloat(0.8), "original")
	if err != nil {
		t.Fatalf("create forecast: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		newDate := time.Now().Add(72 * time.Hour)
		newQty := 150
		newConfidence := decimal.NewFromFloat(0.9)
		newNotes := "updated notes"

		updated, err := svc.UpdateForecast(ctx, created.ID, newDate, newQty, newConfidence, newNotes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updated.ForecastQuantity != newQty {
			t.Errorf("expected qty %d, got %d", newQty, updated.ForecastQuantity)
		}
		if !updated.ConfidenceLevel.Equal(newConfidence) {
			t.Errorf("expected confidence %s, got %s", newConfidence, updated.ConfidenceLevel)
		}
		if updated.Notes != newNotes {
			t.Errorf("expected notes %s, got %s", newNotes, updated.Notes)
		}

		// Verify retrieval
		got, err := svc.GetForecast(ctx, created.ID)
		if err != nil {
			t.Fatalf("get forecast: %v", err)
		}
		if got.Notes != newNotes {
			t.Errorf("expected retrieved notes %s, got %s", newNotes, got.Notes)
		}
	})

	t.Run("get forecast error (not found)", func(t *testing.T) {
		_, err := svc.UpdateForecast(ctx, "nonexistent", time.Now(), 50, decimal.NewFromFloat(0.5), "notes")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("update repo error", func(t *testing.T) {
		mockRepo := &MockDemandForecastRepository{
			DemandForecastRepository: repo,
			updateErr:                errors.New("db update failed"),
		}
		mockSvc := NewDemandPlanningService(mockRepo)

		_, err := mockSvc.UpdateForecast(ctx, created.ID, time.Now(), 50, decimal.NewFromFloat(0.5), "notes")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "db update failed" {
			t.Errorf("expected db update failed error, got: %v", err)
		}
	})
}
