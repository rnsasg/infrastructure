# Header Manipulation

HCM supports manipulating request and response headers at the weighted cluster, route, virtual host, and/or global configuration level.

Note that we can’t modify all headers directly from the configuration. The exception is if we use a Wasm extension. Then, we could modify the `:authority` header, for example.

The immutable headers are the pseudo-headers (prefixed by `:`, such as `:scheme`) and the host header. Additionally, headers such as `:path` and `:authority` can be indirectly modified through `prefix_rewrite`, `regex_rewrite`, and `host_rewrite` configuration.

Envoy applies the headers to requests/responses in the following order:

1. Weighted cluster-level headers
2. Route-level headers
3. Virtual host-level headers
4. Global-level headers

The order means Envoy might overwrite a header set on the weighted cluster level by headers configured at the higher level (route, virtual host, or global).

At each level, we can set the following fields to add/remove request/response headers:

- `response_headers_to_add`: array of headers to add to the response
- `response_headers_to_remove`: array of headers to remove from the response
- `request_headers_to_add`: array of headers to add to the request
- `request_headers_to_remove`: array of headers to remove from the request

In addition to hardcoding the header values, we can also use variables to add dynamic values to the headers. The variable names get delimited by the percent symbol (`%`). The list of supported variable names includes `%DOWNSTREAM_REMOTE_ADDRESS%`, `%UPSTREAM_REMOTE_ADDRESS%`, `%START_TIME%`, `%RESPONSE_FLAGS%`, and many more. You can find the complete list of variables [here](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/header_manipulation_filter.html#header-manipulation-variables).

Let’s look at an example that shows how to add/remove headers from request/response at different levels:

```yaml
route_config:
  response_headers_to_add:
    - header: 
        key: "header_1"
        value: "some_value"
      append: false
  response_headers_to_remove: "header_we_dont_need"
  virtual_hosts:
  - name: hello_vhost
    request_headers_to_add:
      - header: 
          key: "v_host_header"
          value: "from_v_host"
    domains: ["hello.io"]
    routes:
      - match:
          prefix: "/"
        route:
          cluster: hello
        response_headers_to_add:
          - header: 
              key: "route_header"
              value: "%DOWNSTREAM_REMOTE_ADDRESS%"
      - match:
          prefix: "/api"
        route:
          cluster: hello_api
        response_headers_to_add:
          - header: 
              key: "api_route_header"
              value: "api-value"
          - header:
              key: "header_1"
              value: "this_will_be_overwritten"
```

## Standard Headers

Envoy manipulates a set of headers when the request is received (decoding) and when it sends the request to the upstream cluster (encoding).

When using bare-bone Envoy configuration to route the traffic to a single cluster, the following headers are set during the encoding:

- `:authority`, `localhost:10000`
- `:path`, `/`
- `:method`, `GET`
- `:scheme`, `http`
- `user-agent`, `curl/7.64.0`
- `accept`, `*/*`
- `x-forwarded-proto`, `http`
- `x-request-id`, `14f0ac76-128d-4954-ad76-823c3544197e`
- `x-envoy-expected-rq-timeout-ms`, `15000`

On encoding (response), a different set of headers is sent:

- `:status`, `200`
- `x-powered-by`, `Express`
- `content-type`, `text/html; charset=utf-8`
- `content-length`, `563`
- `etag`, `W/"233-b+4UpNDbOtHFiEpLMsDEDK7iTeI"`
- `date`, `Fri, 16 Jul 2021 21:59:52 GMT`
- `x-envoy-upstream-service-time`, `2`
- `server`, `envoy`

The table below explains the different headers set by Envoy either during decoding or encoding.

| Header                           | Description                                                                                                                                                               |
|----------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `:scheme`                        | Set and available to filters and forwarded upstream. (For HTTP/1, `:scheme` header is set either from the absolute URL or from the `x-forwarded-proto` header value)      |
| `user-agent`                     | Usually set by the client, but can be modified when `add_user_agent` is enabled (only if the header is not already set). The value gets determined by the `--service-cluster` command-line option. |
| `x-forwarded-proto`              | The standard header identifies the protocol that a client used to connect to the proxy. The value is either `http` or `https`.                                             |
| `x-request-id`                   | Used by Envoy to uniquely identify a request, also used in access logging and tracing.                                                                                      |
| `x-envoy-expected-rq-timeout-ms` | Specifies the time in milliseconds the router expects the request to be completed. This is read from `x-envoy-upstream-rq-timeout-ms` header value (assuming `respect_expected_rq_timeout` is set) or from the route timeout setting (default is 15 seconds). |
| `x-envoy-upstream-service-time`  | Time in milliseconds spent by the endpoint processing the request and the network latency between Envoy and the upstream host.                                             |
| `server`                         | Set to the value specified in `server_name` field (defaults to envoy).                                                                                                     |

A slew of other headers get set or consumed by Envoy, depending on the scenarios. We’ll call out different headers as we discuss these scenarios and features in the rest of the course.

## Header Sanitization

Header sanitization is a process that involves adding, removing, or modifying request headers for security reasons. There are some headers Envoy will potentially sanitize:

| Header                          | Description                                                                                                                                            |
|---------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `x-envoy-decorator-operation`   | Overrides any locally defined span name generated by the tracing mechanism.                                                                             |
| `x-envoy-downstream-service-cluster` | Contains the service cluster of the caller (removed for external requests). Determined by the `--service-cluster` command line option requires `user_agent` to be set to true. |
| `x-envoy-downstream-service-node`    | Like the previous header, the value is determined by the `--service-node` option.                                                                   |
| `x-envoy-expected-rq-timeout-ms`     | Specifies the time in milliseconds the router expects the request to be completed. This is read from `x-envoy-upstream-rq-timeout-ms` header value (assuming `respect_expected_rq_timeout` is set) or from the route timeout setting (default is 15 seconds). |
| `x-envoy-external-address`      | Trusted client address (see XFF below for details on how this is determined).                                                                           |
| `x-envoy-force-trace`           | Forces traces to be collected.                                                                                                                         |
| `x-envoy-internal`              | Set to true if the request is internal (see XFF below for details on how this is determined).                                                           |
| `x-envoy-ip-tags`               | Set by the HTTP IP tagging filter if the IP tags define the external address.                                                                           |
| `x-envoy-max-retries`           | A maximum number of retries if the retry policy is configured.                                                                                         |
| `x-envoy-retry-grpc-on`         | Retries failed requests for specific gRPC status codes.                                                                                                |
| `x-envoy-retry-on`              | Specifies the retry policy.                                                                                                                            |
| `x-envoy-upstream-alt-stat-name`| Emits upstream response code/timing stats to a dual stat tree.                                                                                         |
| `x-envoy-upstream-rq-per-try-timeout-ms` | Sets a per-try timeout on routed requests.                                                                                                     |
| `x-envoy-upstream-rq-timeout-alt-response` | If present sets a 204 response code (instead of 504) in case of a request timeout.                                                               |
| `x-envoy-upstream-rq-timeout-ms` | Overrides the route configuration timeout.                                                                                                           |
| `x-forwarded-client-cert`       | Indicates certificate information of part of all of the clients/proxies that a request has flowed through.                                              |
| `x-forwarded-for`               | Indicates the IP addresses the request went through. See XFF below for more details.                                                                   |
| `x-forwarded-proto`             | Sets the originating protocol (http or https).                                                                                                         |
| `x-request-id`                  | Used by Envoy to uniquely identify a request. Also used in access logging and tracing.                                                                  |

Whether to sanitize a specific header or not depends on where the request is coming from. Envoy determines whether the request is external or internal by looking at the `x-forwarded-for` header (XFF) and the `internal_address_config` setting.

## XFF

XFF or x-forwaded-for header indicates the IP addresses request went through on its way from the client to the server. Proxies between downstream and upstream services append the IP address of the nearest client to the XFF list before proxying the request.

Envoy doesn’t automatically append the IP address to XFF. Envoy only appends the address if the use_remote_address (default is false) is set to true, and skip_xff_append is set to false.

When use_remote_address is set to true, the HCM uses the real remote address of the client connection when determining whether the origin is internal or external and when modifying headers. This value controls how Envoy determines the trusted client address.

### Trusted client address
A trusted client address is the first source IP address known to be accurate. The source IP address of the downstream node that made a request to the Envoy proxy is considered correct.

Complete XFF sometimes cannot be trusted as malicious agents can forge it. However, if a trusted proxy puts the last address in the XFF, it can be trusted. For example, if we look at the request path IP1 -> IP2 -> IP3 -> Envoy, the IP3 is the node Envoy will consider accurate.

Envoy supports extensions set through the original_ip_detection_extensions field to help determine the original IP address. Currently, there are two extensions, custom_header and xff.

The custom header extension can provide a header name containing the original downstream remote address. We can also tell HCM to treat the detected address as trusted.

With the xff extension, we can specify the number of additional proxy hops starting from the right side of the x-forwarded-for header to trust. If we’d set this value to 1 and use the same example as above, the trusted addresses would be IP2 and IP3.

Envoy uses the trusted client address to determine if the request is internal or external. If we set the use_remote_address to true the request is considered internal if it doesn’t contain XFF and the immediate downstream node’s connection to Envoy has an internal source address. Envoy uses RFC1918 or RFC4193 to determine the internal source address.

If we set the use_remote_address to false (default value), the request is internal only if XFF contains a single internal source address defined by the above two RFCs.

Let’s look at a quick example and set the use_remote_address to true and skip_xff_append to false:

```yaml
...
- filters:
  - name: envoy.filters.network.http_connection_manager
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      use_remote_address: true
      skip_xff_append: false
      ...
```

If we send a request to the proxy from the same machine (i.e., internal request), the headers sent to the upstream will look like this:

```shell
':authority', 'localhost:10000'
':path', '/'
':method', 'GET'
':scheme', 'http'
'user-agent', 'curl/7.64.0'
'accept', '*/*'
'x-forwarded-for', '10.128.0.17'
'x-forwarded-proto', 'http'
'x-envoy-internal', 'true'
'x-request-id', '74513723-9bbd-4959-965a-861e2162555b'
'x-envoy-expected-rq-timeout-ms', '15000'
```

Most of these headers are the same as we saw in the example of the standard header. However, two headers got added – the x-forwarded-for and the x-envoy-internal. The x-forwarded-for will contain the internal IP address and the x-envoy-internal header will get set because we used XFF to determine the address. Instead of figuring out if the request is internal or not by parsing the x-forwarded-for header, we check for the presence of the x-envoy-internal header to quickly determine whether the request is internal or external.

If we send a request from outside of that network, the following headers get sent to the upstream:

```shell
':authority', '35.224.50.133:10000'
':path', '/'
':method', 'GET'
':scheme', 'http'
'user-agent', 'curl/7.64.1'
'accept', '*/*'
'x-forwarded-for', '50.35.69.235'
'x-forwarded-proto', 'http'
'x-envoy-external-address', '50.35.69.235'
'x-request-id', 'dc93fd48-1233-4220-9146-eac52435cdf2'
'x-envoy-expected-rq-timeout-ms', '15000'
```

Notice the :authority value is an actual IP address instead of just localhost. Similarly, the x-forwarded-for header contains the IP address of the called. There’s no x-envoy-internal header because the request is external. However, we do get a new header called x-envoy-external-address. Envoy sets this header only for external requests. The header can be forwarded between internal services and used for analytics based on the origin client’s IP address.