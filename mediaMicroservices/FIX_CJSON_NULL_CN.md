# cjson.null 处理和套接字超时问题的解决方案

## 已解决的问题

### 问题1：cjson.null 处理
在添加init job后运行wrk2时，前端nginx日志显示以下错误：

```
runtime error: /usr/local/openresty/lualib/thrift/TBinaryProtocol.lua:135: bad argument #1 to 'len' (string expected, got userdata)
```

### 问题2：套接字读取超时
在并行用户注册期间，出现超时错误：

```
lua tcp socket read timed out
TTransportException:0: Default (unknown)
```

## 根本原因

### 原因1：空的海报路径值
TMDB电影数据集中有3部电影的 `poster_path` 值为 `null`：
- ID: 564394, 标题: "Crypto"
- ID: 483157, 标题: "Morning Has Broken"
- ID: 595409, 标题: "The Caring City"

当这些null值：
1. 被 write_movie_info.py 放入JSON数组：`[null]`
2. 被Lua的cjson库解码：变成 `cjson.null` (userdata类型)
3. 传递给Thrift的writeString()：失败，因为它期望字符串类型

### 原因2：套接字超时过短
TSocket.lua中的默认Lua套接字超时为1000毫秒（1秒）。当初始化作业发送100个并发用户注册请求时，用户服务的响应时间可能超过1秒，导致超时错误。

## 解决方案

### 解决方案1：cjson.null的两层修复

#### 第一层：Python（预防）
**文件**：`scripts/write_movie_info.py`
- 在添加到数组前检查 `poster_path` 是否非空
- 当poster_path为null时发送空数组 `[]` 而不是 `[null]`

#### 第二层：Lua（防御）
**文件**：`nginx-web-server/lua-scripts/wrk2-api/movie-info/write.lua`
- 添加 `_FilterNulls()` 辅助函数
- 从所有字符串数组中过滤掉 `cjson.null` userdata
- 确保只有有效的字符串传递给Thrift

### 解决方案2：增加套接字超时

**文件**: 
- `nginx-web-server/conf/nginx.conf`
- `openshift/configmaps/nginx.conf`
- `helm-chart/mediamicroservices/templates/configs/nginx/nginx.tpl`

在http块中添加超时指令：
```nginx
lua_socket_connect_timeout 10s;
lua_socket_send_timeout 10s;
lua_socket_read_timeout 10s;
```

这将超时时间从1秒增加到10秒，使后端服务能够在数据初始化期间处理高并发请求。

## 测试方法
部署此修复后运行数据初始化job：
```bash
kubectl logs -n media job/media-data-init-job -f
```

预期结果：
- 所有1000部电影成功处理，无cjson.null错误
- 所有1000个用户成功注册，无超时错误

## 影响
- 修复电影数据初始化期间的崩溃问题（cjson.null问题）
- 防止用户注册期间的超时错误（套接字超时问题）
- 允许wrk2负载测试成功进行
- 不影响其他正常工作的代码路径
