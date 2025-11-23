# Load Generator for Social Network

This Helm chart deploys a load generator for the Social Network microservices using wrk2.

## Prerequisites

- Kubernetes cluster
- Helm 3.x
- Social Network services deployed
- Social graph initialized (users and relationships)

## Building the Load Generator Image

The Social Network already has a Dockerfile-loader. Build and push the image:

```bash
cd socialNetwork
docker build -f Dockerfile-loader -t 10.10.10.240/library/socialnetwork-loader:latest .
docker push 10.10.10.240/library/socialnetwork-loader:latest
```

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
  script: "compose-post.lua"
  endpoint: "/wrk2-api/post/compose"
  threads: 2
  connections: 2
  duration: "30s"
  rate: 10
  additionalArgs: "-D exp -L"
```

### Available Scripts

- `compose-post.lua` - Test composing posts
- `read-home-timeline.lua` - Test reading home timeline
- `read-user-timeline.lua` - Test reading user timeline
- `mixed-workload.lua` - Mixed workload with multiple operations

### Configuration Parameters

- `enabled`: Enable/disable the load generator
- `targetUrl`: Target service URL (nginx-thrift service)
- `script`: Lua script to use for load generation
- `endpoint`: API endpoint path
- `threads`: Number of threads for wrk2
- `connections`: Number of connections to keep open
- `duration`: Duration of the test (e.g., 30s, 5m, 1h)
- `rate`: Requests per second
- `additionalArgs`: Additional arguments to pass to wrk2

## Customizing the Load Test

To change the workload script and parameters:

```bash
helm upgrade socialnetwork ./helm-chart/socialnetwork -n socialnetwork \
  --set load-generator.loadTest.script=mixed-workload.lua \
  --set load-generator.loadTest.threads=10 \
  --set load-generator.loadTest.connections=100 \
  --set load-generator.loadTest.rate=100 \
  --set load-generator.loadTest.duration=5m
```

For read-home-timeline:

```bash
helm upgrade socialnetwork ./helm-chart/socialnetwork -n socialnetwork \
  --set load-generator.loadTest.script=read-home-timeline.lua \
  --set load-generator.loadTest.endpoint=/wrk2-api/home-timeline/read
```

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
