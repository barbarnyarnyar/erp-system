# ERP System CDD Gap Analysis — Phase 6: Dead-Letter Queue

**Source PRD**: docs/PRDs/active/2026-06-06-1557-cdd-gap-analysis.md
**PRD ID**: PRD-2026-06-06-1557
**Phase**: 6 of 6 (Optional, post-PRD)
**Status**: Completed
**Created**: June 06, 2026

---

## Objective

Add dead-letter queue (DLQ) handling for Kafka consumer errors. When a consumer handler fails (invalid payload, processing error), send the failed message to a `*.dead-letter` topic instead of silently dropping it.

## Rationale

This was part of the original Phase 1 plan but is a **new architectural feature**, not a gap fix. CDD contracts do not define DLQ behavior. Moved to a separate phase so it can be independently prioritized.

## Scope

### In Scope

- Add `DeadLetterTopic` constant per service
- Wrap consumer handler calls in try/catch pattern
- Publish failed messages to `*.dead-letter` topic with original message metadata
- Log DLQ publish errors (do not fail on DLQ failure)

### Out of Scope

- DLQ reprocessing (manual or automated replay)
- DLQ monitoring dashboard
- Retry logic before DLQ (single-failure → DLQ)
- Exactly-once semantics

---

## Design

### Topic Naming Convention

```
{topic}.dead-letter
```

Example: `hr.employee.created` DLQ → `hr.employee.created.dead-letter`

### Message Format

```go
type DeadLetterMessage struct {
    OriginalTopic string      `json:"original_topic"`
    OriginalKey   string      `json:"original_key,omitempty"`
    Payload       interface{} `json:"payload"`
    Error         string      `json:"error"`
    FailedAt      time.Time   `json:"failed_at"`
    ServiceName   string      `json:"service_name"`
}
```

### Consumer Wrapper Pattern

```go
func (c *Consumer) handleWithDLQ(topic string, msg []byte, handler func([]byte) error) {
    if err := handler(msg); err != nil {
        dlqMsg := DeadLetterMessage{
            OriginalTopic: topic,
            Payload:       string(msg),
            Error:         err.Error(),
            FailedAt:      time.Now(),
            ServiceName:   c.serviceName,
        }
        dlqPayload, _ := json.Marshal(dlqMsg)
        if dlqErr := c.publisher.Publish(ctx, topic+".dead-letter", dlqPayload); dlqErr != nil {
            log.Printf("ERROR: failed to publish DLQ message for topic %s: %v", topic, dlqErr)
        }
        log.Printf("ERROR: consumer handler failed for topic %s: %v — sent to DLQ", topic, err)
    }
}
```

---

## Implementation Tasks

### Task 1: Create DLQ message type

**Description:** Add `DeadLetterMessage` struct to shared Kafka package or per service.

**File:** `shared/kafka/dead_letter.go` (or per-service variant)

### Task 2: Add DLQ topic constants

**Description:** For each consumer subscription, define a `${Topic}DeadLetter` constant.

**Services with consumers:** FM, HR, SCM, M, CRM, PM (6 services)

### Task 3: Wrap consumer handlers

**Description:** Modify each consumer's `HandleMessage` or equivalent to use the DLQ wrapper pattern.

**Priority order:** M → SCM → FM → HR → CRM → PM (by consumer complexity)

### Task 4: Verify DLQ messages are published

**Description:** Test by sending a malformed message and verifying it appears on the `.dead-letter` topic.

---

## Verification

```bash
# Send invalid message to a consumer topic
# Check DLQ topic for the message
# Verify original consumer still processes valid messages
```

---

## Risks

| Risk | Likelihood | Mitigation |
| ---- | ---------- | ---------- |
| DLQ publish itself fails | Low | Log and continue — don't crash on DLQ failure |
| DLQ topics grow unbounded | Medium | Phase 6 does not include reprocessing; document as known limitation |
| CDD should define DLQ behavior | Medium | Update CDD contracts if DLQ becomes standard |

## Definition of Done

- [x] Task 1: `DeadLetterMessage` type defined
- [x] Task 2: DLQ topic constants added to all 6 consuming services
- [x] Task 3: All consumer handlers wrapped with DLQ
- [x] Task 4: Manual test confirms DLQ messages arrive
- [x] `make build` passes for all services

---

## Handoff Notes

This phase is optional and can be deferred indefinitely. The error logging from Phase 0.5 provides sufficient visibility for development. DLQ becomes important when Kafka is used across service boundaries in production.
