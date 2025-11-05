add_executable(
    UserMentionService
    UserMentionService.cpp
    ${THRIFT_GEN_CPP_DIR}/UserMentionService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    UserMentionService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    
)

target_link_libraries(
    UserMentionService
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

install(TARGETS UserMentionService DESTINATION ./)