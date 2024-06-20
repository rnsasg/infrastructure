# Lab 5: Local Rate Limiter
In this lab, we’ll learn how to configure a local rate limiter. We’ll use the httpbin container running on port 3030:

docker run -d -p 3030:80 kennethreitz/httpbin
Let’s create a rate limiter with five tokens. Every 30 seconds, the rate limiter refills the bucket with five tokens:

```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          route_config:
            name: local_route
            virtual_hosts:
            - name: instance_1
              domains: ["*"]
              routes:
              - match:
                  prefix: /status
                route:
                  cluster: instance_1
              - match:
                  prefix: /headers
                route:
                  cluster: instance_1
                typed_per_filter_config:
                  envoy.filters.http.local_ratelimit:
                    "@type": type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
                    stat_prefix: headers_route
                    token_bucket:
                      max_tokens: 5
                      tokens_per_fill: 5
                      fill_interval: 30s
                    filter_enabled:
                      default_value:
                        numerator: 100
                        denominator: HUNDRED
                    filter_enforced:
                      default_value:
                        numerator: 100
                        denominator: HUNDRED
                    response_headers_to_add:
                      - append: false
                        header:
                          key: x-rate-limited
                          value: OH_NO
          http_filters:
          - name: envoy.filters.http.local_ratelimit
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
              stat_prefix: httpbin_rate_limiter
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
  clusters:
  - name: instance_1
    connect_timeout: 0.25s
    type: STATIC
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 3030
admin:
  address:
    socket_address: 
      address: 127.0.0.1
      port_value: 9901
```

Save the above YAML to 2-lab-5-local-rate-limiter-1.yaml and run Envoy with func-e run -c 2-lab-5-local-rate-limiter-1.yaml.

The above configuration enables a local rate limiter for the route /headers. Additionally, we’ll add a header (x-rate-limited) to the response once the rate limit is reached.

If we make more than five requests to http://localhost:10000/headers within 30 seconds, we’ll get back an HTTP 429 response:

```
$ curl -v localhost:10000/headers
...
> GET /headers HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/7.64.0
> Accept: */*
>
< HTTP/1.1 429 Too Many Requests
< x-rate-limited: OH_NO
...
```


local_rate_limited
Also, notice the x-rate-limited header set by Envoy.

Once rate-limited, we’ll have to wait 30 seconds for the rate limiter to fill the bucket with tokens again. We can also try making requests to /status/200.You’ll notice we won’t get rate limited on that path.

If we open the stats page (localhost:9901/stats/prometheus), we’ll notice the rate-limiting metrics are recorded using the headers_route_rate_limiter stat prefix we configured:

# TYPE envoy_headers_route_http_local_rate_limit_enabled counter
envoy_headers_route_http_local_rate_limit_enabled{} 13

# TYPE envoy_headers_route_http_local_rate_limit_enforced counter
envoy_headers_route_http_local_rate_limit_enforced{} 8

# TYPE envoy_headers_route_http_local_rate_limit_ok counter
envoy_headers_route_http_local_rate_limit_ok{} 5

# TYPE envoy_headers_route_http_local_rate_limit_rate_limited counter
envoy_headers_route_http_local_rate_limit_rate_limited{} 8

## 2lab5localratelimiter1-221021-122936.yaml

```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          route_config:
            name: local_route
            virtual_hosts:
            - name: instance_1
              domains: ["*"]
              routes:
              - match:
                  prefix: /status
                route:
                  cluster: instance_1
              - match:
                  prefix: /headers
                route:
                  cluster: instance_1
                typed_per_filter_config:
                  envoy.filters.http.local_ratelimit:
                    "@type": type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
                    stat_prefix: headers_route
                    token_bucket:
                      max_tokens: 5
                      tokens_per_fill: 5
                      fill_interval: 30s
                    filter_enabled:
                      default_value:
                        numerator: 100
                        denominator: HUNDRED
                    filter_enforced:
                      default_value:
                        numerator: 100
                        denominator: HUNDRED
                    response_headers_to_add:
                      - append: false
                        header:
                          key: x-rate-limited
                          value: OH_NO
          http_filters:
          - name: envoy.filters.http.local_ratelimit
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
              stat_prefix: httpbin_rate_limiter
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
  clusters:
  - name: instance_1
    connect_timeout: 0.25s
    type: STATIC
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 3030
admin:
  address:
    socket_address: 
      address: 127.0.0.1
      port_value: 9901
```