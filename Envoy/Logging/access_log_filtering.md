# Access log filtering
Another feature of access logging in Envoy is the ability to specify filters that determine whether the access log needs to be written or not. For example, we could have an access log filter that logs only the 500 status codes, only log requests that took more than 5 seconds, and so on. The table below shows supported access log filters.

| Access Log Filter Name       | Description                                                                                      |
|------------------------------|--------------------------------------------------------------------------------------------------|
| status_code_filter           | Filters on status code value.                                                                    |
| duration_filter              | Filters on total request duration in milliseconds.                                               |
| not_health_check_filter      | Filters for requests that are not health check requests.                                         |
| traceable_filter             | Filters for requests that are traceable.                                                         |
| runtime_filter               | Filters for random sampling of requests.                                                         |
| and_filter                   | Performs a logical “and” operation on the result of each filter in the list of filters. Filters are evaluated sequentially. |
| or_filter                    | Performs a logical “or” operation on the result of each filter in the list of filters. Filters are evaluated sequentially.  |
| header_filter                | Filters requests based on the presence or value of a request header.                             |
| response_flag_filter         | Filters requests that received responses with an Envoy response flag set.                        |
| grpc_status_filter           | Filters gRPC requests based on their response status.                                            |
| extension_filter             | Use an extension filter that’s statically registered at runtime.                                 |
| metadata_filter              | Filters based on matching dynamic metadata.                                                      |


Each filter has different properties that we have an option to set. Here’s a snippet that shows how to use the status code, header, and an and filter:

```yaml
...
access_log:
- name: envoy.access_loggers.stdout
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
  filter:
    and_filter:
      filters:
        header_filter:
          header:
            name: ":method"
            string_match:
              exact: "GET"
        status_code_filter:
          comparison:
            op: GE
            value:
              default_value: 400
...
```

The above snippet writes a log entry to the standard out for all GET requests with response codes greater than or equal 400.
