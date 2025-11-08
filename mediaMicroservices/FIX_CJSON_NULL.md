# Fix for cjson.null Handling and Socket Timeout Issues

## Problems Addressed

### Problem 1: cjson.null Handling
When the data initialization job runs, Lua scripts crash with the following errors:

```
runtime error: /usr/local/openresty/lualib/thrift/TBinaryProtocol.lua:135: bad argument #1 to 'len' (string expected, got userdata)
```

### Problem 2: Socket Read Timeout
During parallel user registration, timeout errors occur:

```
lua tcp socket read timed out
TTransportException:0: Default (unknown)
```

## Root Causes

### Cause 1: Null Poster Path Values
The TMDB movies dataset contains 3 movies with `null` poster_path values:
- ID: 564394, Title: "Crypto"
- ID: 483157, Title: "Morning Has Broken"
- ID: 595409, Title: "The Caring City"

When these null values are:
1. Placed in JSON arrays by write_movie_info.py: `[null]`
2. Decoded by Lua's cjson library: becomes `cjson.null` (userdata)
3. Passed to Thrift's writeString(): fails because it expects a string

### Cause 2: Short Socket Timeout
The default Lua socket timeout in TSocket.lua is 1000ms (1 second). When the init job sends 100 concurrent user registration requests, the user-service can take longer than 1 second to respond, causing timeout errors.

## Solutions

### Solution 1: Two-layer fix for cjson.null

#### Layer 1: Python (Prevention)
**File**: `scripts/write_movie_info.py`
- Check if `poster_path` is not None before adding to array
- Send empty array `[]` instead of `[null]` when poster_path is null

#### Layer 2: Lua (Defense)
**File**: `nginx-web-server/lua-scripts/wrk2-api/movie-info/write.lua`
- Add `_FilterNulls()` helper function
- Filter out `cjson.null` userdata from all string arrays
- Ensures only valid strings are passed to Thrift

### Solution 2: Increase Socket Timeouts

**Files**: 
- `nginx-web-server/conf/nginx.conf`
- `openshift/configmaps/nginx.conf`
- `helm-chart/mediamicroservices/templates/configs/nginx/nginx.tpl`

Added timeout directives in the http block:
```nginx
lua_socket_connect_timeout 10s;
lua_socket_send_timeout 10s;
lua_socket_read_timeout 10s;
```

This increases timeouts from 1s to 10s, allowing backend services to handle high-concurrency requests during data initialization.

## Testing
Run the data initialization job after deploying this fix:
```bash
kubectl logs -n media job/media-data-init-job -f
```

Expected: 
- All 1000 movies process successfully without cjson.null errors
- All 1000 users register successfully without timeout errors

## Impact
- Fixes crashes during movie data initialization (cjson.null issue)
- Prevents timeout errors during user registration (socket timeout issue)
- Allows wrk2 load testing to proceed successfully
- No functional changes to working code paths
