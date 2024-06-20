# Lab 1: Request Matching
In this lab, we’ll learn how to configure different ways to match requests. We’ll use a direct_response and we won’t involve any Envoy clusters yet.

# Path matching
Here are the rules we’re trying to put into the configuration:

* All requests need to be from the hello.io domain (i.e., we’ll use Host: hello.io header when making requests to the proxy)
* All requests made to path /api will return the string hello - path
* All requests made to the root path (i.e., /) will return the string hello - prefix
* All requests starting with /hello and followed by a number (e.g. /hello/1, /hello/523) should return hello - regex string
Let’s look at the configuration:

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: hello_vhost
              domains: ["hello.io"]
              routes:
              - match:
                  path: "/api"
                direct_response:
                  status: 200
                  body:
                    inline_string: "hello - path"
              - match:
                  safe_regex:
                    google_re2: {}
                    regex: ^/hello/\d+$
                direct_response:
                  status: 200
                  body:
                    inline_string: "hello - regex"
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "hello - prefix"
```

Since we only have a single domain, we’ll use a single virtual host. The domain array will contain a single domain, hello.io. This is the first level of matching Envoy will do.

Then, the virtual host will have multiple route matches. First, we’ll use the path match because we want to exactly match the ‘/api’ path. Second, we define the regex match using the ^/hello/\d+$ regular expression. Finally, we define the prefix match. Note that the order in which these matches are defined matters. If we’d put the prefix match to the top, none of the remaining matches would be evaluated because the prefix match would always be true.

To try this out, save the above YAML to 2-lab-1-request-matching-1.yaml and run func-e run -c 2-lab-1-request-matching-1.yaml to start the Envoy proxy.

From a separate terminal, we can make a couple of test calls:

```shell
$ curl -H "Host: hello.io" localhost:10000
hello - prefix

$ curl -H "Host: hello.io" localhost:10000/api
hello - path

$ curl -H "Host: hello.io" localhost:10000/hello/123
hello - regex
```


## Headers matching

Matching the headers from incoming requests can be combined with path matching to implement complex scenarios. In this example, we’ll use a combination of prefix matches and different header matches.

Let’s come up with some rules:

* All POST requests with the header debug: 1 sent to /1 return a 422 status code
* All requests with header “path” matching the regular expression ^/hello/\d+$ sent to /2 return a 200 status code and message regex
* All requests with header name priority set to a value between 1 and 5, sent to /3 return a 200 status code and message priority
* All requests sent to /4 where the header test is present return a 500
Here are the above rules translated to Envoy configuration:

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: vhost
              domains: ["*"]
              routes:
              - match:
                  path: "/1"
                  headers:
                  - name: ":method"
                    string_match:
                      exact: POST
                  - name: "debug"
                    string_match:
                      exact: "1"
                direct_response:
                  status: 422
              - match:
                  path: "/2"
                  headers:
                  - name: "path"
                    safe_regex_match:
                      google_re2: {}
                      regex: ^/hello/\d+$
                direct_response:
                  status: 200
                  body:
                    inline_string: "regex"
              - match:
                  path: "/3"
                  headers:
                  - name: "priority"
                    range_match:
                      start: 1
                      end: 6
                direct_response:
                  status: 200
                  body:
                    inline_string: "priority"
              - match:
                  path: "/4"
                  headers:
                  - name: "test"
                    present_match: true
                direct_response:
                  status: 500
```

Save the above YAML to 2-lab-1-request-matching-2.yaml and run func-e run -c 2-lab-1-request-matching-2.yaml.

Let’s try sending a couple of requests and test out the rules:

```
$ curl -v -X POST -H "debug: 1" localhost:10000/1
...
> User-Agent: curl/7.64.0
> Accept: */*
> debug: 1
>
< HTTP/1.1 422 Unprocessable Entity

$ curl -H "path: /hello/123" localhost:10000/2
regex

$ curl -H "priority: 3" localhost:10000/3
priority


$ curl -v -H "test: tst" localhost:10000/4
...
> User-Agent: curl/7.64.0
> Accept: */*
> test: tst
>
< HTTP/1.1 500 Internal Server Error
```

## Query parameters matching
In the same way we did path and headers matching, we could also match specific query parameters and their values. Query parameter matching supports the same rules for matching as the other two: matching for exact values, prefixes, and suffixes, using regular expressions, and checking if the query parameter contains a specific value.

Let’s write the following scenarios in the configuration:

All requests sent to the path /1 with a query parameter test present return a 422 status code
All requests sent to the path /2 with a query parameter called env with a value starting with env_ (ignoring the case), return a 200 status code
All requests sent to the path /3 with query parameter debug set to true, return a 500 status code
The above rules translate to the following Envoy configuration:

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: vhost
              domains: ["*"]
              routes:
              - match:
                  path: "/1"
                  query_parameters:
                  - name: test
                    present_match: true
                direct_response:
                  status: 422
              - match:
                  path: "/2"
                  query_parameters:
                  - name: env
                    string_match:
                      prefix: env_
                      ignore_case: true
                direct_response:
                  status: 200
              - match:
                  path: "/3"
                  query_parameters:
                  - name: debug
                    string_match:
                      exact: "true"
                direct_response:
                  status: 500
```

Save the above YAML to 2-lab-1-request-matching-3.yaml and run func-e run -c 2-lab-1-request-matching-3.yaml.

Let’s try sending a couple of requests and test out the rules:

```
$ curl -v localhost:10000/1?test
...
> GET /1?test HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/7.64.0
> Accept: */*
>
< HTTP/1.1 422 Unprocessable Entity

$ curl -v localhost:10000/2?env=eNv_prod
...
> GET /2?env=eNv_prod HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/7.64.0
> Accept: */*
>
< HTTP/1.1 200 OK

$ curl -v localhost:10000/3?debug=true
...
> GET /3?debug=true HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/7.64.0
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
```

## 2lab1requestmatching2-221021-122629.yaml

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: vhost
              domains: ["*"]
              routes:
              - match:
                  path: "/1"
                  headers:
                  - name: ":method"
                    exact_match: POST
                  - name: "debug"
                    exact_match: "1"
                direct_response:
                  status: 422
              - match:
                  path: "/2"
                  headers:
                  - name: "path"
                    safe_regex_match:
                      google_re2: {}
                      regex: ^/hello/\d+$
                direct_response:
                  status: 200
                  body:
                    inline_string: "regex"
              - match:
                  path: "/3"
                  headers:
                  - name: "priority"
                    range_match:
                      start: 1
                      end: 6
                direct_response:
                  status: 200
                  body:
                    inline_string: "priority"
              - match:
                  path: "/4"
                  headers:
                  - name: "test"
                    present_match: true
                direct_response:
                  status: 500
```

## 2lab1requestmatching2envoy120-221021-122629.yaml

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: vhost
              domains: ["*"]
              routes:
              - match:
                  path: "/1"
                  headers:
                  - name: ":method"
                    string_match:
                      exact: POST
                  - name: "debug"
                    string_match:
                      exact: "1"
                direct_response:
                  status: 422
              - match:
                  path: "/2"
                  headers:
                  - name: "path"
                    safe_regex_match:
                      google_re2: {}
                      regex: ^/hello/\d+$
                direct_response:
                  status: 200
                  body:
                    inline_string: "regex"
              - match:
                  path: "/3"
                  headers:
                  - name: "priority"
                    range_match:
                      start: 1
                      end: 6
                direct_response:
                  status: 200
                  body:
                    inline_string: "priority"
              - match:
                  path: "/4"
                  headers:
                  - name: "test"
                    present_match: true
                direct_response:
                  status: 500
```

## /Users/kumarro/Downloads/2lab1requestmatching1-221021-122629.yaml

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: hello_vhost
              domains: ["hello.io"]
              routes:
              - match:
                  path: "/api"
                direct_response:
                  status: 200
                  body:
                    inline_string: "hello - path"
              - match:
                  safe_regex:
                    google_re2: {}
                    regex: ^/hello/\d+$
                direct_response:
                  status: 200
                  body:
                    inline_string: "hello - regex"
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "hello - prefix"
```

## 2lab1requestmatching3-221021-122629.yaml

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
          stat_prefix: hello_service
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: route
            virtual_hosts:
            - name: vhost
              domains: ["*"]
              routes:
              - match:
                  path: "/1"
                  query_parameters:
                  - name: test
                    present_match: true
                direct_response:
                  status: 422
              - match:
                  path: "/2"
                  query_parameters:
                  - name: env
                    string_match:
                      prefix: env_
                      ignore_case: true
                direct_response:
                  status: 200
              - match:
                  path: "/3"
                  query_parameters:
                  - name: debug
                    string_match:
                      exact: "true"
                direct_response:
                  status: 500
```

