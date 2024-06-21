# Statistics
The primary endpoint for the admin interface’s statistics output is accessed through the /stats endpoint. This input is typically used for debugging. We can access the endpoint by either sending a request to /stats endpoint or accessing the same path from the administrative UI.

The endpoint supports filtering the returned stats using the filter query parameter and a regular expression.

Another dimension to filter the output is using the usedonly query parameter. When used, it will output only the statistics that Envoy has updated. For example, counters that have been incremented at least once, gauges changed at least once, and histograms added to at least once.

By default, the stats are written in the StatsD format. Each stat is written to a separate line and the stat name (e.g., cluster_manager.active_clusters) is followed by the stat value (e.g., 15).

For example:

```shell
...
cluster_manager.active_clusters: 15
cluster_manager.cluster_added: 3
cluster_manager.cluster_modified: 4
...
```

The format query parameter controls the output format. Setting it to json will output the stats in JSON format. This format is typically used if we want to access and parse the stats programmatically.

The second format is the Prometheus format (e.g., format=prometheus). This option formats the status in the Prometheus format and can be used to integrate with a Prometheus server. Alternatively, we can use the /stats/prometheus endpoint to get the same output.

## Memory
The /memory endpoint will output the current memory allocation and heap usage in bytes. It’s a subset of information /stats endpoint prints out.

```shell
$ curl localhost:9901/memory
{
 "allocated": "5845672",
 "heap_size": "10485760",
 "pageheap_unmapped": "0",
 "pageheap_free": "3186688",
 "total_thread_cache": "80064",
 "total_physical_bytes": "12699350"
}
```

## Reset counters
Sending a POST request to /reset_counters resets all counters to zero. Note that this won’t reset or drop any data sent to statsd. It affects only the output of the /stats endpoint. The /stats endpoint and the /reset_counters endpoint can be used during debugging.

## Server information and status
The /server_info endpoint outputs the information about the running Envoy server. This includes the version, state, configuration path, log level information, uptime, node information, and more.

The [admin.v3.ServerInfo](https://www.envoyproxy.io/docs/envoy/latest/api-v3/admin/v3/server_info.proto#envoy-v3-api-msg-admin-v3-serverinfo) proto explains the different fields returned by the endpoint.

The /ready endpoint returns a string and an error code reflecting the state of the Envoy. If Envoy is live and ready to accept connections, then it returns the HTTP 200 and the string LIVE. Otherwise, the output will be an HTTP 503. This endpoint can be used as a readiness check.

The /runtime endpoint outputs all runtime values in a JSON format. The output includes the list of active runtime override layers and each key’s stack of layer values. The values can also be modified by sending a POST request to the /runtime_modify endpoint and specifying key/value pairs — for example, POST /runtime_modify?my_key_1=somevalue.

The /hot_restart_version endpoint, together with the --hot-restart-version flag, can be used to determine whether the new binary and the running binary are hot restart compatible.

**Hot restart** is Envoy’s ability to “hot” or “live” restart itself. This means that Envoy can fully reload itself (and configuration) without dropping any existing connections.

# Hystrix event stream
The /hystrix_event_stream endpoint is meant to be used as the stream source for the [Hystrix dashboard](https://github.com/Netflix-Skunkworks/hystrix-dashboard/wiki). Sending a request to the endpoint will trigger a stream of statistics from Envoy in the format that’s expected by the Hystrix dashboard.

Note that we have to configure the Hystrix stats sync in the bootstrap configuration for the endpoint to work.

For example:

```yaml
stats_sinks: 
  - name: envoy.stat_sinks.hystrix
    typed_config:
      "@type": type.googleapis.com/envoy.config.metrics.v3.HystrixSink
      num_buckets: 10
```

## Contention
The /contention endpoint dumps the current Envoy mutex content stats if mutex tracing is enabled.

## CPU and heap profilers
We can enable or disable the CPU/heap profiler using the /cpuprofiler and /heapprofiler endpoints. Note that this requires compiling Envoy with gperftools. The Envoy GitHub repository has [documentation](https://github.com/envoyproxy/envoy/blob/main/bazel/PPROF.md) on how to do this.