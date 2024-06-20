# Lab 6: Global Rate Limiter
In this lab, we’ll learn how to configure a global rate limiter. We’ll use the rate limiter service together with a Redis instance to track the tokens. We’ll use Docker Compose to run Redis and rate limiter service containers.

Let’s create the configuration for the rate limiter service first.

```yaml
domain: my_domain
descriptors:
- key: generic_key
  value: instance_1
  descriptors:
    - key: header_match
      value: get_request
      rate_limit:
        unit: MINUTE
        requests_per_unit: 5
```

We’re specifying a generic key (instance_1) and a single descriptor called header_match with a rate limit set to 5 requests/minute.

Save the above file to the /config/rl-config.yaml folder.

We can now run the Docker Compose file that will start the Redis and rate limiter service.

```yaml
version: "3"
services:
  redis:
    image: redis:alpine
    expose:
      - 6379
    ports:
      - 6379:6379
    networks:
      - ratelimit-network

  # Rate limit service configuration
  ratelimit:
    image:  envoyproxy/ratelimit:bd46f11b
    command: /bin/ratelimit
    ports:
      - 10001:8081
      - 6070:6070
    depends_on:
      - redis
    networks:
      - ratelimit-network
    volumes:
      - $PWD/config:/data/config/config
    environment:
      - USE_STATSD=false
      - LOG_LEVEL=debug
      - REDIS_SOCKET_TYPE=tcp
      - REDIS_URL=redis:6379
      - RUNTIME_ROOT=/data
      - RUNTIME_SUBDIRECTORY=config
```

networks:
  ratelimit-network:
Save the above file to rl-docker-compose.yaml and start all containers using:

$ docker-compose -f rl-docker-compose.yaml up
To make sure the rate limiter service correctly reads the configuration, we can check the output from the containers or use the debug port on the rate limiter service:

$ curl localhost:6070/rlconfig
my_domain.generic_key_instance_1.header_match_get_request: unit=MINUTE requests_per_unit=5
With the rate limiter and Redis up and running, we can start the httpbin container:

docker run -d -p 3030:80 kennethreitz/httpbin
Next, we’ll create the Envoy configuration that defines the rate limit actions. We’ll set the descriptor instance_1 and the get_request whenever a GET request gets sent to the httpbin.

Under the http_filters, we configure the ratelimit filter by specifying the domain name (my_domain) and pointing to the cluster Envoy can use to reach the rate limit service.

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
            - name: namespace.local_service
              domains: ["*"]
              routes:
              - match:
                  prefix: /
                route:
                  cluster: instance_1
                  rate_limits:
                  - actions:
                    - generic_key:
                        descriptor_value: instance_1
                    - header_value_match:
                        descriptor_value: get_request
                        headers:
                        - name: ":method"
                          exact_match: GET
          http_filters:
          - name: envoy.filters.http.ratelimit
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit
              domain: my_domain
              enable_x_ratelimit_headers: DRAFT_VERSION_03
              rate_limit_service:
                transport_api_version: V3
                grpc_service:
                    envoy_grpc:
                      cluster_name: rate-limit
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
  - name: rate-limit
    connect_timeout: 1s
    type: STATIC
    lb_policy: ROUND_ROBIN
    protocol_selection: USE_CONFIGURED_PROTOCOL
    http2_protocol_options: {}
    load_assignment:
      cluster_name: rate-limit
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 10001
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
```

Save the above YAML to 2-lab-6-global-rate-limiter-1.yaml and run the proxy using func-e run -c 2-lab-6-global-rate-limiter-1.yaml.

We can now send more than five requests, and we’ll get rate limited:

```
$ curl -v localhost:10000
...
< HTTP/1.1 429 Too Many Requests
< x-envoy-ratelimited: true
< x-ratelimit-limit: 5, 5;w=60
< x-ratelimit-remaining: 0
< x-ratelimit-reset: 25
...
```

We received HTTP 429 responses together with response headers that indicate that we were rate limited; how many requests can we make before getting rate limited (x-ratelimit-remaining) and when the rate limit resets (x-ratelimit-reset).


## rl-config.yaml

```yaml
domain: my_domain
descriptors:
- key: generic_key
  value: instance_1
  descriptors:
    - key: header_match
      value: get_request
      rate_limit:
        unit: MINUTE
        requests_per_unit: 5
```


## rl-docker-compose.yaml

```yaml
version: "3"
services:
  redis:
    image: redis:alpine
    expose:
      - 6379
    ports:
      - 6379:6379
    networks:
      - ratelimit-network

  # Rate limit service configuration
  ratelimit:
    image:  envoyproxy/ratelimit:bd46f11b
    command: /bin/ratelimit
    ports:
      - 10001:8081
      - 6070:6070
    depends_on:
      - redis
    networks:
      - ratelimit-network
    volumes:
      - $PWD/config:/data/config/config
    environment:
      - USE_STATSD=false
      - LOG_LEVEL=debug
      - REDIS_SOCKET_TYPE=tcp
      - REDIS_URL=redis:6379
      - RUNTIME_ROOT=/data
      - RUNTIME_SUBDIRECTORY=config

networks:
  ratelimit-network:
```

## 2-lab-6-global-rate-limiter-1.yaml

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
            - name: namespace.local_service
              domains: ["*"]
              routes:
              - match:
                  prefix: /
                route:
                  cluster: instance_1
                  rate_limits:
                  - actions:
                    - generic_key:
                        descriptor_value: instance_1
                    - header_value_match:
                        descriptor_value: get_request
                        headers:
                        - name: ":method"
                          exact_match: GET
          http_filters:
          - name: envoy.filters.http.ratelimit
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit
              domain: my_domain
              enable_x_ratelimit_headers: DRAFT_VERSION_03
              rate_limit_service:
                transport_api_version: V3
                grpc_service:
                    envoy_grpc:
                      cluster_name: rate-limit
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
  - name: rate-limit
    connect_timeout: 1s
    type: STATIC
    lb_policy: ROUND_ROBIN
    protocol_selection: USE_CONFIGURED_PROTOCOL
    http2_protocol_options: {}
    load_assignment:
      cluster_name: rate-limit
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 10001
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
```