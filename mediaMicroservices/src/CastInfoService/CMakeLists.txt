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
    yaml-cpp
    opentelemetry_trace
    opentelemetry_resources
    opentelemetry_common
    opentelemetry_exporter_ostream_span
)

target_compile_definitions (
    CastInfoService PRIVATE
    "${MONGOC_DEFINITIONS}"
)

install(TARGETS CastInfoService DESTINATION ./)