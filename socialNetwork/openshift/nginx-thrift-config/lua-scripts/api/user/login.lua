local _M = {}

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

function _M.Login()
  local ngx = ngx
  local GenericObjectPool = require "GenericObjectPool"
  local UserServiceClient = require "social_network_UserService"
  local cjson = require "cjson"

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
  local args = ngx.req.get_uri_args()

  if (_StrIsEmpty(args.username) or _StrIsEmpty(args.password)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.header.content_type = "text/plain"
    ngx.say("Incomplete arguments")
    ngx.say(ngx.var.scheme .. "://" .. ngx.var.server_addr .. ":" .. ngx.var.server_port)
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
    return ngx.redirect("/login.html")
  end

  local client = GenericObjectPool:connection(UserServiceClient, "user-service.social-network.svc.cluster.local", 9090)

  local status, ret = pcall(client.Login, client, req_id,
      args.username, args.password, carrier)
  GenericObjectPool:returnConnection(client)

  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    if (ret.message) then
      ngx.header.content_type = "text/plain"
      ngx.say("User login failure: " .. ret.message)
      ngx.log(ngx.ERR, "User login failure: " .. ret.message)
    else
      ngx.header.content_type = "text/plain"
      ngx.say("User login failure: " .. ret.message)
      ngx.log(ngx.ERR, "User login failure: " .. ret.message)
    end
    ngx.exit(ngx.HTTP_OK)
  else
    ngx.header.content_type = "text/plain"
    ngx.header["Set-Cookie"] = "login_token=" .. ret .. "; Path=/; Expires="
        .. ngx.cookie_time(ngx.time() + ngx.shared.config:get("cookie_ttl"))
    ngx.redirect("../../index.html")
    ngx.exit(ngx.HTTP_OK)
  end
end

return _M