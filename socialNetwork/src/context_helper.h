#ifndef SOCIAL_NETWORK_MICROSERVICES_CONTEXT_HELPER_H
#define SOCIAL_NETWORK_MICROSERVICES_CONTEXT_HELPER_H

#include <opentelemetry/trace/tracer.h>
#include <opentelemetry/trace/scope.h>
#include <functional>

namespace social_network {

/**
 * Wraps a function/Lambda to automatically activate the current Span Context when executed.
 * Used for std::thread, std::async, thread pools, etc.
 */
template<typename Function>
auto WrapAsync(Function&& func) {
    // 1. Capture current Span in the main thread
    auto current_span = opentelemetry::trace::Tracer::GetCurrentSpan();

    // 2. Return a new Lambda that holds the span and activates it during execution
    return [current_span, func = std::forward<Function>(func)]() mutable {
        // 3. Activate Scope in the sub-thread
        opentelemetry::trace::Scope scope(current_span);
        // 4. Execute the original function
        return func();
    };
}

} // namespace social_network

#endif // SOCIAL_NETWORK_MICROSERVICES_CONTEXT_HELPER_H
