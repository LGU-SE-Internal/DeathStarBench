{{- define "mediamicroservices.templates.nginx.nginx.conf"  }}
# Load the ngx_otel_module for OpenTelemetry tracing
# Note: ngx_otel_module v0.1.2 compiled for OpenResty 1.25.3.2 (nginx 1.25.3)
load_module modules/ngx_otel_module.so;

# Checklist: Make sure that worker_processes == #cores you gave to
# nginx process
worker_processes  auto;

error_log  logs/error.log;

# Checklist: Make sure that worker_connections * worker_processes
# is greater than the total connections between the client and Nginx.
events {
  use epoll;
  worker_connections  1024;
}

env fqdn_suffix;

http {
  # OpenTelemetry configuration using ngx_otel_module
  # Note: ngx_otel_module only supports gRPC (port 4317), not HTTP (port 4318)
  otel_exporter {
    endpoint {{ .Values.global.otel.endpoint | replace ":4318" ":4317" }};
  }
  otel_service_name nginx-web-server;
  otel_trace on;
  otel_trace_context propagate;

  include       mime.types;
  default_type  application/octet-stream;

  log_format main '$remote_addr - $remote_user [$time_local] "$request"'
                  '$status $body_bytes_sent "$http_referer" '
                  '"$http_user_agent" "$http_x_forwarded_for"';
  # access_log  logs/access.log  main;

  sendfile        on;
  tcp_nopush      on;
  tcp_nodelay     on;

  # Checklist: Make sure the keepalive_timeout is greateer than
  # the duration of your experiment and keepalive_requests
  # is greateer than the total number of requests sent from
  # the workload generator
  keepalive_timeout  120s;
  keepalive_requests 100000;

  # Docker default hostname resolver
  # resolver 127.0.0.11 ipv6=off;

  # Kubernetes default hostname resolver
  resolver {{ .Values.global.nginx.resolverName }} ipv6=off;

  # Lua socket timeout settings for Thrift connections
  # Increase from default 1s to handle high load during data initialization
  lua_socket_connect_timeout 10s;
  lua_socket_send_timeout 10s;
  lua_socket_read_timeout 10s;

  lua_package_path '/usr/local/openresty/nginx/lua-scripts/?.lua;;';

  server {

    # Checklist: Set up the port that nginx listens to.
    listen       8080 reuseport;
    server_name  localhost;

    # Checklist: Turn of the access_log and error_log if you
    # don't need them.
    access_log  off;
    # error_log off;

    lua_need_request_body on;

    # Checklist: Make sure that the location here is consistent
    # with the location you specified in wrk2.
    location /wrk2-api/user/register {
      content_by_lua '
          local client = require "wrk2-api/user/register"
          client.RegisterUser();
      ';
    }

    location /wrk2-api/movie/register {
      content_by_lua '
          local client = require "wrk2-api/movie/register"
          client.RegisterMovie();
      ';
    }

    location /wrk2-api/review/compose {
      content_by_lua '
          local client = require "wrk2-api/review/compose"
          client.ComposeReview();
      ';
    }

    location /wrk2-api/movie-info/write {
      content_by_lua '
          local client = require "wrk2-api/movie-info/write"
          client.WriteMovieInfo();
      ';
    }

    location /wrk2-api/cast-info/write {
      content_by_lua '
          local client = require "wrk2-api/cast-info/write"
          client.WriteCastInfo();
      ';
    }

    location /wrk2-api/plot/write {
      content_by_lua '
          local client = require "wrk2-api/plot/write"
          client.WritePlot();
      ';
    }
  }
}
{{- end}}
