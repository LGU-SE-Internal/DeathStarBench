#ifndef SOCIAL_NETWORK_MICROSERVICES_LOGGER_H
#define SOCIAL_NETWORK_MICROSERVICES_LOGGER_H

#include <boost/log/trivial.hpp>
#include <boost/log/utility/setup/console.hpp>
#include <boost/log/utility/setup/common_attributes.hpp>
#include <boost/log/attributes/function.hpp>
#include <boost/log/sinks/basic_sink_backend.hpp>
#include <boost/log/sinks/sync_frontend.hpp>
#include <boost/log/expressions/message.hpp>

// OTel Includes
#include <opentelemetry/trace/provider.h>
#include <opentelemetry/trace/tracer.h>
#include <opentelemetry/logs/provider.h>
#include <opentelemetry/logs/logger.h>
#include <opentelemetry/nostd/span.h>

#include <string.h>
#include <atomic>
#include <mutex>

namespace social_network {

// --- 1. Console helper: get TraceID/SpanID ---
std::string get_current_trace_id() {
    // GetCurrentSpan() always returns a valid span object, but it may represent an invalid/no-op span
    auto span = opentelemetry::trace::Tracer::GetCurrentSpan();
    auto ctx = span->GetContext();
    // Check if we're in an active trace context
    if (ctx.IsValid()) {
        char trace_id_hex[32];
        ctx.trace_id().ToLowerBase16(opentelemetry::nostd::span<char, 32>(trace_id_hex));
        return std::string(trace_id_hex, 32);
    }
    return "00000000000000000000000000000000";
}

std::string get_current_span_id() {
    // GetCurrentSpan() always returns a valid span object, but it may represent an invalid/no-op span
    auto span = opentelemetry::trace::Tracer::GetCurrentSpan();
    auto ctx = span->GetContext();
    // Check if we're in an active trace context
    if (ctx.IsValid()) {
        char span_id_hex[16];
        ctx.span_id().ToLowerBase16(opentelemetry::nostd::span<char, 16>(span_id_hex));
        return std::string(span_id_hex, 16);
    }
    return "0000000000000000";
}

// --- 2. OTLP Sink: send logs to Collector ---
namespace sinks = boost::log::sinks;

class OtelOtlpSinkBackend : public sinks::basic_sink_backend<sinks::concurrent_feeding> {
public:
    OtelOtlpSinkBackend() : otel_logger_(nullptr), initialized_(false) {
    }

    void consume(boost::log::record_view const& rec) {
        // Lazy initialization of logger with thread safety
        if (!initialized_.load(std::memory_order_acquire)) {
            std::lock_guard<std::mutex> lock(init_mutex_);
            if (!initialized_.load(std::memory_order_relaxed)) {
                auto provider = opentelemetry::logs::Provider::GetLoggerProvider();
                if (!provider) return; // Provider not set up yet
                otel_logger_ = provider->GetLogger("social_network_logger");
                initialized_.store(true, std::memory_order_release);
            }
        }
        
        if (!otel_logger_) return;

        auto log_record = otel_logger_->CreateLogRecord();
        log_record->SetTimestamp(std::chrono::system_clock::now());
        
        // Map Severity
        auto severity = rec[boost::log::trivial::severity];
        if (severity) log_record->SetSeverity(MapSeverity(severity.get()));

        // Set Body
        log_record->SetBody(rec[boost::log::expressions::smessage].get());

        // [Key] Inject Trace Context
        auto current_span = opentelemetry::trace::Tracer::GetCurrentSpan();
        auto ctx = current_span->GetContext();
        if (ctx.IsValid()) {
            log_record->SetTraceId(ctx.trace_id());
            log_record->SetSpanId(ctx.span_id());
            log_record->SetTraceFlags(ctx.trace_flags());
        }

        otel_logger_->EmitLogRecord(std::move(log_record));
    }

private:
    opentelemetry::nostd::shared_ptr<opentelemetry::logs::Logger> otel_logger_;
    std::atomic<bool> initialized_;
    std::mutex init_mutex_;

    opentelemetry::logs::Severity MapSeverity(boost::log::trivial::severity_level level) {
        using namespace opentelemetry::logs;
        switch (level) {
            case boost::log::trivial::trace: return Severity::kTrace;
            case boost::log::trivial::debug: return Severity::kDebug;
            case boost::log::trivial::info:  return Severity::kInfo;
            case boost::log::trivial::warning: return Severity::kWarn;
            case boost::log::trivial::error: return Severity::kError;
            case boost::log::trivial::fatal: return Severity::kFatal;
            default: return Severity::kInfo;
        }
    }
};

// --- 3. Initialize Logger ---
#define __FILENAME__ \
    (strrchr(__FILE__, '/') ? strrchr(__FILE__, '/') + 1 : __FILE__)
#define LOG(severity) \
    BOOST_LOG_TRIVIAL(severity) << "(" << __FILENAME__ << ":" \
    << __LINE__ << ":" << __FUNCTION__ << ") "

void init_logger() {
  // A. Register global attributes (TraceID, SpanID) for Console output
  boost::log::register_simple_formatter_factory
      <boost::log::trivial::severity_level, char>("Severity");
  boost::log::add_common_attributes();
  boost::log::core::get()->add_global_attribute("TraceID", boost::log::attributes::make_function(&get_current_trace_id));
  boost::log::core::get()->add_global_attribute("SpanID", boost::log::attributes::make_function(&get_current_span_id));

  // B. Configure Console Sink (with TraceID)
  boost::log::add_console_log(
      std::cerr, boost::log::keywords::format =
          "[%TimeStamp%] <%Severity%> [TraceID:%TraceID%]: %Message%");

  // C. Configure OTLP Sink (send to Collector)
  using sink_t = sinks::synchronous_sink<OtelOtlpSinkBackend>;
  boost::shared_ptr<sink_t> otlp_sink = boost::make_shared<sink_t>();
  // Optional: only send Info and above to Collector to avoid excessive traffic
  otlp_sink->set_filter(boost::log::trivial::severity >= boost::log::trivial::info);
  boost::log::core::get()->add_sink(otlp_sink);

  // D. Global filter
  boost::log::core::get()->set_filter (
      boost::log::trivial::severity >= boost::log::trivial::info
  );
}


} //namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_LOGGER_H
