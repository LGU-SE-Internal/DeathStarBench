# Testing the Data Initialization Job

This document describes how to test the data initialization job for the media microservices Helm chart.

## Prerequisites

- Kubernetes cluster (minikube, kind, or cloud provider)
- kubectl configured
- Helm 3.x installed

## Test Deployment

### 1. Deploy the Chart

```bash
# Create namespace
kubectl create namespace media

# Install the chart
cd mediaMicroservices/helm-chart/mediamicroservices
helm install media-microservices . -n media

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app!=data-init-job -n media --timeout=300s
```

### 2. Monitor the Data Initialization Job

```bash
# Check job status
kubectl get jobs -n media
kubectl describe job data-init-job -n media

# Watch job progress
kubectl logs -n media job/data-init-job -f

# Check init container logs (if job is still running)
kubectl logs -n media job/data-init-job -c fetch-datasets
```

### 3. Verify Data Was Loaded

```bash
# Check movie-id-mongodb has data
kubectl exec -n media deployment/movie-id-mongodb -- mongo --eval "db.movie.count()" movie-id

# Should return a number > 0 if data was loaded successfully
```

### 4. Test with wrk2

```bash
# Forward nginx port
kubectl port-forward -n media svc/nginx-web-server 8080:8080 &

# Wait a moment for port forwarding to be established
sleep 5

# Test compose review API (should not get "movie not found" errors)
cd ../../../wrk2
make
./wrk -D exp -t 1 -c 1 -d 10 -L -s ../mediaMicroservices/wrk2/scripts/media-microservices/compose-review.lua http://localhost:8080/wrk2-api/review/compose -R 1

# Kill port forward
pkill -f "port-forward"
```

### 5. Check for Errors

```bash
# Check nginx logs for "unknown reason" errors
kubectl logs -n media deployment/nginx-web-server --tail=100 | grep "unknown reason"

# Check movie-id-service logs for "not found in MongoDB" errors
kubectl logs -n media deployment/movie-id-service --tail=100 | grep "not found in MongoDB"

# If no errors appear, the initialization was successful!
```

## Expected Results

✅ **Success indicators:**
- `data-init-job` completes with status "Completed"
- No "Movie X is not found in MongoDB" errors in movie-id-service logs
- No "unknown reason" errors in nginx logs
- wrk2 tests run without errors

❌ **Failure indicators:**
- Job status shows "Error" or "Failed"
- Movie-id-service logs show "not found in MongoDB" errors
- nginx logs show "unknown reason" errors

## Troubleshooting

### Job Stuck in Pending

Check if init containers are running:
```bash
kubectl describe pod -n media -l job-name=data-init-job
```

### Job Fails to Fetch Datasets

Check fetch-datasets init container logs:
```bash
kubectl logs -n media -l job-name=data-init-job -c fetch-datasets
```

### Job Fails to Initialize Data

Check the main container logs:
```bash
kubectl logs -n media -l job-name=data-init-job -c data-init
```

### Nginx Not Ready

Check nginx pod status:
```bash
kubectl get pods -n media -l service=nginx-web-server
kubectl describe pod -n media -l service=nginx-web-server
```

## Cleanup

```bash
# Uninstall the chart
helm uninstall media-microservices -n media

# Delete the namespace
kubectl delete namespace media
```

## Re-running Initialization

If you need to re-run the data initialization:

```bash
# Option 1: Upgrade the chart (triggers post-upgrade hook)
helm upgrade media-microservices . -n media

# Option 2: Manually run the job
kubectl delete job data-init-job -n media
kubectl create -f <(helm template media-microservices . -n media | grep -A 100 "kind: Job")
```
