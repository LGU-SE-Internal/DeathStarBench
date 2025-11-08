add_executable(
    TextService
    TextService.cpp
    ${THRIFT_GEN_CPP_DIR}/TextService.cpp
    ${THRIFT_GEN_CPP_DIR}/ComposeReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    TextService PRIVATE
    
)

target_link_libraries(
    TextService
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

install(TARGETS TextService DESTINATION ./)