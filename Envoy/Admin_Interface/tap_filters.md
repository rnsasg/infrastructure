# Tap filter
The purpose of the tap filter is to record HTTP traffic based on some matching properties. There are two ways to configure the tap filter: (1) Using the static_config field inside the Envoy configuration or (2) using the admin_config field and specifying the configuration ID. The difference is that we provide everything at once in the static configuration — the match configuration and the output configuration. When using the admin configuration, we provide only the configuration ID and then use the /tap administrative endpoint to configure the filter at runtime.

As we alluded to, the filter configuration is separated into two parts: the **match configuration** and the **output configuration**.

We can specify the matching predicate with the match configuration that tells the tap filter which requests to tap and write to the configured output.

For example, the snippet below shows how to use the any_match to match all requests, regardless of their properties:

```yaml
common_config:
  static_config:
    match:
      any_match: true
...
```

We also have an option to match on request and response headers, trailers, and body.

## Header/trailer match
The header/trailer matchers use the HttpHeadersMatch proto, where we specify an array of headers to match on. For example, this snippet matches any requests where the request header my-header is set precisely to hello.

```yaml
common_config:
  static_config:
    match:
      http_request_headers_match:
        headers:
          name: "my-header"
          string_match:
            exact: "hello"
...
```

> Note that within the string_match we can use the other matchers (e.g., prefix, suffix, safe_regex) as explained earlier.

## Body match
The generic request and response body match uses the HttpGenericBodyMatch to specify a string or binary match. As the name suggests, the string match (string_match) looks for a string within the HTTP body, and the binary match (binary_match) looks for a sequence of bytes to be located in the HTTP body.

For example, the following snippet matches if the response body contains the string hello:

```yaml
common_config:
  static_config:
    match:
      http_response_generic_body_match:
        patterns:
          string_match: "hello"
...
```

## Match predicates
We can combine multiple headers, trailers, and body matchers with match predicates such as or_match, and_match, and not_match.

The or_match and and_match use the MatchSet proto that describes either a logical OR or a logical AND. We specify a list of rules that make up a set in the rules field within the match set.

The example below shows how to use the and_match to ensure that both the response body contains the word hello and the request header my-header is set to hello:

```yaml
common_config:
  static_config:
    match:
      and_match:
        rules:
         - http_response_generic_body_match:
            patterns:
              - string_match: "hello"
          - http_request_headers_match:
              headers:
                name: "my-header"
                string_match:
                  exact: "hello"
...
```

If we wanted to implement the logical OR, then we could replace the and_match field with the or_match field. The configuration within the field would stay the same, as both fields use the MatchSet proto.

Let’s use the same example as that used previously to show how the not_match works. Let’s say we want to tap all requests that don’t have the header my-header: hello set, and for which the response body doesn’t include the string hello.

Here’s how we could write that configuration:

```yaml
common_config:
  static_config:
    match:
      not_match:
        and_match:
          rules:
          - http_response_generic_body_match:
              patterns:
                - string_match: "hello"
            - http_request_headers_match:
                headers:
                  name: "my-header"
                  string_match:
                    exact: "hello"
...
```

The not_match field uses the MatchPredicate proto just like the parent match field. The match field is a recursive structure, and it allows us to create complex nested match configurations.

The last field to mention here is the any_match. This is a Boolean field that, when set to true, will always match.

## Output configuration
Once the requests are tapped, we need to tell the filter where to write the output. At the moment, we can configure a single output sink.

Here’s how a sample output configuration would look:

```yaml
...
output_config:
  sinks:
    - format: JSON_BODY_AS_STRING
      file_per_tap:
        path_prefix: tap
...
```

Using the file_per_tap, we specify that we want to output a single file for every tapped stream. The path_prefix specifies the prefix for the output file. The files are named using the following format:

<path_prefix>_<id>.<pb | json>
The id represents an identifier that allows us to distinguish the recorded trace for stream instances. The file extension (pb or json) depends on the format selection.

The second option for capturing the output is to use the streaming_admin field. This specifies that the /tap admin endpoint will stream the tapped output. Note that to use the /tap admin endpoint for the output, the tap filter must also be configured using the admin_config field. If we statically configure the tap filter, we won’t use the /tap endpoint to get the output.

### Format selection
We have multiple options for the output format that specifies how messages are written. Let’s look at the different formats, starting with the default format, JSON_BODY_AS_BYTES.

The JSON_BODY_AS_BYTES output format outputs the messages as JSON, and any response body data will be in the as_bytes field that contains the base64 encoded string.

For example, here’s how tapped output would look:

```json
{
 "http_buffered_trace": {
  "request": {
   "headers": [
    {
     "key": ":authority",
     "value": "localhost:10000"
    },
    {
     "key": ":path",
     "value": "/"
    },
    {
     "key": ":method",
     "value": "GET"
    },
    {
     "key": ":scheme",
     "value": "http"
    },
    {
     "key": "user-agent",
     "value": "curl/7.64.0"
    },
    {
     "key": "accept",
     "value": "*/*"
    },
    {
     "key": "my-header",
     "value": "hello"
    },
    {
     "key": "x-forwarded-proto",
     "value": "http"
    },
    {
     "key": "x-request-id",
     "value": "67e3e8ac-429a-42fb-945b-ec25927fdcc1"
    }
   ],
   "trailers": []
  },
  "response": {
   "headers": [
    {
     "key": ":status",
     "value": "200"
    },
    {
     "key": "content-length",
     "value": "5"
    },
    {
     "key": "content-type",
     "value": "text/plain"
    },
    {
     "key": "date",
     "value": "Mon, 29 Nov 2021 19:31:43 GMT"
    },
    {
     "key": "server",
     "value": "envoy"
    }
   ],
   "body": {
    "truncated": false,
    "as_bytes": "aGVsbG8="
   },
   "trailers": []
  }
 }
}
```

Note the as_bytes field in the body. The value is a base64 encoded representation of the body data (hello in this example).

The second output format is JSON_BODY_AS_STRING. The difference between the previous format is that with JSON_BODY_AS_STRING, the body data is written in the as_string field as a string. This format is useful when we know that the body is human readable and there’s no need to base64 encode the data.

```json
...
   "body": {
    "truncated": false,
    "as_string": "hello"
   },
...
```

The other three format types are PROTO_BINARY, PROTO_BINARY_LENGTH_DELIMITED, and PROTO_TEXT.

The PROTO_BINARY format writes the output in the binary proto format. This format is not self-delimiting, which means that if the sink writes multiple binary messages without any length information, the data stream will not be useful. If we’re writing one message per file, then the output format will be easier to parse.

We can also use the PROTO_BINARY_LENGTH_DELIMITED format, in which messages are written as sequence tuples. Each tuple is the message length (encoded as 32-bit protobuf varint type), followed by the binary message.

Lastly, we can also use the PROTO_TEXT format, in which the output is written in the protobuf format below.

```json
http_buffered_trace {
  request {
    headers {
      key: ":authority"
      value: "localhost:10000"
    }
    headers {
      key: ":path"
      value: "/"
    }
    headers {
      key: ":method"
      value: "GET"
    }
    headers {
      key: ":scheme"
      value: "http"
    }
    headers {
      key: "user-agent"
      value: "curl/7.64.0"
    }
    headers {
      key: "accept"
      value: "*/*"
    }
    headers {
      key: "debug"
      value: "true"
    }
    headers {
      key: "x-forwarded-proto"
      value: "http"
    }
    headers {
      key: "x-request-id"
      value: "af6e0879-e057-4efc-83e4-846ff4d46efe"
    }
  }
  response {
    headers {
      key: ":status"
      value: "500"
    }
    headers {
      key: "content-length"
      value: "5"
    }
    headers {
      key: "content-type"
      value: "text/plain"
    }
    headers {
      key: "date"
      value: "Mon, 29 Nov 2021 22:32:40 GMT"
    }
    headers {
      key: "server"
      value: "envoy"
    }
    body {
      as_bytes: "hello"
    }
  }
}
```

## Configuring the tap filter statically
We combine the matching config with the output config (using the file_per_tap field) to configure the tap filter statically.

Here’s a snippet that configures the tap filter via static configuration:

```yaml
- name: envoy.filters.http.tap
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.filters.http.tap.v3.Tap
    common_config:
      static_config:
        match_config:
          any_match: true
        output_config:
          sinks:
            - format: JSON_BODY_AS_STRING
              file_per_tap:
                path_prefx: my-tap
```

The above configuration will match all requests and write the output to file names with my-tap prefix.

## Configuring the tap filter using the /tap endpoint
To use the /tap endpoint, we have to specify the admin_config and the config_id in the tap filter configuration:

```yaml
- name: envoy.filters.http.tap
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.filters.http.tap.v3.Tap
    common_config:
      admin_config:
        config_id: my_tap_config_id
```

Once specified, we can send a POST request to the /tap endpoint to configure the tap filter. For example, here’s the POST body that configures the tap filter referenced by the my_tap_config_id name:

```yaml
config_id: my_tap_config_id
tap_config:
  match_config:
    any_match: true
  output_config:
    sinks:
      - streaming_admin: {}
```

The format in which we specify the match configuration is equivalent to how we set it for the statically provided configuration.

The clear advantage of using the admin configuration and the /tap endpoint is that we can update the match configuration at runtime, and we don’t need to restart the Envoy proxy.
