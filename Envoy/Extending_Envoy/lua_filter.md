# Lua Filter
Envoy features a built-in HTTP Lua filter that allows running Lua scripts during request and response flows. Lua is an embeddable scripting language, popular primarily within embedded systems and games. Envoy uses LuaJIT (Just-In-Time compiler for Lua) as the runtime. The highest Lua script version supported by the [LuaJIT](https://luajit.org/) is 5.1, with some features from 5.2.

At runtime, Envoy creates a Lua environment for each worker thread. Because of this, there is no genuinely global data. Any globals created and populated at load time are visible from each worker thread in isolation.

Lua scripts are run as coroutines in a synchronous style, even though they may perform complex asynchronous tasks. This makes it easier to write. Envoy performs all network/async processing via a set of APIs. When an async task is invoked, Envoy suspends the execution of the script and then resumes once the async operation completes.

We shouldn’t be performing any blocking operations from scripts, as that would impact Envoys’ performance. We should use only Envoy APIs for all IO operations.

We can modify and/or inspect request and response headers, body, and trailers using a Lua script. We can also make outbound async HTTP calls to an upstream host or perform a direct response and skip any further filter iteration. For example, within the Lua script, we can make an upstream HTTP call and directly respond without continuing the execution of other filters.

## How to configure the Lua filter
Lua scripts can be defined inline using the inline_code field or by referencing a local file using the source_codes field on the filter:

```yaml
name: envoy.filters.http.lua
typed_config:
  "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
  inline_code: |
    -- Called on the request path.
    function envoy_on_request(request_handle)
      -- Do something.
    end
    -- Called on the response path.
    function envoy_on_response(response_handle)
      -- Do something.
    end
  source_codes:
    myscript.lua:
      filename: /scripts/myscript.lua
```

Envoy treats the above script as a global script, and it executes it for every HTTP request. Two global functions can be defined in each script:

```shell
function envoy_on_request(request_handle)
end
and

function envoy_on_response(response_handle)
end
```

The envoy_on_request function is called on the request path, and the envoy_on_response script is called on the response path. Each function receives a handle that has different methods defined. The script can contain either the response or request function, or both.

We also have an option for disabling or overwriting the scripts on a per-route basis on the virtual host, route, or weighted cluster level.

Disabling or referring to an existing Lua script on host, route, or weighted cluster level is done using the typed_per_filter_config field. For example, here’s how refer to an existing script (e.g. some-script.lua) using the typed_per_filter_config:

```yaml
typed_per_filter_config:
  envoy.filters.http.lua:
    "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.LuaPerRoute
    name: some-script.lua
Similarly, instead of specifying the name field, we could define the source_code and the inline_string field like this:

typed_per_filter_config:
  envoy.filters.http.lua:
    "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.LuaPerRoute
    source_code:
      inline_string: |
        function envoy_on_response(response_handle)
          -- Do something on response.
        end
```

## Stream handle API
We mentioned earlier that the request_handle and response_handle stream handles get passed to the global request and response functions.

The methods available on the stream handle include methods such as headers, body, metadata, various log methods (e.g. logTrace, logInfo, logDebug, …), httpCall, connection, and more. You can find the full list of methods in the [Lua filter source code](https://github.com/envoyproxy/envoy/blob/d79a3ab49f1aa522d0a465385425e3e00c8db147/source/extensions/filters/http/lua/lua_filter.h#L151).

In addition to the stream object, the API supports the following objects:

* [Header object](https://github.com/envoyproxy/envoy/blob/55fc06b43082064cf7551d8dbc08a0e30e2c2f40/source/extensions/filters/http/lua/wrappers.h#L46) (returned by the headers() method)
Buffer object (returned by the body() method)
* [Dynamic metadata object](https://github.com/envoyproxy/envoy/blob/55fc06b43082064cf7551d8dbc08a0e30e2c2f40/source/extensions/filters/http/lua/wrappers.h#L151) (returned by the metadata() method)
* [Stream info object](https://github.com/envoyproxy/envoy/blob/55fc06b43082064cf7551d8dbc08a0e30e2c2f40/source/extensions/filters/http/lua/wrappers.h#L199) (returned by the streamInfo() method)
Connection object (returned bt the connection() method)
* [SSL connection info object](https://github.com/envoyproxy/envoy/blob/0fae6970ddaf93f024908ba304bbd2b34e997a51/source/extensions/filters/common/lua/wrappers.h#L124) (returned by the ssl() method on the connection object)
We’ll see how to use some of the objects and methods in the Lua lab.

