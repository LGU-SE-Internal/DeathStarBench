# Build Instructions for OpenTelemetry Migration

This document provides instructions for building the Docker images after the OpenTelemetry migration.

## Prerequisites

- Docker installed and running
- Sufficient disk space for building images

## Building Order

Due to the dependency on updated libraries (OpenTelemetry C++ SDK), you must build the images in the following order:

### 1. Build Dependency Images

First, build the updated dependency images that include OpenTelemetry C++ SDK:

#### For Social Network:
```bash
cd socialNetwork/docker/thrift-microservice-deps/cpp
docker build -t deathstarbench/social-network-microservices-deps:latest .
```

#### For Media Microservices:
```bash
cd mediaMicroservices/docker/thrift-microservice-deps/cpp
docker build -t deathstarbench/media-microservices-deps:latest .
```

**Note:** Building these dependency images can take 30-60 minutes depending on your machine as they compile OpenTelemetry C++ SDK, Thrift, MongoDB drivers, and other dependencies from source.

### 2. Build Service Images

After the dependency images are built, you can build the main service images:

#### For Social Network:
```bash
cd socialNetwork
docker build -t social-network-microservices:otel .
# Or with your registry:
docker build -t 10.10.10.240/library/social-network-microservices:otel .
```

#### For Media Microservices:
```bash
cd mediaMicroservices
docker build -t media-microservices:otel .
# Or with your registry:
docker build -t 10.10.10.240/library/media-microservices:otel .
```

## Quick Build Script

You can use the following script to build all images:

```bash
#!/bin/bash
set -e

echo "Building dependency images..."

# Build social network dependencies
cd socialNetwork/docker/thrift-microservice-deps/cpp
docker build -t deathstarbench/social-network-microservices-deps:latest .

# Build media microservices dependencies
cd ../../../../mediaMicroservices/docker/thrift-microservice-deps/cpp
docker build -t deathstarbench/media-microservices-deps:latest .

echo "Building service images..."

# Build social network services
cd ../../../../socialNetwork
docker build -t social-network-microservices:otel .

# Build media microservices
cd ../mediaMicroservices
docker build -t media-microservices:otel .

echo "Build complete!"
```

## Alternative: Using Pre-built Base Images

If you have access to the old base images and want to avoid rebuilding dependencies, you can temporarily modify the Dockerfiles:

1. In `socialNetwork/Dockerfile`, change:
   ```dockerfile
   FROM deathstarbench/social-network-microservices-deps:latest
   ```
   back to:
   ```dockerfile
   FROM yg397/thrift-microservice-deps:xenial
   ```

2. In `mediaMicroservices/Dockerfile`, change:
   ```dockerfile
   FROM deathstarbench/media-microservices-deps:latest
   ```
   back to:
   ```dockerfile
   FROM yg397/thrift-microservice-deps:xenial
   ```

**Warning:** This will cause the build to fail because the old base images don't have OpenTelemetry C++ SDK. You must build the new dependency images first.

## Troubleshooting

### Build fails with "opentelemetry header not found"
This means the dependency image wasn't built or isn't available. Build the dependency images first (step 1 above).

### Build takes too long
The dependency images compile many libraries from source. This is expected. You only need to build them once, and they can be cached for future builds.

### Out of disk space
Building these images requires significant disk space (10-20GB). Clean up unused Docker images with:
```bash
docker system prune -a
```

## CI/CD Integration

For CI/CD pipelines, you can:

1. Build and push dependency images to your registry once:
   ```bash
   docker build -t your-registry/social-network-deps:otel socialNetwork/docker/thrift-microservice-deps/cpp/
   docker push your-registry/social-network-deps:otel
   ```

2. Update Dockerfiles to use your registry images:
   ```dockerfile
   FROM your-registry/social-network-deps:otel
   ```

3. Build service images in your pipeline using the pre-built dependencies.

## Verifying the Build

After building, verify the images include OpenTelemetry:

```bash
# Check if OpenTelemetry libraries are present
docker run --rm social-network-microservices:otel ls -la /usr/local/lib/ | grep opentelemetry

# Expected output should show files like:
# libopentelemetry_trace.so
# libopentelemetry_exporter_jaeger_trace.so
# libopentelemetry_exporter_otlp_http.so
# etc.
```
