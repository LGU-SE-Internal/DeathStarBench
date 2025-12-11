#ifndef MEDIA_MICROSERVICES_TEXTHANDLER_H
#define MEDIA_MICROSERVICES_TEXTHANDLER_H

#include <iostream>
#include <string>

#include "../../gen-cpp/TextService.h"
#include "../../gen-cpp/ComposeReviewService.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"

namespace media_service {

class TextHandler : public TextServiceIf {
 public:
  explicit TextHandler(ClientPool<ThriftClient<ComposeReviewServiceClient>> *);
  ~TextHandler() override = default;

  void UploadText(int64_t, const std::string &,
      const std::map<std::string, std::string> &) override;
 private:
  ClientPool<ThriftClient<ComposeReviewServiceClient>> *_compose_client_pool;
};

TextHandler::TextHandler(
    ClientPool<ThriftClient<ComposeReviewServiceClient>> *compose_client_pool) {
  _compose_client_pool = compose_client_pool;
}

void TextHandler::UploadText(
    int64_t req_id,
    const std::string &text,
    const std::map<std::string, std::string> & carrier) {

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
  auto span = tracer->StartSpan("UploadText", options);

  auto scope = tracer->WithActiveSpan(span);

  // Set span attributes following OpenTelemetry semantic conventions
  span->SetAttribute("rpc.system", "thrift");
  span->SetAttribute("rpc.service", "TextService");
  span->SetAttribute("rpc.method", "UploadText");

  // Inject context for downstream services

  std::map<std::string, std::string> writer_text_map;

  TextMapCarrier writer_carrier(writer_text_map);

  propagator->Inject(writer_carrier, opentelemetry::context::RuntimeContext::GetCurrent());

  auto compose_client_wrapper = _compose_client_pool->Pop();
  if (!compose_client_wrapper) {
    ServiceException se;
    se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
    se.message = "Failed to connected to compose-review-service";
    throw se;
  }
  auto compose_client = compose_client_wrapper->GetClient();
  try {
    compose_client->UploadText(req_id, text, writer_text_map);
  } catch (...) {
    _compose_client_pool->Push(compose_client_wrapper);
    LOG(error) << "Failed to upload movie_id to compose-review-service";
    throw;
  }
  _compose_client_pool->Push(compose_client_wrapper);

  span->End();
}
} //namespace media_service

#endif //MEDIA_MICROSERVICES_TEXTHANDLER_H
