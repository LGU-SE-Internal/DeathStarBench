# Media Microservices Helm Chart

本 Helm chart 用于在 Kubernetes 上部署 Media Microservices 应用。

## 前置要求

- Kubernetes 1.19+
- Helm 3.2.0+
- kubectl 已配置可以与集群通信

## 安装

### 基础安装

在 `media` 命名空间安装名为 `media-microservices` 的 release：

```bash
helm install media-microservices . -n media --create-namespace
```

### 自定义安装

可以在安装时覆盖默认值：

```bash
helm install media-microservices . -n media \
  --set global.replicas=2 \
  --set global.otel.endpoint=http://otel-collector.observability.svc.cluster.local:4318
```

## 数据初始化

该 chart 包含一个自动数据初始化 Job（`data-init-job`），它会：
1. 等待 nginx-web-server 就绪
2. 从 DeathStarBench 仓库获取电影和演员数据
3. 用电影信息（标题、演员、剧情、评分）填充 MongoDB
4. 注册测试用户（username_1 到 username_1000）

此 Job 在安装后使用 Helm hooks 自动运行，**超时时间为 15 分钟**（900 秒）。Job 会并行批量注册用户以提高执行速度。可以检查其状态：

```bash
kubectl get jobs -n media
kubectl logs -n media job/data-init-job -f
```

**注意**：初始化过程可能需要 5-10 分钟，具体取决于集群资源和网络状况。

### 禁用数据初始化

如果要跳过自动数据初始化：

```bash
helm install media-microservices . -n media --set data-init-job.enabled=false
```

### 手动数据初始化

如果禁用了自动初始化或需要重新初始化数据：

```bash
# 转发端口以访问 nginx-web-server
kubectl port-forward -n media svc/nginx-web-server 8080:8080

# 在另一个终端运行初始化脚本
cd ../..  # 回到 mediaMicroservices 根目录
python3 scripts/write_movie_info.py \
  -c datasets/tmdb/casts.json \
  -m datasets/tmdb/movies.json \
  --server_address http://localhost:8080

# 注册用户
scripts/register_users.sh
```

## 运行负载测试

数据初始化后，可以运行 wrk2 负载测试：

```bash
# 转发 nginx 端口
kubectl port-forward -n media svc/nginx-web-server 8080:8080

# 运行 compose-review 工作负载
cd ../wrk2
./wrk -D exp -t 2 -c 2 -d 30 -L -s ../mediaMicroservices/wrk2/scripts/media-microservices/compose-review.lua http://localhost:8080/wrk2-api/review/compose -R 10
```

## 验证

验证服务正在运行且数据已初始化：

```bash
# 检查所有 pod 是否运行
kubectl get pods -n media

# 检查数据初始化 Job 是否完成
kubectl get jobs -n media

# 测试 API 端点
kubectl port-forward -n media svc/nginx-web-server 8080:8080
curl -X POST -d "username=username_1&password=password_1&title=Avengers: Endgame&rating=5&text=Great movie!" \
  http://localhost:8080/wrk2-api/review/compose
```

## 卸载

卸载/删除 `media-microservices` 部署：

```bash
helm uninstall media-microservices -n media
```

## 配置参数

下表列出了主要的可配置参数：

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `global.replicas` | 服务的默认副本数 | `1` |
| `global.imagePullPolicy` | 镜像拉取策略 | `IfNotPresent` |
| `global.dockerRegistry` | 镜像的 Docker registry | `10.10.10.240/library` |
| `global.otel.endpoint` | OpenTelemetry collector 端点 | `opentelemetry-kube-stack-deployment-collector.monitoring:4317` |
| `global.otel.samplerParam` | OpenTelemetry 采样率 | `1` |
| `data-init-job.enabled` | 启用自动数据初始化 | `true` |
| `data-init-job.serverAddress` | 初始化的 API 服务器地址 | `http://nginx-web-server.{namespace}.svc.cluster.local:8080` |
| `data-init-job.job.hookTimeout` | Helm hook 超时时间（秒） | `1800`（30 分钟） |
| `data-init-job.job.activeDeadlineSeconds` | Job 超时时间（秒） | `1500`（25 分钟） |
| `data-init-job.job.backoffLimit` | Job 重试次数 | `4` |

## 故障排查

### 安装时数据初始化 Job 超时

**默认超时时间为 30 分钟（1800 秒）**。如果安装仍然超时失败：

**方法 1：先禁用自动初始化安装，然后手动运行**
```bash
# 禁用数据初始化安装
helm install media-microservices ./helm-chart/mediamicroservices -n media \
  --set data-init-job.enabled=false

# 等待所有服务就绪
kubectl wait --for=condition=ready pod -l app!=data-init-job -n media --timeout=300s

# 手动创建并运行初始化 Job
helm template media-microservices ./helm-chart/mediamicroservices -n media \
  --set data-init-job.enabled=true \
  --show-only charts/data-init-job/templates/job.yaml | \
  sed 's/helm.sh\/hook.*//' | kubectl apply -f -

# 监控 Job 执行
kubectl logs -n media job/data-init-job -f
```

**方法 2：显著增加超时时间**
```bash
helm install media-microservices ./helm-chart/mediamicroservices -n media \
  --set data-init-job.job.hookTimeout=3600 \
  --set data-init-job.job.activeDeadlineSeconds=3000
```

**方法 3：检查延迟原因**
```bash
# 检查 Job Pod 是否运行
kubectl get pods -n media -l app=data-init-job

# 查看 Job 日志
kubectl logs -n media job/data-init-job -f

# 检查服务是否就绪
kubectl get pods -n media
```

### 数据初始化 Job 失败

检查 Job 日志：
```bash
kubectl logs -n media job/data-init-job
kubectl logs -n media job/data-init-job -c fetch-datasets
```

### 找不到电影错误

如果看到 "Movie X is not found in MongoDB" 错误，数据初始化可能未完成：
1. 检查 data-init-job 是否成功完成
2. 手动重新运行数据初始化
3. 检查 movie-id-service 日志是否有错误

### 服务未就绪

如果 data-init-job 一直在等待 nginx：
```bash
kubectl get pods -n media
kubectl describe pod -n media <nginx-pod-name>
```

## 问题解决

**问题**: 部署后运行 wrk2 测试时 nginx 日志显示 "unknown reason" 错误，movie-id-service 日志显示 "Movie X is not found in MongoDB"

**原因**: Helm 部署后 MongoDB 数据库为空，没有运行数据初始化步骤

**解决方案**: 本 chart 已添加自动数据初始化 Job，会在部署后自动填充数据库。如果仍有问题：
1. 检查 `kubectl get jobs -n media` 确认 data-init-job 已完成
2. 查看日志 `kubectl logs -n media job/data-init-job` 检查是否有错误
3. 如需重新初始化，删除并重新安装 chart，或手动运行初始化脚本
