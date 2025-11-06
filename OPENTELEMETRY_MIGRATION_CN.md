# OpenTelemetry 迁移指南

本文档描述了 DeathStarBench 项目从 OpenTracing/Jaeger 迁移到 OpenTelemetry 的过程。

## 概述

DeathStarBench 项目已从 OpenTracing/Jaeger 追踪迁移到 OpenTelemetry。此迁移提供：

- 现代化、供应商中立的可观测性框架
- 更好地与云原生生态系统集成
- 通过环境变量配置 OTEL Collector 端点
- 支持不同命名空间中的外部 OTEL Collector

## 已完成的变更

### 1. C++ 服务 (mediaMicroservices, socialNetwork)

#### 代码变更
- 更新 `src/tracing.h` 以使用 OpenTelemetry C++ SDK 替代 Jaeger 客户端
- 用 OpenTelemetry API 替换 OpenTracing API 调用
- 添加对 OTLP HTTP 导出器的支持

#### 构建变更
- 更新 `CMakeLists.txt` 文件链接 OpenTelemetry 库：
  - `opentelemetry_trace`
  - `opentelemetry_exporter_otlp_http`
  - `opentelemetry_exporter_otlp_grpc`
  - `opentelemetry_resources`
  - `opentelemetry_common`

#### Docker 变更
- 更新 `docker/thrift-microservice-deps/cpp/Dockerfile` 以安装 OpenTelemetry C++ v1.8.1
- 移除 Jaeger 客户端和 OpenTracing 依赖

### 2. Nginx/OpenResty 服务 (mediaMicroservices, socialNetwork)

#### Docker 变更
- 从 OpenTelemetry WebServer SDK 迁移到原生 ngx_otel_module 以获得更好的兼容性
- 移除 `opentracing-cpp`、`nginx-opentracing` 和 `jaeger-client-cpp` 的安装
- 更新 `docker/openresty-thrift/xenial/Dockerfile` 以：
  - 从源代码编译 ngx_otel_module v0.1.2
  - 将模块安装到 `/usr/local/openresty/nginx/modules/ngx_otel_module.so`
  - 添加所需依赖：`pkg-config`、`libc-ares-dev`、`libre2-dev` 以支持 gRPC
  - 使用 `--with-compat` 标志构建 OpenResty 以支持动态模块
  - 下载匹配的 nginx 源码（release-1.25.3）用于模块编译
- 移除 OpenTelemetry WebServer SDK 依赖
- 将 OpenResty 从 1.25.3.1 升级到 1.25.3.2

#### Nginx 配置变更
- 用 ngx_otel_module 替换 OpenTelemetry WebServer SDK 模块
- 移除 Jaeger 追踪器配置：
  ```nginx
  # 旧配置（已移除）
  opentracing on;
  opentracing_load_tracer /usr/local/lib/libjaegertracing_plugin.so /usr/local/openresty/nginx/jaeger-config.json;
  ```
- 新的 ngx_otel_module 配置语法：
  ```nginx
  # 新配置 - 加载 ngx_otel_module
  load_module modules/ngx_otel_module.so;
  
  http {
      # 配置 OTEL 导出器（仅支持 gRPC）
      otel_exporter {
          endpoint "otel-collector:4317";  # 注意：gRPC 端口 4317，而非 HTTP 4318
      }
      
      # 启用追踪
      otel_trace on;
      otel_service_name nginx-web-server;
      otel_trace_context propagate;
  }
  ```
- **重要提示：** ngx_otel_module 目前仅支持 gRPC 导出（端口 4317），不支持 HTTP（端口 4318）
- 从 init_by_lua_block 中移除 `opentracing_bridge_tracer` Lua 依赖

#### Helm Chart 变更
- 从 values.yaml 文件中移除 `global.jaeger` 配置部分
- 从 nginx 服务 chart 中移除 `jaeger-config.json` ConfigMap
- 所有 nginx 服务现在使用 `global.otel.endpoint` 进行追踪导出
- **注意：** 端点应使用 gRPC 端口 4317 以兼容 ngx_otel_module

### 3. Go 服务 (hotelReservation)

#### 代码变更
- 更新 `tracing/tracer.go` 以使用 OpenTelemetry Go SDK
- 用 OTLP HTTP 导出器替换 Jaeger 客户端
- 将环境变量从 `JAEGER_*` 改为 `OTEL_*`

### 4. Helm Chart 配置

#### 全局配置
所有 Helm chart 现在使用统一的 OpenTelemetry 配置结构：

```yaml
global:
  otel:
    endpoint: http://otel-collector.observability.svc.cluster.local:4318
    samplerParam: 0.01
    disabled: false
```

#### 环境变量
每个服务部署现在自动接收 `OTEL_EXPORTER_OTLP_ENDPOINT` 环境变量：

```yaml
env:
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "{{ $.Values.global.otel.endpoint }}"
```

#### 移除的依赖
从所有 Chart.yaml 文件中移除 Jaeger 子 chart 依赖：
- mediaMicroservices/Chart.yaml
- socialNetwork/Chart.yaml
- hotelReservation/Chart.yaml

## 配置说明

### 设置 OTEL Collector 端点

您可以通过三种方式配置 OTEL Collector 端点：

#### 1. 通过 Helm Values（推荐）

编辑 `values.yaml`：
```yaml
global:
  otel:
    endpoint: http://otel-collector.your-namespace.svc.cluster.local:4318
    samplerParam: 0.01  # 采样率 (0.01 = 1%)
    disabled: false
```

#### 2. 通过 Helm Install/Upgrade 命令

```bash
helm install media-microservices ./helm-chart/mediamicroservices \
  --set global.otel.endpoint=http://otel-collector.observability.svc.cluster.local:4318 \
  --set global.otel.samplerParam=0.1
```

#### 3. 通过环境变量覆盖

如需要，可以直接在部署中设置 OTEL_EXPORTER_OTLP_ENDPOINT 环境变量。

### 端点格式

端点应指定为：
- HTTP: `http://hostname:4318` (默认 OTLP HTTP 端口)
- HTTPS: `https://hostname:4318`
- 跨命名空间: `http://service.namespace.svc.cluster.local:4318`

**注意：** `/v1/traces` 路径会由代码自动添加，因此不要在端点配置中包含它。

## 从 Jaeger 迁移

如果您之前使用 Jaeger，以下是变更内容：

### 旧配置 (Jaeger)
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

### 新配置 (OpenTelemetry)
```yaml
global:
  otel:
    endpoint: http://otel-collector.observability.svc.cluster.local:4318
    samplerParam: 0.01
    disabled: false
```

### 环境变量变更

| 旧 (Jaeger) | 新 (OpenTelemetry) |
|--------------|---------------------|
| `JAEGER_SAMPLE_RATIO` | `OTEL_SAMPLE_RATIO` |
| `JAEGER_AGENT_HOST` | `OTEL_EXPORTER_OTLP_ENDPOINT` |

## 使用外部 OTEL Collector

要使用部署在不同命名空间的 OTEL Collector：

1. 在 `values.yaml` 中更新端点：
```yaml
global:
  otel:
    endpoint: http://otel-collector.observability.svc.cluster.local:4318
```

2. 确保 OTEL Collector 配置为接收 OTLP HTTP 追踪：
```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
```

3. 如果使用标准 Kubernetes DNS 解析，则不需要额外的网络策略。

## 构建 Docker 镜像

如果需要使用 OpenTelemetry 重新构建 Docker 镜像：

### 对于 mediaMicroservices 和 socialNetwork (C++)：

```bash
cd mediaMicroservices/docker/thrift-microservice-deps/cpp
docker build -t your-registry/media-microservices-deps:latest .

cd ../../..
docker build -t your-registry/media-microservices:latest .
```

### 对于 nginx/OpenResty 镜像：

带有 ngx_otel_module 支持的 nginx 镜像从 `docker/openresty-thrift/xenial` 目录构建：

```bash
# 对于 socialNetwork
cd socialNetwork/docker/openresty-thrift
docker build -f xenial/Dockerfile -t your-registry/openresty-thrift:xenial .

# 对于 mediaMicroservices  
cd mediaMicroservices/docker/openresty-thrift
docker build -f xenial/Dockerfile -t your-registry/openresty-thrift:xenial .
```

**构建流程：**
Dockerfile 会自动：
1. 下载并使用 `--with-compat` 标志构建 OpenResty 1.25.3.2
2. 下载与 OpenResty 版本匹配的 nginx 源码（release-1.25.3）
3. 使用与 OpenResty 相同的参数配置 nginx
4. 克隆并从源代码编译 ngx_otel_module v0.1.2
5. 将编译好的模块安装到 `/usr/local/openresty/nginx/modules/ngx_otel_module.so`

**重要说明：**
- ngx_otel_module 在 Docker 构建过程中编译
- 可以通过修改 `NGX_OTEL_VERSION` 构建参数来更改模块版本
- 该模块需要 gRPC 依赖（pkg-config、libc-ares-dev、libre2-dev），这些依赖已包含在内
- 该模块仅支持 gRPC 导出（端口 4317），不支持 HTTP（端口 4318）

### 对于 hotelReservation (Go)：

更新 `go.mod` 以包含 OpenTelemetry 依赖：
```
require (
    go.opentelemetry.io/otel v1.19.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.19.0
    go.opentelemetry.io/otel/sdk v1.19.0
)
```

然后构建：
```bash
cd hotelReservation
docker build -t your-registry/hotel-reservation:latest .
```

## 验证

验证迁移是否工作：

1. 使用更新的 Helm chart 部署服务
2. 向服务生成一些流量
3. 检查 OTEL Collector 是否有传入的追踪
4. 验证追踪是否出现在可观测性后端（例如 Jaeger、Tempo 等）

### 检查服务日志

服务将记录 OpenTelemetry 初始化：
```
INFO: OpenTelemetry client: adjusted sample ratio 0.01, endpoint: otel-collector.observability.svc.cluster.local:4318
INFO: OpenTelemetry tracer initialized successfully
```

## 故障排除

### 没有追踪出现

1. 检查 OTEL_EXPORTER_OTLP_ENDPOINT 是否正确设置：
```bash
kubectl get pod <pod-name> -o yaml | grep OTEL_EXPORTER_OTLP_ENDPOINT
```

2. 验证 OTEL Collector 是否可访问：
```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl -v http://otel-collector.observability.svc.cluster.local:4318/v1/traces
```

3. 检查服务日志是否有连接错误

### 追踪量过高

调整采样率以减少追踪量：
```yaml
global:
  otel:
    samplerParam: 0.001  # 0.1% 采样
```

## 参考资料

- [OpenTelemetry 文档](https://opentelemetry.io/docs/)
- [OpenTelemetry C++ SDK](https://github.com/open-telemetry/opentelemetry-cpp)
- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [OTLP 规范](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md)
