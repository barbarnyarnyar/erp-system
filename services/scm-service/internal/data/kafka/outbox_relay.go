package kafka

import (
	"context"
	"log"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
)

type OutboxRelayWorker struct {
	repo      domain.TransactionalOutboxRepository
	publisher domain.EventPublisher
	interval  time.Duration
	limit     int
}

func NewOutboxRelayWorker(repo domain.TransactionalOutboxRepository, publisher domain.EventPublisher, interval time.Duration, limit int) *OutboxRelayWorker {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if limit <= 0 {
		limit = 100
	}
	return &OutboxRelayWorker{
		repo:      repo,
		publisher: publisher,
		interval:  interval,
		limit:     limit,
	}
}

func (w *OutboxRelayWorker) Start(ctx context.Context) {
	log.Println("Starting background SCM Outbox Relay Worker...")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping SCM Outbox Relay Worker...")
			return
		case <-ticker.C:
			w.processPending(ctx)
		}
	}
}

func (w *OutboxRelayWorker) processPending(ctx context.Context) {
	records, err := w.repo.GetUnsent(ctx, w.limit)
	if err != nil {
		log.Printf("[SCM-OutboxRelay] Error fetching unsent records: %v", err)
		return
	}

	if len(records) == 0 {
		return
	}

	log.Printf("[SCM-OutboxRelay] Found %d unsent events to process", len(records))

	for _, rec := range records {
		// Attempt to publish to Kafka
		err = w.publisher.Publish(ctx, rec.EventType, rec.AggregateID, rec.Payload)
		if err != nil {
			log.Printf("[SCM-OutboxRelay] Failed to publish event %s (id: %s) to Kafka: %v", rec.EventType, rec.ID, err)

			newRetryCount := rec.RetryCount + 1
			status := domain.OutboxStatusFAILED
			if updateErr := w.repo.UpdateStatus(ctx, rec.ID, status, newRetryCount); updateErr != nil {
				log.Printf("[SCM-OutboxRelay] Failed to update outbox record status to FAILED: %v", updateErr)
			}
			continue
		}

		// On success, update status to SENT
		if updateErr := w.repo.UpdateStatus(ctx, rec.ID, domain.OutboxStatusSENT, rec.RetryCount); updateErr != nil {
			log.Printf("[SCM-OutboxRelay] Failed to update outbox record status to SENT: %v", updateErr)
		}
	}
}
