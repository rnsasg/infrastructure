# Timeouts
Envoy supports numerous configurable timeouts that depend on the scenarios you use the proxy for.

We’ll look at the different configurable timeouts in the HCM section. Note that other filters and components have separate timeouts; we’ll not cover them here.

Some of the timeouts set at a higher-levels in configuration – for example, at the HCM level – can be overwritten at the lower levels, such as the HTTP route level.

Probably the most well-known timeout is the request timeout. The request timeout (request_timeout) specifies the amount of time the Envoy waits for the entire request to be received (e.g., 120s). The timer is activated when the request gets initiated. The timer is deactivated when the last request byte gets sent upstream or when the response gets initiated. By default, the timeout is disabled if not provided or set to 0.

A similar timeout called idle_timeout represents when a downstream or upstream connection gets terminated if there are no active streams. The default idle timeout is set to 1 hour. The idle timeout can be set in the common_http_protocol_options in the HCM configuration as shown below:

```yaml
...
filters:
- name: envoy.filters.network.http_connection_manager
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
    stat_prefix: ingress_http
    common_http_protocol_options:
      # Set idle timeout to 10 minutes
      idle_timeout: 600s
...
```

To configure the idle timeout for upstream connections, we can use the same field common_http_protocol_options, but in the clusters section.

There’s also a timeout that pertains to the headers called request_headers_timeout. This timeout specifies the amount of time Envoy wails for the request headers to be received (e.g. 5s). The timer gets activated upon receiving the first byte of the headers. The time is deactivated when the last byte of the headers is received. By default, the timeout is disabled if not provided or if set to 0.

Some other timeouts are available to set, such as the stream_idle_timeout, drain_timeout, and delayed_close_timeout.

If we move down the hierarchy, the next stop is the route timeout. As mentioned earlier, timeouts at the route level can overwrite the HCM timeouts and a couple of additional timeouts.

The route timeout is the time Envoy waits for the upstream to respond with a complete response. The timer starts once the entire downstream request has been received. The timeout defaults to 15 seconds; however, it isn’t compatible with responses that never end (i.e., streaming). In that case, the timeout needs to be disabled and stream_idle_timeout should be used instead.

We can use the idle_timeout field to overwrite the stream_idle_timeout on the HCM level.

We can also mention the per_try_timeout setting. This timeout is used in connection with retries and specifies a timeout for each try. Usually, the individual tries should use a shorter timeout than the value set by the timeout field.