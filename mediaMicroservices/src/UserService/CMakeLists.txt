add_executable(
    UserService
    UserService.cpp
    ${THRIFT_GEN_CPP_DIR}/UserService.cpp
    ${THRIFT_GEN_CPP_DIR}/ComposeReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)


target_include_directories(
    UserService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    /usr/local/include/jwt
    
)

target_link_libraries(
    UserService
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
    OpenSSL::SSL
)

install(TARGETS UserService DESTINATION ./)