package kafka

import (
	"context"
	"testing"
)

func TestPublisher(t *testing.T) {
	pub := NewPublisher([]string{"localhost:9092"})
	if pub == nil {
		t.Fatal("expected NewPublisher to return non-nil publisher instance")
	}
	if pub.writer == nil {
		t.Error("expected writer to be initialized")
	}

	// Test Publish with cancelled context to cover the code path without blocking
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := pub.Publish(ctx, "test-topic", "key", "payload")
	if err == nil {
		t.Error("expected Publish to fail with cancelled context")
	}

	// Test Publish with invalid json payload to cover marshalling error
	ctx2 := context.Background()
	// Channel cannot be marshalled to JSON, which triggers json.Marshal error
	err = pub.Publish(ctx2, "test-topic", "key", make(chan int))
	if err == nil {
		t.Error("expected Publish to fail for unmarshallable payload")
	}

	// Close connection
	_ = pub.Close()
}
