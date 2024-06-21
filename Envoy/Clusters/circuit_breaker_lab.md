# Circuit Breaker
In this lab, we’ll demonstrate how to use circuit breakers. We’ll run a Python HTTP server and an Envoy proxy in front of it. To start the Python server listening on port 8000, run:

> python3 -m http.server 8000
Next, we’ll create the Envoy configuration with the following circuit breaker:

```yaml
...
    circuit_breakers:
      thresholds:
        max_connections: 20
        max_requests: 100
        max_pending_requests: 20
```
So, if we exceed the 20 connections or 100 requests or 20 pending requests, the circuit breaker will trip. Here’s the full Envoy configuration:

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
          stat_prefix: listener_http
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: vh
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: python_server
  clusters:
  - name: python_server
    connect_timeout: 5s
    circuit_breakers:
      thresholds:
        max_connections: 20
        max_requests: 100
        max_pending_requests: 20
    load_assignment:
      cluster_name: python_server
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8000
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
```

Save the above configuration to 3-lab-1-circuit-breaker.yaml and run the Envoy proxy:

> func-e run -c 3-lab-1-circuit-breaker.yaml
To send multiple concurrent requests to the proxy, we’ll use a tool called hey. By default, hey runs 50 concurrent workers and sends 200 requests, so we don’t even need to pass in any parameters when requesting http://localhost:1000:

```shell
hey http://localhost:10000
...
Status code distribution:
  [200] 107 responses
  [503] 93 responses
```

hey will output numerous stats, but the one we’re interested in is the status code distribution. It shows that we received HTTP 200 for 107 responses and HTTP 503 for 93 responses – this is where the circuit breakers tripped.

We can use the Envoy admin interface (running on port 9901) to look at the detailed metrics, for example:

```shell
...
envoy_cluster_upstream_cx_overflow{envoy_cluster_name="python_server"} 134
envoy_cluster_upstream_rq_pending_overflow{envoy_cluster_name="python_server"} 93
```
## 3lab1circuitbreaker-221021-123757.yaml

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
          stat_prefix: listener_http
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: vh
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: python_server
  clusters:
  - name: python_server
    connect_timeout: 5s
    circuit_breakers:
      thresholds:
        max_connections: 20
        max_requests: 100
        max_pending_requests: 20
    load_assignment:
      cluster_name: python_server
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8000
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
```

