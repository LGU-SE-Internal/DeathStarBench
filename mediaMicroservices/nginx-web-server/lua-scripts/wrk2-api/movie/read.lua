local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

-- Helper function to convert Thrift objects to plain Lua tables for JSON encoding
-- Thrift objects have method functions (read, write) and userdata that cjson cannot serialize
local function _ThriftToTable(obj)
  if type(obj) ~= "table" then
    return obj
  end
  
  local result = {}
  for k, v in pairs(obj) do
    local vtype = type(v)
    -- Skip functions (like read, write methods from Thrift) and userdata
    if vtype ~= "function" and vtype ~= "userdata" and vtype ~= "thread" then
      if vtype == "table" then
        -- Check if it's an array-like table
        if #v > 0 then
          result[k] = {}
          for i, item in ipairs(v) do
            result[k][i] = _ThriftToTable(item)
          end
        else
          result[k] = _ThriftToTable(v)
        end
      else
        result[k] = v
      end
    end
  end
  return result
end

function _M.ReadMoviePage()
  local ngx = ngx
  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local carrier = {}
  
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
    ngx.say(require("cjson").encode(_ThriftToTable(ret)))
  end
end

function _M.ReadMovieInfo()
  local ngx = ngx
  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local carrier = {}
  
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
    ngx.say(require("cjson").encode(_ThriftToTable(ret)))
  end
end

return _M
