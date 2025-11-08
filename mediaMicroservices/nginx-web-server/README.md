# OpenResty with ngx_otel_module - Build Instructions

This directory contains all files needed to build the OpenResty image with OpenTelemetry support using ngx_otel_module.

## Quick Start

To build the image, run from this directory:

```bash
cd mediaMicroservices/nginx-web-server
docker build -t openresty-thrift:focal .
```

Or with a custom registry:

```bash
cd mediaMicroservices/nginx-web-server
docker build -t 10.10.10.240/library/openresty-thrift:focal .
```

## Directory Structure

```
nginx-web-server/
├── Dockerfile              # Build instructions for OpenResty + ngx_otel_module
├── lua-thrift/             # Lua Thrift libraries
├── lua-json/               # JSON library for Lua
├── lua-scripts/            # Application Lua scripts (embedded in image)
├── nginx.conf              # Nginx configuration template
├── nginx.vh.default.conf   # Default virtual host configuration
└── README.md               # This file
```

**Note**: `gen-lua` directory (Thrift generated code) is copied from parent directory during build.

## Overview

This build migrates from the OpenTelemetry WebServer SDK to the native ngx_otel_module, which provides better compatibility with OpenResty.

**Base Image**: Ubuntu 20.04 (Focal) - provides modern system dependencies (CMake 3.16.3, c-ares 1.15.0) required by ngx_otel_module.

## Features

- **Base Image**: Ubuntu 20.04 LTS (Focal)
- OpenResty 1.25.3.2 (nginx 1.25.3)
- ngx_otel_module v0.1.2 (compiled from source)
- CMake 3.16.3 (from Ubuntu Focal)
- c-ares 1.15.0 (from Ubuntu Focal)
- OpenSSL 1.1.1w
- PCRE 8.45
- LuaRocks 3.9.2
- Lua libraries: lua-resty-jwt (installed from GitHub source)

## ngx_otel_module Configuration

The ngx_otel_module is built during the Docker image construction and installed to `/usr/local/openresty/nginx/modules/ngx_otel_module.so`.

To use the module in nginx configuration, add the following directives:

```nginx
# Load the module
load_module modules/ngx_otel_module.so;

http {
    # Configure OTEL exporter
    otel_exporter {
        endpoint "your-otel-collector:4317";  # gRPC endpoint
    }

    # Enable tracing globally or per-location
    otel_trace on;
    otel_service_name your-service-name;
    otel_trace_context propagate;
}
```

**Note:** The ngx_otel_module currently only supports gRPC export (port 4317), not HTTP (port 4318).

## Layer Caching

The Dockerfile is structured to optimize Docker layer caching:

1. **System dependencies** - Cached until base image or package list changes
2. **OpenSSL & PCRE** - Cached until versions change
3. **lua-resty-hmac** - Cached until git repo changes
4. **OpenResty** - Cached until version changes (built with --with-compat flag)
5. **Nginx source** - Cached until version changes (required for module compilation)
6. **ngx_otel_module** - Compiled from source and installed
7. **LuaRocks** - Cached until version changes
8. **Lua dependencies** - Cached until luarocks packages change
9. **Application files** - Rebuilt when source files change

This structure ensures fast rebuilds when only application code changes.

## Migration from OpenTelemetry WebServer SDK

This build replaces the previous OpenTelemetry WebServer SDK implementation with ngx_otel_module for better compatibility. Key changes:

1. **Module compilation**: ngx_otel_module is compiled from source against the matching nginx version
2. **No external SDK**: Removed dependency on opentelemetry-webserver-sdk
3. **gRPC only**: The module currently supports gRPC export only (not HTTP)
4. **Configuration syntax**: Uses standard nginx directives instead of NginxModule* directives

## Requirements

The ngx_otel_module requires:
- nginx configured with `--with-compat` flag (included in this build)
- pkg-config, libc-ares-dev, libre2-dev for gRPC support (included in dependencies)
