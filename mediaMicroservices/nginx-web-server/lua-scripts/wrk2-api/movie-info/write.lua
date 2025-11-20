local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
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

function _M.WriteMovieInfo()
  local GenericObjectPool = require "GenericObjectPool"
  local MovieInfoServiceClient = require 'media_service_MovieInfoService'
  local ttypes = require("media_service_ttypes")
  local Cast = ttypes.Cast
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

  local movie_info = cjson.decode(data)
  if (movie_info["movie_id"] == nil or movie_info["title"] == nil or
      movie_info["casts"] == nil or movie_info["plot_id"] == nil or
      movie_info["thumbnail_ids"] == nil or movie_info["photo_ids"] == nil or
      movie_info["video_ids"] == nil or movie_info["avg_rating"] == nil or
      movie_info["num_rating"] == nil) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local casts = {}
  for _,cast in ipairs(movie_info["casts"]) do
    local new_cast = Cast:new{}
    new_cast["character"]=cast["character"]
    new_cast["cast_id"]=cast["cast_id"]
    new_cast["cast_info_id"]=cast["cast_info_id"]
    table.insert(casts, new_cast)
  end

  -- Filter out cjson.null values from string arrays to prevent Thrift writeString errors
  local thumbnail_ids = _FilterNulls(movie_info["thumbnail_ids"])
  local photo_ids = _FilterNulls(movie_info["photo_ids"])
  local video_ids = _FilterNulls(movie_info["video_ids"])

  local client = GenericObjectPool:connection(MovieInfoServiceClient, "movie-info-service" .. k8s_suffix , 9090)
  client:WriteMovieInfo(req_id, movie_info["movie_id"], movie_info["title"],
      casts, movie_info["plot_id"], thumbnail_ids,
      photo_ids, video_ids, tostring(movie_info["avg_rating"]),
      movie_info["num_rating"], carrier)
  ngx.say(movie_info["avg_rating"])
  GenericObjectPool:returnConnection(client)

end

return _M