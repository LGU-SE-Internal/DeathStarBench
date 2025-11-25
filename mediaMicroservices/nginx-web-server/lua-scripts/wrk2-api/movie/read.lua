local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

function _M.ReadMoviePage()
  local ngx = ngx
  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = ngx.var.opentracing_binary_context
  local carrier = {}
  carrier["ot-tracer-traceid"] = ngx.var.opentracing_context_traceid
  carrier["ot-tracer-spanid"] = ngx.var.opentracing_context_spanid
  carrier["ot-tracer-sampled"] = ngx.var.opentracing_context_sampled
  
  local movie_id = ngx.var.arg_movie_id or ""
  local review_start = tonumber(ngx.var.arg_review_start) or 0
  local review_stop = tonumber(ngx.var.arg_review_stop) or 10
  
  local GenericObjectPool = require "GenericObjectPool"
  local PageServiceClient = require 'media_service_PageService'
  local page_client = GenericObjectPool:connection(
    PageServiceClient, "page-service" .. k8s_suffix, 9090)
  
  local status, ret = pcall(page_client.ReadPage, page_client, 
    req_id, movie_id, review_start, review_stop, carrier)
  GenericObjectPool:returnConnection(page_client)
  
  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    ngx.say("Failed to read movie page: " .. ret.message)
    ngx.log(ngx.ERR, "Failed to read movie page: " .. ret.message)
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  else
    ngx.say(require("cjson").encode(ret))
  end
end

function _M.ReadMovieInfo()
  local ngx = ngx
  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = ngx.var.opentracing_binary_context
  local carrier = {}
  carrier["ot-tracer-traceid"] = ngx.var.opentracing_context_traceid
  carrier["ot-tracer-spanid"] = ngx.var.opentracing_context_spanid
  carrier["ot-tracer-sampled"] = ngx.var.opentracing_context_sampled
  
  local movie_id = ngx.var.arg_movie_id or ""
  
  local GenericObjectPool = require "GenericObjectPool"
  local MovieInfoServiceClient = require 'media_service_MovieInfoService'
  local movie_info_client = GenericObjectPool:connection(
    MovieInfoServiceClient, "movie-info-service" .. k8s_suffix, 9090)
  
  local status, ret = pcall(movie_info_client.ReadMovieInfo, movie_info_client,
    req_id, movie_id, carrier)
  GenericObjectPool:returnConnection(movie_info_client)
  
  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    ngx.say("Failed to read movie info: " .. ret.message)
    ngx.log(ngx.ERR, "Failed to read movie info: " .. ret.message)
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  else
    ngx.say(require("cjson").encode(ret))
  end
end

return _M
