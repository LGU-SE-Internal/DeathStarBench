--
-- Stub/dummy OpenTracing bridge tracer module
-- This is a no-op implementation since tracing is now handled by ngx_otel_module at the nginx level
-- Keeps existing Lua code compatible without making extensive changes
--

local _M = {}

-- Dummy span object that does nothing
local DummySpan = {}
DummySpan.__index = DummySpan

function DummySpan:finish() end
function DummySpan:set_tag(key, value) end
function DummySpan:log_kv(table) end
function DummySpan:set_baggage_item(key, value) end
function DummySpan:get_baggage_item(key) return nil end
function DummySpan:context() return {} end

-- Dummy tracer object
local DummyTracer = {}
DummyTracer.__index = DummyTracer

function DummyTracer:start_span(operation_name, options)
  return setmetatable({}, DummySpan)
end

function DummyTracer:binary_extract(carrier)
  return {}
end

function DummyTracer:http_headers_extract(carrier)
  return {}
end

function DummyTracer:binary_inject(span_context, carrier)
end

function DummyTracer:http_headers_inject(span_context, carrier)
end

-- Module functions
function _M.new_from_global()
  return setmetatable({}, DummyTracer)
end

function _M.new(tracer)
  return setmetatable({}, DummyTracer)
end

return _M
