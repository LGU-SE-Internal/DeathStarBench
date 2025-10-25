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
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	loggerProvider *sdklog.LoggerProvider
	otelLogger     log.Logger
	loggerMutex    sync.RWMutex
)

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
func sendLogToOtel(logger log.Logger, logEntry map[string]interface{}) {
	ctx := context.Background()

	// Get the current span context if available
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	// Extract log level and message
	level, _ := logEntry["level"].(string)
	message, _ := logEntry["message"].(string)

	// Map level to OpenTelemetry severity
	severity := mapLevelToSeverity(level)

	// Create log record
	var logRecord log.Record
	logRecord.SetTimestamp(time.Now())
	logRecord.SetBody(log.StringValue(message))
	logRecord.SetSeverity(severity)
	logRecord.SetSeverityText(strings.ToUpper(level))

	// Prepare attributes including trace context
	attrs := make([]log.KeyValue, 0)
	
	// Add trace context if available
	if spanCtx.HasTraceID() {
		attrs = append(attrs, log.String("trace_id", spanCtx.TraceID().String()))
	}
	if spanCtx.HasSpanID() {
		attrs = append(attrs, log.String("span_id", spanCtx.SpanID().String()))
	}

	// Add other fields as attributes
	for k, v := range logEntry {
		if k == "level" || k == "message" || k == "time" {
			continue
		}
		attrs = append(attrs, log.String(k, toString(v)))
	}
	
	logRecord.AddAttributes(attrs...)

	// Emit the log record
	logger.Emit(ctx, logRecord)
}

// mapLevelToSeverity maps zerolog level to OpenTelemetry severity
func mapLevelToSeverity(level string) log.Severity {
	switch level {
	case "trace":
		return log.SeverityTrace
	case "debug":
		return log.SeverityDebug
	case "info":
		return log.SeverityInfo
	case "warn":
		return log.SeverityWarn
	case "error":
		return log.SeverityError
	case "fatal":
		return log.SeverityFatal
	case "panic":
		return log.SeverityFatal4
	default:
		return log.SeverityInfo
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

	// Create logger with the dual writer
	logger := zerolog.New(otelWriter).With().Timestamp().Caller().Logger()
	logger.Info().Msg("OpenTelemetry logger initialized successfully")

	return tracer, logger, nil
}

// CtxWithTraceID returns a logger with trace and span IDs from context
func CtxWithTraceID(ctx context.Context) zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		spanCtx := span.SpanContext()
		newLogger := *logger
		if spanCtx.HasTraceID() {
			newLogger = newLogger.With().Str("trace_id", spanCtx.TraceID().String()).Logger()
		}
		if spanCtx.HasSpanID() {
			newLogger = newLogger.With().Str("span_id", spanCtx.SpanID().String()).Logger()
		}
		return newLogger
	}
	
	return *logger
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
