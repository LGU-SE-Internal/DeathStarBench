# cjson.null 处理问题的解决方案

## 问题描述
在添加init job后运行wrk2时，前端nginx日志显示以下错误：

```
runtime error: /usr/local/openresty/lualib/thrift/TBinaryProtocol.lua:135: bad argument #1 to 'len' (string expected, got userdata)
```

## 根本原因
TMDB电影数据集中有3部电影的 `poster_path` 值为 `null`：
- ID: 564394, 标题: "Crypto"
- ID: 483157, 标题: "Morning Has Broken"
- ID: 595409, 标题: "The Caring City"

当这些null值：
1. 被 write_movie_info.py 放入JSON数组：`[null]`
2. 被Lua的cjson库解码：变成 `cjson.null` (userdata类型)
3. 传递给Thrift的writeString()：失败，因为它期望字符串类型

## 解决方案
采用两层防护措施：

### 第一层：Python（预防）
**文件**：`scripts/write_movie_info.py`
- 在添加到数组前检查 `poster_path` 是否非空
- 当poster_path为null时发送空数组 `[]` 而不是 `[null]`

### 第二层：Lua（防御）
**文件**：`nginx-web-server/lua-scripts/wrk2-api/movie-info/write.lua`
- 添加 `_FilterNulls()` 辅助函数
- 从所有字符串数组中过滤掉 `cjson.null` userdata
- 确保只有有效的字符串传递给Thrift

## 测试方法
部署此修复后运行数据初始化job：
```bash
kubectl logs -n media job/media-data-init-job -f
```

预期结果：所有1000部电影成功处理，无错误。

## 影响
- 修复数据初始化期间的崩溃问题
- 允许wrk2负载测试正常进行
- 不影响其他正常工作的代码路径
