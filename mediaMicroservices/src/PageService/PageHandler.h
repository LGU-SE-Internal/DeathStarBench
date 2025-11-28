#ifndef MEDIA_MICROSERVICES_SRC_COMPOSEPAGESERVICE_COMPOSEPAGEHANDLER_H_
#define MEDIA_MICROSERVICES_SRC_COMPOSEPAGESERVICE_COMPOSEPAGEHANDLER_H_

#include <iostream>
#include <string>
#include <future>

#include "../../gen-cpp/PageService.h"
#include "../../gen-cpp/MovieReviewService.h"
#include "../../gen-cpp/MovieInfoService.h"
#include "../../gen-cpp/CastInfoService.h"
#include "../../gen-cpp/PlotService.h"
#include "../logger.h"
#include "../tracing.h"
#include "../context_helper.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"

namespace media_service {

class PageHandler : public PageServiceIf {
 public:
  PageHandler(
      ClientPool<ThriftClient<MovieReviewServiceClient>> *,
      ClientPool<ThriftClient<MovieInfoServiceClient>> *,
      ClientPool<ThriftClient<CastInfoServiceClient>> *,
      ClientPool<ThriftClient<PlotServiceClient>> *);
  ~PageHandler() override = default;

  void ReadPage(Page& _return, int64_t req_id, const std::string& movie_id,
                int32_t review_start, int32_t review_stop,
                const std::map<std::string, std::string> & carrier) override;

 private:
  ClientPool<ThriftClient<MovieReviewServiceClient>> *_movie_review_client_pool;
  ClientPool<ThriftClient<MovieInfoServiceClient>> *_movie_info_client_pool;
  ClientPool<ThriftClient<CastInfoServiceClient>> *_cast_info_client_pool;
  ClientPool<ThriftClient<PlotServiceClient>> *_plot_client_pool;
};
PageHandler::PageHandler(
    ClientPool<ThriftClient<MovieReviewServiceClient>> *movie_review_client_pool,
    ClientPool<ThriftClient<MovieInfoServiceClient>> *movie_info_client_pool,
    ClientPool<ThriftClient<CastInfoServiceClient>> *cast_info_client_pool,
    ClientPool<ThriftClient<PlotServiceClient>> *plot_client_pool) {
  _movie_review_client_pool = movie_review_client_pool;
  _movie_info_client_pool = movie_info_client_pool;
  _cast_info_client_pool = cast_info_client_pool;
  _plot_client_pool = plot_client_pool;
}
void PageHandler::ReadPage(
    Page &_return,
    int64_t req_id,
    const std::string &movie_id,
    int32_t review_start,
    int32_t review_stop,
    const std::map<std::string, std::string> &carrier) {

  // Get tracer and propagator

  auto tracer = opentelemetry::trace::Provider::GetTracerProvider()->GetTracer("media_service");

  auto propagator = opentelemetry::context::propagation::GlobalTextMapPropagator::GetGlobalPropagator();

  

  // Extract context from carrier

  std::map<std::string, std::string> carrier_copy = carrier;

  TextMapCarrier carrier_reader(carrier_copy);

  auto current_ctx = opentelemetry::context::RuntimeContext::GetCurrent();


  auto parent_ctx = propagator->Extract(carrier_reader, current_ctx);

  

  // Start span with extracted context as parent

  opentelemetry::trace::StartSpanOptions options;

  options.kind = opentelemetry::trace::SpanKind::kServer;

  options.parent = parent_ctx;
  auto span = tracer->StartSpan("ReadPage", options);

  auto scope = tracer->WithActiveSpan(span);

  

  // Inject context for downstream services

  std::map<std::string, std::string> writer_text_map;

  TextMapCarrier writer_carrier(writer_text_map);

  propagator->Inject(writer_carrier, opentelemetry::context::RuntimeContext::GetCurrent());

  std::future<std::vector<Review>> movie_review_future;
  std::future<MovieInfo> movie_info_future;
  std::future<std::vector<CastInfo>> cast_info_future;
  std::future<std::string> plot_future;

  movie_info_future = std::async(std::launch::async, [&](){
    MovieInfo _reture_movie_info;
    auto movie_info_client_wrapper = _movie_info_client_pool->Pop();
    if (!movie_info_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connected to movie-info-service";
      throw se;
    }
    auto movie_info_client = movie_info_client_wrapper->GetClient();
    try {
      movie_info_client->ReadMovieInfo(_reture_movie_info,
          req_id, movie_id, writer_text_map);
    } catch (...) {
      _movie_info_client_pool->Push(movie_info_client_wrapper);
      LOG(error) << "Failed to read movie_info to movie-info-service";
      throw;
    }
    _movie_info_client_pool->Push(movie_info_client_wrapper);
    return _reture_movie_info;
  });

  movie_review_future = std::async(std::launch::async, [&](){
    std::vector<Review> _return_movie_reviews;
    auto movie_review_client_wrapper = _movie_review_client_pool->Pop();
    if (!movie_review_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connected to movie-review-service";
      throw se;
    }
    auto movie_review_client = movie_review_client_wrapper->GetClient();
    try {
      movie_review_client->ReadMovieReviews(_return_movie_reviews,
          req_id, movie_id, review_start, review_stop, writer_text_map);
    } catch (...) {
      _movie_review_client_pool->Push(movie_review_client_wrapper);
      LOG(error) << "Failed to read reviews to movie-review-service";
      throw;
    }
    _movie_review_client_pool->Push(movie_review_client_wrapper);
    return _return_movie_reviews;
  });

  try {
    _return.movie_info = movie_info_future.get();
  } catch (...) {
    throw;
  }
  
  std::vector<int64_t> cast_info_ids;
  std::set<int64_t> cast_info_ids_set;
  for (auto &cast : _return.movie_info.casts) {
    // Deduplicate cast_info_ids to avoid "cast_info_ids are duplicated" error
    if (cast_info_ids_set.find(cast.cast_info_id) == cast_info_ids_set.end()) {
      cast_info_ids_set.insert(cast.cast_info_id);
      cast_info_ids.emplace_back(cast.cast_info_id);
    }
  }

  cast_info_future = std::async(std::launch::async, [&](){
    std::vector<CastInfo> _return_cast_infos;
    auto cast_info_client_wrapper = _cast_info_client_pool->Pop();
    if (!cast_info_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connected to cast-info-service";
      throw se;
    }
    auto cast_info_client = cast_info_client_wrapper->GetClient();
    try {
      cast_info_client->ReadCastInfo(_return_cast_infos, req_id,
          cast_info_ids, writer_text_map);
    } catch (...) {
      _cast_info_client_pool->Push(cast_info_client_wrapper);
      LOG(error) << "Failed to read cast-info to cast-info-service";
      throw;
    }
    _cast_info_client_pool->Push(cast_info_client_wrapper);
    return _return_cast_infos;
  });

  plot_future = std::async(std::launch::async, [&](){
    std::string _return_plot;
    auto plot_client_wrapper = _plot_client_pool->Pop();
    if (!plot_client_wrapper) {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
      se.message = "Failed to connected to plot-service";
      throw se;
    }
    auto plot_client = plot_client_wrapper->GetClient();
    try {
      plot_client->ReadPlot(_return_plot, req_id, _return.movie_info.plot_id,
          writer_text_map);
    } catch (...) {
      _plot_client_pool->Push(plot_client_wrapper);
      LOG(error) << "Failed to read plot to plot-service";
      throw;
    }
    _plot_client_pool->Push(plot_client_wrapper);
    return _return_plot;
  });

  try {
    _return.reviews = movie_review_future.get();
    _return.plot = plot_future.get();
    _return.cast_infos = cast_info_future.get();
  } catch (...) {
    throw;
  }
  span->End();
}
} //namespace media_service

#endif //MEDIA_MICROSERVICES_SRC_COMPOSEPAGESERVICE_COMPOSEPAGEHANDLER_H_
