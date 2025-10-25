# OpenTelemetry Log Integration for Hotel Reservation

## 概述 / Overview

本文档描述了 hotelReservation 服务如何集成 OpenTelemetry 日志导出功能，实现日志与追踪(trace)的关联。

This document describes how the hotelReservation services integrate OpenTelemetry log export functionality to correlate logs with traces.

## 功能特性 / Features

### 1. 日志导出 / Log Export

- **OTLP HTTP 协议**: 所有日志通过 OTLP HTTP 协议发送到 OpenTelemetry Collector
- **JSON 格式**: 日志以 JSON 格式收集，便于结构化查询
- **批量处理**: 使用批处理器优化性能，减少网络开销

### 2. 追踪关联 / Trace Correlation

- **Trace ID**: 每条日志自动包含 `trace_id` 字段
- **Span ID**: 每条日志自动包含 `span_id` 字段  
- **自动注入**: 在有活动 span 的上下文中，trace 信息自动注入日志

### 3. 双输出 / Dual Output

- **控制台输出**: 日志继续输出到控制台，便于本地调试
- **OTLP 导出**: 同时异步发送到 OpenTelemetry Collector

## 实现原理 / Implementation

### 架构 / Architecture

```
zerolog.Logger
    ↓
OtelLogWriter (intercepts JSON output)
    ↓
    ├→ Console Output (for debugging)
    └→ OTLP Log Exporter (for observability backend)
```

### 核心组件 / Core Components

#### 1. OtelLogWriter

位置: `hotelReservation/tracing/logger.go`

功能:
- 实现 `io.Writer` 接口
- 拦截 zerolog 的 JSON 输出
- 解析日志条目并提取信息
- 创建 OpenTelemetry LogRecord
- 添加 trace context (trace_id, span_id)

#### 2. InitWithLogging()

```go
func InitWithLogging(serviceName, host string) (trace.Tracer, zerolog.Logger, error)
```

功能:
- 初始化 trace provider 和 log provider
- 配置 OTLP 导出器
- 返回配置好的 tracer 和 logger

### 日志格式 / Log Format

每条日志包含以下字段:

```json
{
  "level": "info",
  "message": "Request processed",
  "timestamp": "2024-10-25T18:00:00Z",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "caller": "server.go:123",
  // ... other fields
}
```

## 使用方法 / Usage

### 服务初始化 / Service Initialization

所有服务已更新使用新的初始化方法:

```go
// 初始化 OpenTelemetry（包含日志和追踪）
tracer, logger, err := tracing.InitWithLogging("service-name", jaegerAddr)
if err != nil {
    log.Panic().Msgf("Failed to initialize OpenTelemetry: %v", err)
}

// 设置全局 logger
log.Logger = logger
```

### 在代码中使用 / Using in Code

#### 方式 1: 使用全局 logger

```go
log.Info().Msg("Processing request")
```

#### 方式 2: 使用带 trace context 的 logger

```go
func (s *Server) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
    // 从 context 提取 trace 信息并创建 logger
    logger := zerolog.Ctx(ctx)
    if logger.GetLevel() == zerolog.Disabled {
        globalLogger := log.Logger
        logger = &globalLogger
    }
    
    // 添加 trace context
    span := trace.SpanFromContext(ctx)
    if span.IsRecording() {
        spanCtx := span.SpanContext()
        if spanCtx.HasTraceID() {
            newLogger := logger.With().Str("trace_id", spanCtx.TraceID().String()).Logger()
            logger = &newLogger
        }
        if spanCtx.HasSpanID() {
            newLogger := logger.With().Str("span_id", spanCtx.SpanID().String()).Logger()
            logger = &newLogger
        }
    }
    
    logger.Info().Msg("Processing request with trace context")
    
    // ... business logic
}
```

## 配置 / Configuration

### 环境变量 / Environment Variables

- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP Collector 端点地址
  - 示例: `otel-collector.observability.svc.cluster.local:4318`
  - 默认: 使用 jaegerAddr 参数值

- `OTEL_SAMPLE_RATIO`: 采样率
  - 范围: 0.0 - 1.0
  - 默认: 0.01 (1%)

### Helm Values

在 `values.yaml` 中配置:

```yaml
global:
  otel:
    endpoint: http://otel-collector.observability.svc.cluster.local:4318
    samplerParam: 0.01
    disabled: false
```

## OpenTelemetry Collector 配置 / Collector Configuration

确保 Collector 配置接收 OTLP 日志:

```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024

exporters:
  # 配置你的后端 (如 Loki, Elasticsearch, etc.)
  loki:
    endpoint: http://loki:3100/loki/api/v1/push

service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [loki]
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [jaeger]
```

## 验证 / Verification

### 1. 检查日志输出

服务启动时会看到:

```
INFO: OpenTelemetry tracer and logger initialized
```

### 2. 检查 Collector 接收

查看 Collector 日志确认接收到日志:

```bash
kubectl logs -n observability otel-collector-xxx | grep "logs"
```

### 3. 查询后端

在你的可观测性后端（如 Grafana）中查询:

```
{service_name="search", trace_id="..."}
```

### 4. 验证 Trace 关联

1. 生成请求并获取 trace_id
2. 在日志后端搜索相同的 trace_id
3. 确认可以看到该请求的所有相关日志

## 性能考虑 / Performance Considerations

1. **异步发送**: 日志通过 goroutine 异步发送，不阻塞主逻辑
2. **批量处理**: 使用 BatchProcessor 减少网络请求
3. **采样**: 通过 OTEL_SAMPLE_RATIO 控制日志量
4. **失败处理**: OTLP 发送失败不影响控制台输出

## 故障排除 / Troubleshooting

### 问题: 日志没有发送到 Collector

1. 检查 OTEL_EXPORTER_OTLP_ENDPOINT 是否正确
2. 验证 Collector 是否可达:
   ```bash
   curl -v http://otel-collector:4318/v1/logs
   ```
3. 检查服务日志是否有错误信息

### 问题: trace_id 缺失

1. 确认请求通过 gRPC 拦截器处理
2. 检查 span context 是否正确传播
3. 验证 logger 是否从 context 正确提取

### 问题: 性能影响

1. 增加采样率: 减小 OTEL_SAMPLE_RATIO
2. 调整批处理大小
3. 考虑使用本地 Collector agent

## 依赖版本 / Dependencies

```
go.opentelemetry.io/otel v1.32.0
go.opentelemetry.io/otel/log v0.8.0
go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.8.0
go.opentelemetry.io/otel/sdk/log v0.8.0
github.com/rs/zerolog v1.31.0
```

## 参考资料 / References

- [OpenTelemetry Logs Specification](https://opentelemetry.io/docs/specs/otel/logs/)
- [OTLP Protocol](https://opentelemetry.io/docs/specs/otlp/)
- [Zerolog Documentation](https://github.com/rs/zerolog)

## 未来改进 / Future Improvements

1. 支持日志级别动态调整
2. 添加更多结构化字段
3. 实现日志过滤器
4. 支持 OTLP gRPC 协议
