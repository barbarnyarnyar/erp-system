package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/timandy/routine"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	GlobalTracer trace.Tracer
	traceIDTLS   = routine.NewThreadLocal[string]()
)

func init() {
	// Set standard logger output to our custom log bridge to automatically correlate log messages
	log.SetOutput(&LogBridge{originalWriter: os.Stderr})
}

// InitTracer initializes OpenTelemetry trace provider and composite propagator.
func InitTracer(serviceName string) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	GlobalTracer = otel.Tracer(serviceName)
}

// SetTraceID sets the trace ID in the current goroutine's local storage.
func SetTraceID(traceID string) {
	traceIDTLS.Set(traceID)
}

// GetTraceID retrieves the trace ID from the current goroutine's local storage.
func GetTraceID() string {
	return traceIDTLS.Get()
}

// ClearTraceID clears the trace ID from the current goroutine's local storage.
func ClearTraceID() {
	traceIDTLS.Remove()
}

// GetTraceIDFromContext extracts trace ID from the OpenTelemetry span in the context.
func GetTraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// LogBridge intercepts writes to the standard logger to inject trace IDs.
type LogBridge struct {
	originalWriter io.Writer
}

func (l *LogBridge) Write(p []byte) (n int, err error) {
	traceID := GetTraceID()
	if traceID == "" {
		return l.originalWriter.Write(p)
	}

	// Standard log format prefix: "YYYY/MM/DD HH:MM:SS " (20 characters)
	if len(p) >= 20 && p[4] == '/' && p[7] == '/' && p[10] == ' ' && p[13] == ':' && p[16] == ':' {
		timestamp := p[:20]
		msg := p[20:]

		decorated := make([]byte, 0, len(p)+len(traceID)+16)
		decorated = append(decorated, timestamp...)
		decorated = append(decorated, fmt.Sprintf("[trace_id: %s] ", traceID)...)
		decorated = append(decorated, msg...)

		_, err = l.originalWriter.Write(decorated)
		return len(p), err
	}

	decorated := []byte(fmt.Sprintf("[trace_id: %s] %s", traceID, string(p)))
	_, err = l.originalWriter.Write(decorated)
	return len(p), err
}

// TracingMiddleware extracts OpenTelemetry span/trace parent context from the request header,
// starts an HTTP span, and registers the trace ID in goroutine-local storage.
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if GlobalTracer == nil {
			InitTracer(serviceName)
		}

		// Extract trace context from request headers
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Check if a trace ID exists in "X-Trace-ID" header as a fallback
		xTraceID := c.GetHeader("X-Trace-ID")
		var span trace.Span
		if xTraceID != "" && !trace.SpanContextFromContext(ctx).IsValid() {
			// Fallback: If no standard traceparent header exists, but X-Trace-ID exists,
			// generate a dummy span context with the given TraceID.
			traceID, err := trace.TraceIDFromHex(xTraceID)
			if err == nil {
				spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    traceID,
					Remote:     true,
					TraceFlags: trace.FlagsSampled,
				})
				ctx = trace.ContextWithSpanContext(ctx, spanCtx)
			}
		}

		// Start a new span for the HTTP handler
		ctx, span = GlobalTracer.Start(ctx, c.FullPath())
		defer span.End()

		// Update the request context with the new span context
		c.Request = c.Request.WithContext(ctx)

		traceID := span.SpanContext().TraceID().String()
		SetTraceID(traceID)
		defer ClearTraceID()

		// Also set the trace ID in the Gin context for convenience
		c.Set("trace_id", traceID)

		// Set response header
		c.Header("X-Trace-ID", traceID)

		// Set trace ID in logger fields
		logger := GetLogger(c).WithField("trace_id", traceID)
		c.Set("logger", logger)

		c.Next()
	}
}
