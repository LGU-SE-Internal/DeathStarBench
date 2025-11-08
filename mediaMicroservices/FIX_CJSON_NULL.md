# Fix for cjson.null Handling in Movie Info Write Operations

## Problem
When the data initialization job runs, Lua scripts crash with the following errors:

```
runtime error: /usr/local/openresty/lualib/thrift/TBinaryProtocol.lua:135: bad argument #1 to 'len' (string expected, got userdata)
```

## Root Cause
The TMDB movies dataset contains 3 movies with `null` poster_path values:
- ID: 564394, Title: "Crypto"
- ID: 483157, Title: "Morning Has Broken"
- ID: 595409, Title: "The Caring City"

When these null values are:
1. Placed in JSON arrays by write_movie_info.py: `[null]`
2. Decoded by Lua's cjson library: becomes `cjson.null` (userdata)
3. Passed to Thrift's writeString(): fails because it expects a string

## Solution
Two-layer fix for robustness:

### Layer 1: Python (Prevention)
**File**: `scripts/write_movie_info.py`
- Check if `poster_path` is not None before adding to array
- Send empty array `[]` instead of `[null]` when poster_path is null

### Layer 2: Lua (Defense)
**File**: `nginx-web-server/lua-scripts/wrk2-api/movie-info/write.lua`
- Add `_FilterNulls()` helper function
- Filter out `cjson.null` userdata from all string arrays
- Ensures only valid strings are passed to Thrift

## Testing
Run the data initialization job after deploying this fix:
```bash
kubectl logs -n media job/media-data-init-job -f
```

Expected: All 1000 movies process successfully without errors.

## Impact
- Fixes crashes during data initialization
- Allows wrk2 load testing to proceed
- No functional changes to working code paths
