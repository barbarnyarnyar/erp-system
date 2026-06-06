package testing

import (
	"context"
	"errors"
)

// MockEvent captures the components of a published event for assertion in tests.
type MockEvent struct {
	Topic   string
	Key     string
	Payload interface{}
}

// MockPublisher is a test fixture that implements EventPublisher and tracks published events.
type MockPublisher struct {
	Events      []MockEvent
	FailPublish bool
}

// Publish records the event or returns an error if FailPublish is toggled.
func (m *MockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	if m.FailPublish {
		return errors.New("failed to publish")
	}
	m.Events = append(m.Events, MockEvent{
		Topic:   topic,
		Key:     key,
		Payload: payload,
	})
	return nil
}
