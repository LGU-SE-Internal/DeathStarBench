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
#include "opentelemetry/sdk/trace/batch_span_processor.h"
#include "opentelemetry/exporters/otlp/otlp_http_exporter.h"
#include "opentelemetry/exporters/otlp/otlp_grpc_exporter.h"
#include "opentelemetry/sdk/resource/resource.h"
#include "opentelemetry/trace/provider.h"
#include "opentelemetry/context/propagation/global_propagator.h"
#include "opentelemetry/context/propagation/text_map_propagator.h"
#include "opentelemetry/trace/propagation/http_trace_context.h"

namespace media_service {

namespace trace_api = opentelemetry::trace;
namespace trace_sdk = opentelemetry::sdk::trace;
namespace otlp = opentelemetry::exporter::otlp;
namespace resource = opentelemetry::sdk::resource;

void SetUpTracer(
    const std::string &config_file_path,
    const std::string &service) {
  auto configYAML = YAML::LoadFile(config_file_path);
  
  // Get OTEL endpoint from environment variable or config
  std::string otel_endpoint;
  const char* env_endpoint = std::getenv("OTEL_EXPORTER_OTLP_ENDPOINT");
  if (env_endpoint != nullptr) {
    otel_endpoint = std::string(env_endpoint);
  } else if (configYAML["endpoint"]) {
    otel_endpoint = configYAML["endpoint"].as<std::string>();
  } else {
    otel_endpoint = "http://localhost:4318"; // Default HTTP endpoint
  }

  // Get sampling ratio
  double sampling_ratio = 0.01; // Default
  if (configYAML["samplerParam"]) {
    sampling_ratio = configYAML["samplerParam"].as<double>();
  }

  // Create resource attributes
  auto resource_attributes = resource::ResourceAttributes{
    {"service.name", service},
  };
  auto resource_ptr = resource::Resource::Create(resource_attributes);

  bool r = false;
  while (!r) {
    try {
      // Create OTLP HTTP exporter
      otlp::OtlpHttpExporterOptions exporter_options;
      exporter_options.url = otel_endpoint + "/v1/traces";
      
      auto exporter = std::unique_ptr<trace_sdk::SpanExporter>(
        new otlp::OtlpHttpExporter(exporter_options));

      // Create batch span processor
      trace_sdk::BatchSpanProcessorOptions processor_options;
      auto processor = std::unique_ptr<trace_sdk::SpanProcessor>(
        new trace_sdk::BatchSpanProcessor(std::move(exporter), processor_options));

      // Create tracer provider
      auto provider = std::shared_ptr<trace_api::TracerProvider>(
        new trace_sdk::TracerProvider(std::move(processor), resource_ptr));

      // Set global tracer provider
      trace_api::Provider::SetTracerProvider(provider);

      // Set global propagator
      opentelemetry::context::propagation::GlobalTextMapPropagator::SetGlobalPropagator(
        std::make_shared<opentelemetry::trace::propagation::HttpTraceContext>());
      
      r = true;
    } catch(...) {
      LOG(error) << "Failed to connect to OTEL collector, retrying ...";
      sleep(1);
    }
  }
}

} //namespace media_service

#endif //MEDIA_MICROSERVICES_TRACING_H
