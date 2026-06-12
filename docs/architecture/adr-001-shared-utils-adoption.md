# ADR-001: Adoption of Shared Utilities Submodules

## Status
Accepted

## Context
Across our 7 Go microservices (Auth, CRM, FM, HR, Manufacturing, Projects, SCM), there was significant code duplication around:
1. **Event Publishing**: Each service hand-rolled its Kafka publisher constructor, leading to duplicate, inconsistent initialization.
2. **Event Error Logging**: Log lines for publishing failures were duplicated across dozens of business logic files.
3. **ID Generation**: Services used collision-prone `fmt.Sprintf("xxx_%d", time.Now().UnixNano())` logic.
4. **Mock Publishers**: Duplicate testing mocks with drifted features were declared inside different service test suites.
5. **HTTP Error Formatting**: Repetitive `c.JSON(http.StatusBadRequest, gin.H{"error": ...})` calls populated handlers.

## Decision
We establish and enforce standard shared modules located under the `shared/` directory:
- **`shared/utils/idgen.go`**: High-entropy, randomized + timestamped ID generator: `utils.NewID(prefix)`.
- **`shared/utils/publish_log.go`**: Standardized event logging decorator helper: `utils.LogPublishErr(serviceName, topic, err)`.
- **`shared/utils/isany.go`**: Generic helper for enum validations: `utils.IsAny(val, val1, val2...)`.
- **`shared/utils/response.go`**: Wrapper utility `ResponseHelper` to enforce consistent JSON responses.
- **`shared/utils/logger.go`**: Enforces structured log formats.
- **`shared/kafka/publisher.go`**: Canonical Kafka client constructor with proper TCP configurations, least bytes balancing, and automatic topic creation.
- **`shared/testing/mockpublisher.go`**: Consolidates testing mocks with event capture and manual failure trigger support.

### Go Module Integration
For each Go service:
1. Include a `replace erp-system/shared => ../../shared` statement in its `go.mod`.
2. Import needed submodules under `"erp-system/shared/utils"`, `"erp-system/shared/kafka"`, or `"erp-system/shared/testing"`.

## How to Adopt for New Services / Handlers

### 1. ID Generation
Instead of UnixNano formatting, use:
```go
import "erp-system/shared/utils"

id := utils.NewID("prefix")
```

### 2. Handlers and Responses
Every API handler struct should inject `*utils.ResponseHelper`:
```go
type MyHandler struct {
    svc      *service.MyService
    response *utils.ResponseHelper
}

func NewMyHandler(svc *service.MyService, response *utils.ResponseHelper) *MyHandler {
    return &MyHandler{svc: svc, response: response}
}
```

In handlers, use response helper methods rather than manual `c.JSON(http.Status..., gin.H{"error":...})`:
```go
// bad request validation
if !h.response.BindAndValidate(c, &req) {
    return
}

// 404 response
h.response.NotFound(c, "resource not found")

// 500 response
h.response.InternalErr(c, err)
```

In `cmd/main.go`, instantiate the helper and initialize the logger:
```go
utils.InitLogger("my-service")
responseHelper := utils.NewResponseHelper("my-service")

myHandler := handlers.NewMyHandler(mySvc, responseHelper)
```
