# Build the dependencies with OpenTelemetry
FROM ubuntu:16.04 AS deps-builder

ARG LIB_MONGOC_VERSION=1.14.0
ARG LIB_THRIFT_VERSION=0.12.0
ARG LIB_JSON_VERSION=3.6.1
ARG LIB_OPENTELEMETRY_VERSION=1.8.1
ARG LIB_YAML_VERSION=0.6.2
ARG LIB_CPP_JWT_VERSION=1.1.1
ARG LIB_CPP_REDIS_VERSION=4.3.1

ARG BUILD_DEPS="ca-certificates g++ wget git libmemcached-dev automake bison flex libboost-all-dev libevent-dev libssl-dev libtool make pkg-config"
ARG CMAKE_VERSION=3.20.0

RUN apt-get update \
  && apt-get install -y ${BUILD_DEPS} --no-install-recommends \
  # Install newer CMake (OpenTelemetry requires CMake >= 3.7)
  && cd /tmp \
  && wget https://github.com/Kitware/CMake/releases/download/v${CMAKE_VERSION}/cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz \
  && tar -zxf cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz \
  && cp -r cmake-${CMAKE_VERSION}-linux-x86_64/bin/* /usr/local/bin/ \
  && cp -r cmake-${CMAKE_VERSION}-linux-x86_64/share/* /usr/local/share/ \
  && rm -rf cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz cmake-${CMAKE_VERSION}-linux-x86_64 \
  # Install mongo-c-driver
  && cd /tmp \
  && wget https://github.com/mongodb/mongo-c-driver/releases/download/${LIB_MONGOC_VERSION}/mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
  && tar -zxf mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
  && cd mongo-c-driver-${LIB_MONGOC_VERSION} \
  && mkdir -p cmake-build \
  && cd cmake-build \
  && cmake -DENABLE_TESTS=0 -DENABLE_EXAMPLES=0 .. \
  && make \
  && make install \
  && cd /tmp \
  # Install lib-thrift
  && wget -O thrift-${LIB_THRIFT_VERSION}.tar.gz https://github.com/apache/thrift/archive/v${LIB_THRIFT_VERSION}.tar.gz \
  && tar -zxf thrift-${LIB_THRIFT_VERSION}.tar.gz \
  && cd thrift-${LIB_THRIFT_VERSION} \
  && mkdir -p cmake-build \
  && cd cmake-build \
  && cmake -DBUILD_TESTING=0 .. \
  && make \
  && make install \
  && cd /tmp \
  # Install /nlohmann/json
  && wget -O json-${LIB_JSON_VERSION}.tar.gz https://github.com/nlohmann/json/archive/v${LIB_JSON_VERSION}.tar.gz \
  && tar -zxf json-${LIB_JSON_VERSION}.tar.gz \
  && cd json-${LIB_JSON_VERSION} \
  && mkdir -p cmake-build \
  && cd cmake-build \
  && cmake -DBUILD_TESTING=0 .. \
  && make \
  && make install \
  && cd /tmp \
  # Install yaml-cpp
  && wget -O yaml-cpp-${LIB_YAML_VERSION}.tar.gz https://github.com/jbeder/yaml-cpp/archive/yaml-cpp-${LIB_YAML_VERSION}.tar.gz \
  && tar -zxf yaml-cpp-${LIB_YAML_VERSION}.tar.gz \
  && cd yaml-cpp-yaml-cpp-${LIB_YAML_VERSION} \
  && mkdir -p cmake-build \
  && cd cmake-build \
  && cmake -DCMAKE_CXX_FLAGS="-fPIC" -DYAML_CPP_BUILD_TESTS=0 .. \
  && make \
  && make install \
  && cd /tmp \
  # Install Protobuf (required for OpenTelemetry OTLP)
  && wget -O protobuf-3.19.4.tar.gz https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protobuf-all-3.19.4.tar.gz \
  && tar -zxf protobuf-3.19.4.tar.gz \
  && cd protobuf-3.19.4 \
  && ./configure --prefix=/usr/local \
  && make -j$(nproc) \
  && make install \
  && ldconfig \
  && cd /tmp \
  # Install gRPC (required for OpenTelemetry OTLP)
  && git clone --recurse-submodules -b v1.43.0 --depth 1 https://github.com/grpc/grpc.git \
  && cd grpc \
  && mkdir -p cmake-build \
  && cd cmake-build \
  && cmake -DgRPC_INSTALL=ON \
    -DgRPC_BUILD_TESTS=OFF \
    -DCMAKE_BUILD_TYPE=Release \
    -DgRPC_PROTOBUF_PROVIDER=package \
    -DgRPC_SSL_PROVIDER=package \
    .. \
  && make -j$(nproc) \
  && make install \
  && cd /tmp \
  # Install OpenTelemetry C++
  && wget -O opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}.tar.gz https://github.com/open-telemetry/opentelemetry-cpp/archive/v${LIB_OPENTELEMETRY_VERSION}.tar.gz \
  && tar -zxf opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}.tar.gz \
  && cd opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION} \
  && mkdir -p cmake-build \
  && cd cmake-build \
  && cmake -DCMAKE_CXX_FLAGS="-fPIC" \
    -DBUILD_TESTING=OFF \
    -DWITH_OTLP=ON \
    -DWITH_OTLP_GRPC=ON \
    -DWITH_OTLP_HTTP=ON \
    -DWITH_EXAMPLES=OFF \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
    .. \
  && make -j$(nproc) \
  && make install \
  && cd /tmp \
  && wget -O cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz https://github.com/arun11299/cpp-jwt/archive/v${LIB_CPP_JWT_VERSION}.tar.gz \
  && tar -zxf cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz \
  && cd cpp-jwt-${LIB_CPP_JWT_VERSION} \
  && cp -R include/jwt /usr/local/include \
  # use the dependency in /usr/local/include instead of in jwt/json
  && rm -rf /usr/local/include/jwt/json \
  && sed -i 's/\#include \"jwt\/json\/json.hpp\"/\#include \<nlohmann\/json\.hpp\>/g' /usr/local/include/jwt/jwt.hpp \
  # Install cpp_redis
  && cd /tmp \
  && git clone https://github.com/cpp-redis/cpp_redis.git \
  && cd cpp_redis && git checkout ${LIB_CPP_REDIS_VERSION} \
  && git submodule init && git submodule update \
  && mkdir cmake-build && cd cmake-build \
  && cmake .. -DCMAKE_BUILD_TYPE=Release \
  && make \
  && make install \
  && cd /tmp \
  && rm -rf \
    mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
    mongo-c-driver-${LIB_MONGOC_VERSION} \
    thrift-${LIB_THRIFT_VERSION}.tar.gz \
    thrift-${LIB_THRIFT_VERSION} \
    json-${LIB_JSON_VERSION}.tar.gz \
    json-${LIB_JSON_VERSION} \
    yaml-cpp-${LIB_YAML_VERSION}.tar.gz \
    yaml-cpp-yaml-cpp-${LIB_YAML_VERSION} \
    protobuf-3.19.4.tar.gz \
    protobuf-3.19.4 \
    grpc \
    opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION}.tar.gz \
    opentelemetry-cpp-${LIB_OPENTELEMETRY_VERSION} \
    cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz \
    cpp-jwt-${LIB_CPP_JWT_VERSION} \
    cpp_redis

ENV LD_LIBRARY_PATH="/usr/local/lib:${LD_LIBRARY_PATH:-}"
RUN ldconfig

# Build the application
FROM deps-builder AS app-builder

COPY ./ /media-microservices
RUN cd /media-microservices \
    && mkdir -p build \
    && cd build \
    && cmake .. \
    && make \
    && make install

WORKDIR /media-microservices