# Extensibility Overview
One way to extend Envoy is by implementing different filters that process or augment the requests. These filters can generate statistics, translate protocols, modify the requests, and so on.

An example of filters is HTTP filters, such as the external authz filter and other filters built into the Envoy binary.

Additionally, we can also write our filters that Envoy dynamically loads and runs. We can decide where we want to run the filter in the filter chain by declaring it in the correct order.

We have a couple of options for extending Envoy. By default, Envoy filters are written in C++. However, we can write them in Lua script or use WebAssembly (WASM) to develop Envoy filters in other programming languages.

Note that the Lua and Wasm filters are limited in their APIs compared to the C++ filters.

## Native C++ API
The first option is to write native C++ filters and then package them with Envoy. This would require us to recompile Envoy and maintain our version of it. Taking this route makes sense if we’re trying to solve complex or high-performance use cases.

## Lua filter
The second option is using the Lua script. There is an HTTP filter in Envoy that allows us to define a Lua script either inline or as an external file and execute it during both the request and response flows.

## Wasm filter
The last option is Wasm-based filters. We write the filter as a separate Wasm module with this option, and Envoy loads it dynamically at run time.

In the upcoming modules, we’ll learn more about the Lua and Wasm filters.