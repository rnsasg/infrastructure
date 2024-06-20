## Rate limiting statistics
Envoy will emit metrics described in the table below, whether using global or local rate-limiting. We can set the stats prefix using the stat_prefix field when configuring the filters.

Each metric name is prefixed either with <stat_prefix>.http_local_rate_limit.<metric_name> when using a local rate limiter and cluster.<route_target_cluster>.ratelimit.<metric_name> when using a global rate limiter.

| Rate limiter         | Metric name            | Description                                                   |
|----------------------|------------------------|---------------------------------------------------------------|
| Local                | enabled                | Total number of requests for which the rate limiter was called |
| Local/Global         | ok                     | Total under-limit responses from the token bucket              |
| Local                | rate_limited           | Total responses without an available token                    |
| Local                | enforced               | Total number of rate-limited requests (e.g., HTTP 429 returned)|
| Global               | over_limit             | Total over-limit responses from the rate limit service         |
| Global               | error                  | Total errors contacting the rate limit service                 |
| Global               | failure_mode_allowed   | Total requests that were errors but allowed due to failure_mode_deny setting |


