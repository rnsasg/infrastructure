# HTTP Inspector Listener Filter
The [HTTP inspector listener filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/http_inspector) (envoy.filters.listener.http_inspector) allows us to detect if the application protocol appears to be HTTP. If the protocol is not HTTP, the listener filter will pass the packet.

If the application protocol is determined to be HTTP, it also detects the corresponding HTTP protocol (e.g., HTTP/1.x or HTTP/2).

We can check the result of the HTTP inspection filter using the application_protocols field in the filter chain match.

Let’s consider the following snippet:

```yaml
...
    listener_filters:
    - name: envoy.filters.listener.http_inspector
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.filters.listener.http_inspector.v3.HttpInspector
    filter_chains:
    - filter_chain_match:
        application_protocols: ["h2"]
      filters:
      - name: my_http2_filter
        ... 
    - filter_chain_match:
        application_protocols: ["http/1.1"]
      filters:
      - name: my_http1_filter
...
```

We’ve added the http_inspector filter under the listener_filters field to inspect the connection and determine the application protocol. If the HTTP protocol is HTTP/2 (h2c), Envoy matches the first network filter chain (starting with my_http2_filter).

Alternatively, if the downstream HTTP protocol is HTTP/1.1 (http/1.1), Envoy matches the second filter chain and runs the filter chain starting with the filter called my_http1_filter.