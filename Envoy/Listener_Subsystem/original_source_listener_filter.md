# Original Source Listener Filter
The [original source filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/original_src_filter)(envoy.filters.listener.original_src) replicates the downstream (host connecting to Envoy) remote address of the connection on the upstream (host receiving requests from Envoy) side of Envoy.

For example, if we connect to Envoy with 10.0.0.1, Envoy connects to the upstream with source IP 10.0.0.1. The address is determined from the proxy protocol filter (explained next), or it can come from trusted HTTP headers.

```yaml
- name: envoy.filters.listener.original_src
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.filters.listener.original_src.v3.OriginalSrc
    mark: 100
```

The filter also allows us to set the SO_MARK option on the upstream connection’s socket. The SO_MARK option is used for marking each packet sent through the socket and allows us to do mark-based routing (we can match the mark later on).

The snippet above sets the mark to 100. Using this mark, we can ensure that non-local addresses may be routed back through the Envoy proxy when binding to the original source address.