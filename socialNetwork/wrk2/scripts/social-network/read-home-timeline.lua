local socket = require("socket")
local time = socket.gettime()*1000
math.randomseed(time)
math.random(); math.random(); math.random()

-- load env vars
local max_user_index = tonumber(os.getenv("max_user_index")) or 962

-- Build URL from wrk's command line parameters
local function get_url()
  local scheme = wrk.scheme or "http"
  local host = wrk.host or "localhost"
  local port = wrk.port or "8080"
  return scheme .. "://" .. host .. ":" .. port
end

local url = get_url()

request = function()
  local user_id = tostring(math.random(0, max_user_index - 1))
  local start = tostring(math.random(0, 100))
  local stop = tostring(start + 10)

  local args = "user_id=" .. user_id .. "&start=" .. start .. "&stop=" .. stop
  local method = "GET"
  local headers = {}
  headers["Content-Type"] = "application/x-www-form-urlencoded"
  local path = url .. "/wrk2-api/home-timeline/read?" .. args
  return wrk.format(method, path, headers, nil)

end
