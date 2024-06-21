# Configuring access loggers
We can configure access loggers on the HTTP or TCP filter level and the listener level. We can also configure multiple access logs with different logging formats and logging sinks. A logging sink is an abstract term for the location the logs write to, for example, to the console (stdout, stderr), a file, or a network service.

A scenario in which we’d configure multiple access logs is when we’d like to see high-level information in the console (standard out) and full request details written to a file on the disk. The field used to configure the access loggers is called access_log.

Let’s look at an example of enabling access logging to standard out (StdoutAccessLog) on the HTTP connection manager (HCM) level:

```yaml
- filters:
  - name: envoy.filters.network.http_connection_manager
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      stat_prefix: ingress_http
      access_log:
      - name: envoy.access_loggers.stdout
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
```

Envoy aims to have portable and extensible configuration: typed config. A side effect of this is verbose names for configuration. For example, to enable access logging, we find the name of the HTTP configuration type and then the type corresponding to the console (StdoutAccessLog).

The StdoutAccessLog configuration writes the log entries to the standard out (the console). Other supported access logging sinks are the following:

File (FileAccessLog)
gRPC (HttpGrpcAccessLogConfig and TcpGrpcAccessLogConfig)
Standard error (StderrAccessLog)
Wasm (WasmAccessLog)
Open Telemetry
The file access log allows us to write log entries to a file we specify in the configuration. For example:

```yaml
- filters:
  - name: envoy.filters.network.http_connection_manager
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      stat_prefix: ingress_http
      access_log:
      - name: envoy.access_loggers.file
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
          path: ./envoy-access-logs.log
```

Note the change in the name (envoy.access_loggers.file) and the type (file.v3.FileAccessLog). Additionally, we’ve provided the path where we want Envoy to store the access logs.

The gRPC access logging sink sends the logs to an HTTP or TCP gRPC logging service. To use the gRPC logging sink, we have to build a gRPC server with an endpoint that implements the [MetricsService](https://github.com/envoyproxy/envoy/blob/main/api/envoy/service/metrics/v3/metrics_service.proto), specifically the StreamMetrics function. Then, Envoy can connect to the gRPC server and send the logs to it.

Earlier, we mentioned the default access log format that’s comprised of different command operators:

```shell
[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%"
%RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION%
%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%"
"%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
```

The format of the log entries is configurable and can be modified using the log_format field. Using the log_format, we can configure which values the log entry includes and specify whether we want logs in plain text or JSON format.

Let’s say we want to log only the start time, response code, and user agent. We’d configure it like this:

```yaml
- filters:
  - name: envoy.filters.network.http_connection_manager
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      stat_prefix: ingress_http
      access_log:
      - name: envoy.access_loggers.stdout
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
          log_format:
            text_format_source:
              inline_string: "%START_TIME% %RESPONSE_CODE% %REQ(USER-AGENT)%"
```

A sample log entry using the above format would look like this:

2021-11-01T21:32:27.170Z 404 curl/7.64.0
Similarly, instead of providing a text format, we can also set up the JSON format string if we want the logs to be in a structured format such as JSON.

To use the JSON format, we have to provide a format dictionary instead of a single string, as in plain text format.

Here’s an example of using the same log format but writing the log entries in JSON instead:

```yaml
- filters:
  - name: envoy.filters.network.http_connection_manager
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      stat_prefix: ingress_http
      access_log:
      - name: envoy.access_loggers.stdout
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
          log_format:
            json_format:
              start_time: "%START_TIME%"
              response_code: "%RESPONSE_CODE%"
              user_agent: "%REQ(USER-AGENT)%"
```

The above snippet would generate the following log entry:

{"user_agent":"curl/7.64.0","response_code":404,"start_time":"2021-11-01T21:37:59.979Z"}
Certain command operators, such as FILTER_STATE or DYNAMIC_METADATA, might produce nested JSON log entries.

The log format can also use formatter plugins specified through the formatters field. There are two known formatter extensions in the current version: the metadata (envoy.formatter.metadata) and request without query (envoy.formatter.req_without_query) extension.

The metadata formatter extension implements the METADATA command operator that allows us to output different types of metadata (DYNAMIC, CLUSTER, or ROUTE).

Similarly, the req_without_query formatter allows us to use the REQ_WITHOUT_QUERY command operator, which works the same way as the REQ command operator but removes the query string. The command operator is used to avoid logging any sensitive information into the access log.

Here’s an example of how to provide a formatter and how to use it in the inline_string:

```yaml
- filters:
  - name: envoy.filters.network.http_connection_manager
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      stat_prefix: ingress_http
      access_log:
      - name: envoy.access_loggers.stdout
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
          log_format:
            text_format_source:
              inline_string: "[%START_TIME%] %REQ(:METHOD)% %REQ_WITHOUT_QUERY(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%"
            formatters:
            - name: envoy.formatter.req_without_query
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.formatter.req_without_query.v3.ReqWithoutQuery
```
The above configuration with this request curl localhost:10000/?hello=1234 would generate a log entry that doesn’t include the query parameters (hello=1234):

[2021-11-01T21:48:55.941Z] GET / HTTP/1.1