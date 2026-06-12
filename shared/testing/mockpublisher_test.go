package testing

import (
	"context"
	"testing"
)

func TestMockPublisher(t *testing.T) {
	pub := &MockPublisher{}

	ctx := context.Background()
	err := pub.Publish(ctx, "test.topic", "key1", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pub.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.Events))
	}
	if pub.Events[0].Topic != "test.topic" || pub.Events[0].Key != "key1" || pub.Events[0].Payload != "hello" {
		t.Errorf("unexpected event captured: %+v", pub.Events[0])
	}

	pub.FailPublish = true
	err = pub.Publish(ctx, "test.topic", "key1", "hello")
	if err == nil {
		t.Error("expected publishing failure when FailPublish is enabled")
	}
}
