# Build Instructions for OpenTelemetry Migration

This document provides instructions for building the Docker images after the OpenTelemetry migration.

## Prerequisites

- Docker installed and running
- Sufficient disk space for building images (20-30GB recommended)

## Quick Start

The Dockerfiles now include all dependencies in a single build, eliminating the need for separate dependency images.

### Using the Build Script (Recommended)

```bash
# Build both services with default registry (10.10.10.240/library)
./build-docker.sh

# Build with custom registry
DOCKER_REGISTRY=your-registry.com/project ./build-docker.sh

# Build with custom tag
DOCKER_TAG=v1.0 ./build-docker.sh

# Build only social network
./build-docker.sh --social-only

# Build only media microservices
./build-docker.sh --media-only
```

### Manual Build

#### For Social Network:
```bash
cd socialNetwork
docker build -t 10.10.10.240/library/social-network-microservices:otel .
```

#### For Media Microservices:
```bash
cd mediaMicroservices
docker build -t 10.10.10.240/library/media-microservices:otel .
```

## Build Details

Each Dockerfile now performs a complete build in a single stage:

1. **Dependency Installation** - Compiles and installs all required libraries:
   - MongoDB C driver
   - Apache Thrift
   - nlohmann/json
   - yaml-cpp
   - OpenTracing C++
   - Jaeger client C++
   - **OpenTelemetry C++ SDK** (with OTLP and Jaeger exporters)
   - JWT, Redis, AMQP libraries

2. **Service Build** - Compiles the microservices using the installed dependencies

3. **Final Image** - Creates a minimal runtime image with only binaries and libraries

## Build Time

**First build:** 30-60 minutes per service (compiles all dependencies from source)

**Subsequent builds:** Much faster due to Docker layer caching (unless dependencies change)

## Verifying the Build

After building, verify the images include OpenTelemetry:

```bash
# Check if OpenTelemetry libraries are present
docker run --rm 10.10.10.240/library/social-network-microservices:otel \
    ls -la /usr/local/lib/ | grep opentelemetry

# Expected output should show files like:
# libopentelemetry_trace.so
# libopentelemetry_exporter_jaeger_trace.so
# libopentelemetry_exporter_otlp_http.so
# libopentelemetry_resources.so
# libopentracing_shim.so
```

## Pushing to Registry

```bash
# Push social network image
docker push 10.10.10.240/library/social-network-microservices:otel

# Push media microservices image
docker push 10.10.10.240/library/media-microservices:otel
```

## Troubleshooting

### Build fails with "failed to download"
Network issues downloading dependencies. Check your internet connection and retry.

### Build takes too long
Building from source takes time. This is expected on first build. Use `docker build --no-cache` if you need to force a fresh build.

### Out of disk space
Building these images requires significant disk space. Clean up unused Docker images:
```bash
docker system prune -a
```

### Build fails with compilation errors
Ensure you have sufficient RAM (8GB+ recommended) for parallel compilation. You can reduce parallel jobs by modifying `-j$(nproc)` to `-j2` in the Dockerfiles.

## CI/CD Integration

For CI/CD pipelines:

1. **Use Docker Build Cache**: Configure your CI to cache Docker layers between builds
2. **Build Arguments**: Pass build arguments for version pins if needed
3. **Multi-stage**: The Dockerfiles use multi-stage builds to keep final images small

Example GitLab CI configuration:
```yaml
build:
  script:
    - docker build -t $CI_REGISTRY/social-network-microservices:$CI_COMMIT_TAG socialNetwork/
    - docker push $CI_REGISTRY/social-network-microservices:$CI_COMMIT_TAG
  cache:
    paths:
      - /var/lib/docker
```

## Customization

### Change Library Versions

Edit the `ARG` declarations at the top of the Dockerfile:

```dockerfile
ARG LIB_OPENTELEMETRY_VERSION=1.14.2  # Change to desired version
```

### Optimize Build Time

Use BuildKit for faster builds:
```bash
DOCKER_BUILDKIT=1 docker build -t image-name .
```

## Differences from Previous Approach

**Before:** Required building separate dependency images first, then service images.

**Now:** Single Dockerfile builds everything in one go, simplifying the build process and making it easier to customize dependency versions.
