# Load Generator for Media Microservices

This Helm chart deploys a load generator for the Media Microservices using wrk2.

## Prerequisites

- Kubernetes cluster
- Helm 3.x
- Media Microservices deployed
- Movie data initialized

## Building the Load Generator Image

The Dockerfile builds a modified wrk2 with native continuous mode support (duration=0 for infinite running).

```bash
cd mediaMicroservices
docker build -f Dockerfile-loader -t 10.10.10.240/library/mediamicroservices-loader:latest .
docker push 10.10.10.240/library/mediamicroservices-loader:latest
```

**Note:** The wrk2 source in this repository has been modified to support infinite duration when `duration: "0"` is set. See `/CUSTOM_WRK2.md` for details on the modifications.

## Installation

The load generator is included as a subchart in the Media Microservices Helm chart. You can enable/disable it through the main chart values:

```bash
# Install with load generator enabled (default)
helm install media ./helm-chart/mediamicroservices -n media --create-namespace

# Disable load generator
helm install media ./helm-chart/mediamicroservices -n media \
  --set load-generator.enabled=false
```

## Configuration

The load generator can be configured through the `values.yaml` file:

```yaml
loadTest:
  enabled: true
  targetUrl: "http://nginx-web-server:8080"
  targetHost: "nginx-web-server"
  targetPort: 8080
  script: "compose-review.lua"
  endpoint: "/wrk2-api/review/compose"
  threads: 2
  connections: 2
  duration: "30s"
  rate: 10
  additionalArgs: "-D exp -L"
  continuous: false  # Set to true for continuous load testing
```

### Available Scripts

- `compose-review.lua` - Test composing movie reviews

### Configuration Parameters

- `enabled`: Enable/disable the load generator
- `targetUrl`: Target service URL (nginx-web-server service)
- `targetHost`: Target service hostname for health checks
- `targetPort`: Target service port for health checks
- `script`: Lua script to use for load generation
- `endpoint`: API endpoint path
- `threads`: Number of threads for wrk2
- `connections`: Number of connections to keep open
- `duration`: Duration of each test cycle (e.g., 30s, 5m, 1h)
- `rate`: Requests per second
- `additionalArgs`: Additional arguments to pass to wrk2
- `continuous`: If true, runs load test continuously in a loop with automatic restarts on failure. If false, runs once then sleeps.

## Customizing the Load Test

To customize the load test parameters:

```bash
helm upgrade media ./helm-chart/mediamicroservices -n media \
  --set load-generator.loadTest.threads=10 \
  --set load-generator.loadTest.connections=100 \
  --set load-generator.loadTest.rate=100 \
  --set load-generator.loadTest.duration=5m
```

### Enable Continuous Load Testing

To run the load test continuously (it will keep running and restart automatically on failure):

```bash
helm upgrade media ./helm-chart/mediamicroservices -n media \
  --set load-generator.loadTest.continuous=true
```

In continuous mode:
- The load test runs in an infinite loop
- Each cycle runs for the specified `duration`
- If a cycle fails, it waits 5 seconds before restarting
- If a cycle succeeds, it immediately starts the next cycle
- The pod will automatically restart if it crashes

## Viewing Logs

To view the load generator logs:

```bash
kubectl logs -n media -l app=load-generator -f
```

## Uninstallation

The load generator will be removed when the parent chart is uninstalled:

```bash
helm uninstall media -n media
```
