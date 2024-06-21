# Health checking

Envoy supports two types of health checks, active and passive. We can use both types of health checking at the same time. With active health checking, Envoy periodically sends a request to the endpoints to check its status. With passive health checking, Envoy monitors how the endpoints respond to connections. It enables Envoy to detect an unhealthy endpoint even before the active health check marks it as unhealthy. Passive health checking in Envoy is realized through outlier detection.

## Active health checking
Envoy supports different active health check methods on endpoints: HTTP, TCP, gRPC, and Redis health check. The health check method can be configured for each cluster separately. We can configure the health checks with the health_checks field in the cluster configuration.

Regardless of the selected health check method, a couple of common configuration settings need to be defined.

The timeout (timeout) represents the time allotted to wait for a health check response. If the response is not reached within the time value specified in this field, the health check attempt will be considered a failure. The interval specifies the time cadence between health checks. For example, an interval of 5 seconds will trigger a health check every 5 seconds.

The other two required settings can determine when a specific endpoint is considered healthy or unhealthy. The healthy_threshold specifies the number of “healthy” health checks (e.g., HTTP 200 response) required before an endpoint is marked healthy. The unhealthy_threshold does the same, but with “unhealthy” health checks – it specifies the number of unhealthy health checks required before an endpoint is marked unhealthy.

1. HTTP health checks
Envoy sends an HTTP request to the endpoint. If the endpoint responds with an HTTP 200, Envoy considers it healthy. The 200 response is the default response regarded as a healthy response. Using the expected_statuses field, we can customize that by providing a range of healthy HTTP statuses.

If the endpoint responds with an HTTP 503, the unhealthy_threshold is ignored, and the endpoint is considered unhealthy immediately.

```yaml
  clusters:
  - name: my_cluster_name
    health_checks:
      - timeout: 1s
        interval: 0.25s
        unhealthy_threshold: 5
        healthy_threshold: 2
        http_health_check:
          path: "/health"
          expected_statuses:
            - start: 200
              end: 299
      ...
```

For example, the above snippet defines an HTTP health check where Envoy will send an HTTP request to the /health path to the endpoints in the cluster. Envoy sends the request every 0.25s (interval) and waits for 1s (timeout) before timing out. To be considered healthy, the endpoint must respond with a status between 200 and 299 (expected_statuses) twice (healthy_threshold). The endpoint needs to respond with any other status code five times (unhealthy_threshold) to be considered unhealthy. Additionally, if the endpoint responds with HTTP 503, it’s immediately considered unhealthy (the unhealthy_threshold setting is ignored).

2. TCP health check
We specify a Hex-encoded payload (e.g. 68656C6C6F) that gets sent to the endpoint. If we set an empty payload, Envoy will do a connect-only health check where it only attempts to connect to the endpoint and considers it a success if the connection succeeds.

In addition to the payload that gets sent, we also need to specify the responses. Envoy will perform a fuzzy match on the response, and if the response matches the request, the endpoint is considered healthy.

```yaml
  clusters:
  - name: my_cluster_name
    health_checks:
      - timeout: 1s
        interval: 0.25s
        unhealthy_threshold: 1
        healthy_threshold: 1
        tcp_health_check:
          send:
            text: "68656C6C6F"
          receive:
            - text: "68656C6C6F"
      ...
```

3. gRPC health check
This health check follows the grpc.health.v1.Health health checking protocol. Check the GRPC health checking protocol document for more information on how it works.

The two optional configuration values we can set are the service_name and the authority. The service name is the value set to the service field in the HealthCheckRequest from grpc.health.v1.Health. The authority is the value of the :authority header. If it’s empty, Envoy uses the name of the cluster.

```yaml
  clusters:
  - name: my_cluster_name
    health_checks:
      - timeout: 1s
        interval: 0.25s
        unhealthy_threshold: 1
        healthy_threshold: 1
        grpc_health_check: {}
      ...
```

4. Redis health check
The Redis health check sends a Redis PING command to the endpoint and expects a PONG response. If the upstream Redis endpoint responds with anything other than PONG, it immediately causes the health check to fail. We can also specify a key, and Envoy will perform an EXIST <key> command instead of the PING command. The endpoint is healthy if the return value from Redis is 0 (i.e., the key doesn’t exist). Any other response is considered a failure.

```yaml
  clusters:
  - name: my_cluster_name
    health_checks:
      - timeout: 1s
        interval: 0.25s
        unhealthy_threshold: 1
        healthy_threshold: 1
        redis_health_check:
          key: "maintenance"
      ...
```

The above example checks for the key “maintenance” (e.g. EXIST maintenance), and if the key doesn’t exist, the health check passes.

## HTTP health-checking filter
The HTTP health-checking filter can be used to limit the amount of health-checking traffic that gets generated. The filter can run in different modes of operation that control how and whether the traffic is passed to the local service or not (i.e., no pass-through or pass-through).

1. Non-pass-through mode
When running in the non-pass-through mode, the health check request is never sent to the local service. Envoy responds either with HTTP 200 or HTTP 503, depending on the current draining state of the server.

A variation of the non-pass-through mode is where the HTTP 200 is returned if at least a specified percentage of endpoints are available in the upstream cluster. The percentage of endpoints can be configured with the cluster_min_healthy_percentages field:
```yaml
...
  pass_through_mode: false
  cluster_min_healthy_percentages:
    value: 15
...
```

2. Pass-through mode
Envoy passes every health check request to the local service in the pass-through mode. The service can either respond with an HTTP 200 or HTTP 503.

An additional setting for the pass-through mode is to use caching. Envoy passes the health check request to the service and caches the result for a period of time (cache_time). Any subsequent health check requests will use the cached value. Once the cache is invalidated, the next health check request is passed to the service.

```yaml
...
  pass_through_mode: true
  cache_time: 5m
...
```

The above snippet enables the pass-through mode with a cache that expires in 5 minutes.