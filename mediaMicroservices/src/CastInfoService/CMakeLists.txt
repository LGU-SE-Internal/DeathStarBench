add_executable(
    CastInfoService
    CastInfoService.cpp
    ${THRIFT_GEN_CPP_DIR}/CastInfoService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    CastInfoService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    /usr/local/include/opentelemetry
)

target_link_libraries(
    CastInfoService
    ${MONGOC_LIBRARIES}
    ${LIBMEMCACHED_LIBRARIES}
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    opentelemetry_trace opentelemetry_exporter_otlp_grpc opentelemetry_resources opentelemetry_common opentelemetry_otlp_recordable opentelemetry_opentracing_shim opentracing yaml-cpp
)

target_compile_definitions (
    CastInfoService PRIVATE
    "${MONGOC_DEFINITIONS}"
)

install(TARGETS CastInfoService DESTINATION ./)