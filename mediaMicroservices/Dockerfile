FROM ubuntu:20.04


ARG LIB_MONGOC_VERSION=1.14.0
ARG LIB_THRIFT_VERSION=0.12.0
ARG LIB_JSON_VERSION=3.6.1
ARG LIB_YAML_VERSION=0.6.2
ARG LIB_OPENTRACING_VERSION=1.6.0
ARG LIB_PROTOBUF_VERSION=3.19.4
ARG LIB_OPENTELEMETRY_VERSION=1.23.0
ARG LIB_CPP_JWT_VERSION=1.1.1
ARG LIB_CPP_REDIS_VERSION=4.3.1

ARG BUILD_DEPS="ca-certificates g++ cmake wget git libmemcached-dev automake bison flex libboost-all-dev libevent-dev libssl-dev libtool make pkg-config libcurl4-openssl-dev"

# Install build dependencies
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ${BUILD_DEPS} --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

# Install mongo-c-driver
RUN cd /tmp && \
    wget https://github.com/mongodb/mongo-c-driver/releases/download/${LIB_MONGOC_VERSION}/mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz && \
    tar -zxf mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz && \
    cd mongo-c-driver-${LIB_MONGOC_VERSION} && \
    mkdir -p cmake-build && cd cmake-build && \
    cmake -DENABLE_TESTS=0 -DENABLE_EXAMPLES=0 .. && \
    make && make install && \
    cd /tmp && rm -rf mongo-c-driver-${LIB_MONGOC_VERSION}*

# Install lib-thrift
RUN cd /tmp && \
    wget -O thrift-${LIB_THRIFT_VERSION}.tar.gz https://github.com/apache/thrift/archive/v${LIB_THRIFT_VERSION}.tar.gz && \
    tar -zxf thrift-${LIB_THRIFT_VERSION}.tar.gz && \
    cd thrift-${LIB_THRIFT_VERSION} && \
    mkdir -p cmake-build && cd cmake-build && \
    cmake -DBUILD_TESTING=0 .. && \
    make && make install && \
    cd /tmp && rm -rf thrift-${LIB_THRIFT_VERSION}*

# Install nlohmann/json
RUN cd /tmp && \
    wget -O json-${LIB_JSON_VERSION}.tar.gz https://github.com/nlohmann/json/archive/v${LIB_JSON_VERSION}.tar.gz && \
    tar -zxf json-${LIB_JSON_VERSION}.tar.gz && \
    cd json-${LIB_JSON_VERSION} && \
    mkdir -p cmake-build && cd cmake-build && \
    cmake -DBUILD_TESTING=0 .. && \
    make && make install && \
    cd /tmp && rm -rf json-${LIB_JSON_VERSION}*

# Install yaml-cpp
RUN cd /tmp && \
    wget -O yaml-cpp-${LIB_YAML_VERSION}.tar.gz https://github.com/jbeder/yaml-cpp/archive/yaml-cpp-${LIB_YAML_VERSION}.tar.gz && \
    tar -zxf yaml-cpp-${LIB_YAML_VERSION}.tar.gz && \
    cd yaml-cpp-yaml-cpp-${LIB_YAML_VERSION} && \
    mkdir -p cmake-build && cd cmake-build && \
    cmake -DCMAKE_CXX_FLAGS="-fPIC" -DYAML_CPP_BUILD_TESTS=0 .. && \
    make && make install && \
    cd /tmp && rm -rf yaml-cpp-${LIB_YAML_VERSION}*

# Install opentracing-cpp
RUN cd /tmp && \
    wget -O opentracing-cpp-${LIB_OPENTRACING_VERSION}.tar.gz https://github.com/opentracing/opentracing-cpp/archive/v${LIB_OPENTRACING_VERSION}.tar.gz && \
    tar -zxf opentracing-cpp-${LIB_OPENTRACING_VERSION}.tar.gz && \
    cd opentracing-cpp-${LIB_OPENTRACING_VERSION} && \
    mkdir -p cmake-build && cd cmake-build && \
    cmake -DCMAKE_CXX_FLAGS="-fPIC" -DBUILD_TESTING=0 .. && \
    make && make install && \
    cd /tmp && rm -rf opentracing-cpp-${LIB_OPENTRACING_VERSION}*



# Install protobuf (required by OpenTelemetry OTLP exporters)
RUN cd /tmp && \
    wget -O protobuf-${LIB_PROTOBUF_VERSION}.tar.gz https://github.com/protocolbuffers/protobuf/releases/download/v${LIB_PROTOBUF_VERSION}/protobuf-all-${LIB_PROTOBUF_VERSION}.tar.gz && \
    tar -zxf protobuf-${LIB_PROTOBUF_VERSION}.tar.gz && \
    cd protobuf-${LIB_PROTOBUF_VERSION} && \
    ./configure --prefix=/usr/local && \
    make -j$(nproc) && make install && ldconfig && \
    cd /tmp && rm -rf protobuf-${LIB_PROTOBUF_VERSION}*

# Install opentelemetry-cpp (Ubuntu 20.04 comes with CMake 3.16+ which is sufficient)
# Note: Only building OTLP HTTP support (collector uses HTTP endpoint :4318)
# WITH_OPENTRACING=ON builds the OpenTracing shim for backward compatibility
RUN cd /tmp && \
    wget -O opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}.tar.gz https://github.com/open-telemetry/opentelemetry-cpp/archive/v${LIB_OPENTELEMETRY_VERSION}.tar.gz && \
    tar -zxf opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}.tar.gz && \
    cd opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION} && \
    mkdir -p cmake-build && cd cmake-build && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" \
      -DBUILD_TESTING=OFF \
      -DWITH_OTLP_HTTP=ON \
      -DWITH_OPENTRACING=ON \
      -DWITH_EXAMPLES=OFF \
      -DBUILD_SHARED_LIBS=ON \
      -DOPENTELEMETRY_INSTALL=ON \
      -DWITH_DEPRECATED_SDK_FACTORY=OFF .. && \
    make && make install && \
    cd /tmp && rm -rf opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}*

# Install jwt
RUN cd /tmp && \
    wget -O cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz https://github.com/arun11299/cpp-jwt/archive/v${LIB_CPP_JWT_VERSION}.tar.gz && \
    tar -zxf cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz && \
    cd cpp-jwt-${LIB_CPP_JWT_VERSION} && \
    cp -R include/jwt /usr/local/include && \
    rm -rf /usr/local/include/jwt/json && \
    sed -i 's/\#include \"jwt\/json\/json.hpp\"/\#include \<nlohmann\/json\.hpp\>/g' /usr/local/include/jwt/jwt.hpp && \
    cd /tmp && rm -rf cpp-jwt-${LIB_CPP_JWT_VERSION}*

# Install cpp_redis
RUN cd /tmp && \
    git clone https://github.com/cpp-redis/cpp_redis.git && \
    cd cpp_redis && git checkout ${LIB_CPP_REDIS_VERSION} && \
    git submodule init && git submodule update && \
    mkdir cmake-build && cd cmake-build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make && make install && \
    cd /tmp && rm -rf cpp_redis

# Update library cache
ENV LD_LIBRARY_PATH=/usr/local/lib:${LD_LIBRARY_PATH}
RUN ldconfig

# Build media microservices
COPY ./ /media-microservices
RUN cd /media-microservices && \
    mkdir -p build && cd build && \
    cmake .. && \
    make && \
    make install

WORKDIR /media-microservices