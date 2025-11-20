add_executable(
    MediaService
    MediaService.cpp
    ${THRIFT_GEN_CPP_DIR}/MediaService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    MediaService PRIVATE
    /usr/local/include/opentelemetry
)

target_link_libraries(
    MediaService
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    opentelemetry_trace opentelemetry_exporter_otlp_grpc opentelemetry_resources opentelemetry_common opentelemetry_otlp_recordable opentelemetry_opentracing_shim opentracing yaml-cpp
)

install(TARGETS MediaService DESTINATION ./)