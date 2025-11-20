add_executable(
    UniqueIdService
    UniqueIdService.cpp
    ${THRIFT_GEN_CPP_DIR}/UniqueIdService.cpp
    ${THRIFT_GEN_CPP_DIR}/ComposeReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    UniqueIdService PRIVATE
    /usr/local/include/opentelemetry
)

target_link_libraries(
    UniqueIdService
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    opentelemetry_trace opentelemetry_exporter_otlp_grpc opentelemetry_resources opentelemetry_common opentelemetry_otlp_recordable opentelemetry_opentracing_shim opentracing yaml-cpp
)

install(TARGETS UniqueIdService DESTINATION ./)