# Circuit breakers
Circuit breaking is an important pattern that can help with service resiliency. The circuit breaker pattern prevents additional failures by controlling and managing access to the failing services. It allows us to fail quickly and apply the back pressure downstream as soon as possible.

Let’s look at a snippet that defines circuit breaking:

```yaml
...
  clusters:
  - name: my_cluster_name
  ...
    circuit_breakers:
      thresholds:
        - priority: DEFAULT
          max_connections: 1000
        - priority: HIGH
          max_requests: 2000
...
```

We can configure the circuit breaker thresholds for each route priority separately. For example, the higher-priority routes should have higher thresholds than the default priority. If any thresholds are exceeded, the circuit breaker trips, and the downstream host receives the HTTP 503 response.

There are multiple options we can configure the circuit breakers with:

1. Maximum connections (max_connections)
Specifies the maximum number of connections that Envoy will make to all endpoints in the cluster. If this number is exceeded, the circuit breaker trips and increments the upstream_cx_overflow metric for the cluster. The default value is 1024.

2. Maximum pending requests (max_pending_requests)
Specifies the maximum number of requests queued while waiting for a ready connection pool connection. When the threshold is exceeded, Envoy increments the stat upstream_rq_pending_overflow for the cluster. The default value is 1024.

3. Maximum requests (max_requests)
Specifies the maximum number of parallel requests that Envoy makes to all endpoints in the cluster. The default value is 1024.

4. Maximum retries (max_retries)
Specifies the maximum number of parallel retries that Envoy allows to all endpoints in the cluster. The default value is 3. If this circuit breaker overflows, the upstream_rq_retry_overflow counter is incremented.

Optionally, we can combine the circuit breakers with a retry budget (retry_budget). Specifying a retry budget, we can limit the concurrent retries to the number of active requests.