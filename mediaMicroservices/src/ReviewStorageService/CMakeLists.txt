add_executable(
    ReviewStorageService
    ReviewStorageService.cpp
    ${THRIFT_GEN_CPP_DIR}/ReviewStorageService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    ReviewStorageService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    
)

target_link_libraries(
    ReviewStorageService
    ${MONGOC_LIBRARIES}
    ${LIBMEMCACHED_LIBRARIES}
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    opentelemetry_trace
    opentelemetry_resources
    opentelemetry_common
)

install(TARGETS ReviewStorageService DESTINATION ./)