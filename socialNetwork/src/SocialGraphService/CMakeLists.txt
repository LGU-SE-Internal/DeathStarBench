add_executable(
    SocialGraphService
    SocialGraphService.cpp
    ${THRIFT_GEN_CPP_DIR}/SocialGraphService.cpp
    ${THRIFT_GEN_CPP_DIR}/UserService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    SocialGraphService PRIVATE
    ${MONGOC_INCLUDE_DIRS}
    /usr/local/include/opentelemetry
    /usr/local/include/hiredis
    /usr/local/include/sw
)

target_link_libraries(
    SocialGraphService
    ${MONGOC_LIBRARIES}
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    nlohmann_json::nlohmann_json
    Boost::log
    Boost::log_setup
    Boost::program_options
    /usr/local/lib/libopentelemetry_trace opentelemetry_exporter_otlp_http opentelemetry_resources opentelemetry_common opentelemetry_otlp_recordable opentelemetry_opentelemetry_opentracing_shim opentracing
    /usr/local/lib/libhiredis.a
    /usr/local/lib/libhiredis_ssl.a
    /usr/local/lib/libredis++.a
    OpenSSL::SSL
)

install(TARGETS SocialGraphService DESTINATION ./)