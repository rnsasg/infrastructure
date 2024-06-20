# Traffic Splitting
We’ll learn how to configure traffic splitting with Envoy using runtime fractions and weighted clusters in this lab.


## Using runtime fractions
Runtime fractions are an excellent approach for traffic splitting when we have only two upstream clusters. Runtime fractions work by providing a runtime fraction (e.g., a numerator and a denominator) that represents the fraction of the traffic we want to route to a specific cluster. Then, we provide a second match using the same conditions (i.e., the same prefix in our example) but a different upstream cluster.

Let’s create an Envoy configuration that returns a direct response with status 201 for 70% of the traffic. The remainder of the traffic returns a status 202.

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
            - name: hello_vhost
              domains: ["*"]
              routes:
                - match:
                    prefix: "/"
                    runtime_fraction:
                      default_value:
                        numerator: 70
                        denominator: HUNDRED
                      runtime_key: routing.hello_io
                  direct_response:
                    status: 201
                    body:
                      inline_string: "v1"
                - match:
                    prefix: "/"
                  direct_response:
                    status: 202
                    body:
                      inline_string: "v2"
```

Save the above Envoy configuration to 2-lab-2-traffic-splitting-1.yaml and run Envoy with this configuration:

func-e run -c 2-lab-2-traffic-splitting-1.yaml
With Envoy running, we can use the hey tool to send 200 requests to the proxy:

```shell
$ hey http://localhost:10000
...
Status code distribution:
  [201] 142 responses
  [202] 58 responses
```

Looking at the status code distribution, we’ll notice that we received HTTP 201 responses ~71% of the time, and the remainder were HTTP 202 responses.

Using weighted clusters
When we have more than two upstream clusters to which we want to split the traffic, we can use the weighted clusters approach. Here, we individually assign weights to each upstream cluster. We used multiple matches with the previous approach, whereas we’ll use a single route and multiple weighted clusters in the weighted cluster.

For this approach, we’ll have to define the actual upstream clusters. We’ll run three instances of the httpbin image.

Let’s run three different instances on ports 3030, 4040, and 5050; we’ll refer to them in the Envoy configuration as instance_1, instance_2, and instance_3:

docker run -d -p 3030:80 kennethreitz/httpbin
docker run -d -p 4040:80 kennethreitz/httpbin
docker run -d -p 5050:80 kennethreitz/httpbin
An upstream cluster can be defined with the following snippet:

  clusters:
  - name: instance_1
    connect_timeout: 5s
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 3030
Let’s create the Envoy configuration that splits 50% of the traffic to instance_1, 30% to instance_2, and 20% to instance_3.

We’ll also enable the admin interface to retrieve the metrics that will show us the number of requests made to different clusters:

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
                  weighted_clusters:
                    clusters:
                      - name: instance_1
                        weight: 50
                      - name: instance_2
                        weight: 30
                      - name: instance_3
                        weight: 20
  clusters:
  - name: instance_1
    connect_timeout: 5s
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 3030
  - name: instance_2
    connect_timeout: 5s
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 4040
  - name: instance_3
    connect_timeout: 5s
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 5050
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
```

Save the above YAML to 2-lab-2-traffic-splitting-2.yaml and run the proxy: func-e run -c 2-lab-2-traffic-splitting-2.yaml.

Once the proxy is running, we’ll use hey to send 200 requests:

hey http://localhost:10000
The response from hey won’t help us determine the split because each upstream cluster responds with an HTTP 200.

To see the traffic split, open the stats on http://localhost:9901/stats/prometheus. From the list of metrics, look for the envoy_cluster_external_upstream_rq metric that counts the number of external upstream requests. We should see a split similar to this one:

# TYPE envoy_cluster_external_upstream_rq counter
envoy_cluster_external_upstream_rq{envoy_response_code="200",envoy_cluster_name="instance_1"} 99
envoy_cluster_external_upstream_rq{envoy_response_code="200",envoy_cluster_name="instance_2"} 63
envoy_cluster_external_upstream_rq{envoy_response_code="200",envoy_cluster_name="instance_3"} 38
If we calculate the percentages, we’ll notice that they correspond to the percentages we set in the configuration.


## 2lab2trafficsplitting2-221021-122723.yaml

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
                  weighted_clusters:
                    clusters:
                      - name: instance_1
                        weight: 50
                      - name: instance_2
                        weight: 30
                      - name: instance_3
                        weight: 20
  clusters:
  - name: instance_1
    connect_timeout: 5s
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 3030
  - name: instance_2
    connect_timeout: 5s
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 4040
  - name: instance_3
    connect_timeout: 5s
    load_assignment:
      cluster_name: instance_1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 5050
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
```

## 2lab2trafficsplitting1-221021-122723.yaml

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
            - name: hello_vhost
              domains: ["*"]
              routes:
                - match:
                    prefix: "/"
                    runtime_fraction:
                      default_value:
                        numerator: 70
                        denominator: HUNDRED
                      runtime_key: routing.hello_io
                  direct_response:
                    status: 201
                    body:
                      inline_string: "v1"
                - match:
                    prefix: "/"
                  direct_response:
                    status: 202
                    body:
                      inline_string: "v2"
```


