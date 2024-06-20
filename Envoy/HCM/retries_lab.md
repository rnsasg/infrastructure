# Lab 4: Retries
In this lab, we’ll learn how to configure different retry policies. We’ll use the httpbin Docker image because we can send requests to different paths (e.g. /status/[statuscode]), and the httpbin will respond with that status code.

Make sure you have the httpbin container listening on port 3030:

```shell
docker run -d -p 3030:80 kennethreitz/httpbin
```

Let’s create the Envoy configuration that defines a simple retry policy on 5xx responses. We are also enabling the admin interface, so we can see the retries being reported in the metrics:

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            virtual_hosts:
            - name: httpbin
              domains: ["*"]
              routes:
                - match:
                    prefix: "/"
                  route:
                    cluster: httpbin
                    retry_policy:
                      retry_on: "5xx"
                      num_retries: 5
  clusters:
  - name: httpbin
    connect_timeout: 5s
    load_assignment:
      cluster_name: single_cluster
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

Save the above YAML to 2-lab-4-retries-1.yaml and run the Envoy proxy: func-e run -c 2-lab-4-retries-1.yaml.

Let’s send a single request to /status/500 path:

```shell
$ curl localhost:10000/status/500
...
< HTTP/1.1 500 Internal Server Error
< server: envoy
...
< content-length: 0
< x-envoy-upstream-service-time: 276
As expected, we received a 500 response. Also, notice the x-envoy-upstream-service-time (time in milliseconds spent by the upstream host processing the request) is significantly bigger than if we’d send a /status/200 request:

$ curl localhost:10000/status/200
...
< HTTP/1.1 200 OK
< server: envoy
...
< content-length: 0
< x-envoy-upstream-service-time: 2
```

This is because Envoy performs the retries, which fail in the end. Similarly, if we open the stats page on the admin interface (http://localhost:9901/stats/prometheus), we’ll notice the metric that represents the number of retries (envoy_cluster_retry_upstream_rq) has a value of 5:

# TYPE envoy_cluster_retry_upstream_rq counter
envoy_cluster_retry_upstream_rq{envoy_response_code="500",envoy_cluster_name="httpbin"} 5

## 28296947-lab-4-retries.yaml

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            virtual_hosts:
            - name: httpbin
              domains: ["*"]
              routes:
                - match:
                    prefix: "/"
                  route:
                    cluster: httpbin
                    retry_policy:
                      retry_on: "5xx"
                      num_retries: 5
  clusters:
  - name: httpbin
    connect_timeout: 5s
    load_assignment:
      cluster_name: single_cluster
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
