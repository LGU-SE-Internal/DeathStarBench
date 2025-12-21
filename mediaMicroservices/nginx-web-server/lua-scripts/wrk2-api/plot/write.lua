local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

function _M.WritePlot()
  local GenericObjectPool = require "GenericObjectPool"
  local PlotServiceClient = require 'media_service_PlotService'
  local ngx = ngx
  local cjson = require("cjson")

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
  local data = ngx.req.get_body_data()

  -- If body is not in memory, it might be in a temp file
  if not data then
    local body_file = ngx.req.get_body_file()
    if body_file then
      local file = io.open(body_file, "r")
      if file then
        data = file:read("*all")
        file:close()
      end
    end
  end

  if not data then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Empty body")
    ngx.log(ngx.ERR, "Empty body")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local plot = cjson.decode(data)
  if (plot["plot_id"] == nil or plot["plot"] == nil) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local client = GenericObjectPool:connection(PlotServiceClient, "plot-service" .. k8s_suffix, 9090)
  client:WritePlot(req_id, plot["plot_id"], plot["plot"], carrier)
  GenericObjectPool:returnConnection(client)
end

return _M