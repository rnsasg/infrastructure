# Envoy component logs
So far, we’ve talked about logs generated as a result of sending requests to Envoy. However, Envoy also produces logs as part of the startup and during execution.

We can see the Envoy component logs each time we run Envoy:

```
...
[2021-11-03 17:22:43.361][1678][info][main] [source/server/server.cc:368] initializing epoch 0 (base id=0, hot restart version=11.104)
[2021-11-03 17:22:43.361][1678][info][main] [source/server/server.cc:370] statically linked extensions:
[2021-11-03 17:22:43.361][1678][info][main] [source/server/server.cc:372]   envoy.filters.network: envoy.client_ssl_auth, envoy.echo, envoy.ext_authz, envoy.filters.network.client_ssl_auth
...
```

The default format string for a component log is [%Y-%m-%d %T.%e][%t][%l][%n] [%g:%#] %v. The first part of the format string represents the date and time, followed by thread ID (%t), the log level of the message (%l), logger name (%n), the relative path of the source file and line number (%g:%#), and the actual log message (%v).

When starting Envoy, we can use the --log-format command-line option to customize the format. For example, if we wanted to log the time logger name, source function name, and the log message, then we could write the format string like so: [%T.%e][%n][%!] %v.

Then, when starting Envoy, we can set the format string as follows:

func-e run -c someconfig.yaml --log-format '[%T.%e][%n][%!] %v'
If we use the format string, the log entries look like this:

```
[17:43:15.963][main][initialize]   response trailer map: 160 bytes: grpc-message,grpc-status
[17:43:15.965][main][createRuntime] runtime: {}
[17:43:15.965][main][initialize] No admin address given, so no admin HTTP server started.
[17:43:15.966][config][initializeTracers] loading tracing configuration
[17:43:15.966][config][initialize] loading 0 static secret(s)
[17:43:15.966][config][initialize] loading 0 cluster(s)
[17:43:15.966][config][initialize] loading 1 listener(s)
[17:43:15.969][config][initializeStatsConfig] loading stats configuration
[17:43:15.969][runtime][onRtdsReady] RTDS has finished initialization
[17:43:15.969][upstream][maybeFinishInitialize] cm init: all clusters initialized
[17:43:15.969][main][onRuntimeReady] there is no configured limit to the number of allowed active connections. Set a limit via the runtime key overload.global_downstream_max_connections
[17:43:15.970][main][operator()] all clusters initialized. initializing init manager
[17:43:15.970][config][startWorkers] all dependencies initialized. starting workers
[17:43:15.971][main][run] starting main dispatch loop
```

Envoy features multiple loggers, and for each logger (e.g. main, config, http, …), we can control the logging level (info, debug, trace). We can look at the names of all active loggers if we enable the Envoy admin interface and send a request to /logging path. Another way to look at all available loggers is via the [source code](https://github.com/envoyproxy/envoy/blob/82261f5a401418df13626ca3fa52fa65fea10c81//source/common/common/logger.h).

Here’s how the default output from /logging endpoint looks:

```yaml
active loggers:
  admin: info
  alternate_protocols_cache: info
  aws: info
  assert: info
  backtrace: info
  cache_filter: info
  client: info
  config: info
  connection: info
  conn_handler: info
  decompression: info
  dns: info
  dubbo: info
  envoy_bug: info
  ext_authz: info
  rocketmq: info
  file: info
  filter: info
  forward_proxy: info
  grpc: info
  hc: info
  health_checker: info
  http: info
  http2: info
  hystrix: info
  init: info
  io: info
  jwt: info
  kafka: info
  key_value_store: info
  lua: info
  main: info
  matcher: info
  misc: info
  mongo: info
  quic: info
  quic_stream: info
  pool: info
  rbac: info
  redis: info
  router: info
  runtime: info
  stats: info
  secret: info
  tap: info
  testing: info
  thrift: info
  tracing: info
  upstream: info
  udp: info
  wasm: info
```

Notice that the default logging level for every logger is set to info. The other log levels are the following:

```
trace
debug
info
warning/warn
error
critical
off
```

To configure log levels, we can use the --log-level option or the --component-log-level to control the log level for each component separately. The component log levels can be written in the format of log_name:log_level. If we’re setting log levels for multiple components, then we separate them with a comma. For example: upstream:critical,secret:error,router:trace.

For example, to set the main log level to trace, config log level to error, and turn off all other loggers, we could type the following:

> func-e run -c someconfig.yaml --log-level off --component-log-level main:trace,config:error

By default, all Envoy application logs are written to the standard error (stderr). To change that, we can provide an output file using the --log-path option:

> func-e run -c someconfig.yaml --log-path app-logs.log

In one of the labs, we’ll also show how Envoy can be configured to write application logs to the Google Cloud operations suite (formerly known as Stackdriver).