# cjson.null 处理和套接字超时问题的解决方案

## 已解决的问题

### 问题1：cjson.null 处理
在组合带有媒体附件的帖子时，Lua脚本可能崩溃并显示以下错误：

```
runtime error: /usr/local/openresty/lualib/thrift/TBinaryProtocol.lua:135: bad argument #1 to 'len' (string expected, got userdata)
```

### 问题2：套接字读取超时
在并行用户注册和帖子组合期间，出现超时错误：

```
lua tcp socket read timed out
TTransportException:0: Default (unknown)
```

## 根本原因

### 原因1：空的媒体值
当 media_ids 或 media_types 数组包含 `null` 值时：
1. 它们被客户端或脚本放入JSON数组：`[null]`
2. 被Lua的cjson库解码：变成 `cjson.null` (userdata类型)
3. 传递给Thrift的writeString()：失败，因为它期望字符串类型

### 原因2：套接字超时过短
TSocket.lua中的默认Lua套接字超时为1000毫秒（1秒）。当初始化脚本发送并发请求（用户注册、帖子组合、关注）时，后端服务的响应时间可能超过1秒，导致超时错误。

## 解决方案

### 解决方案1：cjson.null的两层修复

#### 第一层：Python（预防）
**文件**：`scripts/init_social_graph.py`
- 添加了注释说明需要避免数组中的None/null值
- 脚本已经只生成有效的字符串值

#### 第二层：Lua（防御）
**文件**：
- `nginx-web-server/lua-scripts/api/post/compose.lua`
- `nginx-web-server/lua-scripts/wrk2-api/post/compose.lua`
- `openshift/nginx-thrift-config/lua-scripts/api/post/compose.lua`
- `openshift/nginx-thrift-config/lua-scripts/wrk2-api/post/compose.lua`

添加了 `_FilterNulls()` 辅助函数：
- 从所有字符串数组中过滤掉 `cjson.null` userdata
- 确保只有有效的字符串传递给Thrift
- 应用于传递给ComposePost之前的media_ids和media_types数组

### 解决方案2：增加套接字超时

**文件**: 
- `nginx-web-server/conf/nginx.conf`
- `openshift/nginx-thrift-config/nginx.conf`
- `helm-chart/socialnetwork/templates/configs/nginx/nginx.tpl`

在http块中添加超时指令：
```nginx
lua_socket_connect_timeout 10s;
lua_socket_send_timeout 10s;
lua_socket_read_timeout 10s;
```

这将超时时间从1秒增加到10秒，使后端服务能够在数据初始化期间处理高并发请求。

## 测试方法
部署此修复后运行社交图初始化脚本：
```bash
cd socialNetwork/scripts
python3 init_social_graph.py --ip <nginx-ip> --port 8080 --graph socfb-Reed98
```

预期结果：
- 所有用户成功注册，无超时错误
- 所有关注关系成功创建，无超时错误
- 带有媒体附件的帖子成功组合，无cjson.null错误

## 影响
- 修复带有媒体的帖子组合期间的潜在崩溃问题（cjson.null问题）
- 防止用户注册和社交图初始化期间的超时错误（套接字超时问题）
- 允许wrk2负载测试成功进行
- 不影响其他正常工作的代码路径
