add_executable(
    WriteHomeTimelineService
    WriteHomeTimelineService.cpp
    ${THRIFT_GEN_CPP_DIR}/SocialGraphService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    WriteHomeTimelineService PRIVATE
    /usr/local/include/opentelemetry
    /usr/local/include/cpp_redis
    ${LIBEVENT_INCLUDE_DIRS}
)


target_link_libraries(
    WriteHomeTimelineService
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    nlohmann_json::nlohmann_json
    Boost::log
    Boost::log_setup
    OpenSSL::SSL
    opentelemetry_trace opentelemetry_exporter_otlp_grpc opentelemetry_resources opentelemetry_common opentelemetry_otlp_recordable opentelemetry_opentracing_shim opentracing yaml-cpp
    /usr/local/lib/libamqpcpp.so
    ${LIBEVENT_LIBRARIES}
    /usr/local/lib/libcpp_redis.a
    /usr/local/lib/libtacopie.a
)

install(TARGETS WriteHomeTimelineService DESTINATION ./)