package tracing

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	loggerProvider *sdklog.LoggerProvider
	otelLogger     otellog.Logger
	loggerMutex    sync.RWMutex
)

// TraceContextHook is a zerolog hook that automatically adds trace context to logs
type TraceContextHook struct{}

// Run implements zerolog.Hook interface
func (h TraceContextHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}
	
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	
	spanCtx := span.SpanContext()
	if spanCtx.HasTraceID() {
		e.Str("trace_id", spanCtx.TraceID().String())
	}
	if spanCtx.HasSpanID() {
		e.Str("span_id", spanCtx.SpanID().String())
	}
}

// LoggerWithTraceContext returns a logger with trace context from the span
// This should be called at the start of each request handler to get a logger
// that automatically includes trace_id and span_id in all log entries via the TraceContextHook
func LoggerWithTraceContext(ctx context.Context, baseLogger zerolog.Logger) *zerolog.Logger {
	// Store the logger in the context so the hook can access the context
	ctx = baseLogger.WithContext(ctx)
	// Return the logger from context
	return zerolog.Ctx(ctx)
}

// OtelLogWriter is a writer that sends logs to OpenTelemetry
type OtelLogWriter struct {
	consoleWriter io.Writer
}

// Write implements io.Writer interface
func (w *OtelLogWriter) Write(p []byte) (n int, err error) {
	// Always write to console first
	n, err = w.consoleWriter.Write(p)

	// Parse the JSON log entry
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		// If we can't parse, just return - console log was already written
		return n, nil
	}

	// Send to OpenTelemetry if logger is initialized
	loggerMutex.RLock()
	logger := otelLogger
	loggerMutex.RUnlock()

	if logger != nil {
		go sendLogToOtel(logger, logEntry)
	}

	return n, nil
}

// sendLogToOtel sends a log entry to OpenTelemetry
func sendLogToOtel(logger otellog.Logger, logEntry map[string]interface{}) {
	ctx := context.Background()

	// Extract trace context from log entry fields and create a context with span context
	// These fields are added by CtxWithTraceID or manually in service code
	var traceID trace.TraceID
	var spanID trace.SpanID
	var hasTraceID bool
	
	if traceIDStr, ok := logEntry["trace_id"].(string); ok && traceIDStr != "" {
		if parsedTraceID, err := trace.TraceIDFromHex(traceIDStr); err == nil {
			traceID = parsedTraceID
			hasTraceID = true
		}
	}
	if spanIDStr, ok := logEntry["span_id"].(string); ok && spanIDStr != "" {
		if parsedSpanID, err := trace.SpanIDFromHex(spanIDStr); err == nil {
			spanID = parsedSpanID
		}
	}

	// If we have at least a trace ID, create a span context and add it to the context
	// This allows the OpenTelemetry logger to properly associate logs with traces
	// Note: A valid trace context requires at least a trace ID
	if hasTraceID {
		spanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
			// Use FlagsSampled as a reasonable default since these logs are being exported
			// The actual sampling decision was already made when the trace was created
			TraceFlags: trace.FlagsSampled,
		})
		ctx = trace.ContextWithSpanContext(ctx, spanContext)
	}

	// Extract log level and message
	level, _ := logEntry["level"].(string)
	message, _ := logEntry["message"].(string)

	// Map level to OpenTelemetry severity
	severity := mapLevelToSeverity(level)

	// Create log record
	var logRecord otellog.Record
	logRecord.SetTimestamp(time.Now())
	logRecord.SetBody(otellog.StringValue(message))
	logRecord.SetSeverity(severity)
	logRecord.SetSeverityText(strings.ToUpper(level))

	// Prepare attributes (excluding trace_id and span_id as they're now set via context)
	attrs := make([]otellog.KeyValue, 0)
	
	// Add other fields as attributes
	for k, v := range logEntry {
		if k == "level" || k == "message" || k == "time" || k == "trace_id" || k == "span_id" {
			continue
		}
		attrs = append(attrs, otellog.String(k, toString(v)))
	}
	
	logRecord.AddAttributes(attrs...)

	// Emit the log record with the context containing trace information
	logger.Emit(ctx, logRecord)
}

// mapLevelToSeverity maps zerolog level to OpenTelemetry severity
func mapLevelToSeverity(level string) otellog.Severity {
	switch level {
	case "trace":
		return otellog.SeverityTrace
	case "debug":
		return otellog.SeverityDebug
	case "info":
		return otellog.SeverityInfo
	case "warn":
		return otellog.SeverityWarn
	case "error":
		return otellog.SeverityError
	case "fatal":
		return otellog.SeverityFatal
	case "panic":
		return otellog.SeverityFatal4
	default:
		return otellog.SeverityInfo
	}
}

// toString converts interface{} to string
func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// InitLogger initializes the OpenTelemetry logger provider
func InitLogger(serviceName, endpoint string) error {
	ctx := context.Background()

	// Create OTLP HTTP log exporter
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		return err
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return err
	}

	// Create logger provider
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	// Set global logger provider
	global.SetLoggerProvider(lp)

	// Store logger provider and logger
	loggerMutex.Lock()
	loggerProvider = lp
	otelLogger = lp.Logger("hotelReservation")
	loggerMutex.Unlock()

	return nil
}

// InitWithLogging initializes both tracing and logging, and returns a configured logger
func InitWithLogging(serviceName, host string) (trace.Tracer, zerolog.Logger, error) {
	// Get OTEL endpoint from environment variable or use host parameter
	endpoint := host
	if val, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT"); ok {
		endpoint = val
	}

	// Remove protocol and path if present for the endpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimSuffix(endpoint, "/v1/traces")
	endpoint = strings.TrimSuffix(endpoint, "/v1/logs")

	// Initialize tracing first
	tracer, err := Init(serviceName, endpoint)
	if err != nil {
		return nil, zerolog.Logger{}, err
	}

	// Create console writer
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	// Initialize logging
	err = InitLogger(serviceName, endpoint)
	if err != nil {
		// Log error but don't fail - tracing is more critical
		// Return logger with just console output
		logger := zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()
		logger.Error().Err(err).Msg("Failed to initialize OpenTelemetry logger, continuing with console logging only")
		return tracer, logger, nil
	}

	// Create OtelLogWriter that writes to both console and OpenTelemetry
	otelWriter := &OtelLogWriter{
		consoleWriter: consoleWriter,
	}

	// Create logger with the dual writer and trace context hook
	// The hook automatically adds trace_id and span_id to all logs when a span context is available
	logger := zerolog.New(otelWriter).
		With().
		Timestamp().
		Caller().
		Logger().
		Hook(TraceContextHook{})
	logger.Info().Msg("OpenTelemetry logger initialized successfully")

	return tracer, logger, nil
}

// CtxWithTraceID returns a logger that will automatically include trace_id and span_id
// via the TraceContextHook. The logger must have been created with the hook (via InitWithLogging).
// Usage: logger := tracing.CtxWithTraceID(ctx); logger.Info().Msg("message")
func CtxWithTraceID(ctx context.Context) *zerolog.Logger {
	// Always use the global logger and store it in the context
	// so the TraceContextHook can access the context via e.GetCtx()
	ctx = log.Logger.WithContext(ctx)
	return zerolog.Ctx(ctx)
}

// ShutdownLogger gracefully shuts down the logger provider
func ShutdownLogger(ctx context.Context) error {
	loggerMutex.RLock()
	lp := loggerProvider
	loggerMutex.RUnlock()

	if lp != nil {
		return lp.Shutdown(ctx)
	}
	return nil
}
