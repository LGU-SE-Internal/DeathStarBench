# Media Microservices Helm Chart

This Helm chart deploys the Media Microservices application on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- kubectl configured to communicate with your cluster

## Installation

### Basic Installation

To install the chart with the release name `media-microservices` in the `media` namespace:

```bash
helm install media-microservices . -n media --create-namespace
```

### Custom Installation

You can override default values during installation:

```bash
helm install media-microservices . -n media \
  --set global.replicas=2 \
  --set global.otel.endpoint=http://otel-collector.observability.svc.cluster.local:4318
```

## Data Initialization

The chart includes an automatic data initialization job (`data-init-job`) that:
1. Waits for nginx-web-server to be ready
2. Fetches movie and cast data from the DeathStarBench repository
3. Populates MongoDB with movie information (titles, casts, plots, ratings)
4. Registers test users (username_1 through username_1000)

This job runs automatically after installation using Helm hooks with a **15-minute timeout** (900 seconds). The job registers users in parallel batches for faster execution. You can check its status:

```bash
kubectl get jobs -n media
kubectl logs -n media job/data-init-job -f
```

**Note**: The initialization process may take 5-10 minutes depending on cluster resources and network conditions.

### Disabling Data Initialization

If you want to skip automatic data initialization:

```bash
helm install media-microservices . -n media --set data-init-job.enabled=false
```

### Manual Data Initialization

If you disabled the automatic initialization or need to re-initialize data:

```bash
# Forward port to access nginx-web-server
kubectl port-forward -n media svc/nginx-web-server 8080:8080

# In another terminal, run the initialization script
cd ../..  # Go to mediaMicroservices root
python3 scripts/write_movie_info.py \
  -c datasets/tmdb/casts.json \
  -m datasets/tmdb/movies.json \
  --server_address http://localhost:8080

# Register users
scripts/register_users.sh
```

## Running Load Tests

After the data is initialized, you can run wrk2 load tests:

```bash
# Forward nginx port
kubectl port-forward -n media svc/nginx-web-server 8080:8080

# Run compose-review workload
cd ../wrk2
./wrk -D exp -t 2 -c 2 -d 30 -L -s ../mediaMicroservices/wrk2/scripts/media-microservices/compose-review.lua http://localhost:8080/wrk2-api/review/compose -R 10
```

## Verification

To verify that the services are running and data is initialized:

```bash
# Check all pods are running
kubectl get pods -n media

# Check data initialization job completed
kubectl get jobs -n media

# Test API endpoint
kubectl port-forward -n media svc/nginx-web-server 8080:8080
curl -X POST -d "username=username_1&password=password_1&title=Avengers: Endgame&rating=5&text=Great movie!" \
  http://localhost:8080/wrk2-api/review/compose
```

## Uninstallation

To uninstall/delete the `media-microservices` deployment:

```bash
helm uninstall media-microservices -n media
```

## Configuration

The following table lists the main configurable parameters:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.replicas` | Default number of replicas for services | `1` |
| `global.imagePullPolicy` | Image pull policy | `IfNotPresent` |
| `global.dockerRegistry` | Docker registry for images | `10.10.10.240/library` |
| `global.otel.endpoint` | OpenTelemetry collector endpoint | `opentelemetry-kube-stack-deployment-collector.monitoring:4317` |
| `global.otel.samplerParam` | OpenTelemetry sampling rate | `1` |
| `data-init-job.enabled` | Enable automatic data initialization | `true` |
| `data-init-job.serverAddress` | API server address for initialization | `http://nginx-web-server.{namespace}.svc.cluster.local:8080` |
| `data-init-job.job.activeDeadlineSeconds` | Job timeout in seconds | `900` (15 minutes) |
| `data-init-job.job.backoffLimit` | Job retry attempts | `4` |

## Troubleshooting

### Data Initialization Job Times Out

If the job times out during installation:

```bash
# Check job status and logs
kubectl get jobs -n media
kubectl logs -n media job/data-init-job -f

# If needed, increase the timeout and reinstall
helm install media-microservices ./helm-chart/mediamicroservices -n media \
  --set data-init-job.job.activeDeadlineSeconds=1200
```

### Data Initialization Job Fails

Check the job logs:
```bash
kubectl logs -n media job/data-init-job
kubectl logs -n media job/data-init-job -c fetch-datasets
```

### Movies Not Found Errors

If you see "Movie X is not found in MongoDB" errors, the data initialization may not have completed:
1. Check if the data-init-job completed successfully
2. Manually re-run the data initialization
3. Check movie-id-service logs for errors

### Services Not Ready

If the data-init-job is stuck waiting for nginx:
```bash
kubectl get pods -n media
kubectl describe pod -n media <nginx-pod-name>
```

## Architecture

The chart deploys the following services:
- nginx-web-server (API gateway)
- unique-id-service
- movie-id-service (with MongoDB and Memcached)
- text-service
- rating-service (with Redis)
- user-service (with MongoDB and Memcached)
- compose-review-service (with Memcached)
- review-storage-service (with MongoDB and Memcached)
- user-review-service (with MongoDB and Redis)
- movie-review-service (with MongoDB and Redis)
- cast-info-service (with MongoDB and Memcached)
- page-service
- plot-service (with MongoDB and Memcached)
- movie-info-service (with MongoDB and Memcached)

Plus the data-init-job for automatic database initialization.
