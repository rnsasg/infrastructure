# Retries
We can define retry policies on the virtual host and the route level. A retry policy set on the virtual host level will apply for all routes in that virtual host. If there’s a retry policy defined on the route level, it will take precedence over the virtual host policy, and it gets treated separately – i.e., the route-level retry policy doesn’t inherit values from the virtual host level retry policy. Even though Envoy treats the retry policies independently, the configuration is the same.

In addition to setting the retry policy in the configuration, we can also configure it through request headers (i.e., x-envoy-retry-on header).

Within the Envoy configuration, we can configure the following:

1. Maximum number of retries
Envoy will retry requests up to a configured maximum. The exponential backoff algorithm is the default used to determine the intervals between retries. Another way to determine the interval between retries is through the headers (e.g. x-envoy-upstream-rq-per-try-timeout-ms). All retries are also contained within the overall request timeout, the request_timeout configuration setting. By default, Envoy sets the number of retries to one.

2. Retry conditions
We can retry the requests based on different conditions. For example, we can only retry 5xx response codes, gateway failures, 4xx response codes, etc.

3. Retry budgets
The retry budget specifies a limit on concurrent requests in relation to the number of active requests. This can help prevent retry traffic from contributing to the traffic volume.

4. Host selection retry plugins
The host selection during retries usually follows the same process as the original request. Using a retry plugin, we can change this behavior and specify a host or priority predicate that will reject a specific host and cause the host selection to be reattempted.

Let’s look at a couple of configuration examples on how to define a retry policy. We’re using httpbin and matching the /status/500 path that returns the 500 response code.

```yaml
  route_config:
    name: 5xx_route
    virtual_hosts:
    - name: httpbin
      domains: ["*"]
      routes:
      - match:
          path: /status/500
        route:
          cluster: httpbin
          retry_policy:
            retry_on: "5xx"
            num_retries: 5
```

Within the retry_policy field, we’re setting the retry condition (retry_on) to 500, which means that we want to retry only if the upstream returns an HTTP 500 (which it will). Envoy will retry the request five times. This is configured via the num_retries field.

If we run Envoy and send a request, the request will fail (HTTP 500), and the following log entry will be created:

```shell
[2021-07-26T18:43:29.515Z] "GET /status/500 HTTP/1.1" 500 URX 0 0 269 269 "-" "curl/7.64.0" "1ae9ffe2-21f2-43f7-ab80-79be4a95d6d4" "localhost:10000" "127.0.0.1:5000"
```

Notice the 500 URX section telling us that the upstream responded with 500, and the URX response flag means that Envoy rejected the request because the upstream retry limit was reached.

The retry condition can be set to one or more values, separated by a comma, specified in the table below.

| Retry condition           | Description                                                                                             |
|---------------------------|---------------------------------------------------------------------------------------------------------|
| 5xx                       | Retry on 5xx response code or if the upstream doesn’t respond (includes connect-failure and refused-stream). |
| gatewayerror              | Retry on 502, 503, or 504 response codes.                                                                |
| reset                     | Retry if upstream doesn’t respond at all.                                                                |
| connect-failure           | Retry if the request fails due to a connection failure to the upstream server (e.g., connect timeout).   |
| envoy-ratelimited         | Retry if `x-envoy-ratelimited` header is present.                                                        |
| retriable-4xx             | Retry if upstream responds with a retriable 4xx response code (only HTTP 409 at the moment).              |
| refused-stream            | Retry if upstream resets the stream with a `REFUSED_STREAM` error code.                                   |
| retriable-status-codes    | Retry if upstream responds with any response code matching one defined in the `x-envoy-retriable-status-codes` header (e.g., comma-delimited list of integers, for example "502,409"). |
| retriable-headers         | Retry if the upstream response includes any headers matching in the `x-envoy-retriable-header-names` header. |

In addition to controlling the responses on which Envoy retries the request, we can also configure the host selection logic for the retries. We can specify the retry_host_predicate that Envoy uses when selecting a host for retries.

We can keep track of previously attempted hosts (envoy.retry_host_predicates.previous_host) and reject them if they’ve already been attempted. Or, we can reject any hosts marked as canary hosts (e.g., any hosts marked with canary: true) using envoy.retry_host_predicates.canary_hosts predicate.

For example, here’s how to configure the previous_hosts plugin to reject any previously attempted hosts and retry the host selection a maximum of 5 times:

```yaml
  route_config:
    name: 5xx_route
    virtual_hosts:
    - name: httpbin
      domains: ["*"]
      routes:
      - match:
          path: /status/500
        route:
          cluster: httpbin
          retry_policy:
            retry_host_predicate:
            - name: envoy.retry_host_predicates.previous_hosts
            host_selection_retry_max_attempts: 5
```

With multiple endpoints in the cluster defined, we’d see the retries sent to different hosts with each retry.

## Request hedging

The idea behind request hedging is to send multiple requests simultaneously to different hosts and use the upstream results that respond first. Note that we usually configure this for idempotent requests where making the same call multiple times has the same effect.

We can configure the request hedging by specifying a hedge policy. Currently, Envoy performs hedging only in response to a request timeout. So when an initial request times out, a retry request is issued without canceling the original timed-out request. Envoy will return the first good response based on the retry policy to the downstream.

The hedging can be configured can be enabled by setting the hedge_on_per_try_timeout field to true. Just like the retry policy, it can be enabled on the virtual host or route level:

```yaml  route_config:
    name: 5xx_route
    virtual_hosts:
    - name: httpbin
      domains: ["*"]
      hedge_policy:
        hedge_on_per_try_timeout: true
      routes:
      - match:
      ...
```

