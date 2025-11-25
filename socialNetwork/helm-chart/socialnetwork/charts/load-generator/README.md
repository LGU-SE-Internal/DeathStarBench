# Load Generator for Social Network

This Helm chart deploys a load generator for the Social Network microservices using wrk2.

## Prerequisites

- Kubernetes cluster
- Helm 3.x
- Social Network services deployed
- Social graph initialized (users and relationships)

## Building the Load Generator Image

The Dockerfile builds a modified wrk2 with native continuous mode support (duration=0 for infinite running).

```bash
cd socialNetwork
docker build -f Dockerfile-loader -t 10.10.10.240/library/socialnetwork-loader:latest .
docker push 10.10.10.240/library/socialnetwork-loader:latest
```

**Note:** The wrk2 source in this repository has been modified to support infinite duration when `duration: "0"` is set. See `/CUSTOM_WRK2.md` for details on the modifications.

## Installation

The load generator is included as a subchart in the Social Network Helm chart. You can enable/disable it through the main chart values:

```bash
# Install with load generator enabled (default)
helm install socialnetwork ./helm-chart/socialnetwork -n socialnetwork --create-namespace

# Disable load generator
helm install socialnetwork ./helm-chart/socialnetwork -n socialnetwork \
  --set load-generator.enabled=false
```

## Configuration

The load generator can be configured through the `values.yaml` file:

```yaml
loadTest:
  enabled: true
  targetUrl: "http://nginx-thrift:8080"
  targetHost: "nginx-thrift"
  targetPort: 8080
  script: "mixed-workload.lua"  # Default: comprehensive mixed workload
  endpoint: "/wrk2-api/post/compose"  # Not used by mixed-workload.lua
  threads: 2
  connections: 2
  duration: "0"  # Use "0" for infinite continuous mode, or "30s", "5m", etc. for finite duration
  rate: 10
  additionalArgs: "-D exp -L"
```

### Available Scripts

- `mixed-workload.lua` - **Recommended (Default)** - Comprehensive mixed workload covering 9+ microservices:
  - 50% Read home timeline (HomeTimelineService, PostStorageService, SocialGraphService)
  - 25% Read user timeline (UserTimelineService, PostStorageService)
  - 10% Compose post (ComposePostService, TextService, MediaService, UserMentionService, UrlShortenService, UniqueIdService, WriteHomeTimelineService, UserTimelineService)
  - 10% Follow user (SocialGraphService, UserService)
  - 5% Unfollow user (SocialGraphService, UserService)
- `compose-post.lua` - Write-only workload for composing posts
- `read-home-timeline.lua` - Read-only workload for home timeline
- `read-user-timeline.lua` - Read-only workload for user timeline

### Configuration Parameters

- `enabled`: Enable/disable the load generator
- `targetUrl`: Target service URL (nginx-thrift service)
- `targetHost`: Target service hostname for health checks
- `targetPort`: Target service port for health checks
- `script`: Lua script to use for load generation
- `endpoint`: API endpoint path (only used by single-operation scripts like compose-post.lua)
- `threads`: Number of threads for wrk2
- `connections`: Number of connections to keep open
- `duration`: Duration of the test. Use `"0"` for infinite continuous mode (modified wrk2), or specify duration like `"30s"`, `"5m"`, `"1h"` for finite tests
- `rate`: Requests per second
- `additionalArgs`: Additional arguments to pass to wrk2

## Customizing the Load Test

### Using Mixed Workload (Recommended)

The default configuration uses the comprehensive mixed workload. To customize parameters:

```bash
helm upgrade socialnetwork ./helm-chart/socialnetwork -n socialnetwork \
  --set load-generator.loadTest.threads=10 \
  --set load-generator.loadTest.connections=100 \
  --set load-generator.loadTest.rate=100 \
  --set load-generator.loadTest.duration=0  # 0 for infinite, or specify like "5m"
```

### Using Single-Operation Scripts

To switch to a single-operation script like read-home-timeline:

```bash
helm upgrade socialnetwork ./helm-chart/socialnetwork -n socialnetwork \
  --set load-generator.loadTest.script=read-home-timeline.lua \
  --set load-generator.loadTest.endpoint=/wrk2-api/home-timeline/read
```

### Enable Continuous Load Testing

The modified wrk2 supports native continuous mode. Set `duration: "0"` for infinite running:

```bash
helm upgrade socialnetwork ./helm-chart/socialnetwork -n socialnetwork \
  --set load-generator.loadTest.duration=0
```

With `duration: "0"`:
- wrk2 runs continuously without stopping
- Single process handles infinite load generation
- Pod will automatically restart if it crashes (Kubernetes deployment behavior)

## Viewing Logs

To view the load generator logs:

```bash
kubectl logs -n socialnetwork -l app=load-generator -f
```

## Uninstallation

The load generator will be removed when the parent chart is uninstalled:

```bash
helm uninstall socialnetwork -n socialnetwork
```
