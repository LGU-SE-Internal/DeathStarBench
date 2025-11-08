local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

-- Helper function to filter out cjson.null values from arrays
-- cjson.null is userdata, not a string, which causes writeString to fail
local function _FilterNulls(arr)
  local cjson = require("cjson")
  local filtered = {}
  if type(arr) == "table" then
    for _, v in ipairs(arr) do
      if v ~= cjson.null and type(v) == "string" then
        table.insert(filtered, v)
      end
    end
  end
  return filtered
end

function _M.ComposePost()

  local ngx = ngx
  local cjson = require "cjson"
  local jwt = require "resty.jwt"

  local GenericObjectPool = require "GenericObjectPool"
  local social_network_ComposePostService = require "social_network_ComposePostService"
  local ComposePostServiceClient = social_network_ComposePostService.ComposePostServiceClient

  GenericObjectPool:setMaxTotal(512)

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)

  ngx.req.read_body()
  local post = ngx.req.get_post_args()

  if (_StrIsEmpty(post.post_type) or _StrIsEmpty(post.text)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  if (_StrIsEmpty(ngx.var.cookie_login_token)) then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.redirect("../../index.html")
    ngx.exit(ngx.HTTP_OK)
  end

  local login_obj = jwt:verify(ngx.shared.config:get("secret"), ngx.var.cookie_login_token)
  if not login_obj["verified"] then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.say(login_obj.reason);
    ngx.redirect("../../index.html")
    ngx.exit(ngx.HTTP_OK)
  end
  -- get user id/name from login obj
  local timestamp = tonumber(login_obj["payload"]["timestamp"])
  local ttl = tonumber(login_obj["payload"]["ttl"])
  local user_id = tonumber(login_obj["payload"]["user_id"])
  local username = login_obj["payload"]["username"]

  if (timestamp + ttl < ngx.time()) then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.header.content_type = "text/plain"
    ngx.say("Login token expired, please log in again")
    ngx.redirect("../../index.html")
    ngx.exit(ngx.HTTP_OK)
  else
    local status, ret
    local client = GenericObjectPool:connection(
      ComposePostServiceClient, "compose-post-service" .. k8s_suffix, 9090)

    local carrier = {}

    if (not _StrIsEmpty(post.media_ids) and not _StrIsEmpty(post.media_types)) then
      -- Filter out cjson.null values from arrays to prevent Thrift writeString errors
      local media_ids = _FilterNulls(cjson.decode(post.media_ids))
      local media_types = _FilterNulls(cjson.decode(post.media_types))
      status, ret = pcall(client.ComposePost, client,
          req_id, username, tonumber(user_id), post.text,
          media_ids, media_types,
          tonumber(post.post_type), carrier)
    else
      status, ret = pcall(client.ComposePost, client,
          req_id, username, tonumber(user_id), post.text,
          {}, {}, tonumber(post.post_type), carrier)
    end

    if not status then
      ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
      if (ret.message) then
        ngx.say("compost_post failure: " .. ret.message)
        ngx.log(ngx.ERR, "compost_post failure: " .. ret.message)
      else
        ngx.say("compost_post failure: " .. ret)
        ngx.log(ngx.ERR, "compost_post failure: " .. ret)
      end
      client.iprot.trans:close()
      ngx.exit(ngx.status)
    end

    GenericObjectPool:returnConnection(client)
    ngx.status = ngx.HTTP_OK
    ngx.say("Successfully upload post")
      ngx.exit(ngx.status)
  end
end

return _M