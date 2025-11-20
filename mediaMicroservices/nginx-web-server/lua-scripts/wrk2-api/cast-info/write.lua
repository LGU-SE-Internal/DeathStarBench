local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

function _M.WriteCastInfo()
  local GenericObjectPool = require "GenericObjectPool"
  local CastInfoServiceClient = require 'media_service_CastInfoService'
  local ngx = ngx
  local cjson = require("cjson")

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local carrier = {}

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

  local cast_info = cjson.decode(data)
  if (cast_info["cast_info_id"] == nil or cast_info["name"] == nil or
      cast_info["gender"] == nil or cast_info["intro"] == nil) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local client = GenericObjectPool:connection(CastInfoServiceClient, "cast-info-service" .. k8s_suffix, 9090)
  client:WriteCastInfo(req_id, cast_info["cast_info_id"], cast_info["name"],
      cast_info["gender"], cast_info["intro"],  carrier)
  GenericObjectPool:returnConnection(client)

end

return _M