# 问题解决方案 - Media Microservices Helm 部署数据初始化

## 问题描述

使用 Helm 部署 media 微服务后，虽然所有 Pod 都正常运行（`kubectl get pods -n media` 显示所有 pod 状态为 Running），但运行 wrk2 测试时出现以下错误：

**nginx 日志错误：**
```
2025/11/08 11:33:42 [error] 466#466: *4352 lua user thread aborted: runtime error: unknown reason
stack traceback:
coroutine 0:
        [C]: in function 'error'
        /gen-lua/media_service_MovieIdService.lua:293: in function 'recv_UploadMovieId'
        /gen-lua/media_service_MovieIdService.lua:266: in function 'UploadMovieId'
        .../openresty/nginx/lua-scripts/wrk2-api/review/compose.lua:34: in function
```

**movie-id-service 日志错误：**
```
[2025-Nov-08 11:33:42.309825] <error>: (MovieIdHandler.h:166:UploadMovieId) Movie Deadfall is not found in MongoDB
[2025-Nov-08 11:33:42.422029] <error>: (MovieIdHandler.h:166:UploadMovieId) Movie Doctor Strange is not found in MongoDB
[2025-Nov-08 11:33:42.672707] <error>: (MovieIdHandler.h:166:UploadMovieId) Movie Dr. No is not found in MongoDB
[2025-Nov-08 11:33:42.766699] <error>: (MovieIdHandler.h:166:UploadMovieId) Movie Homeward Bound: The Incredible Journey is not found in MongoDB
```

## 根本原因

这个问题**不是因为 nginx 从 OpenTracing 迁移到 OpenTelemetry 导致的配置问题**，而是 Helm 部署缺少了一个关键的数据初始化步骤。

在原始的 Docker Compose 部署流程中：
1. 运行 `docker-compose up -d` 启动所有服务
2. **手动步骤**：运行以下命令初始化数据：
   ```bash
   python3 scripts/write_movie_info.py -c datasets/tmdb/casts.json -m datasets/tmdb/movies.json --server_address http://127.0.0.1:8080
   scripts/register_users.sh
   scripts/register_movies.sh
   ```
3. 然后才能成功运行 wrk2 测试

而在 Helm 部署中，缺少了步骤 2，导致 MongoDB 数据库为空，所以 movie-id-service 找不到电影信息。

## 解决方案

已经添加了一个自动数据初始化 Job（`data-init-job`），它会在 Helm 部署后自动运行，完成以下操作：

1. 等待 nginx-web-server 服务就绪
2. 从 GitHub 仓库获取数据集（casts.json 和 movies.json）
3. 向 MongoDB 填充电影信息：
   - 演员和剧组信息
   - 电影元数据（标题、评分、缩略图）
   - 剧情信息
   - 电影 ID 映射
4. 注册 100 个测试用户（username_1 到 username_100）

## 如何使用

### 方法 1：重新部署（推荐）

如果您已经部署了 Helm chart，请重新部署以启用自动初始化：

```bash
# 卸载现有部署
helm uninstall media-microservices -n media

# 重新安装（会自动运行数据初始化）
helm install media-microservices ./helm-chart/mediamicroservices -n media

# 查看初始化 Job 的状态
kubectl get jobs -n media
kubectl logs -n media job/data-init-job -f

# 等待 Job 完成（状态应该显示 "Completed"）
kubectl wait --for=condition=complete job/data-init-job -n media --timeout=600s
```

### 方法 2：手动初始化（如果已部署）

如果您不想重新部署，可以手动运行初始化：

```bash
# 转发 nginx 端口
kubectl port-forward -n media svc/nginx-web-server 8080:8080 &

# 在 mediaMicroservices 目录运行
cd mediaMicroservices
python3 scripts/write_movie_info.py \
  -c datasets/tmdb/casts.json \
  -m datasets/tmdb/movies.json \
  --server_address http://localhost:8080

# 注册用户
scripts/register_users.sh

# 停止端口转发
pkill -f "port-forward"
```

## 验证解决方案

### 1. 检查初始化 Job 是否完成

```bash
kubectl get jobs -n media
# 应该看到 data-init-job 的状态为 "Completed"

kubectl get pods -n media | grep data-init-job
# 应该看到 pod 状态为 "Completed"
```

### 2. 检查是否还有错误

```bash
# 查看 movie-id-service 日志，应该没有 "not found in MongoDB" 错误
kubectl logs -n media deployment/movie-id-service --tail=50

# 查看 nginx 日志，应该没有 "unknown reason" 错误
kubectl logs -n media deployment/nginx-web-server --tail=50
```

### 3. 运行 wrk2 测试

```bash
# 转发端口
kubectl port-forward -n media svc/nginx-web-server 8080:8080 &

# 运行测试
cd wrk2
./wrk -D exp -t 2 -c 2 -d 30 -L \
  -s ../mediaMicroservices/wrk2/scripts/media-microservices/compose-review.lua \
  http://localhost:8080/wrk2-api/review/compose -R 10

# 应该看到成功的响应，没有错误
```

## 配置选项

如果需要禁用自动初始化（例如使用自己的数据）：

```bash
helm install media-microservices ./helm-chart/mediamicroservices -n media \
  --set data-init-job.enabled=false
```

## 更多文档

详细文档请参阅：
- 英文文档：`helm-chart/mediamicroservices/README.md`
- 中文文档：`helm-chart/mediamicroservices/README_CN.md`
- 测试指南：`helm-chart/mediamicroservices/TESTING.md`
- 更改摘要：`helm-chart/mediamicroservices/CHANGES.md`

## 总结

- ✅ 问题已完全解决
- ✅ 添加了自动数据初始化 Job
- ✅ 不需要手动运行初始化脚本
- ✅ 部署后可以立即运行 wrk2 测试
- ✅ 没有破坏性更改，可以禁用自动初始化
- ✅ 修复了 Chart.yaml 中缺失的依赖项
- ✅ 提供了完整的中英文文档

这个解决方案确保 Helm 部署的 media 微服务与 Docker Compose 部署一样，在启动后数据库就已经填充好了测试数据，可以直接运行负载测试。
