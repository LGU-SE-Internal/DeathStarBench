# Build the dependencies with OpenTelemetry
FROM ubuntu:16.04 AS deps-builder

ARG LIB_MONGOC_VERSION=1.15.0
ARG LIB_THRIFT_VERSION=0.12.0
ARG LIB_JSON_VERSION=3.6.1
ARG LIB_OPENTELEMETRY_VERSION=1.8.1
ARG LIB_YAML_VERSION=0.6.2
ARG LIB_CPP_JWT_VERSION=1.1.1
ARG LIB_CPP_REDIS_VERSION=4.3.1
ARG LIB_AMQP_CPP_VERSION=4.1.4
ARG LIB_SIMPLEAMQPCLIENT_VERSION=2.4.0
ARG LIB_HIREDIS_VERSION=1.0.0
ARG LIB_REDIS_PLUS_PLUS_VERSION=1.2.3

ARG BUILD_DEPS="ca-certificates g++ wget git libmemcached-dev automake bison flex libboost-all-dev libevent-dev libssl-dev libtool make pkg-config librabbitmq-dev python3-dev python3-pip python3-setuptools python3-wheel"
ARG CMAKE_VERSION=3.20.0

# Install build dependencies and CMake
RUN apt-get update && apt-get install -y ${BUILD_DEPS} --no-install-recommends

# Install newer CMake (OpenTelemetry requires CMake >= 3.7)
RUN cd /tmp \
  && wget https://github.com/Kitware/CMake/releases/download/v${CMAKE_VERSION}/cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz \
  && tar -zxf cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz \
  && cp -r cmake-${CMAKE_VERSION}-linux-x86_64/bin/* /usr/local/bin/ \
  && cp -r cmake-${CMAKE_VERSION}-linux-x86_64/share/* /usr/local/share/ \
  && rm -rf cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz cmake-${CMAKE_VERSION}-linux-x86_64

# Install mongo-c-driver
RUN cd /tmp \
  && wget https://github.com/mongodb/mongo-c-driver/releases/download/${LIB_MONGOC_VERSION}/mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
  && tar -zxf mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
  && cd mongo-c-driver-${LIB_MONGOC_VERSION} \
  && mkdir -p cmake-build && cd cmake-build \
  && cmake -DCMAKE_BUILD_TYPE=Release -DENABLE_TESTS=0 -DENABLE_EXAMPLES=0 .. \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf mongo-c-driver-${LIB_MONGOC_VERSION}*

# Install lib-thrift
RUN cd /tmp \
  && wget -O thrift-${LIB_THRIFT_VERSION}.tar.gz https://github.com/apache/thrift/archive/v${LIB_THRIFT_VERSION}.tar.gz \
  && tar -zxf thrift-${LIB_THRIFT_VERSION}.tar.gz \
  && cd thrift-${LIB_THRIFT_VERSION} \
  && mkdir -p cmake-build && cd cmake-build \
  && cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=0 .. \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf thrift-${LIB_THRIFT_VERSION}*

# Install nlohmann/json
RUN cd /tmp \
  && wget -O json-${LIB_JSON_VERSION}.tar.gz https://github.com/nlohmann/json/archive/v${LIB_JSON_VERSION}.tar.gz \
  && tar -zxf json-${LIB_JSON_VERSION}.tar.gz \
  && cd json-${LIB_JSON_VERSION} \
  && mkdir -p cmake-build && cd cmake-build \
  && cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=0 .. \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf json-${LIB_JSON_VERSION}*

# Install yaml-cpp
RUN cd /tmp \
  && wget -O yaml-cpp-${LIB_YAML_VERSION}.tar.gz https://github.com/jbeder/yaml-cpp/archive/yaml-cpp-${LIB_YAML_VERSION}.tar.gz \
  && tar -zxf yaml-cpp-${LIB_YAML_VERSION}.tar.gz \
  && cd yaml-cpp-yaml-cpp-${LIB_YAML_VERSION} \
  && mkdir -p cmake-build && cd cmake-build \
  && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DYAML_CPP_BUILD_TESTS=0 .. \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf yaml-cpp-${LIB_YAML_VERSION}*

# Install Protobuf (required for OpenTelemetry OTLP)
RUN cd /tmp \
  && wget -O protobuf-3.19.4.tar.gz https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protobuf-all-3.19.4.tar.gz \
  && tar -zxf protobuf-3.19.4.tar.gz \
  && cd protobuf-3.19.4 \
  && ./configure --prefix=/usr/local \
  && make -j$(nproc) && make install \
  && ldconfig \
  && cd /tmp && rm -rf protobuf-3.19.4*

# Install gRPC (required for OpenTelemetry OTLP)
RUN cd /tmp \
  && git clone --recurse-submodules -b v1.43.0 --depth 1 https://github.com/grpc/grpc.git \
  && cd grpc && mkdir -p cmake-build && cd cmake-build \
  && cmake -DgRPC_INSTALL=ON \
    -DgRPC_BUILD_TESTS=OFF \
    -DCMAKE_BUILD_TYPE=Release \
    -DgRPC_PROTOBUF_PROVIDER=package \
    -DgRPC_SSL_PROVIDER=package \
    .. \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf grpc

# Install OpenTelemetry C++
# Build without OTLP to avoid abseil namespace conflicts with gRPC
RUN cd /tmp \
  && wget -O opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}.tar.gz https://github.com/open-telemetry/opentelemetry-cpp/archive/v${LIB_OPENTELEMETRY_VERSION}.tar.gz \
  && tar -zxf opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}.tar.gz \
  && cd opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION} \
  && mkdir -p cmake-build && cd cmake-build \
  && cmake -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_CXX_FLAGS="-fPIC" \
    -DBUILD_TESTING=OFF \
    -DWITH_OTLP=OFF \
    -DWITH_EXAMPLES=OFF \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
    .. \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}*

# Install cpp-jwt
RUN cd /tmp \
  && wget -O cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz https://github.com/arun11299/cpp-jwt/archive/v${LIB_CPP_JWT_VERSION}.tar.gz \
  && tar -zxf cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz \
  && cd cpp-jwt-${LIB_CPP_JWT_VERSION} \
  && cp -R include/jwt /usr/local/include \
  && rm -rf /usr/local/include/jwt/json \
  && sed -i 's/\#include \"jwt\/json\/json.hpp\"/\#include \<nlohmann\/json\.hpp\>/g' /usr/local/include/jwt/jwt.hpp \
  && cd /tmp && rm -rf cpp-jwt-${LIB_CPP_JWT_VERSION}*

# Install cpp_redis
RUN cd /tmp \
  && git clone https://github.com/cpp-redis/cpp_redis.git \
  && cd cpp_redis && git checkout ${LIB_CPP_REDIS_VERSION} \
  && git submodule init && git submodule update \
  && mkdir cmake-build && cd cmake-build \
  && cmake .. -DCMAKE_BUILD_TYPE=Release \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf cpp_redis

# Install AMQP-CPP
RUN cd /tmp \
  && git clone https://github.com/CopernicaMarketingSoftware/AMQP-CPP.git \
  && cd AMQP-CPP && git checkout v${LIB_AMQP_CPP_VERSION} \
  && mkdir cmake-build && cd cmake-build \
  && cmake .. -DCMAKE_BUILD_TYPE=Release -DAMQP-CPP_BUILD_SHARED=on -DAMQP-CPP_LINUX_TCP=on \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf AMQP-CPP

# Install SimpleAmqpClient
RUN cd /tmp \
  && git clone https://github.com/alanxz/SimpleAmqpClient.git \
  && cd SimpleAmqpClient && git checkout v${LIB_SIMPLEAMQPCLIENT_VERSION} \
  && mkdir cmake-build && cd cmake-build \
  && cmake .. -DCMAKE_BUILD_TYPE=Release \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf SimpleAmqpClient

# Install Hiredis
RUN cd /tmp \
  && git clone https://github.com/redis/hiredis.git \
  && cd hiredis && git checkout v${LIB_HIREDIS_VERSION} \
  && make -j$(nproc) USE_SSL=1 && make USE_SSL=1 install \
  && cd /tmp && rm -rf hiredis

# Install Redis plus plus
RUN cd /tmp \
  && git clone https://github.com/sewenew/redis-plus-plus.git \
  && cd redis-plus-plus && git checkout ${LIB_REDIS_PLUS_PLUS_VERSION} \
  && sed -i '/Transaction transaction/i\\    ShardsPool* get_shards_pool(){\n        return &_pool;\n    }\n' \
     src/sw/redis++/redis_cluster.h \
  && cmake -DREDIS_PLUS_PLUS_USE_TLS=ON . \
  && make -j$(nproc) && make install \
  && cd /tmp && rm -rf redis-plus-plus

# Install pyyaml (use older version compatible with Python 3.5)
RUN pip3 install 'PyYAML<6.0'

ENV LD_LIBRARY_PATH="/usr/local/lib:${LD_LIBRARY_PATH:-}"
RUN ldconfig

# Build the application
FROM deps-builder AS builder

COPY ./ /social-network-microservices
RUN cd /social-network-microservices \
    && mkdir -p build \
    && cd build \
    && cmake -DCMAKE_BUILD_TYPE=Debug .. \
    && make -j$(nproc) \
    && make install

FROM ubuntu:16.04

# Copy compiled C++ binaries and dependencies
COPY --from=builder /usr/local/bin/* /usr/local/bin/
COPY --from=builder /usr/local/lib/* /usr/local/lib/

# Install system dependencies
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        openssl \
        ca-certificates \
        libsasl2-2 \
        libmemcached11 \
        libmemcachedutil2 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /social-network-microservices
