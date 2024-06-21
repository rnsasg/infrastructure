# Lab 
In this lab, we’ll learn to use logging filters to log only certain requests based on some request properties.

Let’s come up with a couple of requirements for logging that we want to implement with filters:

log all requests with HTTP 404 status code to a log file called not_found.log
log all requests with env=debug header value to a log file called debug.log
log all POST requests to the standard output
Based on these requirements, we’ll have two access loggers logging to files (not_found.log and debug.log) and a single access logger that writes to stdout.

The access_log field is an array, and we can define multiple loggers underneath it. Inside the individual logger, we can use the filter field to specify when the strings should be written to the log.

We’ll use the status code filter for the first requirement, and for the second one, we’ll use the header filter. Here’s how the configuration looks:

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
          access_log:
          - name: envoy.access_loggers.file
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
              path: ./debug.log
            filter:
              header_filter:
                header:
                  name: env
                  string_match:
                    exact: debug
          - name: envoy.access_loggers.file
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
              path: ./not_found.log
            filter:
              status_code_filter:
                comparison:
                  value:
                    default_value: 404
                    runtime_key: ingress_http_status_code_filter
          - name: envoy.access_loggers.stdout
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
            filter:
              header_filter:
                header:
                  name: ":method"
                  string_match:
                    exact: POST
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: my_first_route
            virtual_hosts:
            - name: direct_response_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/404"
                direct_response:
                  status: 404
                  body:
                    inline_string: "404"
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
```yaml

Save the above YAML to 6-lab-1-logging-filters.yaml and run Envoy in the background:

func-e run -c 6-lab-1-logging-filters.yaml &
With Envoy running, if we send a request to http://localhost:10000, then we’ll notice that nothing is written to the standard out nor the log files.

Let’s try out the POST request next:

```shell
$ curl -X POST localhost:10000
[2021-11-03T21:52:36.398Z] "POST / HTTP/1.1" 200 - 0 3 0 - "-" "curl/7.64.0" "528335ae-8f0d-4d22-934a-02d4702a9c62" "localhost:10000" "-"
```

You’ll notice that the log entry is written to the standard out as defined in the configuration.

Next, let’s send a header env: debug with the following request:

curl -H "env: debug"  localhost:10000
Like the first example, nothing will be written to the standard out (it’s not a POST request). However, if we look in the debug.log file, then we’ll see the log entry:

```shell
$ cat debug.log
[2021-11-03T21:54:49.357Z] "GET / HTTP/1.1" 200 - 0 3 0 - "-" "curl/7.64.0" "ea2a11d6-6ccb-4f13-9686-4d30dbc3136e" "localhost:10000" "-"
```

Similarly, let’s send a request to /404 and look in the not_found.log file:

```shell
$ curl localhost:10000/404
404
```

```shell
$ cat not_found.log
[2021-11-03T21:55:37.891Z] "GET /404 HTTP/1.1" 404 - 0 3 0 - "-" "curl/7.64.0" "59bf1a1a-62b2-49e4-9226-7c49516ec390" "localhost:10000" "-"
```

In a case when multiple filter conditions are satisfied (e.g., we have a POST request and we’re sending the request to /404), the logs will be written to standard out and not_found.log in this case.

## 6lab1loggingfilters-221021-124811.yaml

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
          access_log:
          - name: envoy.access_loggers.file
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
              path: ./debug.log
            filter:
              header_filter:
                header:
                  name: env
                  string_match:
                    exact: debug
          - name: envoy.access_loggers.file
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
              path: ./not_found.log
            filter:
              status_code_filter:
                comparison:
                  value:
                    default_value: 404
                    runtime_key: ingress_http_status_code_filter
          - name: envoy.access_loggers.stdout
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
            filter:
              header_filter:
                header:
                  name: ":method"
                  string_match:
                    exact: POST
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: my_first_route
            virtual_hosts:
            - name: direct_response_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/404"
                direct_response:
                  status: 404
                  body:
                    inline_string: "404"
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
```

## /Users/kumarro/Downloads/6lab2grpcals-221021-124853.yaml


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
          access_log:
          - name: envoy.access_loggers.http_grpc
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.access_loggers.grpc.v3.HttpGrpcAccessLogConfig
              common_config:
                log_name: "mygrpclog"
                transport_api_version: V3
                grpc_service: 
                  envoy_grpc:
                    cluster_name: grpc_als_cluster
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: my_first_route
            virtual_hosts:
            - name: direct_response_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/404"
                direct_response:
                  status: 404
                  body:
                    inline_string: "404"
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
  clusters:
    - name: grpc_als_cluster
      connect_timeout: 5s
      type: STRICT_DNS
      http2_protocol_options: {}
      load_assignment:
        cluster_name: grpc_als_cluster
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: 127.0.0.1
                  port_value: 5000
```
