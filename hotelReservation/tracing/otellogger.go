package tracing

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	globalOtelLogger *OtelLogger
	loggerMutex      sync.RWMutex
	// Log is the global logger instance for backward compatibility
	Log *OtelLogger
)

// OtelLogger wraps OpenTelemetry log provider with a convenient API
type OtelLogger struct {
	logger  log.Logger
	ctx     context.Context
	attrs   []log.KeyValue
	level   log.Severity
	message string
}

// Logger returns the global OtelLogger instance
func Logger() *OtelLogger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	if globalOtelLogger == nil {
		// Return a no-op logger if not initialized
		return &OtelLogger{
			ctx:   context.Background(),
			attrs: make([]log.KeyValue, 0),
		}
	}
	return &OtelLogger{
		logger: globalOtelLogger.logger,
		ctx:    context.Background(),
		attrs:  make([]log.KeyValue, 0),
	}
}

// WithContext returns a new logger with the given context
func (l *OtelLogger) WithContext(ctx context.Context) *OtelLogger {
	newLogger := &OtelLogger{
		logger: l.logger,
		ctx:    ctx,
		attrs:  make([]log.KeyValue, len(l.attrs)),
	}
	copy(newLogger.attrs, l.attrs)
	
	// Extract trace context from span if available
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		spanCtx := span.SpanContext()
		if spanCtx.HasTraceID() {
			newLogger.attrs = append(newLogger.attrs, log.String("trace_id", spanCtx.TraceID().String()))
		}
		if spanCtx.HasSpanID() {
			newLogger.attrs = append(newLogger.attrs, log.String("span_id", spanCtx.SpanID().String()))
		}
	}
	
	return newLogger
}

// With returns a new logger with additional attributes
func (l *OtelLogger) With() *OtelLoggerChain {
	return &OtelLoggerChain{
		logger: l,
		attrs:  make([]log.KeyValue, len(l.attrs)),
	}
}

// Trace starts a trace level log
func (l *OtelLogger) Trace() *OtelLogEvent {
	return &OtelLogEvent{
		logger:   l,
		severity: log.SeverityTrace,
		attrs:    make([]log.KeyValue, len(l.attrs)),
	}
}

// Debug starts a debug level log
func (l *OtelLogger) Debug() *OtelLogEvent {
	return &OtelLogEvent{
		logger:   l,
		severity: log.SeverityDebug,
		attrs:    make([]log.KeyValue, len(l.attrs)),
	}
}

// Info starts an info level log
func (l *OtelLogger) Info() *OtelLogEvent {
	return &OtelLogEvent{
		logger:   l,
		severity: log.SeverityInfo,
		attrs:    make([]log.KeyValue, len(l.attrs)),
	}
}

// Warn starts a warn level log
func (l *OtelLogger) Warn() *OtelLogEvent {
	return &OtelLogEvent{
		logger:   l,
		severity: log.SeverityWarn,
		attrs:    make([]log.KeyValue, len(l.attrs)),
	}
}

// Error starts an error level log
func (l *OtelLogger) Error() *OtelLogEvent {
	return &OtelLogEvent{
		logger:   l,
		severity: log.SeverityError,
		attrs:    make([]log.KeyValue, len(l.attrs)),
	}
}

// Fatal starts a fatal level log
func (l *OtelLogger) Fatal() *OtelLogEvent {
	return &OtelLogEvent{
		logger:   l,
		severity: log.SeverityFatal,
		attrs:    make([]log.KeyValue, len(l.attrs)),
		isFatal:  true,
	}
}

// Panic starts a panic level log
func (l *OtelLogger) Panic() *OtelLogEvent {
	return &OtelLogEvent{
		logger:   l,
		severity: log.SeverityFatal4,
		attrs:    make([]log.KeyValue, len(l.attrs)),
		isPanic:  true,
	}
}

// OtelLoggerChain is used for building logger with attributes
type OtelLoggerChain struct {
	logger *OtelLogger
	attrs  []log.KeyValue
}

// Str adds a string attribute
func (c *OtelLoggerChain) Str(key, val string) *OtelLoggerChain {
	c.attrs = append(c.attrs, log.String(key, val))
	return c
}

// Int adds an int attribute
func (c *OtelLoggerChain) Int(key string, val int) *OtelLoggerChain {
	c.attrs = append(c.attrs, log.Int64(key, int64(val)))
	return c
}

// Err adds an error attribute
func (c *OtelLoggerChain) Err(err error) *OtelLoggerChain {
	if err != nil {
		c.attrs = append(c.attrs, log.String("error", err.Error()))
	}
	return c
}

// Timestamp adds a timestamp
func (c *OtelLoggerChain) Timestamp() *OtelLoggerChain {
	// Timestamp is automatically added by OpenTelemetry
	return c
}

// Caller adds caller information
func (c *OtelLoggerChain) Caller() *OtelLoggerChain {
	// Caller information can be added if needed
	return c
}

// Logger returns a new logger with the chained attributes
func (c *OtelLoggerChain) Logger() *OtelLogger {
	newLogger := &OtelLogger{
		logger: c.logger.logger,
		ctx:    c.logger.ctx,
		attrs:  make([]log.KeyValue, len(c.logger.attrs)+len(c.attrs)),
	}
	copy(newLogger.attrs, c.logger.attrs)
	copy(newLogger.attrs[len(c.logger.attrs):], c.attrs)
	return newLogger
}

// OtelLogEvent represents a log event being built
type OtelLogEvent struct {
	logger   *OtelLogger
	severity log.Severity
	attrs    []log.KeyValue
	isFatal  bool
	isPanic  bool
}

// Str adds a string attribute to the log event
func (e *OtelLogEvent) Str(key, val string) *OtelLogEvent {
	e.attrs = append(e.attrs, log.String(key, val))
	return e
}

// Int adds an int attribute to the log event
func (e *OtelLogEvent) Int(key string, val int) *OtelLogEvent {
	e.attrs = append(e.attrs, log.Int64(key, int64(val)))
	return e
}

// Float64 adds a float64 attribute to the log event
func (e *OtelLogEvent) Float64(key string, val float64) *OtelLogEvent {
	e.attrs = append(e.attrs, log.Float64(key, val))
	return e
}

// Err adds an error attribute to the log event
func (e *OtelLogEvent) Err(err error) *OtelLogEvent {
	if err != nil {
		e.attrs = append(e.attrs, log.String("error", err.Error()))
	}
	return e
}

// Msg emits the log with the given message
func (e *OtelLogEvent) Msg(msg string) {
	e.emit(msg)
}

// Msgf emits the log with a formatted message
func (e *OtelLogEvent) Msgf(format string, args ...interface{}) {
	e.emit(fmt.Sprintf(format, args...))
}

// Send emits the log (alias for Msg with empty string)
func (e *OtelLogEvent) Send() {
	e.emit("")
}

func (e *OtelLogEvent) emit(msg string) {
	if e.logger.logger == nil {
		// Fallback to console if logger not initialized
		timestamp := time.Now().Format(time.RFC3339)
		severityText := severityToString(e.severity)
		fmt.Fprintf(os.Stdout, "%s %s %s\n", timestamp, severityText, msg)
		
		if e.isFatal {
			os.Exit(1)
		}
		if e.isPanic {
			panic(msg)
		}
		return
	}
	
	var record log.Record
	record.SetTimestamp(time.Now())
	record.SetBody(log.StringValue(msg))
	record.SetSeverity(e.severity)
	record.SetSeverityText(severityToString(e.severity))
	
	// Combine logger attrs and event attrs
	allAttrs := make([]log.KeyValue, 0, len(e.logger.attrs)+len(e.attrs))
	allAttrs = append(allAttrs, e.logger.attrs...)
	allAttrs = append(allAttrs, e.attrs...)
	record.AddAttributes(allAttrs...)
	
	e.logger.logger.Emit(e.logger.ctx, record)
	
	// Also print to console
	timestamp := time.Now().Format(time.RFC3339)
	severityText := severityToString(e.severity)
	
	// Format attributes for console output
	attrStr := ""
	if len(allAttrs) > 0 {
		parts := make([]string, 0, len(allAttrs))
		for _, attr := range allAttrs {
			parts = append(parts, fmt.Sprintf("%s=%v", attr.Key, attr.Value.AsString()))
		}
		attrStr = " " + strings.Join(parts, " ")
	}
	
	fmt.Fprintf(os.Stdout, "%s %s %s%s\n", timestamp, severityText, msg, attrStr)
	
	if e.isFatal {
		os.Exit(1)
	}
	if e.isPanic {
		panic(msg)
	}
}

func severityToString(sev log.Severity) string {
	switch sev {
	case log.SeverityTrace:
		return "TRC"
	case log.SeverityDebug:
		return "DBG"
	case log.SeverityInfo:
		return "INF"
	case log.SeverityWarn:
		return "WRN"
	case log.SeverityError:
		return "ERR"
	case log.SeverityFatal, log.SeverityFatal2, log.SeverityFatal3, log.SeverityFatal4:
		return "FTL"
	default:
		return "INF"
	}
}

// InitOtelLogger initializes the OpenTelemetry logger
func InitOtelLogger(serviceName, endpoint string) error {
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

	// Create and store global logger
	loggerMutex.Lock()
	globalOtelLogger = &OtelLogger{
		logger: lp.Logger("hotelReservation"),
		ctx:    context.Background(),
		attrs:  make([]log.KeyValue, 0),
	}
	// Set the global Log variable for backward compatibility
	Log = globalOtelLogger
	loggerMutex.Unlock()

	return nil
}

// InitWithOtelLogging initializes both tracing and OpenTelemetry logging
func InitWithOtelLogging(serviceName, host string) (trace.Tracer, *OtelLogger, error) {
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
		return nil, nil, err
	}

	// Initialize OpenTelemetry logging
	err = InitOtelLogger(serviceName, endpoint)
	if err != nil {
		// Return a no-op logger if initialization fails
		fmt.Fprintf(os.Stderr, "Failed to initialize OpenTelemetry logger: %v\n", err)
		loggerMutex.Lock()
		globalOtelLogger = &OtelLogger{
			ctx:   context.Background(),
			attrs: make([]log.KeyValue, 0),
		}
		Log = globalOtelLogger
		loggerMutex.Unlock()
	}

	return tracer, Logger(), nil
}
