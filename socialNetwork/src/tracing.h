#include <utility>

#ifndef SOCIAL_NETWORK_MICROSERVICES_TRACING_H
#define SOCIAL_NETWORK_MICROSERVICES_TRACING_H

#include <string>
#include <cstdlib>
#include <yaml-cpp/yaml.h>
#include <map>

#include <opentracing/propagation.h>
#include <opentelemetry/sdk/trace/tracer_provider_factory.h>
#include <opentelemetry/sdk/trace/simple_processor_factory.h>
#include <opentelemetry/sdk/resource/resource.h>
#include <opentelemetry/exporters/otlp/otlp_http_exporter_factory.h>
#include <opentelemetry/trace/provider.h>
#include <opentelemetry/opentracingshim/tracer_shim.h>
#include "logger.h"

namespace social_network {

using opentracing::expected;
using opentracing::string_view;

class TextMapReader : public opentracing::TextMapReader {
 public:
  explicit TextMapReader(const std::map<std::string, std::string> &text_map)
      : _text_map(text_map) {}

  expected<void> ForeachKey(
      std::function<expected<void>(string_view key, string_view value)> f)
  const override {
    for (const auto& key_value : _text_map) {
      auto result = f(key_value.first, key_value.second);
      if (!result) return result;
    }
    return {};
  }

 private:
  const std::map<std::string, std::string>& _text_map;
};

class TextMapWriter : public opentracing::TextMapWriter {
 public:
  explicit TextMapWriter(std::map<std::string, std::string> &text_map)
    : _text_map(text_map) {}

  expected<void> Set(string_view key, string_view value) const override {
    _text_map[key] = value;
    return {};
  }

 private:
  std::map<std::string, std::string>& _text_map;
};

void SetUpTracer(
    const std::string &config_file_path,
    const std::string &service) {
  auto configYAML = YAML::LoadFile(config_file_path);

  bool r = false;
  while (!r) {
    try {
      std::vector<std::unique_ptr<opentelemetry::sdk::trace::SpanProcessor>> processors;
      
      // Use OTLP HTTP exporter
      // Get endpoint from environment variable or use default
      const char* otlp_endpoint_env = std::getenv("OTEL_EXPORTER_OTLP_ENDPOINT");
      std::string otlp_endpoint = otlp_endpoint_env != nullptr ? otlp_endpoint_env : "http://localhost:4318";
      
      opentelemetry::exporter::otlp::OtlpHttpExporterOptions otlp_options;
      otlp_options.url = otlp_endpoint + "/v1/traces";
      
      auto exporter = opentelemetry::exporter::otlp::OtlpHttpExporterFactory::Create(otlp_options);
      auto processor = opentelemetry::sdk::trace::SimpleSpanProcessorFactory::Create(std::move(exporter));
      processors.push_back(std::move(processor));
      
      LOG(info) << "Using OpenTelemetry OTLP HTTP exporter: " << otlp_options.url;
      
      // Create resource with service name
      auto resource_attributes = opentelemetry::sdk::resource::ResourceAttributes{
        {"service.name", service}
      };
      auto resource = opentelemetry::sdk::resource::Resource::Create(resource_attributes);
      
      auto provider = opentelemetry::sdk::trace::TracerProviderFactory::Create(
        std::move(processors), resource);
      
      // Set the global tracer provider
      opentelemetry::trace::Provider::SetTracerProvider(std::move(provider));
      
      // Create OpenTracing shim
      auto tracer_shim = opentelemetry::opentracingshim::TracerShim::createTracerShim();
      opentracing::Tracer::InitGlobal(tracer_shim);
      
      r = true;
    }
    catch(const std::exception& e) {
      LOG(error) << "Failed to setup tracer: " << e.what() << ", retrying ...";
      sleep(1);
    }
    catch(...) {
      LOG(error) << "Failed to setup tracer, retrying ...";
      sleep(1);
    }
  }
}


} //namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_TRACING_H
