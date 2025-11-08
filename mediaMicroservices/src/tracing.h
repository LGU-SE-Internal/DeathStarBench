#include <utility>

#ifndef MEDIA_MICROSERVICES_TRACING_H
#define MEDIA_MICROSERVICES_TRACING_H

#include <string>
#include <yaml-cpp/yaml.h>
#include <map>
#include <cstdlib>
#include "logger.h"

#include "opentelemetry/sdk/trace/tracer_provider.h"
#include "opentelemetry/sdk/trace/simple_processor.h"
#include "opentelemetry/exporters/ostream/span_exporter.h"
#include "opentelemetry/sdk/resource/resource.h"
#include "opentelemetry/trace/provider.h"

namespace media_service {

namespace trace_api = opentelemetry::trace;
namespace trace_sdk = opentelemetry::sdk::trace;
namespace resource = opentelemetry::sdk::resource;

void SetUpTracer(
    const std::string &config_file_path,
    const std::string &service) {
  auto configYAML = YAML::LoadFile(config_file_path);

  // Create resource attributes
  auto resource_attributes = resource::ResourceAttributes{
    {"service.name", service},
  };
  auto resource_ptr = resource::Resource::Create(resource_attributes);

  bool r = false;
  while (!r) {
    try {
      // Create console exporter (simple non-OTLP exporter)
      auto exporter = std::unique_ptr<trace_sdk::SpanExporter>(
        new opentelemetry::exporter::trace::OStreamSpanExporter());

      // Create simple span processor
      auto processor = std::unique_ptr<trace_sdk::SpanProcessor>(
        new trace_sdk::SimpleSpanProcessor(std::move(exporter)));

      // Create tracer provider
      auto provider = std::shared_ptr<trace_api::TracerProvider>(
        new trace_sdk::TracerProvider(std::move(processor), resource_ptr));

      // Set global tracer provider
      trace_api::Provider::SetTracerProvider(provider);
      
      r = true;
    } catch(...) {
      LOG(error) << "Failed to initialize OTEL tracer, retrying ...";
      sleep(1);
    }
  }
}

} //namespace media_service

#endif //MEDIA_MICROSERVICES_TRACING_H
