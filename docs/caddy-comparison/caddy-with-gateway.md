# Architecture: Using Caddy with API Gateway Service

This document describes how Caddy could be integrated into the DeathStarBench architecture if OpenResty were to be replaced.

## Current Architecture (OpenResty)

```
┌─────────────┐
│   Client    │
│  (Browser)  │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────────────────────────────────────┐
│           OpenResty/Nginx                   │
│  ┌───────────────────────────────────────┐  │
│  │       Lua Business Logic              │  │
│  │  - JWT Verification                   │  │
│  │  - Input Validation                   │  │
│  │  - Session Management                 │  │
│  │  - CORS Handling                      │  │
│  │  - Tracing Context Injection          │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │     Thrift Client Libraries           │  │
│  │  - UserService Client                 │  │
│  │  - ComposePostService Client          │  │
│  │  - TimelineService Client             │  │
│  │  - Connection Pooling                 │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │   OpenTelemetry Integration           │  │
│  │  - HTTP Span Creation                 │  │
│  │  - Thrift RPC Span Creation           │  │
│  │  - Context Propagation                │  │
│  └───────────────────────────────────────┘  │
└────────┬─────────────┬──────────────────────┘
         │ Thrift      │ Thrift
         │ RPC         │ RPC
         ▼             ▼
    ┌────────┐    ┌─────────────┐
    │  User  │    │ ComposePost │
    │Service │    │   Service   │
    └────────┘    └─────────────┘
```

## Proposed Architecture (Caddy + API Gateway)

```
┌─────────────┐
│   Client    │
│  (Browser)  │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────────────────────────────────────┐
│                Caddy                        │
│  ┌───────────────────────────────────────┐  │
│  │      Static File Serving              │  │
│  │      TLS Termination                  │  │
│  │      HTTP Reverse Proxy               │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │   OpenTelemetry Integration           │  │
│  │  - HTTP Span Creation (native)        │  │
│  │  - OTLP Export                        │  │
│  └───────────────────────────────────────┘  │
└────────┬────────────────────────────────────┘
         │ HTTP
         ▼
┌─────────────────────────────────────────────┐
│          API Gateway Service                │
│         (NEW - Go/Java/Node.js)             │
│  ┌───────────────────────────────────────┐  │
│  │       Business Logic                  │  │
│  │  - JWT Verification                   │  │
│  │  - Input Validation                   │  │
│  │  - Session Management                 │  │
│  │  - CORS Handling                      │  │
│  │  - Tracing Context Injection          │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │     Thrift Client Libraries           │  │
│  │  - UserService Client                 │  │
│  │  - ComposePostService Client          │  │
│  │  - TimelineService Client             │  │
│  │  - Connection Pooling                 │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │   OpenTelemetry SDK                   │  │
│  │  - Receives HTTP span from Caddy      │  │
│  │  - Creates child spans for Thrift     │  │
│  │  - Exports to OTLP Collector          │  │
│  └───────────────────────────────────────┘  │
└────────┬─────────────┬──────────────────────┘
         │ Thrift      │ Thrift
         │ RPC         │ RPC
         ▼             ▼
    ┌────────┐    ┌─────────────┐
    │  User  │    │ ComposePost │
    │Service │    │   Service   │
    └────────┘    └─────────────┘
```

## Comparison

### Benefits of Caddy Approach

1. **Simpler Configuration**
   - Caddyfile is easier to read than Nginx config
   - Native OTEL support without modules

2. **Automatic HTTPS**
   - Built-in Let's Encrypt integration
   - Automatic certificate renewal

3. **Type-Safe Business Logic**
   - API Gateway in Go/Java is type-safe
   - Better IDE support and refactoring
   - Compile-time error checking

4. **Better Testing**
   - Can unit test API Gateway service
   - Lua scripts are harder to test

### Drawbacks of Caddy Approach

1. **Additional Network Hop**
   - Client → Caddy → API Gateway → Backend Service
   - Current: Client → OpenResty → Backend Service
   - Increased latency: ~1-5ms per request

2. **Additional Service to Manage**
   - API Gateway deployment
   - API Gateway scaling
   - API Gateway monitoring
   - API Gateway updates

3. **Development Effort**
   - 4-6 weeks to build and test API Gateway
   - Port ~20+ Lua scripts to Go/Java
   - Implement Thrift client logic
   - Comprehensive testing required

4. **Operational Complexity**
   - Another service to deploy and scale
   - More containers to orchestrate
   - More failure points

5. **Performance**
   - Extra serialization/deserialization
   - Extra network round trip
   - Potential bottleneck at API Gateway

6. **Loss of Dynamic Scripting**
   - Lua allows runtime changes
   - Go/Java requires recompilation and redeployment

## Code Comparison

### Current OpenResty Lua (api/post/compose.lua)
```lua
function ComposePost()
  local jwt = require "resty.jwt"
  local client = GenericObjectPool:connection(ComposePostServiceClient, "compose-post-service", 9090)
  
  ngx.req.read_body()
  local post = ngx.req.get_post_args()
  
  -- Verify JWT
  local login_obj = jwt:verify(secret, ngx.var.cookie_login_token)
  if not login_obj["verified"] then
    return ngx.HTTP_UNAUTHORIZED
  end
  
  -- Extract user info
  local user_id = login_obj["payload"]["user_id"]
  
  -- Create tracing span
  local span = tracer:start_span("compose_post_client")
  local carrier = {}
  tracer:text_map_inject(span:context(), carrier)
  
  -- Call Thrift service
  local status, ret = pcall(client.ComposePost, client,
      req_id, username, user_id, post.text,
      media_ids, media_types, post_type, carrier)
  
  GenericObjectPool:returnConnection(client)
  span:finish()
  
  return status and ngx.HTTP_OK or ngx.HTTP_INTERNAL_SERVER_ERROR
end
```

### Proposed API Gateway (Go)
```go
package main

import (
    "context"
    "net/http"
    "github.com/golang-jwt/jwt/v5"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    thrift "github.com/apache/thrift/lib/go/thrift"
)

type APIGateway struct {
    thriftPool *ThriftClientPool
    jwtSecret  []byte
    tracer     trace.Tracer
}

func (g *APIGateway) ComposePost(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse request body
    var post PostRequest
    if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Verify JWT
    cookie, err := r.Cookie("login_token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
        return g.jwtSecret, nil
    })
    if err != nil || !token.Valid {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    
    claims := token.Claims.(jwt.MapClaims)
    userID := int64(claims["user_id"].(float64))
    username := claims["username"].(string)
    
    // Create tracing span
    ctx, span := g.tracer.Start(ctx, "compose_post_client")
    defer span.End()
    
    // Get Thrift client from pool
    client, err := g.thriftPool.Get()
    if err != nil {
        http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
        return
    }
    defer g.thriftPool.Put(client)
    
    // Inject tracing context
    carrier := propagation.MapCarrier{}
    otel.GetTextMapPropagator().Inject(ctx, carrier)
    
    // Call Thrift service
    err = client.ComposePost(
        context.Background(),
        reqID,
        username,
        userID,
        post.Text,
        post.MediaIDs,
        post.MediaTypes,
        post.PostType,
        carrier,
    )
    
    if err != nil {
        http.Error(w, "Failed to compose post", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Successfully upload post"))
}
```

### Caddyfile
```caddyfile
:8080 {
  log {
    output stdout
    format json
  }
  
  handle /api/* {
    tracing {
      span "{http.request.method} {http.request.uri.path}"
    }
    reverse_proxy api-gateway:8080
  }
  
  handle {
    root * /usr/share/nginx/html
    file_server
  }
}
```

As you can see, the Caddy configuration is much simpler, but that's because all the complexity moved to the API Gateway service.

## Deployment Changes

### Current Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-thrift
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: nginx-thrift
        image: openresty/openresty:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: lua-scripts
          mountPath: /usr/local/openresty/nginx/lua-scripts
```

### Proposed Deployment
```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: caddy
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: caddy
        image: caddy:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: caddyfile
          mountPath: /etc/caddy/Caddyfile
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway  # NEW SERVICE
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: api-gateway
        image: myorg/api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://otel-collector:4318"
```

## Conclusion

Using Caddy instead of OpenResty is **technically possible** but requires:

1. **Building a new API Gateway microservice** (4-6 weeks)
2. **Additional operational overhead** (deploy, scale, monitor)
3. **Performance impact** (+1 network hop)
4. **Loss of dynamic scripting** (Lua → compiled language)

**Recommendation:** The current OpenResty setup is already using OpenTelemetry successfully. Unless there's a compelling reason to refactor the entire API layer, it's better to keep OpenResty.

## When Caddy Makes Sense

Caddy would be a good choice if:

1. **Starting a new project from scratch**
   - Design with HTTP/gRPC from the beginning
   - No Lua legacy code to port

2. **Backend services are all HTTP/gRPC**
   - No Thrift protocol
   - Simple reverse proxying sufficient

3. **No complex business logic at gateway**
   - Just TLS termination and routing
   - Authentication handled by backend services

4. **Want to leverage service mesh**
   - Istio/Linkerd handle observability
   - Gateway just does TLS and static content

None of these conditions apply to DeathStarBench's current architecture.
