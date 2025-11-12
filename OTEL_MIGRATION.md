# OpenTracing to OpenTelemetry Migration

This document describes the migration from OpenTracing to OpenTelemetry for the socialNetwork and mediaMicroservices components.

## Overview

The migration uses the OpenTelemetry C++ SDK with the OpenTracing shim layer to provide backward compatibility. This allows existing OpenTracing API calls to continue working while using OpenTelemetry as the underlying telemetry implementation.

## Changes Made

### 1. Dependency Updates

#### Dockerfiles
- **socialNetwork/docker/thrift-microservice-deps/cpp/Dockerfile**
- **mediaMicroservices/docker/thrift-microservice-deps/cpp/Dockerfile**

Added OpenTelemetry C++ SDK v1.14.2 installation with the following features enabled:
- Jaeger exporter support (`WITH_JAEGER=ON`)
- OTLP exporter support (`WITH_OTLP=ON`, `WITH_OTLP_GRPC=ON`, `WITH_OTLP_HTTP=ON`)
- OpenTracing shim for backward compatibility

The OpenTracing library (v1.5.1) is kept for shim compatibility.

### 2. Tracing Implementation Updates

#### tracing.h files
- **socialNetwork/src/tracing.h**
- **mediaMicroservices/src/tracing.h**

Replaced Jaeger client-cpp with OpenTelemetry SDK:
- Uses OpenTelemetry SDK for trace provider and span processors
- Configures Jaeger exporter for trace export
- Creates OpenTracing shim to wrap OpenTelemetry tracer
- Maintains backward compatibility with existing OpenTracing API calls

Key changes:
```cpp
// Old: Using Jaeger client directly
auto tracer = jaegertracing::Tracer::make(service, config, logger);
opentracing::Tracer::InitGlobal(std::static_pointer_cast<opentracing::Tracer>(tracer));

// New: Using OpenTelemetry with OpenTracing shim
auto exporter = opentelemetry::exporter::jaeger::JaegerExporterFactory::Create(jaeger_options);
auto processor = opentelemetry::sdk::trace::SimpleSpanProcessorFactory::Create(std::move(exporter));
auto provider = opentelemetry::sdk::trace::TracerProviderFactory::Create(std::move(processors), resource);
opentelemetry::trace::Provider::SetTracerProvider(std::move(provider));
auto tracer_shim = opentracing::shim::TracerShim::createTracerShim();
opentracing::Tracer::InitGlobal(tracer_shim);
```

### 3. Build Configuration Updates

#### CMakeLists.txt files
Updated all service CMakeLists.txt files to link against OpenTelemetry libraries instead of jaegertracing:

**Old libraries:**
- `jaegertracing`

**New libraries:**
- `opentelemetry_trace` - Core tracing functionality
- `opentelemetry_exporter_jaeger_trace` - Jaeger exporter
- `opentelemetry_resources` - Resource attributes
- `opentelemetry_common` - Common utilities
- `opentelemetry_otlp_recordable` - OTLP support
- `opentracing_shim` - OpenTracing compatibility layer

## Benefits

1. **Future-proof**: OpenTelemetry is the industry standard for observability
2. **Backward Compatible**: Existing OpenTracing instrumentation continues to work
3. **More Features**: Access to OTLP protocol and other modern exporters
4. **Better Performance**: OpenTelemetry has optimized performance characteristics
5. **Active Development**: OpenTelemetry has more active development and community support

## Migration Strategy

Following the [official OpenTelemetry migration guide](https://opentelemetry.io/docs/migration/opentracing/), this migration:

1. ✅ Installed OpenTelemetry SDK and removed direct Jaeger client dependency
2. ✅ Installed OpenTracing Shim for backward compatibility
3. ✅ Configured OpenTelemetry to export via Jaeger (maintains existing infrastructure)
4. ✅ Added support for OTLP HTTP exporter via environment variable
5. ✅ Configured Helm charts to use OpenTelemetry Collector endpoint
6. ⏭️ Future: Can progressively rewrite instrumentation using OpenTelemetry API

## OpenTelemetry Collector Integration

The services now support dual export modes, configurable via environment variables:

### OTLP Exporter (Recommended)
When `OTEL_EXPORTER_OTLP_ENDPOINT` is set, traces are sent to the OpenTelemetry Collector via OTLP HTTP protocol:

**Helm Configuration:**
```yaml
# In values.yaml
global:
  opentelemetry:
    enabled: true
    collectorEndpoint: "http://opentelemetry-kube-stack-deployment-collector.monitoring:4318"
```

The environment variable is automatically injected into all C++ microservice pods.

**How it works:**
- Services check for `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable at startup
- If set, uses OTLP HTTP exporter to send traces to the collector
- If not set, falls back to Jaeger agent (backward compatible)

### Jaeger Exporter (Legacy)
Without `OTEL_EXPORTER_OTLP_ENDPOINT`, services continue using the Jaeger agent endpoint from `jaeger-config.yml`.

## Testing

To test the migration:

1. Build the Docker images with the updated dependencies
2. Deploy the services with Helm:
   ```bash
   # The OTEL collector endpoint is configured in values.yaml
   helm install socialnetwork ./socialNetwork/helm-chart/socialnetwork
   ```
3. Generate some traffic
4. Verify traces appear in your observability backend:
   - Via OpenTelemetry Collector (if configured)
   - Or directly in Jaeger UI (legacy mode)
5. Confirm trace context propagation works between services

## Configuration Options

You can disable OTLP and use only Jaeger by setting:
```yaml
global:
  opentelemetry:
    enabled: false
```

Or change the collector endpoint:
```yaml
global:
  opentelemetry:
    collectorEndpoint: "http://your-collector:4318"
```

## Next Steps (Optional)

1. ✅ **Migrate to OTLP**: Already implemented - set via environment variable
2. **Use OpenTelemetry API**: Gradually replace OpenTracing API calls with OpenTelemetry API
3. **Add Metrics and Logs**: Leverage OpenTelemetry's unified observability (traces, metrics, logs)
4. **Enhanced Collector Pipeline**: Configure advanced processing, sampling, and routing in the collector

## References

- [OpenTelemetry Migration Guide](https://opentelemetry.io/docs/migration/opentracing/)
- [OpenTelemetry C++ SDK](https://github.com/open-telemetry/opentelemetry-cpp)
- [OpenTracing Shim Documentation](https://github.com/open-telemetry/opentelemetry-cpp/tree/main/opentracing-shim)
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)
