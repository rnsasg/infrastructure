## HTTP Tap Filter Lab

In this lab, we’ll show how to use and configure the HTTP tap filter. We’ll configure a /error route that returns a direct response with a body with the value error. Then, in the tap filter, we’ll configure the matcher to match any requests where the response body contains the string error and the request header debug: true. If both conditions evaluate to be true, then we’ll tap the request and write the output to a file with the prefix tap_debug.

Let’s start by creating the match config. We’ll use two matchers, one to match the request headers (http_request_headers_match) and the second to match the response body (http_response_generic_body_match). We’ll combine these two conditions with a logical AND.

Here’s how the match configuration will look:

```yaml
- name: envoy.filters.http.tap
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.filters.http.tap.v3.Tap
    common_config:
      static_config:
        match:
          and_match:
            rules:
              - http_request_headers_match:
                  headers:
                    name: debug
                    string_match:
                      exact: "true"
              - http_response_generic_body_match:
                  patterns:
                    - string_match: error
```
We’ll use the JSON_BODY_AS_STRING format and write the output to files prefixed with tap_debug:

```yaml
output_config:
  sinks:
    - format: JSON_BODY_AS_STRING
      file_per_tap:
        path_prefix: tap_debug
```

Let’s put both pieces together and create a full configuration. We’ll use direct_response, so we don’t need to set up or run any additional services:

```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          http_filters:
          - name: envoy.filters.http.tap
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.tap.v3.Tap
              common_config:
                static_config:
                  match:
                    and_match:
                      rules:
                        - http_request_headers_match:
                            headers:
                              name: debug
                              string_match:
                                exact: "true"
                        - http_response_generic_body_match:
                            patterns:
                              - string_match: error
                  output_config:
                    sinks:
                      - format: JSON_BODY_AS_STRING
                        file_per_tap:
                          path_prefix: tap_debug
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match:
                  path: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: hello
              - match:
                  path: "/error"
                direct_response:
                  status: 500
                  body:
                    inline_string: error
```

Save the above YAML to 7-lab-1-tap-filter-1.yaml file and run it using func-e CLI:

func-e run -c 7-lab-1-tap-filter-1.yaml &
If we send a request to http://localhost:10000, we’ll receive the response that matches the first route (HTTP 200 and body hello). The request won’t be tapped because we didn’t provide any headers, nor did the response body contain the value error.

Let’s try setting the debug header and sending a request to /error endpoint:

$ curl -H "debug: true" localhost:10000/error
error
This time, a JSON file with the contents of the tapped request is created in the same folder. Here’s how the contents of that file should look:

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
     "value": "/error"
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
     "key": "debug",
     "value": "true"
    },
    {
     "key": "x-forwarded-proto",
     "value": "http"
    },
    {
     "key": "x-request-id",
     "value": "4855ee5d-7798-4c50-8692-a6989e72ca9b"
    }
   ],
   "trailers": []
  },
  "response": {
   "headers": [
    {
     "key": ":status",
     "value": "500"
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
     "value": "Mon, 29 Nov 2021 22:38:32 GMT"
    },
    {
     "key": "server",
     "value": "envoy"
    }
   ],
   "body": {
    "truncated": false,
    "as_string": "error"
   },
   "trailers": []
  }
 }
}
```

The output shows all request headers and trailers, as well as the response we received.

We’ll use the same scenario for the next example, but we’ll implement it using the /tap admin endpoint.

First, let’s create the Envoy configuration:

```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          http_filters:
          - name: envoy.filters.http.tap
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.tap.v3.Tap
              common_config:
                admin_config:
                  config_id: my_tap_id
          - name: envoy.filters.http.router
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match:
                  path: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: hello
              - match:
                  path: "/error"
                direct_response:
                  status: 500
                  body:
                    inline_string: error
admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901
```

This time, we’re using the admin_config field and specifying the configuration ID. Additionally, we’re enabling the admin interface to use the /tap endpoint.

Save the above YAML to 7-lab-1-tap-filter-2.yaml and run it using func-e CLI:

func-e run -c 7-lab-1-tap-filter-2.yaml &
We’ll get the expected responses if we try sending the requests to the / and error paths. We have to send a POST request to the /tap endpoint with the tap configuration to enable the request tapping.

Let’s use this tap configuration that matches any request:

```json
{
  "config_id": "my_tap_id",
  "tap_config": {
    "match": {
      "any_match": true
    },
    "output_config": {
      "sinks": [
        {
          "streaming_admin": {}
        }
      ]
    }
  }
}
```

Note that we’re providing the config ID that matches the ID we’ve defined in the Envoy configuration. If we give an invalid config ID, then we’ll get an error when sending a POST request to /tap endpoint:

Unknown config id 'some_tap_id'. No extension has registered with this id.
We’re also using the streaming_admin field as the output sink, which means that if the POST request to /tap is accepted, then Envoy will stream the serialized JSON messages until we terminate the request.

Let’s save the above JSON to tap-config-any.json and then use cURL to send a POST request to the /tap endpoint:

curl -X POST -d @tap-config-any.json http://localhost:9901/tap
We’ll open a second terminal window and send a cURL request to localhost:10000 to test the configuration. Since we’re matching on all requests, we’ll see the streamed tapped output in the first terminal window:

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
     "value": "POST"
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
     "key": "content-length",
     "value": "198"
    },
    {
     "key": "content-type",
     "value": "application/x-www-form-urlencoded"
    },
    {
     "key": "x-forwarded-proto",
     "value": "http"
    },
    {
     "key": "x-request-id",
     "value": "59ca4c38-6112-444d-9b64-ff30e1326338"
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
     "value": "Mon, 29 Nov 2021 23:09:25 GMT"
    },
    {
     "key": "server",
     "value": "envoy"
    },
    {
     "key": "connection",
     "value": "close"
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
