#ifndef MEDIA_MICROSERVICES_CONTEXT_HELPER_H
#define MEDIA_MICROSERVICES_CONTEXT_HELPER_H

#include <opentelemetry/trace/tracer.h>
#include <opentelemetry/trace/scope.h>
#include <opentelemetry/context/runtime_context.h>
#include <functional>

namespace media_service {

/**
 * Wraps a function/Lambda to automatically propagate the current OpenTelemetry Context when executed.
 * Used for std::thread, std::async, thread pools, etc.
 * This captures the full Context (not just Span), ensuring proper trace propagation across threads.
 */
template<typename Function>
auto WrapAsync(Function&& func) {
    // 1. Capture current Context in the main thread (includes active span, baggage, etc.)
    auto current_ctx = opentelemetry::context::RuntimeContext::GetCurrent();

    // 2. Return a new Lambda that holds the context and activates it during execution
    return [current_ctx, func = std::forward<Function>(func)]() mutable {
        // 3. Attach the captured Context in the sub-thread
        // Token ensures Context is restored when it goes out of scope
        auto token = opentelemetry::context::RuntimeContext::Attach(current_ctx);
        
        // 4. Execute the original function
        return func();
    };
}

} // namespace media_service

#endif // MEDIA_MICROSERVICES_CONTEXT_HELPER_H
