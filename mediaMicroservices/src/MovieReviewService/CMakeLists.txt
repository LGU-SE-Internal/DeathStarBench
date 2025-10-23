add_executable(
    MovieReviewService
    MovieReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/MovieReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/ReviewStorageService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    MovieReviewService PRIVATE
    ${MONGOC_INCLUDE_DIRS}
    
    /usr/local/include/cpp_redis
)

target_link_libraries(
    MovieReviewService
    ${MONGOC_LIBRARIES}
    nlohmann_json::nlohmann_json
  ${THRIFT_LIB}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    opentelemetry_trace
    opentelemetry_exporter_otlp_http
    opentelemetry_exporter_otlp_grpc
    opentelemetry_resources
    opentelemetry_common
    /usr/local/lib/libcpp_redis.a
    /usr/local/lib/libtacopie.a
)

install(TARGETS MovieReviewService DESTINATION ./)