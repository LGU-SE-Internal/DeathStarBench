#include <utility>

#ifndef SOCIAL_NETWORK_MICROSERVICES_TRACING_H
#define SOCIAL_NETWORK_MICROSERVICES_TRACING_H

#include <string>
#include <cstdlib>
#include <map>
#include <vector>
#include <iostream>

// OTel SDK Includes
#include <opentelemetry/sdk/trace/tracer_provider_factory.h>
#include <opentelemetry/sdk/trace/simple_processor_factory.h>
#include <opentelemetry/sdk/resource/resource.h>
#include <opentelemetry/exporters/otlp/otlp_http_exporter_factory.h>
#include <opentelemetry/trace/provider.h>
#include <opentelemetry/context/propagation/global_propagator.h>
#include <opentelemetry/context/propagation/text_map_propagator.h>
#include <opentelemetry/trace/propagation/http_trace_context.h>
#include <opentelemetry/context/runtime_context.h>

// OTel Logs Includes
#include <opentelemetry/logs/provider.h>
#include <opentelemetry/sdk/logs/logger_provider_factory.h>
#include <opentelemetry/sdk/logs/simple_log_record_processor_factory.h>
#include <opentelemetry/exporters/otlp/otlp_http_log_record_exporter_factory.h>
#include "logger.h"

namespace social_network {

// --- OTel Carrier Implementation for Context Propagation ---
class TextMapCarrier : public opentelemetry::context::propagation::TextMapCarrier {
public:
    explicit TextMapCarrier(std::map<std::string, std::string>& headers) : headers_(headers) {}
    
    // Get (used for Extract)
    opentelemetry::nostd::string_view Get(opentelemetry::nostd::string_view key) const noexcept override {
        auto it = headers_.find(std::string(key));
        if (it != headers_.end()) {
            return opentelemetry::nostd::string_view(it->second);
        }
        return "";
    }

    // Set (used for Inject)
    void Set(opentelemetry::nostd::string_view key, opentelemetry::nostd::string_view value) noexcept override {
        headers_[std::string(key)] = std::string(value);
    }

private:
    std::map<std::string, std::string>& headers_;
};

void SetUpLogProvider(const std::string &service) {
  // 1. Get OTLP Endpoint (reuse environment variable logic)
  const char* otlp_endpoint_env = std::getenv("OTEL_EXPORTER_OTLP_ENDPOINT");
  std::string otlp_endpoint = otlp_endpoint_env != nullptr ? otlp_endpoint_env : "http://localhost:4318";

  // 2. Configure Log Exporter
  opentelemetry::exporter::otlp::OtlpHttpLogRecordExporterOptions otlp_options;
  otlp_options.url = otlp_endpoint + "/v1/logs";
  auto exporter = opentelemetry::exporter::otlp::OtlpHttpLogRecordExporterFactory::Create(otlp_options);

  // 3. Create Processor
  auto processor = opentelemetry::sdk::logs::SimpleLogRecordProcessorFactory::Create(std::move(exporter));

  // 4. Create Resource (keep consistent with Trace)
  auto resource_attributes = opentelemetry::sdk::resource::ResourceAttributes{
    {"service.name", service}
  };
  auto resource = opentelemetry::sdk::resource::Resource::Create(resource_attributes);

  // 5. Create and set global Provider
  auto provider = opentelemetry::sdk::logs::LoggerProviderFactory::Create(
    std::move(processor), resource);
  
  // Convert unique_ptr to shared_ptr in two steps to avoid ambiguity:
  // 1. Move unique_ptr into std::shared_ptr (handles upcast from SDK to API)
  std::shared_ptr<opentelemetry::logs::LoggerProvider> std_provider(std::move(provider));
  // 2. Convert std::shared_ptr to nostd::shared_ptr
  opentelemetry::logs::Provider::SetLoggerProvider(
    opentelemetry::nostd::shared_ptr<opentelemetry::logs::LoggerProvider>(std_provider));
}

void SetUpTracer(const std::string &service) {
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
      opentelemetry::trace::Provider::SetTracerProvider(
        std::shared_ptr<opentelemetry::trace::TracerProvider>(std::move(provider)));
      
      // Set the global propagator for context propagation (W3C Trace Context)
      opentelemetry::context::propagation::GlobalTextMapPropagator::SetGlobalPropagator(
        opentelemetry::nostd::shared_ptr<opentelemetry::context::propagation::TextMapPropagator>(
          new opentelemetry::trace::propagation::HttpTraceContext()));
      
      // Initialize Log Provider
      SetUpLogProvider(service);
      
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
