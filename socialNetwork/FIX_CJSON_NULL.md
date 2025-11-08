# Fix for cjson.null Handling and Socket Timeout Issues

## Problems Addressed

### Problem 1: cjson.null Handling
When composing posts with media attachments, Lua scripts can crash with the following errors:

```
runtime error: /usr/local/openresty/lualib/thrift/TBinaryProtocol.lua:135: bad argument #1 to 'len' (string expected, got userdata)
```

### Problem 2: Socket Read Timeout
During parallel user registration and post composition, timeout errors occur:

```
lua tcp socket read timed out
TTransportException:0: Default (unknown)
```

## Root Causes

### Cause 1: Null Media Values
When media_ids or media_types arrays contain `null` values:
1. They are placed in JSON arrays by clients or scripts: `[null]`
2. Decoded by Lua's cjson library: becomes `cjson.null` (userdata)
3. Passed to Thrift's writeString(): fails because it expects a string

### Cause 2: Short Socket Timeout
The default Lua socket timeout in TSocket.lua is 1000ms (1 second). When the init script sends concurrent requests (user registration, post composition, follows), backend services can take longer than 1 second to respond, causing timeout errors.

## Solutions

### Solution 1: Two-layer fix for cjson.null

#### Layer 1: Python (Prevention)
**File**: `scripts/init_social_graph.py`
- Added comments explaining the need to avoid None/null values in arrays
- The script already generates only valid string values

#### Layer 2: Lua (Defense)
**Files**: 
- `nginx-web-server/lua-scripts/api/post/compose.lua`
- `nginx-web-server/lua-scripts/wrk2-api/post/compose.lua`
- `openshift/nginx-thrift-config/lua-scripts/api/post/compose.lua`
- `openshift/nginx-thrift-config/lua-scripts/wrk2-api/post/compose.lua`

Added `_FilterNulls()` helper function that:
- Filters out `cjson.null` userdata from all string arrays
- Ensures only valid strings are passed to Thrift
- Applied to media_ids and media_types arrays before passing to ComposePost

### Solution 2: Increase Socket Timeouts

**Files**: 
- `nginx-web-server/conf/nginx.conf`
- `openshift/nginx-thrift-config/nginx.conf`
- `helm-chart/socialnetwork/templates/configs/nginx/nginx.tpl`

Added timeout directives in the http block:
```nginx
lua_socket_connect_timeout 10s;
lua_socket_send_timeout 10s;
lua_socket_read_timeout 10s;
```

This increases timeouts from 1s to 10s, allowing backend services to handle high-concurrency requests during data initialization.

## Testing
Run the social graph initialization script after deploying this fix:
```bash
cd socialNetwork/scripts
python3 init_social_graph.py --ip <nginx-ip> --port 8080 --graph socfb-Reed98
```

Expected: 
- All users register successfully without timeout errors
- All follows are created successfully without timeout errors
- Posts with media attachments compose successfully without cjson.null errors

## Impact
- Fixes potential crashes during post composition with media (cjson.null issue)
- Prevents timeout errors during user registration and social graph initialization (socket timeout issue)
- Allows wrk2 load testing to proceed successfully
- No functional changes to working code paths
