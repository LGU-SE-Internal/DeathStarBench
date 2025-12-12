# Load Generator for Hotel Reservation

This Helm chart deploys a load generator for the Hotel Reservation microservices using wrk2.

## Prerequisites

- Kubernetes cluster
- Helm 3.x
- Hotel Reservation services deployed

## Building the Load Generator Image

The Dockerfile builds a modified wrk2 with native continuous mode support (duration=0 for infinite running).

```bash
cd hotelReservation
docker build -f Dockerfile-loader -t 10.10.10.240/library/hotelreservation-loader:latest .
docker push 10.10.10.240/library/hotelreservation-loader:latest
```

**Note:** The wrk2 source in this repository has been modified to support infinite duration when `duration: "0"` is set. See `/CUSTOM_WRK2.md` for details on the modifications.

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
  targetServicePrefix: "frontend"  # Service name to connect to
  targetPort: 5000
  script: "mixed-workload_type_1_with_attractions.lua"  # Default: comprehensive mixed workload
  threads: 2
  connections: 2
  duration: "0"  # Use "0" for infinite continuous mode, or "30s", "5m", etc. for finite duration
  rate: 10
  additionalArgs: "-D exp -L"
```

### Available Scripts

- `mixed-workload_type_1_with_attractions.lua` - **Recommended (Default)** - Comprehensive mixed workload covering most services:
  - 50% Search hotels (geo, profile, rate services)
  - 30% Recommendations (rate, geo services)
  - 0.5% User login (user service)
  - 0.5% Reserve (reservation service)
  - 6% Restaurants (attractions service)
  - 6% Museums (attractions service)
  - 7% Cinemas (attractions service)
- `mixed-workload_type_1.lua` - Standard mixed workload without attraction services

### Configuration Parameters

- `enabled`: Enable/disable the load generator
- `targetServicePrefix`: Service name to connect to (e.g., `frontend`)
- `targetPort`: Target service port for health checks
- `script`: Lua script to use for load generation
- `threads`: Number of threads for wrk2
- `connections`: Number of connections to keep open
- `duration`: Duration of the test. Use `"0"` for infinite continuous mode (modified wrk2), or specify duration like `"30s"`, `"5m"`, `"1h"` for finite tests
- `rate`: Requests per second
- `additionalArgs`: Additional arguments to pass to wrk2

## Customizing the Load Test

### Using Mixed Workload (Recommended)

The default configuration uses the comprehensive mixed workload. To customize parameters:

```bash
helm upgrade hotelreservation ./helm-chart/hotelreservation -n hotelreservation \
  --set load-generator.loadTest.threads=10 \
  --set load-generator.loadTest.connections=100 \
  --set load-generator.loadTest.rate=100 \
  --set load-generator.loadTest.duration=0  # 0 for infinite, or specify like "5m"
```

### Enable Continuous Load Testing

The modified wrk2 supports native continuous mode. Set `duration: "0"` for infinite running:

```bash
helm upgrade hotelreservation ./helm-chart/hotelreservation -n hotelreservation \
  --set load-generator.loadTest.duration=0
```

With `duration: "0"`:
- wrk2 runs continuously without stopping
- Single process handles infinite load generation
- Pod will automatically restart if it crashes (Kubernetes deployment behavior)

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
