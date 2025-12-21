local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

function _M.RegisterUser()
  local ngx = ngx
  local GenericObjectPool = require "GenericObjectPool"
  local UserServiceClient = require "social_network_UserService".UserServiceClient

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local carrier = {}
  -- Use nginx's current span context (not incoming client headers)
  if ngx.var.otel_trace_id and ngx.var.otel_span_id then
    carrier["traceparent"] = string.format("00-%s-%s-01", ngx.var.otel_trace_id, ngx.var.otel_span_id)
    carrier["tracestate"] = ngx.var.http_tracestate or ""
  else
    carrier["traceparent"] = ngx.var.http_traceparent or ""
    carrier["tracestate"] = ngx.var.http_tracestate or ""
  end

  ngx.req.read_body()
  local post = ngx.req.get_post_args()

  if (_StrIsEmpty(post.first_name) or _StrIsEmpty(post.last_name) or
      _StrIsEmpty(post.username) or _StrIsEmpty(post.password)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local client = GenericObjectPool:connection(UserServiceClient, "user-service" .. k8s_suffix, 9090)

  local status, err = pcall(client.RegisterUser, client, req_id, post.first_name,
      post.last_name, post.username, post.password, carrier)
  GenericObjectPool:returnConnection(client)

  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    if (err.message) then
      ngx.say("User registration failure: " .. err.message)
      ngx.log(ngx.ERR, "User registration failure: " .. err.message)
    else
      ngx.say("User registration failure: " .. err.message)
      ngx.log(ngx.ERR, "User registration failure: " .. err.message)
    end
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  else
    ngx.redirect("../../index.html")
    ngx.exit(ngx.HTTP_OK)
  end
end

return _M
