local _M = {}

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

function _M.Unfollow()
  local ngx = ngx
  local GenericObjectPool = require "GenericObjectPool"
  local SocialGraphServiceClient = require "social_network_SocialGraphService"

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

  local client = GenericObjectPool:connection(
      SocialGraphServiceClient, "social-graph-service.social-network.svc.cluster.local", 9090)

  local status
  local err
  if (not _StrIsEmpty(post.user_id) and not _StrIsEmpty(post.followee_id)) then
    status, err = pcall(client.Unfollow, client,req_id,
        tonumber(post.user_id), tonumber(post.followee_id), carrier )
  elseif (not _StrIsEmpty(post.user_name) and not _StrIsEmpty(post.followee_name)) then
    status, err = pcall(client.UnfollowWithUsername, client,req_id,
        post.user_name, post.followee_name, carrier )
  else
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    ngx.say("Unfollow Failed: " .. err.message)
    ngx.log(ngx.ERR, "Unfollow Failed: " .. err.message)
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  end
  GenericObjectPool:returnConnection(client)

end

return _M