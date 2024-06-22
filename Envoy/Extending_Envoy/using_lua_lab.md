# Lua Lab

In this lab, we’ll write a Lua script that adds a header to response headers and uses a global script defined in a file.

We’ll create an Envoy configuration and a Lua script that adds a header to the response handle. Since we won’t use the request path, we don’t have to define the envoy_on_request function. The response function looks like this:

```lua
function envoy_on_response(response_handle)
  response_handle:headers():add("hello", "world")
end
```
We’re calling the add(<header-name>, <header-value>) function on the headers object returned from the headers() function.

Let’s define this script inline within the Envoy configuration. To simplify the configuration, we’ll use a direct_response, instead of a cluster.

```yaml
static_resources:
  listeners:
  - name: main
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
          route_config:
            name: some_route
            virtual_hosts:
            - name: some_service
              domains:
              - "*"
              routes:
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
          http_filters:
          - name: envoy.filters.http.lua
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
              inline_code: |
                function envoy_on_response(response_handle)
                  response_handle:headers():add("hello", "world")
                end
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```

Save the above YAML to 8-lab-1-lua-script.yaml and run it:

func-e run -c 8-lab-1-lua-script.yaml &
To test out the function, we can send a request to localhost:10000 and inspect the response headers:

```shell
$ curl -v localhost:10000
...
> GET / HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/7.64.0
> Accept: */*
>
< HTTP/1.1 200 OK
< content-length: 3
< content-type: text/plain
< hello: world
< date: Tue, 23 Nov 2021 21:37:01 GMT
< server: envoy
<
200
```

The output should include the hello: world header we added to other standard headers.

Let’s look at a more complex scenario. For incoming requests, we want to check if they have a header called my-request-id, and if the header doesn’t exist, then we want to add a my-request-id header for all GET requests.

Because we want to check the method and a header in the envoy_on_response function, we’ll use the dynamic metadata to store those values in the envoy_on_request function. Then, in the response function, we can read the metadata, check whether the header is set and whether the method is GET, and add the my-request-id header.

Here’s what the code looks like:

```lua
function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  local metadata = request_handle:streamInfo():dynamicMetadata()
  metadata:set("envoy.filters.http.lua", "requestInfo", {
      requestId = headers:get("my-request-id"),
      method = headers:get(":method"),
    })
end
function envoy_on_response(response_handle)
  local requestInfoObj = response_handle:streamInfo():dynamicMetadata():get("envoy.filters.http.lua")["requestInfo"]

  local requestId = requestInfoObj.requestId
  local method = requestInfoObj.method
  if (requestId == nil or requestId == '') and (method == 'GET') then
    response_handle:logInfo("Adding request ID header")
    response_handle:headers():add("my-request-id", "some_id_here")
  end
end
```

Note that at the moment, we’re using some_id_here for the my-request-id value, and later we’ll create a function that generates an ID for us. Here’s how the complete Envoy configuration looks:

```yaml
static_resources:
  listeners:
  - name: main
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
          route_config:
            name: some_route
            virtual_hosts:
            - name: some_service
              domains:
              - "*"
              routes:
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
          http_filters:
          - name: envoy.filters.http.lua
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
              inline_code: |
                function envoy_on_request(request_handle)
                  local headers = request_handle:headers()
                  local metadata = request_handle:streamInfo():dynamicMetadata()
                  metadata:set("envoy.filters.http.lua", "requestInfo", {
                      requestId = headers:get("my-request-id"),
                      method = headers:get(":method"),
                    })
                end
                function envoy_on_response(response_handle)
                  local requestInfoObj = response_handle:streamInfo():dynamicMetadata():get("envoy.filters.http.lua")["requestInfo"]

                  local requestId = requestInfoObj.requestId
                  local method = requestInfoObj.method
                  if (requestId == nil or requestId == '') and (method == 'GET') then
                    response_handle:logInfo("Adding request ID header")
                    response_handle:headers():add("my-request-id", "some_id_here")
                  end
                end
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```
Save the above YAML to 8-lab-1-lua-script-1.yaml and run it:

func-e run -c 8-lab-1-lua-script-1.yaml &
Let’s try out a couple of scenarios. First, we’ll send a GET request without my-request-id header set:

```shell
$ curl -v localhost:10000
...
[2021-11-23 22:59:35.932][2258][info][lua] [source/extensions/filters/http/lua/lua_filter.cc:795] script log: Adding request ID header
< HTTP/1.1 200 OK
< content-length: 3
< content-type: text/plain
< my-request-id: some_id_here
< date: Tue, 23 Nov 2021 22:59:35 GMT
< server: envoy
<
* Connection #0 to host localhost left intact
200
```

We know the Lua code ran because we see the log entry and the my-request-id header set.

Let’s try sending a POST request:

```shell
$ curl -X POST -v localhost:10000
...
> POST / HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/7.64.0
> Accept: */*
>
< HTTP/1.1 200 OK
< content-length: 3
< content-type: text/plain
< date: Mon, 29 Nov 2021 23:49:58 GMT
< server: envoy
Notice the my-request-id header was not included in the headers. Finally, let’s also try sending a GET request but also provide the my-request-id header. In this case, the my-request-id header shouldn’t be included in the response either:

$ curl -v -H "my-request-id: something" localhost:10000
...
> GET / HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/7.64.0
> Accept: */*
> my-request-id: something
>
< HTTP/1.1 200 OK
< content-length: 3
< content-type: text/plain
< date: Mon, 29 Nov 2021 23:51:57 GMT
< server: envoy
<
* Connection #0 to host localhost left intact
```

As a final exercise, we’ll create a separate .lua script that generates a simple random string we can use for request IDs. We’ll load the script and then call it in the response function to get the request ID.

Let’s create a library.lua file with the following contents:

```lua
LIBRARY = {}

function LIBRARY.RandomString()
  local result = ""
  for i = 1, 24 do
    result = result .. string.char(math.random(97, 122))
  end
  return result
end

return LIBRARY
```

We’re declaring a table called LIBRARY and a function called RandomString on it.

Save the above Lua script to a file called library.lua and put it in the same folder where your Envoy process will be running.

Luajit runtime looks for Lua modules in the working directory of the process and in the /usr/local/share/lua/5.1/ folder.

The existing code we have in the Envoy config will largely stay the same. We’ll only need to load the library.lua and invoke the RandomString function.

Here’s the updated Envoy config:

```yaml
static_resources:
  listeners:
  - name: main
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
          route_config:
            name: some_route
            virtual_hosts:
            - name: some_service
              domains:
              - "*"
              routes:
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
          http_filters:
          - name: envoy.filters.http.lua
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
              inline_code: |
                local library = require("library")
                function envoy_on_request(request_handle)
                  local headers = request_handle:headers()
                  local metadata = request_handle:streamInfo():dynamicMetadata()
                  metadata:set("envoy.filters.http.lua", "requestInfo", {
                      requestId = headers:get("my-request-id"),
                      method = headers:get(":method"),
                    })
                end
                function envoy_on_response(response_handle)
                  local requestInfoObj = response_handle:streamInfo():dynamicMetadata():get("envoy.filters.http.lua")["requestInfo"]

                  local requestId = requestInfoObj.requestId
                  local method = requestInfoObj.method
                  if (requestId == nil or requestId == '') and (method == 'GET') then
                    response_handle:logInfo("Adding request ID header")
                    response_handle:headers():add("my-request-id", library.RandomString())
                  end
                end
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```

Save the above YAML to 8-lab-1-lua-script-2.yaml and run it with func-e.

To try it out, let’s send a request to localhost:10000:

```shell
$ curl -v localhost:10000
...
[2021-11-23 23:14:18.206][2526][info][lua] [source/extensions/filters/http/lua/lua_filter.cc:795] script log: Adding request ID header
< HTTP/1.1 200 OK
< content-length: 3
< content-type: text/plain
< my-request-id: usptcritlocbzsezhjmroule
< date: Tue, 23 Nov 2021 23:14:18 GMT
< server: envoy
<
* Connection #0 to host localhost left intact
200
```

The output will include the my-request-id and the random string we generated that was called from the library.lua file.

## Files

### /Users/kumarro/Downloads/8lab1luascript1-221021-125246.yaml

```yaml
static_resources:
  listeners:
  - name: main
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
          route_config:
            name: some_route
            virtual_hosts:
            - name: some_service
              domains:
              - "*"
              routes:
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
          http_filters:
          - name: envoy.filters.http.lua
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
              inline_code: |
                function envoy_on_request(request_handle)
                  local headers = request_handle:headers()
                  local metadata = request_handle:streamInfo():dynamicMetadata()
                  metadata:set("envoy.filters.http.lua", "requestInfo", {
                      requestId = headers:get("my-request-id"),
                      method = headers:get(":method"),
                    })
                end
                function envoy_on_response(response_handle)
                  local requestInfoObj = response_handle:streamInfo():dynamicMetadata():get("envoy.filters.http.lua")["requestInfo"]

                  local requestId = requestInfoObj.requestId
                  local method = requestInfoObj.method
                  if (requestId == nil or requestId == '') and (method == 'GET') then
                    response_handle:logInfo("Adding request ID header")
                    response_handle:headers():add("my-request-id", "some_id_here")
                  end
                end
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```

### /Users/kumarro/Downloads/8lab1luascript2-221021-125246.yaml

```yaml
static_resources:
  listeners:
  - name: main
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
          route_config:
            name: some_route
            virtual_hosts:
            - name: some_service
              domains:
              - "*"
              routes:
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
          http_filters:
          - name: envoy.filters.http.lua
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
              inline_code: |
                local library = require("library")
                function envoy_on_request(request_handle)
                  local headers = request_handle:headers()
                  local metadata = request_handle:streamInfo():dynamicMetadata()
                  metadata:set("envoy.filters.http.lua", "requestInfo", {
                      requestId = headers:get("my-request-id"),
                      method = headers:get(":method"),
                    })
                end
                function envoy_on_response(response_handle)
                  local requestInfoObj = response_handle:streamInfo():dynamicMetadata():get("envoy.filters.http.lua")["requestInfo"]

                  local requestId = requestInfoObj.requestId
                  local method = requestInfoObj.method
                  if (requestId == nil or requestId == '') and (method == 'GET') then
                    response_handle:logInfo("Adding request ID header")
                    response_handle:headers():add("my-request-id", library.RandomString())
                  end
                end
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```


### 8lab1luascript-221021-125246.yaml

```yaml
static_resources:
  listeners:
  - name: main
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
          route_config:
            name: some_route
            virtual_hosts:
            - name: some_service
              domains:
              - "*"
              routes:
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
          http_filters:
          - name: envoy.filters.http.lua
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
              inline_code: |
                function envoy_on_response(response_handle)
                  response_handle:headers():add("hello", "world")
                end
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```

### library-211220-095407.lua

```lua
LIBRARY = {}

function LIBRARY.RandomString()
  local result = ""
  for i = 1, 24 do
    result = result .. string.char(math.random(97, 122))
  end
  return result
end

return LIBRARY
```
