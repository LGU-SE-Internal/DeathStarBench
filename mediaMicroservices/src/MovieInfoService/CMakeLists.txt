add_executable(
    MovieInfoService
    MovieInfoService.cpp
    ${THRIFT_GEN_CPP_DIR}/MovieInfoService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    MovieInfoService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    /usr/local/include/opentelemetry
)

target_link_libraries(
    MovieInfoService
    ${MONGOC_LIBRARIES}
    ${LIBMEMCACHED_LIBRARIES}
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    opentelemetry_trace opentelemetry_exporter_jaeger_trace opentelemetry_exporter_otlp_http opentelemetry_resources opentelemetry_common opentelemetry_otlp_recordable opentracing_shim
)

target_compile_definitions (
    MovieInfoService PRIVATE
    "${MONGOC_DEFINITIONS}"
)

install(TARGETS MovieInfoService DESTINATION ./)