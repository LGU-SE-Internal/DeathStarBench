local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

function _M.RegisterMovie()
  local GenericObjectPool = require "GenericObjectPool"
  local MovieIdServiceClient = require'media_service_MovieIdService'
  local ngx = ngx

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

  if (_StrIsEmpty(post.title) or _StrIsEmpty(post.movie_id)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local client = GenericObjectPool:connection(MovieIdServiceClient,"movie-id-service" .. k8s_suffix ,9090)

  local status, err = pcall(client.RegisterMovieId, client, req_id, post.title, tostring(post.movie_id), carrier)

  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    if (err.message) then
      ngx.say("Movie registration failure: " .. err.message)
      ngx.log(ngx.ERR, "Movie registration failure: " .. err.message)
    else
      ngx.say("Movie registration failure: " .. err)
      ngx.log(ngx.ERR, "Movie registration failure: " .. err)
    end
    client.iprot.trans:close()
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  end

  GenericObjectPool:returnConnection(client)

end

return _M