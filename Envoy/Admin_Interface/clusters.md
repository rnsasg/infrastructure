# Clusters
The clusters endpoint (/clusters) will show the list of configured clusters and includes the following information:

* Per-host statistics
* Per-host health status
* Circuit breaker settings
* Per-host weight and locality information

Host in this context refers to every discovered host that’s part of the upstream clusters.

The snippet below shows what the information looks like (note that the output is trimmed):

```json
{
 "cluster_statuses": [
  {
   "name": "api_google_com",
   "host_statuses": [
    {
     "address": {
      "socket_address": {
       "address": "10.0.0.1",
       "port_value": 8080
      }
     },
     "stats": [
      {
       "value": "23",
       "name": "cx_total"
      },
      {
       "name": "rq_error"
      },
      {
       "value": "51",
       "name": "rq_success"
      },
      ...
     ],
     "health_status": {
      "eds_health_status": "HEALTHY"
     },
     "weight": 1,
     "locality": {}
    }
   ],
   "circuit_breakers": {
    "thresholds": [
     {
      "max_connections": 1024,
      "max_pending_requests": 1024,
      "max_requests": 1024,
      "max_retries": 3
     },
     {
      "priority": "HIGH",
      "max_connections": 1024,
      "max_pending_requests": 1024,
      "max_requests": 1024,
      "max_retries": 3
     }
    ]
   },
   "observability_name": "api_google_com"
  },
  ...
```

> To get the JSON output, we can append the ?format=json when making the request or opening the URL in the browser.

## Host statistics
The output includes the statistics for each host, as explained in the table below:

| Metric Name       | Description                            |
|-------------------|----------------------------------------|
| cx_total          | Total connections                      |
| cx_active         | Total active connections               |
| cx_connect_fail   | Total connection failures              |
| rq_total          | Total requests                         |
| rq_timeout        | Total timed out requests               |
| rq_success        | Total requests with non-5xx responses  |
| rq_error          | Total requests with 5xx responses      |
| rq_active         | Total active requests                  |


## Host health status
The host health status gets reported under the health_status field. The values in the health status depend on whether the health checking is enabled. Assuming active and passive (circuit breaker) health checking is enabled, the table shows the boolean fields that might be included in the health_status field.

Note that the fields from the table are reported only if set to true. For example, if the host is healthy, then the health status will look like this:

| Field Name                          | Description                                                                                   |
|-------------------------------------|-----------------------------------------------------------------------------------------------|
| failed_active_health_check           | True, if the host is currently failing active health checks.                                   |
| failed_outlier_check                 | True, if the host is presently considered an outlier and has been ejected.                     |
| failed_active_degraded_check         | True, if the host is presently being marked as degraded through active health checking.        |
| pending_dynamic_removal              | True, if the host has been removed from service discovery but is being stabilized.              |
| pending_active_hc                    | True, if the host has not yet been health checked.                                             |
| excluded_via_immediate_hc_fail       | True, if the host should be excluded from panic, spillover, etc. calculations.                  |
| active_hc_timeout                    | True, if the host failed active health check due to timeout.                                    |
| eds_health_status                    | By default, set to healthy (if not using EDS). Can be set to unhealthy or degraded.             |



```json
"health_status": {
    "eds_health_status": "HEALTHY"
}
```

If an active health check is configured and the host is failing, then the status will look like this:

```json
"health_status": {
    "failed_active_health_check": true,
    "eds_health_status": "HEALTHY"
}
```