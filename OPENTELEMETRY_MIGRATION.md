# OpenTelemetry Migration Guide

This document describes the migration from OpenTracing/Jaeger to OpenTelemetry for the DeathStarBench project.

## Overview

The DeathStarBench project has been migrated from OpenTracing/Jaeger tracing to OpenTelemetry. This migration provides:

- Modern, vendor-neutral observability framework
- Better integration with cloud-native ecosystems
- Configurable OTEL Collector endpoint via environment variables
- Support for external OTEL Collectors in different namespaces

## Changes Made

### 1. C++ Services (mediaMicroservices, socialNetwork)

#### Code Changes
- Updated `src/tracing.h` to use OpenTelemetry C++ SDK instead of Jaeger client
- Replaced OpenTracing API calls with OpenTelemetry API
- Added support for OTLP HTTP exporter

#### Build Changes
- Updated `CMakeLists.txt` files to link against OpenTelemetry libraries:
  - `opentelemetry_trace`
  - `opentelemetry_exporter_otlp_http`
  - `opentelemetry_exporter_otlp_grpc`
  - `opentelemetry_resources`
  - `opentelemetry_common`

#### Docker Changes
- Updated `docker/thrift-microservice-deps/cpp/Dockerfile` to install OpenTelemetry C++ v1.8.1
- Removed Jaeger client and OpenTracing dependencies

### 2. Go Services (hotelReservation)

#### Code Changes
- Updated `tracing/tracer.go` to use OpenTelemetry Go SDK
- Replaced Jaeger client with OTLP HTTP exporter
- Changed environment variables from `JAEGER_*` to `OTEL_*`

### 3. Helm Chart Configuration

#### Global Values
All Helm charts now use a unified OpenTelemetry configuration structure:

```yaml
global:
  otel:
    endpoint: http://otel-collector.observability.svc.cluster.local:4318
    samplerParam: 0.01
    disabled: false
```

#### Environment Variables
Each service deployment now automatically receives the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable:

```yaml
env:
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "{{ $.Values.global.otel.endpoint }}"
```

#### Removed Dependencies
Removed Jaeger subchart dependency from all Chart.yaml files:
- mediaMicroservices/Chart.yaml
- socialNetwork/Chart.yaml
- hotelReservation/Chart.yaml

## Configuration

### Setting the OTEL Collector Endpoint

You can configure the OTEL Collector endpoint in three ways:

#### 1. Via Helm Values (Recommended)

Edit `values.yaml`:
```yaml
global:
  otel:
    endpoint: http://otel-collector.your-namespace.svc.cluster.local:4318
    samplerParam: 0.01  # Sampling ratio (0.01 = 1%)
    disabled: false
```

#### 2. Via Helm Install/Upgrade Command

```bash
helm install media-microservices ./helm-chart/mediamicroservices \
  --set global.otel.endpoint=http://otel-collector.observability.svc.cluster.local:4318 \
  --set global.otel.samplerParam=0.1
```

#### 3. Via Environment Variable Override

The OTEL_EXPORTER_OTLP_ENDPOINT environment variable can be set directly in the deployment if needed.

### Endpoint Format

The endpoint should be specified as:
- HTTP: `http://hostname:4318` (default OTLP HTTP port)
- HTTPS: `https://hostname:4318`
- Cross-namespace: `http://service.namespace.svc.cluster.local:4318`

**Note:** The `/v1/traces` path is automatically appended by the code, so don't include it in the endpoint configuration.

## Migration from Jaeger

If you were previously using Jaeger, here's what changed:

### Old Configuration (Jaeger)
```yaml
global:
  jaeger:
    localAgentHostPort: jaeger:6831
    queueSize: 1000000
    bufferFlushInterval: 10
    samplerType: probabilistic
    samplerParam: 0.01
    disabled: false
    logSpans: false
```

### New Configuration (OpenTelemetry)
```yaml
global:
  otel:
    endpoint: http://otel-collector.observability.svc.cluster.local:4318
    samplerParam: 0.01
    disabled: false
```

### Environment Variable Changes

| Old (Jaeger) | New (OpenTelemetry) |
|--------------|---------------------|
| `JAEGER_SAMPLE_RATIO` | `OTEL_SAMPLE_RATIO` |
| `JAEGER_AGENT_HOST` | `OTEL_EXPORTER_OTLP_ENDPOINT` |

## Using External OTEL Collector

To use an OTEL Collector deployed in a different namespace:

1. Update the endpoint in `values.yaml`:
```yaml
global:
  otel:
    endpoint: http://otel-collector.observability.svc.cluster.local:4318
```

2. Ensure the OTEL Collector is configured to receive OTLP HTTP traces:
```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
```

3. No additional network policies are needed if using standard Kubernetes DNS resolution.

## Building Docker Images

If you need to rebuild the Docker images with OpenTelemetry:

### For mediaMicroservices and socialNetwork (C++):

```bash
cd mediaMicroservices/docker/thrift-microservice-deps/cpp
docker build -t your-registry/media-microservices-deps:latest .

cd ../../..
docker build -t your-registry/media-microservices:latest .
```

### For hotelReservation (Go):

Update your `go.mod` to include OpenTelemetry dependencies:
```
require (
    go.opentelemetry.io/otel v1.19.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.19.0
    go.opentelemetry.io/otel/sdk v1.19.0
)
```

Then build:
```bash
cd hotelReservation
docker build -t your-registry/hotel-reservation:latest .
```

## Verification

To verify the migration is working:

1. Deploy the services with the updated Helm charts
2. Generate some traffic to the services
3. Check the OTEL Collector for incoming traces
4. Verify traces appear in your observability backend (e.g., Jaeger, Tempo, etc.)

### Check Service Logs

Services will log OpenTelemetry initialization:
```
INFO: OpenTelemetry client: adjusted sample ratio 0.01, endpoint: otel-collector.observability.svc.cluster.local:4318
INFO: OpenTelemetry tracer initialized successfully
```

## Troubleshooting

### No Traces Appearing

1. Check the OTEL_EXPORTER_OTLP_ENDPOINT is set correctly:
```bash
kubectl get pod <pod-name> -o yaml | grep OTEL_EXPORTER_OTLP_ENDPOINT
```

2. Verify the OTEL Collector is accessible:
```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl -v http://otel-collector.observability.svc.cluster.local:4318/v1/traces
```

3. Check service logs for connection errors

### High Trace Volume

Adjust the sampling ratio to reduce volume:
```yaml
global:
  otel:
    samplerParam: 0.001  # 0.1% sampling
```

## References

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OpenTelemetry C++ SDK](https://github.com/open-telemetry/opentelemetry-cpp)
- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [OTLP Specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md)
