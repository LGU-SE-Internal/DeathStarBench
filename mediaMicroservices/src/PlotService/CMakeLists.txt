add_executable(
    PlotService
    PlotService.cpp
    ${THRIFT_GEN_CPP_DIR}/PlotService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    PlotService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    
)

target_link_libraries(
    PlotService
    ${MONGOC_LIBRARIES}
    ${LIBMEMCACHED_LIBRARIES}
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    opentelemetry_trace
    opentelemetry_exporter_otlp_http
    opentelemetry_exporter_otlp_grpc
    opentelemetry_resources
    opentelemetry_common
)

target_compile_definitions (
    PlotService PRIVATE
    "${MONGOC_DEFINITIONS}"
)

install(TARGETS PlotService DESTINATION ./)