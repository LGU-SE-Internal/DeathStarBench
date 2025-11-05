# OpenResty with OpenTelemetry - Build Instructions

This directory contains the Dockerfile for building an OpenResty image with OpenTelemetry support.

## Build Command

To build the image, run the following command from this directory:

```bash
docker build -f Dockerfile -t openresty-thrift:xenial .
```

Or with a custom registry:

```bash
docker build -f Dockerfile -t your-registry/openresty-thrift:xenial .
```

## Build Context

The Dockerfile uses `../` relative paths to access files in the parent directory:
- `../lua-thrift` - Lua Thrift libraries
- `../lua-bridge-tracer` - OpenTracing bridge tracer for Lua
- `../lua-json` - JSON library for Lua
- `../nginx.conf` - Nginx configuration file
- `../nginx.vh.default.conf` - Default virtual host configuration

## Layer Caching

The Dockerfile is structured to optimize Docker layer caching:

1. **System dependencies** - Cached until base image or package list changes
2. **OpenSSL & PCRE** - Cached until versions change
3. **lua-resty-hmac** - Cached until git repo changes
4. **OpenTelemetry SDK** - Cached until version changes
5. **OpenResty** - Cached until version changes
6. **LuaRocks** - Cached until version changes
7. **Lua dependencies** - Cached until luarocks packages change
8. **Application files** - Rebuilt when source files change

This structure ensures fast rebuilds when only application code changes.

## Features

- OpenResty 1.25.3.1 (nginx 1.25.3)
- OpenTelemetry WebServer SDK v1.0.3
- OpenSSL 1.1.1w
- PCRE 8.45
- LuaRocks 3.9.2
- Lua libraries: lua-resty-jwt (installed from GitHub source)
- Built from source: liblualongnumber, libluabitwise, libluabpack (from lua-thrift)

## Environment Variables

The image sets the following environment variables:
- `PATH` - Includes OpenResty binaries
- `LD_LIBRARY_PATH` - Includes OpenTelemetry SDK libraries
- `LUA_PATH` - Lua module search paths
- `LUA_CPATH` - Lua C module search paths
