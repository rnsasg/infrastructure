# Request ID Generation
Unique request IDs are crucial for tracing requests through multiple services, visualizing request flows, and pinpointing sources of latency.

We can configure how the request ID gets generated through the request_id_extension field. If we don’t provide any configuration, Envoy uses the default extension called UuidRequestIdConfig.

The default extension generates a unique identifier (UUID4) and populates the x-request-id HTTP header. Envoy uses the 14th nibble of the UUID to determine what happens with the trace.

If the 14th nibble is set to 9, the tracing should be sampled. If it’s set to a it should be forced traced due to server-side override (a), or if set to b it should be force traced due to client-side request ID joining.

The 14th nibble is chosen because it’s fixed to 4 by design. Therefore, 4 indicates a default UUID and no trace status, for example 7b674932-635d-**4**ceb-b907-12674f8c7267.

The two configuration options we have in the UuidRequestIdconfig are the pack_trace_reason and use_request_id_for_trace_sampling.

```yaml
...
..
  route_config:
    name: local_route
  request_id_extension:
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.request_id.uuid.v3.UuidRequestIdConfig
      pack_trace_reason: false
      use_request_id_for_trace_sampling: false
  http_filters:
  - name: envoy.filters.http.router
...
```

The pack trace reason is a boolean value that controls whether the implementation alters the UUID to contain the trace sampling decision, as mentioned above. The default value is true. The use_request_id_for_trace_sampling sets whether to use x-request-id for sampling or not. The default value is true as well.