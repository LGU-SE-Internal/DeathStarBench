# Load Generator for Hotel Reservation

This Helm chart deploys a load generator for the Hotel Reservation microservices using wrk2.

## Prerequisites

- Kubernetes cluster
- Helm 3.x
- Hotel Reservation services deployed

## Building the Load Generator Image

First, build and push the load generator Docker image:

```bash
cd hotelReservation
docker build -f Dockerfile-loader -t 10.10.10.240/library/hotelreservation-loader:latest .
docker push 10.10.10.240/library/hotelreservation-loader:latest
```

## Installation

The load generator is included as a subchart in the Hotel Reservation Helm chart. You can enable/disable it through the main chart values:

```bash
# Install with load generator enabled (default)
helm install hotelreservation ./helm-chart/hotelreservation -n hotelreservation --create-namespace

# Disable load generator
helm install hotelreservation ./helm-chart/hotelreservation -n hotelreservation \
  --set load-generator.enabled=false
```

## Configuration

The load generator can be configured through the `values.yaml` file:

```yaml
loadTest:
  enabled: true
  targetUrl: "http://frontend:5000"
  targetHost: "frontend"
  targetPort: 5000
  script: "mixed-workload_type_1_with_attractions.lua"  # Use the with_attractions version
  threads: 2
  connections: 2
  duration: "30s"
  rate: 10
  additionalArgs: "-D exp -L"
  continuous: false  # Set to true for continuous load testing
```

### Available Scripts

- `mixed-workload_type_1.lua` - Standard mixed workload
- `mixed-workload_type_1_with_attractions.lua` - Mixed workload including attractions (restaurants, museums, cinemas)

### Configuration Parameters

- `enabled`: Enable/disable the load generator
- `targetUrl`: Target service URL (frontend service)
- `targetHost`: Target service hostname for health checks
- `targetPort`: Target service port for health checks
- `script`: Lua script to use for load generation
- `threads`: Number of threads for wrk2
- `connections`: Number of connections to keep open
- `duration`: Duration of each test cycle (e.g., 30s, 5m, 1h)
- `rate`: Requests per second
- `additionalArgs`: Additional arguments to pass to wrk2
- `continuous`: If true, runs load test continuously in a loop with automatic restarts on failure. If false, runs once then sleeps.

## Customizing the Load Test

To customize the load test parameters:

```bash
helm upgrade hotelreservation ./helm-chart/hotelreservation -n hotelreservation \
  --set load-generator.loadTest.threads=10 \
  --set load-generator.loadTest.connections=100 \
  --set load-generator.loadTest.rate=100 \
  --set load-generator.loadTest.duration=5m
```

### Enable Continuous Load Testing

To run the load test continuously (it will keep running and restart automatically on failure):

```bash
helm upgrade hotelreservation ./helm-chart/hotelreservation -n hotelreservation \
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
kubectl logs -n hotelreservation -l app=load-generator -f
```

## Uninstallation

The load generator will be removed when the parent chart is uninstalled:

```bash
helm uninstall hotelreservation -n hotelreservation
```
