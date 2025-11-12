#include <utility>

#ifndef MEDIA_MICROSERVICES_TRACING_H
#define MEDIA_MICROSERVICES_TRACING_H

#include <string>
#include <yaml-cpp/yaml.h>
#include <map>

#include <opentracing/propagation.h>
#include <opentelemetry/sdk/trace/tracer_provider_factory.h>
#include <opentelemetry/sdk/trace/simple_processor_factory.h>
#include <opentelemetry/exporters/jaeger/jaeger_exporter_factory.h>
#include <opentelemetry/trace/provider.h>
#include "opentracing-shim/include/tracer.h"

namespace media_service {

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
  
  // Parse Jaeger endpoint configuration
  std::string jaeger_endpoint = "localhost:6831";
  if (configYAML["reporter"] && configYAML["reporter"]["localAgentHostPort"]) {
    jaeger_endpoint = configYAML["reporter"]["localAgentHostPort"].as<std::string>();
  }

  // Configure OpenTelemetry Jaeger exporter
  opentelemetry::exporter::jaeger::JaegerExporterOptions jaeger_options;
  jaeger_options.endpoint = jaeger_endpoint;
  
  auto exporter = opentelemetry::exporter::jaeger::JaegerExporterFactory::Create(jaeger_options);
  auto processor = opentelemetry::sdk::trace::SimpleSpanProcessorFactory::Create(std::move(exporter));
  
  std::vector<std::unique_ptr<opentelemetry::sdk::trace::SpanProcessor>> processors;
  processors.push_back(std::move(processor));
  
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
  auto tracer_shim = opentracing::shim::TracerShim::createTracerShim();
  opentracing::Tracer::InitGlobal(tracer_shim);
}


} //namespace media_service

#endif //MEDIA_MICROSERVICES_TRACING_H
