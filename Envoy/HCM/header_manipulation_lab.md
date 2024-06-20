## Lab 3: Header Manipulation
In this lab, we’ll learn how to manipulate request and response headers at different configuration levels.

We’ll use a single example that does the following:

* Adds a response header lab: 3 for all requests
* Adds a request header vh: one to the virtual host
* Adds a response header called json for the /json route match. The response header has the value from the request header called hello

We’ll only have a single upstream cluster called single_cluster, listening on port 3030. Let’s run the httpbin container listening on that port:

```shell
docker run -d -p 3030:80 kennethreitz/httpbin
```

Let’s create the Envoy configuration that follows the above rules:
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
            response_headers_to_add:
              - header:
                  key: "lab"
                  value: "3"
            virtual_hosts:
            - name: vh_1
              request_headers_to_add:
                - header: 
                    key: vh
                    value: "one"
              domains: ["*"]
              routes:
                - match:
                    prefix: "/json"
                  route:
                    cluster: single_cluster
                  response_headers_to_add:
                    - header: 
                        key: "json"
                        value: "%REQ(hello)%"
                - match:
                    prefix: "/"
                  route:
                    cluster: single_cluster
  clusters:
  - name: single_cluster
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
```

Save the above YAML to 2-lab-3-header-manipulation-1.yaml and run the Envoy proxy: func-e run -c 2-lab-3-header-manipulation-1.yaml.

Let’s start by making a simple request to /headers - this will match the / prefix:

```shell
$ curl -v localhost:10000/headers
...
< x-envoy-upstream-service-time: 2
< lab: 3
<
{
  "headers": {
    "Accept": "*/*",
    "Host": "localhost:10000",
    "User-Agent": "curl/7.64.0",
    "Vh": "one",
    "X-Envoy-Expected-Rq-Timeout-Ms": "15000"
  }
}
```

We’ll notice the response header lab: 3 was set. This comes from the route configuration and will be added to all requests we make.

In the response, the Vh: one header (note the capitalization is coming from the httpbin code) was added to the request. The response from httpbin shows the headers received (i.e., request headers).

Let’s try making a request to the /json path. The requests sent to httpbin on that path will return a sample JSON. Additionally, this time we’ll include a request header called hello: world:

```shell
$ curl -v -H "hello: world" localhost:10000/json
> GET /json HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/7.64.0
> Accept: */*
> hello: world
>
< HTTP/1.1 200 OK
< server: envoy
< json: world
< lab: 3
...
```

Notice this time we have the request header we set (hello: world), and on the response path, we see the json: world header that gets its value from the request header we set. Similarly, the lab: 3 response header is set. The request header vh: 1 is too, but we can’t see it this time because it’s not being outputted anywhere.

## 2lab3headermanipulation1-221021-122812.yaml

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
            response_headers_to_add:
              - header:
                  key: "lab"
                  value: "3"
            virtual_hosts:
            - name: vh_1
              request_headers_to_add:
                - header: 
                    key: vh
                    value: "one"
              domains: ["*"]
              routes:
                - match:
                    prefix: "/json"
                  route:
                    cluster: single_cluster
                  response_headers_to_add:
                    - header: 
                        key: "json"
                        value: "%REQ(hello)%"
                - match:
                    prefix: "/"
                  route:
                    cluster: single_cluster
  clusters:
  - name: single_cluster
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
```

