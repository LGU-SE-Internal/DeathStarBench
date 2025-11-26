math.randomseed(os.time())
math.random(); math.random(); math.random()

local charset = {'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's',
  'd', 'f', 'g', 'h', 'j', 'k', 'l', 'z', 'x', 'c', 'v', 'b', 'n', 'm', 'Q',
  'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P', 'A', 'S', 'D', 'F', 'G', 'H',
  'J', 'K', 'L', 'Z', 'X', 'C', 'V', 'B', 'N', 'M', '1', '2', '3', '4', '5',
  '6', '7', '8', '9', '0'}

-- Movie IDs are generated from titles, stored as "movie_<hash>"
-- For testing, we'll use a set of common movie IDs
local movie_ids = {}
for i = 0, 99 do
  table.insert(movie_ids, "movie_" .. tostring(i))
end

-- Usernames for testing
local usernames = {}
for i = 0, 99 do
  table.insert(usernames, "username_" .. tostring(i))
end

local function stringRandom(length)
  if length > 0 then
    return stringRandom(length - 1) .. charset[math.random(1, #charset)]
  else
    return ""
  end
end

local function urlEncode(s)
  s = string.gsub(s, "([^%w%.%- ])", function(c) return string.format("%%%02X", string.byte(c)) end)
  return string.gsub(s, " ", "+")
end

-- Build URL from wrk's command line parameters
local function get_url()
  local scheme = wrk.scheme or "http"
  local host = wrk.host or "localhost"
  local port = wrk.port or "8080"
  return scheme .. "://" .. host .. ":" .. port
end

local url = get_url()

-- Compose a new review (write operation)
local function compose_review()
  local username = usernames[math.random(1, #usernames)]
  local password = "password"
  local movie_id = movie_ids[math.random(1, #movie_ids)]
  local rating = math.random(0, 10)
  local text = stringRandom(256)
  
  local title = "Movie " .. tostring(math.random(0, 99))
  
  local path = "/wrk2-api/review/compose"
  local method = "POST"
  local headers = {}
  local body = "username=" .. username .. "&password=" .. password .. "&title=" ..
                  urlEncode(title) .. "&rating=" .. rating .. "&text=" .. urlEncode(text)
  headers["Content-Type"] = "application/x-www-form-urlencoded"
  
  return wrk.format(method, path, headers, body)
end

-- Read movie page with reviews (read operation)
local function read_movie_page()
  local movie_id = movie_ids[math.random(1, #movie_ids)]
  local review_start = 0
  local review_stop = 10
  
  local args = "movie_id=" .. movie_id .. "&review_start=" .. review_start .. 
               "&review_stop=" .. review_stop
  local method = "GET"
  local headers = {}
  headers["Content-Type"] = "application/x-www-form-urlencoded"
  local path = "/wrk2-api/movie/read-page?" .. args
  return wrk.format(method, path, headers, nil)
end

-- Read movie info (read operation)
local function read_movie_info()
  local movie_id = movie_ids[math.random(1, #movie_ids)]
  
  local args = "movie_id=" .. movie_id
  local method = "GET"
  local headers = {}
  headers["Content-Type"] = "application/x-www-form-urlencoded"
  local path = "/wrk2-api/movie/read-info?" .. args
  return wrk.format(method, path, headers, nil)
end

-- Main request function with mixed workload
-- Distribution: 50% read movie page, 30% read movie info, 20% compose review
request = function()
  local read_page_ratio = 0.50
  local read_info_ratio = 0.30
  local compose_ratio = 0.20
  
  local coin = math.random()
  if coin < read_page_ratio then
    return read_movie_page()
  elseif coin < read_page_ratio + read_info_ratio then
    return read_movie_info()
  else
    return compose_review()
  end
end
