# Logging

Whether it’s used as a traffic gateway or a sidecar in the service mesh, Envoy is in a unique position to reveal what’s going on in your network. A common path to understanding is logging, be it for analytics, auditing, or troubleshooting. Logging also carries a problem of volume and the potential to reveal secrets.


## What is access logging?
Whenever you open your browser and visit Google or other websites, the server on the other side collects the information about your visit. Specifically, it’s collecting and storing the data about web pages you requested from the server. In most cases, this data includes the origin (i.e., host information), date and time you requested the page, the request properties (method, path, headers, body, etc.), the status the server returns, the size of the request, and more. All this data typically gets stored in text files called access logs.

Access log entries from web servers or proxies typically follow a standardized common logging format. Different proxies and servers can use their own default access log formats. Envoy has its default logging format. We can customize the default format and configure it to write the logs in the same format as other servers, such as Apache or NGINX. The same access log format allows us to use different servers and combine data logging and analysis using a single tool.

This module will explain how access logging works in Envoy and how to configure and customize it.

Capturing and reading access logs
We can configure the capture of any access requests made to the Envoy proxy and write them to so-called access logs. Let’s look at an example of a couple of access log entries:

```shell
[2021-11-01T20:37:45.204Z] "GET / HTTP/1.1" 200 - 0 3 0 - "-" "curl/7.64.0" "9c08a41b-805f-42c0-bb17-40ec50a3377a" "localhost:10000" "-"
[2021-11-01T21:08:18.274Z] "POST /hello HTTP/1.1" 200 - 0 3 0 - "-" "curl/7.64.0" "6a593d31-c9ac-453a-80e9-ab805d07ae20" "localhost:10000" "-"
[2021-11-01T21:09:42.717Z] "GET /test HTTP/1.1" 404 NR 0 0 0 - "-" "curl/7.64.0" "1acc3559-50eb-463c-ae21-686fe34abbe8" "localhost:10000" "-"
```

The output contains three different log entries and follows the same default log format. The default log format looks like this:

```shell
[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%"
%RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION%
%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%"
"%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
```

The values such as %RESPONSE_FLAGS%, %REQ(:METHOD)% and others are called **command operators**.

## Command operators
The command operators extract the relevant data and insert it in the log entry for both TCP and HTTP. If the values are not set or unavailable (for example, RESPONSE_CODE in TCP), the logs will contain the character - (or "-" for JSON logs).

Each command operator starts and ends with the character %, for example, %START_TIME%. If the command operator accepts any parameters, we can provide them in parentheses. For example, if we wanted to log only day, month, and year using the START_TIME command operator, then we could configure it by specifying the values in the parentheses like this: %START_TIME(%d-%m-%Y)%.

Let’s look at the different command operators. We’ve tried to group them into separate tables based on their common properties.

| Command Operator                   | Description                                                                                                      | Example                                      |
|------------------------------------|------------------------------------------------------------------------------------------------------------------|----------------------------------------------|
| `START_TIME`                       | Request start time including milliseconds.                                                                        | `%START_TIME(%Y/%m/%dT%H:%M:%S%z ^%)%`       |
| `PROTOCOL`                         | Protocol (either HTTP/1.1, HTTP/2 or HTTP/3).                                                                     | `%PROTOCOL%`                                 |
| `RESPONSE_CODE`                    | HTTP response code. Response code gets set to 0 if the downstream client disconnected.                            | `%RESPONSE_CODE%`                            |
| `RESPONSE_CODE_DETAILS`            | Additional information about the HTTP response (e.g., who set it and why).                                        | `%RESPONSE_CODE_DETAILS%`                    |
| `CONNECTION_TERMINATION_DETAILS`   | Provides additional information about why Envoy terminated the connection for L4 reasons.                         | `%CONNECTION_TERMINATION_DETAILS%`           |
| `ROUTE_NAME`                       | Name of the route.                                                                                               | `%ROUTE_NAME%`                               |
| `CONNECTION_ID`                    | An identifier for the downstream connection. It can cross-reference TCP access logs across multiple log sinks or cross-reference timer-based reports for the same connection. The identifier is unique with a high likelihood within an execution but can duplicate across multiple instances or between restarts. | `%CONNECTION_ID%`                            |
| `GRPC_STATUS`                      | gRPC status code, including text message and a number.                                                           | `%GRPC_STATUS%`                              |
| `HOSTNAME`                         | The system hostname.                                                                                             | `%HOSTNAME%`                                 |
| `LOCAL_REPLY_BODY`                 | The body text for the requests rejected by Envoy.                                                                | `%LOCAL_REPLY_BODY%`                         |
| `FILTER_CHAIN_NAME`                | The network filter chain name of the downstream connection.                                                      | `%FILTER_CHAIN_NAME%`                        |


## Sizes
This group contains all command operators representing sizes — from request and response header bytes to bytes received and sent.

| Command Operator                   | Description                                                                                                      | Example                                      |
|------------------------------------|------------------------------------------------------------------------------------------------------------------|----------------------------------------------|
| `REQUEST_HEADER_BYTES`             | Uncompressed bytes of request headers.                                                                           | `%REQUEST_HEADER_BYTES%`                     |
| `RESPONSE_HEADERS_BYTES`           | Uncompressed bytes of response headers.                                                                          | `%RESPONSE_HEADERS_BYTES%`                   |
| `RESPONSE_TRAILERS_BYTES`          | Uncompressed bytes of response trailers.                                                                         | `%RESPONSE_TRAILERS_BYTES%`                  |
| `BYTES_SENT`                       | Body bytes sent for HTTP, and downstream bytes sent on connection for TCP.                                        | `%BYTES_SENT%`                               |
| `BYTES_RECEIVED`                   | Body bytes received.                                                                                             | `%BYTES_RECEIVED%`                           |
| `UPSTREAM_WIRE_BYTES_SENT`         | The total number of bytes sent upstream by the HTTP stream.                                                      | `%UPSTREAM_WIRE_BYTES_SENT%`                 |
| `UPSTREAM_WIRE_BYTES_RECEIVED`     | Total number of bytes received from the upstream HTTP stream.                                                    | `%UPSTREAM_WIRE_BYTES_RECEIVED%`             |
| `UPSTREAM_HEADER_BYTES_SENT`       | The number of header bytes sent upstream by the HTTP stream.                                                     | `%UPSTREAM_HEADER_BYTES_SENT%`               |
| `UPSTREAM_HEADER_BYTES_RECEIVED`   | The number of header bytes received from the upstream by the HTTP stream.                                        | `%UPSTREAM_HEADER_BYTES_RECEIVED%`           |
| `DOWNSTREAM_WIRE_BYTES_SENT`       | The total number of bytes sent downstream by the HTTP stream.                                                    | `%DOWNSTREAM_WIRE_BYTES_SENT%`               |
| `DOWNSTREAM_WIRE_BYTES_RECEIVED`   | The total number of bytes received from the downstream by the HTTP stream.                                       | `%DOWNSTREAM_WIRE_BYTES_RECEIVED%`           |
| `DOWNSTREAM_HEADER_BYTES_SENT`     | The number of header bytes sent downstream by the HTTP stream.                                                   | `%DOWNSTREAM_HEADER_BYTES_SENT%`             |
| `DOWNSTREAM_HEADER_BYTES_RECEIVED` | The number of header bytes received from the downstream by the HTTP stream.                                      | `%DOWNSTREAM_HEADER_BYTES_RECEIVED%`         |


## Durations

| Command Operator        | Description                                                                                                      | Example                    |
|-------------------------|------------------------------------------------------------------------------------------------------------------|----------------------------|
| `DURATION`              | The total duration of the request (in milliseconds) from the start time to the last byte out.                    | `%DURATION%`               |
| `REQUEST_DURATION`      | The total duration of the request (in milliseconds) from the start time to the last byte of the request received from downstream. | `%REQUEST_DURATION%`       |
| `REQUEST_TX_DURATION`   | The total duration of the request (in milliseconds) from the start time to the last byte sent upstream.          | `%REQUEST_TX_DURATION%`    |
| `RESPONSE_DURATION`     | The total duration of the request (in milliseconds) from the start time to the first byte read from the upstream host. | `%RESPONSE_DURATION%`      |
| `RESPONSE_TX_DURATION`  | The total duration of the request (in milliseconds) from the first byte read from the upstream host to the last byte sent downstream. | `%RESPONSE_TX_DURATION%`   |


## Response flags
The RESPONSE_FLAGS command operator contains additional details about the response or connection. The following list shows the response flags’ values and their meaning for HTTP and TCP connections.

## HTTP and TCP

* UH: No healthy upstream hosts in an upstream cluster in addition to 503 response code.
* UF: Upstream connection failure in addition to 503 response code.
* UO: Upstream overflow (circuit breaking) in addition to 503 response code.
* NR: No route configured for a given request in addition to 404 response code or no matching filter chain for a downstream connection.
* URX: The request was rejected because the upstream retry limit (HTTP) or maximum connection attempts (TCP) was reached.
* NC: Upstream cluster not found.
* DT: When a request or connection exceeded max_connection_duration or max_downstream_connection_duration.

## HTTP only

* DC: Downstream connection termination.
* LH: Local service failed health check request in addition to 503 response code.
* UT: Upstream request timeout in addition to 504 response code.
* LR: Connection local reset in addition to 503 response code.
* UR: Upstream remote reset in addition to 503 response code.
* UC: Upstream connection termination in addition to 503 response code.
* DI: The request processing was delayed for a period specified via fault injection.
* FI: The request was aborted with a response code specified via fault injection.
* RL: The request was rate-limited locally by the HTTP rate limit filter in addition to the 429 response code.
* UAEX: The request was denied by the external authorization service.
* RLSE: The request was rejected because of an error in the rate limit service.
* IH: The request was rejected because it set an invalid value for a strictly checked header in addition to 400 response code.
* SI: Stream idle timeout in addition to 408 response code.
* DPE: The downstream request had an HTTP protocol error.
* UPE: The upstream response had an HTTP protocol error.
* UMSDR: The upstream request reached the max stream duration.
* OM: Overload Manager terminated the request.

## Upstream information

| Command Operator                     | Description                                                                                                                     | Example                   |
|--------------------------------------|---------------------------------------------------------------------------------------------------------------------------------|---------------------------|
| `UPSTREAM_HOST`                      | Upstream host URL or tcp://ip:port for TCP connections.                                                                         | `%UPSTREAM_HOST%`         |
| `UPSTREAM_CLUSTER`                   | Upstream cluster to which the upstream host belongs. If runtime feature envoy.reloadable_features.use_observable_cluster_name is enabled, then alt_stat_name will be used if provided. | `%UPSTREAM_CLUSTER%`      |
| `UPSTREAM_LOCAL_ADDRESS`             | The local address of the upstream connection. If it’s an IP address, it includes both the address and port.                     | `%UPSTREAM_LOCAL_ADDRESS%`|
| `UPSTREAM_TRANSPORT_FAILURE_REASON`  | Provides the failure reason from the transport socket if the connection failed due to the transport socket.                     |                           |


## Downstream information

| Command Operator                                   | Description                                                                                                                                                                                                                                               | Example                                   |
|----------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------|
| `DOWNSTREAM_REMOTE_ADDRESS`                        | Remote address of the downstream connection. If it’s an IP address, it includes both the address and port.                                                                                                                                                | `%DOWNSTREAM_REMOTE_ADDRESS%`             |
| `DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT`           | Remote address of the downstream connection. If it’s an IP address, then it includes the address only.                                                                                                                                                   | `%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%`|
| `DOWNSTREAM_DIRECT_REMOTE_ADDRESS`                 | Direct remote address of the downstream connection. If it’s an IP address, it includes both the address and port.                                                                                                                                         | `%DOWNSTREAM_DIRECT_REMOTE_ADDRESS%`      |
| `DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT`    | Direct remote address of the downstream connection. If it’s an IP address, then it includes the address only.                                                                                                                                             | `%DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT%` |
| `DOWNSTREAM_LOCAL_ADDRESS`                         | The local address of the downstream connection. If it’s an IP address, it includes both the address and port. If the original connection was redirected by iptables REDIRECT, this value represents the original destination address restored by the original destination filter. If redirected by iptables TPROXY and the listener’s transparent option was set to true, then this represents the original destination address and port. | `%DOWNSTREAM_LOCAL_ADDRESS%`              |
| `DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT`            | Same as DOWNSTREAM_LOCAL_ADDRESS excluding port, if the address is an IP address.                                                                                                                                                                        | `%DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT%` |
| `DOWNSTREAM_LOCAL_PORT`                            | Similar to DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT but only extracts the port portion of the DOWNSTREAM_LOCAL_ADDRESS.                                                                                                                                      | `%DOWNSTREAM_LOCAL_PORT%`                 |


## Headers and trailers
The REQ, RESP, and TRAILER command operator allows us to extract request, response, and trailer header information and include it in the logs.

| Command Operator      | Description                                                                                                                                                                                                                                                        | Example                                                       |
|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------|
| `REQ(X?Y):Z`          | HTTP request header where X is the main HTTP header, Y is the alternative one, and Z is an optional parameter denoting string truncation up to Z characters long. If the value from header X is not set, then request header Y is used. If none of the headers are present, - will be in the logs. | `%REQ(HELLO?BYE):5%` includes the value of the header hello. If not set, uses the value from the header bye. It truncates the value to 5 characters. |
| `RESP(X?Y):Z`         | Same as REQ but taken from HTTP response headers.                                                                                                                                                                                                                   | `%RESP(HELLO?BYE):5%` includes the value of the header hello. If not set, uses the value from the header bye. It truncates the value to 5 characters. |
| `TRAILER(X?Y):Z`      | Same as REQ but taken from HTTP response trailers.                                                                                                                                                                                                                 | `%TRAILER(HELLO?BYE):5%` includes the value of the header hello. If not set, uses the value from the header bye. It truncates the value to 5 characters. |

## Metadata

| Command Operator                      | Description                                                                                                                                                                                                                                                                                                                                                                                   | Example                                                                                                                                                                                                                                                                                                   |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DYNAMIC_METADATA(NAMESPACE:KEY*):Z`  | Dynamic metadata info, where NAMESPACE is the filter used when setting the metadata. KEY is an optional lookup key in the namespace with the option of specifying nested keys separated by :. Z is an optional parameter denoting string truncation up to Z characters long.                                                                                                                       | For example, the `my_filter: {"my_key": "hello", "json_object": {"some_key": "foo"}}` metadata can be logged using `%DYNAMIC_METADATA(my_filter)%`. To log a specific key, we could write `%DYNAMIC_METADATA(my_filter:my_key)%`.                                                                            |
| `CLUSTER_METADATA(NAMESPACE:KEY*):Z`  | Upstream cluster metadata info, where NAMESPACE is the filter namespace used when setting the metadata, KEY is an optional lookup key in the namespace with the option of specifying nested keys separated by :. Z is an optional parameter denoting string truncation up to Z characters long.                                                                                               | See example for `DYNAMIC_METADATA`.                                                                                                                                                                                                                                                                       |
| `FILTER_STATE(KEY:F):Z`               | Filter state info, where the KEY is required to look up the filter state object. The serialized proto will be logged as a JSON string if possible. If the serialized proto is unknown, it will be logged as a protobuf debug string. F is an optional parameter indicating which method `FilterState` uses for serialization. If PLAIN is set, then the filter state object will be serialized as an unstructured string. If TYPED is set or no F is provided, then the filter state object will be serialized as a JSON string. Z is an optional parameter denoting string truncation up to Z characters long. | For example, `%FILTER_STATE(my_key:PLAIN):50%` includes the value of the filter state object identified by `my_key`, serialized as an unstructured string and truncated to 50 characters.                                                                                                                    |

## TLS

| Command Operator                      | Description                                                                                                               | Example                              |
|---------------------------------------|---------------------------------------------------------------------------------------------------------------------------|--------------------------------------|
| REQUESTED_SERVER_NAME                 | String value set on SSL connection socket for Server Name Indication (SNI).                                                | %REQUESTED_SERVER_NAME%              |
| DOWNSTREAM_LOCAL_URI_SAN              | The URIs present in the SAN of the local certificate used to establish the downstream TLS connection.                      | %DOWNSTREAM_LOCAL_URI_SAN%           |
| DOWNSTREAM_PEER_URI_SAN               | The URIs present in the SAN of the peer certificate used to establish the downstream TLS connection.                       | %DOWNSTREAM_PEER_URI_SAN%            |
| DOWNSTREAM_LOCAL_SUBJECT              | The subject present in the local certificate used to establish the downstream TLS connection.                              | %DOWNSTREAM_LOCAL_SUBJECT%           |
| DOWNSTREAM_PEER_SUBJECT               | The subject present in the peer certificate used to establish the downstream TLS connection.                               | %DOWNSTREAM_PEER_SUBJECT%            |
| DOWNSTREAM_PEER_ISSUER                | The issuer present in the peer certificate used to establish the downstream TLS connection.                                | %DOWNSTREAM_PEER_ISSUER%             |
| DOWNSTREAM_TLS_SESSION_ID             | The session ID for the established downstream TLS connection.                                                             | %DOWNSTREAM_TLS_SESSION_ID%          |
| DOWNSTREAM_TLS_CIPHER                 | The OpenSSL name for the set of ciphers used to establish the downstream TLS connection.                                   | %DOWNSTREAM_TLS_CIPHER%              |
| DOWNSTREAM_TLS_VERSION                | The TLS version (TLSv1.2 or TLSv1.3) used to establish the downstream TLS connection.                                      | %DOWNSTREAM_TLS_VERSION%             |
| DOWNSTREAM_PEER_FINGERPRINT_256       | The hex-encoded SHA256 fingerprint of the client certificate used to establish the downstream TLS connection.              | %DOWNSTREAM_PEER_FINGERPRINT_256%    |
| DOWNSTREAM_PEER_FINGERPRINT_1         | The hex-encoded SHA1 fingerprint of the client certificate used to establish the downstream TLS connection.                | %DOWNSTREAM_PEER_FINGERPRINT_1%      |
| DOWNSTREAM_PEER_SERIAL                | The serial number of the client certificate used to establish the downstream TLS connection.                               | %DOWNSTREAM_PEER_SERIAL%             |
| DOWNSTREAM_PEER_CERT                  | The client certificate in the URL-safe encoded PEM format used to establish the downstream TLS connection.                 | %DOWNSTREAM_PEER_CERT%               |
| DOWNSTREAM_PEER_CERT_V_START          | The validity start date of the client certificate used to establish the downstream TLS connection.                         | %DOWNSTREAM_PEER_CERT_V_START%       |
| DOWNSTREAM_PEER_CERT_V_END            | The validity end date of the client certificate used to establish the downstream TLS connection.                           | %DOWNSTREAM_PEER_CERT_V_END%         |
