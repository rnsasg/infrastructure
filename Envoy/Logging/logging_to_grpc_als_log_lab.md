# gRPC Lab

This lab covers how to configure Envoy to use a separate gRPC access log service (ALS). We’ll use a basic gRPC server that implements the StreamAccessLogs function and outputs the received logs from Envoy to standard out.

Let’s start by running the ALS server as a Docker container:


```shell
docker run -dit -p 5000:5000 gcr.io/tetratelabs/envoy-als:0.1.0
```

The ALS server listens on port 5000 by default, so if we look at the logs from the Docker container, they should look similar to the following:

```shell
$ docker logs [container-id]
Creating new ALS server
2021/11/05 20:24:03 Listening on :5000
```

The output is telling us that the ALS is listening on port 5000.

There are two requirements that need to be configured in the Envoy configuration. The first is the access log using the access_log field and a logger type called HttpGrpcAccessLogConfig. Secondly, within the access log configuration, we have to refer to the gRPC server. One way to do that is to define an Envoy cluster.

Here’s the snippet that configures the logger and points to an Envoy cluster called grpc_als_cluster:

```yaml
...
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
...
```

The next snippet is the cluster configuration, which at this point we should already be familiar with. In our case, we’re running the gRPC server on the same machine on port 5000.

```yaml
...
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
...
```

Let’s put both pieces together and come up with a sample Envoy configuration that uses a gRPC access log service:

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
Save the above YAML to 6-lab-2-grpc-als.yaml and start the Envoy proxy using func-e:

> func-e run -c 6-lab-2-grpc-als.yaml &
We’re running the Docker container and Envoy in the background, so we can now use curl to send a couple of requests to the Envoy proxy:

```shell
$ curl localhost:10000
200
```

The proxy responds with 200, as that’s what we’ve defined in the configuration. You’ll notice there weren’t any logs outputted to the standard out, which is expected.

To see the logs, we’ll have to look at the logs from the Docker container. You can use docker ps to get the container ID and then run the logs command:

```shell
$ docker logs 96f
Creating new ALS server
2021/11/05 20:24:03 Listening on :5000
2021/11/05 20:33:52 Received value
2021/11/05 20:33:52 {"identifier":{"node":{"userAgentName":"envoy","userAgentBuildVersion":{"version":{"majorNumber":1,"minorNumber":20},"metadata":{"fields":{"build.type":{"stringValue":"RELEASE"},"revision.sha":{"stringValue":"96701cb24611b0f3aac1cc0dd8bf8589fbdf8e9e"},"revision.status":{"stringValue":"Clean"},"ssl.version":{"stringValue":"BoringSSL"}}}},"extensions":[{"name":"envoy.matching.common_inputs.environment_variable","category":"envoy.matching.common_inputs"},{"name":"envoy.access_loggers.file","category":"envoy.access_loggers"},{"name":"envoy.access_loggers.http_grpc","category":"envoy.access_loggers"},{"name":"envoy.access_loggers.open_telemetry","category":"envoy.access_loggers"},{"name":"envoy.access_loggers.stderr","category":"envoy.access_loggers"},{"name":"envoy.access_loggers.stdout","category":"envoy.access_loggers"},{"name":"envoy.acc...
...
```

We’ll then notice the log entries sent from the Envoy proxy to our gRPC server. The code in the gRPC server is straightforward and only converts the received values to a string and outputs them.

Here’s what the complete StreamAccessLogs function looks like:

```go
func (s *server) StreamAccessLogs(stream v3.AccessLogService_StreamAccessLogsServer) error {
  for {
    in, err := stream.Recv()
    log.Println("Received value")
    if err == io.EOF {
      return nil
    }
    if err != nil {
      return err
    }
    str, _ := s.marshaler.MarshalToString(in)
    log.Println(str)
  }
}
```

At this point, we can parse the specific values from the received stream and decide how to format them and where to send them.