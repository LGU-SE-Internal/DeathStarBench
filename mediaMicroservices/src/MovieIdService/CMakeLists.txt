add_executable(
    MovieIdService
    MovieIdService.cpp
    ${THRIFT_GEN_CPP_DIR}/MovieIdService.cpp
    ${THRIFT_GEN_CPP_DIR}/RatingService.cpp
    ${THRIFT_GEN_CPP_DIR}/ComposeReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    MovieIdService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    
)

target_link_libraries(
    MovieIdService
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
    MovieIdService PRIVATE
    "${MONGOC_DEFINITIONS}"
)

install(TARGETS MovieIdService DESTINATION ./)