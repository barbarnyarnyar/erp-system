package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/data/memory"
)

type MockEventPublisher struct {
	PublishFunc func(ctx context.Context, topic string, key string, payload interface{}) error
	Calls       []PublishCall
}

type PublishCall struct {
	Topic   string
	Key     string
	Payload interface{}
}

func (m *MockEventPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	m.Calls = append(m.Calls, PublishCall{Topic: topic, Key: key, Payload: payload})
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, key, payload)
	}
	return nil
}

func TestOutboxRelayWorker_ProcessPending(t *testing.T) {
	ctx := context.Background()

	t.Run("successful relay", func(t *testing.T) {
		repo := memory.NewMemoryTransactionalOutboxRepo()
		pub := &MockEventPublisher{}

		rec1 := &domain.TransactionalOutbox{
			ID:          "ob_1",
			EventType:   "topic.test.event",
			AggregateID: "agg_1",
			Payload:     "payload_1",
			Status:      domain.OutboxStatusPENDING,
			CreatedAt:   time.Now(),
		}
		rec2 := &domain.TransactionalOutbox{
			ID:          "ob_2",
			EventType:   "topic.test.event",
			AggregateID: "agg_2",
			Payload:     "payload_2",
			Status:      domain.OutboxStatusFAILED,
			CreatedAt:   time.Now(),
		}
		_ = repo.Create(ctx, rec1)
		_ = repo.Create(ctx, rec2)

		worker := NewOutboxRelayWorker(repo, pub, 0, 0) // default interval/limit will apply

		// Run processPending
		worker.processPending(ctx)

		// Assert publication
		if len(pub.Calls) != 2 {
			t.Fatalf("expected 2 publisher calls, got %d", len(pub.Calls))
		}
		hasAgg1 := false
		hasAgg2 := false
		for _, call := range pub.Calls {
			if call.Key == "agg_1" {
				hasAgg1 = true
			}
			if call.Key == "agg_2" {
				hasAgg2 = true
			}
		}
		if !hasAgg1 || !hasAgg2 {
			t.Errorf("expected publication keys agg_1 and agg_2, got %+v", pub.Calls)
		}

		// Assert status updated to SENT in repository
		pending, _ := repo.GetPending(ctx, 10)
		if len(pending) != 0 {
			t.Errorf("expected 0 pending events left, got %d", len(pending))
		}
	})

	t.Run("failed relay status update", func(t *testing.T) {
		repo := memory.NewMemoryTransactionalOutboxRepo()
		pub := &MockEventPublisher{
			PublishFunc: func(ctx context.Context, topic string, key string, payload interface{}) error {
				return errors.New("kafka connection failure")
			},
		}

		rec1 := &domain.TransactionalOutbox{
			ID:          "ob_fail",
			EventType:   "topic.test.event",
			AggregateID: "agg_fail",
			Payload:     "payload_fail",
			Status:      domain.OutboxStatusPENDING,
			CreatedAt:   time.Now(),
		}
		_ = repo.Create(ctx, rec1)

		worker := NewOutboxRelayWorker(repo, pub, 50*time.Millisecond, 10)

		// Run processPending
		worker.processPending(ctx)

		// Assert publisher called
		if len(pub.Calls) != 1 {
			t.Fatalf("expected 1 publisher call, got %d", len(pub.Calls))
		}

		// Assert status updated to FAILED in repository
		pending, _ := repo.GetPending(ctx, 10)
		if len(pending) != 1 {
			t.Fatalf("expected 1 pending/failed event, got %d", len(pending))
		}
		if pending[0].Status != domain.OutboxStatusFAILED {
			t.Errorf("expected status to be FAILED, got %s", pending[0].Status)
		}
	})

	t.Run("empty outbox does nothing", func(t *testing.T) {
		repo := memory.NewMemoryTransactionalOutboxRepo()
		pub := &MockEventPublisher{}
		worker := NewOutboxRelayWorker(repo, pub, 50*time.Millisecond, 10)

		worker.processPending(ctx)
		if len(pub.Calls) != 0 {
			t.Errorf("expected 0 publisher calls, got %d", len(pub.Calls))
		}
	})

	t.Run("worker start and stop", func(t *testing.T) {
		repo := memory.NewMemoryTransactionalOutboxRepo()
		pub := &MockEventPublisher{}
		worker := NewOutboxRelayWorker(repo, pub, 10*time.Millisecond, 10)

		workerCtx, cancel := context.WithTimeout(ctx, 25*time.Millisecond)
		defer cancel()

		// Start background run
		worker.Start(workerCtx) // exits when context times out/cancelled
	})
}
