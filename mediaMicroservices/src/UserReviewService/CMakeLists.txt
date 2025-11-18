add_executable(
    UserReviewService
    UserReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/UserReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/ReviewStorageService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    UserReviewService PRIVATE
    ${MONGOC_INCLUDE_DIRS}
    /usr/local/include/opentelemetry
    /usr/local/include/cpp_redis
)

target_link_libraries(
    UserReviewService
    ${MONGOC_LIBRARIES}
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    opentelemetry_trace opentelemetry_exporter_otlp_http opentelemetry_resources opentelemetry_common opentelemetry_otlp_recordable opentelemetry_opentracing_shim opentracing yaml-cpp
    /usr/local/lib/libcpp_redis.a
    /usr/local/lib/libtacopie.a
)

install(TARGETS UserReviewService DESTINATION ./)