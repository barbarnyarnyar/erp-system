package utils

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewID(t *testing.T) {
	id1 := NewID("test")
	id2 := NewID("test")

	if id1 == id2 {
		t.Errorf("expected generated IDs to be unique, got duplicate %s", id1)
	}
	if !strings.HasPrefix(id1, "test_") {
		t.Errorf("expected ID prefix 'test_', got %s", id1)
	}
}

func TestIsAny(t *testing.T) {
	if !IsAny("apple", "orange", "apple", "banana") {
		t.Error("expected apple to match list containing apple")
	}
	if IsAny("grape", "orange", "apple", "banana") {
		t.Error("expected grape not to match list")
	}
}

func TestLogPublishErr(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // reset standard output

	err := errors.New("connection failed")
	LogPublishErr("test-service", "test.topic", err)

	logged := buf.String()
	if !strings.Contains(logged, "ERROR: failed to publish event") {
		t.Errorf("expected log output to contain error message, got %q", logged)
	}
}

func TestLogger(t *testing.T) {
	InitLogger("test-service")
	if Logger == nil {
		t.Fatal("expected Logger to be initialized")
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	Logger.WithField("user_id", "123").WithError(errors.New("db error")).Info("running test %d", 1)
	Logger.Debug("debug msg")
	Logger.Warn("warn msg")
	Logger.Error("error msg")

	logged := buf.String()
	if !strings.Contains(logged, "[test-service]") ||
		!strings.Contains(logged, "fields=map[user_id:123]") ||
		!strings.Contains(logged, "error=db error") ||
		!strings.Contains(logged, "running test 1") ||
		!strings.Contains(logged, "DEBUG") ||
		!strings.Contains(logged, "WARN") ||
		!strings.Contains(logged, "ERROR") {
		t.Errorf("unexpected log output: %q", logged)
	}

	// Test GetLogger
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	l1 := GetLogger(c)
	if l1 != Logger {
		t.Error("expected GetLogger to return default logger when context has none")
	}

	entry := &LoggerEntry{serviceName: "ctx-service"}
	c.Set("logger", entry)
	l2 := GetLogger(c)
	if l2.serviceName != "ctx-service" {
		t.Error("expected GetLogger to retrieve logger from context")
	}

	// Test GetLogger with no global logger
	oldLogger := Logger
	Logger = nil
	defer func() { Logger = oldLogger }()
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	l3 := GetLogger(c2)
	if l3.serviceName != "unknown" {
		t.Errorf("expected serviceName unknown, got %s", l3.serviceName)
	}
}

func TestResponseHelper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	helper := NewResponseHelper("test-service")

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.Success(c, "operation successful", map[string]string{"foo": "bar"})

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "operation successful") || !strings.Contains(w.Body.String(), "test-service") {
			t.Errorf("unexpected body: %s", w.Body.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.Error(c, http.StatusConflict, "failed", nil)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("BadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.BadRequest(c, "invalid params")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.Unauthorized(c, "need auth")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.NotFound(c, "item missing")

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("InternalServerError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.InternalServerError(c, "unexpected crash", errors.New("null pointer"))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("Health", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.Health(c, "8080")

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("ValidateJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"invalid`))
		c.Request = req

		var obj struct {
			Name string `json:"name"`
		}
		valid := helper.ValidateJSON(c, &obj)
		if valid {
			t.Error("expected validation to fail for invalid JSON")
		}
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("BindAndValidate", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"name": "test"}`))
		c.Request = req

		var obj struct {
			Name string `json:"name"`
		}
		valid := helper.BindAndValidate(c, &obj)
		if !valid {
			t.Error("expected validation to succeed")
		}
	})

	t.Run("NotFoundErr", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.NotFoundErr(c, errors.New("missing"))

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("ConflictErr", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.ConflictErr(c, errors.New("collision"))

		if w.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("InternalErr", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		helper.InternalErr(c, errors.New("internal failure"))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("SuccessResponse", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SuccessResponse(c, "global success", "data")

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}
