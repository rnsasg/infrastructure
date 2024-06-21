# Extending Envoy Using Wasm-and-go

In this lab, we’ll use [TinyGo](https://tinygo.org/),[proxy-wasm-go-sdk](https://github.com/tetratelabs/proxy-wasm-go-sdk), and [func-e CLI](https://func-e.io/) to build and test an Envoy Wasm extension.

We’ll write a simple Wasm module that adds a header to response headers. Later, we’ll show how to read configuration and add custom metrics. We’ll use Golang and compile it with the TinyGo compiler.

Installing TinyGo
Let’s download and install TinyGo:

```shell
wget https://github.com/tinygo-org/tinygo/releases/download/v0.21.0/tinygo_0.21.0_amd64.deb
sudo dpkg -i tinygo_0.21.0_amd64.deb
```

You can run tinygo version to check thatthe installation is successful:

```shell
$ tinygo version
tinygo version 0.21.0 linux/amd64 (using go version go1.17.2 and LLVM version 11.0.0)
```

Scaffolding the Wasm module
We’ll start by creating a new folder for our extension, initializing the Go module, and downloading the SDK dependency:

```shell
$ mkdir header-filter && cd header-filter
$ go mod init header-filter
$ go mod edit -require=github.com/tetratelabs/proxy-wasm-go-sdk@main
$ go mod download github.com/tetratelabs/proxy-wasm-go-sdk
```

Next, let’s create the main.go file where the code for our WASM extension will live:

```go
package main

import (
  "github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
  "github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
  proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
  // Embed the default VM context here,
  // so that we don't need to reimplement all the methods.
  types.DefaultVMContext
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
  return &pluginContext{}
}

type pluginContext struct {
  // Embed the default plugin context here,
  // so that we don't need to reimplement all the methods.
  types.DefaultPluginContext
}

// Override types.DefaultPluginContext.
func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
  return &httpHeaders{contextID: contextID}
}

type httpHeaders struct {
  // Embed the default http context here,
  // so that we don't need to reimplement all the methods.
  types.DefaultHttpContext
  contextID uint32
}

func (ctx *httpHeaders) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
  proxywasm.LogInfo("OnHttpRequestHeaders")
  return types.ActionContinue
}

func (ctx *httpHeaders) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
  proxywasm.LogInfo("OnHttpResponseHeaders")
  return types.ActionContinue
}

func (ctx *httpHeaders) OnHttpStreamDone() {
  proxywasm.LogInfof("%d finished", ctx.contextID)
}
```

Save the above contents to a file called main.go.

Let’s build the filter to check that everything is good:

tinygo build -o main.wasm -scheduler=none -target=wasi main.go
The build command should run successfully and generate a file called main.wasm.

We’ll use func-e to run a local Envoy instance to test our built extension.

First, we need an Envoy config that will configure the extension:

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
            - name: envoy.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                stat_prefix: ingress_http
                codec_type: auto
                route_config:
                  name: local_route
                  virtual_hosts:
                    - name: local_service
                      domains:
                        - "*"
                      routes:
                        - match:
                            prefix: "/"
                          direct_response:
                            status: 200
                            body:
                              inline_string: "hello world\n"
                http_filters:
                  - name: envoy.filters.http.wasm
                    typed_config:
                      "@type": type.googleapis.com/udpa.type.v1.TypedStruct
                      type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
                      value:
                        config:
                          vm_config:
                            runtime: "envoy.wasm.runtime.v8"
                            code:
                              local:
                                filename: "main.wasm"
                  - name: envoy.filters.http.router
                    typed_config: 
                      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
```

Save the above to 8-lab-2-wasm-config.yaml file.

The Envoy configuration sets up a single listener on port 10000 that returns a direct response (HTTP 200) with body hello world. Inside the http_filters section, we’re configuring the envoy.filters.http.wasm filter and referencing the local WASM file (main.wasm) we built earlier.

Let’s run Envoy with this configuration in the background:

func-e run -c 8-lab-2-wasm-config.yaml &
The Envoy instance should start without any issues. Once it’s started, we can send a request to the port Envoy is listening on (10000):

```shell
$ curl localhost:10000
[2021-11-04 22:41:19.982][91521][info][wasm] [source/extensions/common/wasm/context.cc:1167] wasm log: OnHttpRequestHeaders
[2021-11-04 22:41:19.982][91521][info][wasm] [source/extensions/common/wasm/context.cc:1167] wasm log: OnHttpResponseHeaders
[2021-11-04 22:41:19.983][91521][info][wasm] [source/extensions/common/wasm/context.cc:1167] wasm log: 2 finished
hello world
```

The output shows the two log entries: one from the OnHttpRequestHeaders handler and the second one from the OnHttpResponseHeaders handler. The last line is the example response returned by the direct response configuration in the filter.

You can stop the proxy by bringing the process to the foreground with fg and pressing CTRL+C to stop it.

## Setting additional headers on HTTP response
Let’s open the main.go file and add a header to the response headers. We’ll be updating the OnHttpResponseHeaders function to do that.

We’ll call the AddHttpResponseHeader function to add a new header. Update the OnHttpResponseHeaders function to look like this:

```go
func (ctx *httpHeaders) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
  proxywasm.LogInfo("OnHttpResponseHeaders")
  err := proxywasm.AddHttpResponseHeader("my-new-header", "some-value-here")
  if err != nil {
    proxywasm.LogCriticalf("failed to add response header: %v", err)
  }
  return types.ActionContinue
}
```

Let’s rebuild the extension:

tinygo build -o main.wasm -scheduler=none -target=wasi main.go
And we can now re-run the Envoy proxy with the updated extension:

func-e run -c 8-lab-2-wasm-config.yaml &
Now, if we send a request again (make sure to add the -v flag), we’ll see the header that got added to the response:

```shell
$ curl -v localhost:10000
...
< HTTP/1.1 200 OK
< content-length: 13
< content-type: text/plain
< my-new-header: some-value-here
< date: Mon, 22 Jun 2021 17:02:31 GMT
< server: envoy
<
hello world
```

Reading values from configuration
Hardcoding values like this in code is never a good idea. Let’s see how we can read the additional headers.

Add the additionalHeaders and contextID to the pluginContext struct:
```go
type pluginContext struct {
  // Embed the default plugin context here,
  // so that we don't need to reimplement all the methods.
  types.DefaultPluginContext
  additionalHeaders map[string]string
  contextID         uint32
}
Update the NewPluginContext function to initialize the values:
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
  return &pluginContext{contextID: contextID, additionalHeaders: map[string]string{}}
}
In the OnPluginStart function, we can now read in values from the Envoy configuration and store the key/value pairs in the additionalHeaders map:
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
  // Get the plugin configuration
  config, err := proxywasm.GetPluginConfiguration()
  if err != nil && err != types.ErrorStatusNotFound {
    proxywasm.LogCriticalf("failed to load config: %v", err)
    return types.OnPluginStartStatusFailed
  }

  // Read the config
  scanner := bufio.NewScanner(bytes.NewReader(config))
  for scanner.Scan() {
    line := scanner.Text()
    if strings.HasPrefix(line, "#") {
      continue
    }
    // Each line in the config is in the "key=value" format
    if tokens := strings.Split(scanner.Text(), "="); len(tokens) == 2 {
      ctx.additionalHeaders[tokens[0]] = tokens[1]
    }
  }
  return types.OnPluginStartStatusOK
}
To access the configuration values we’ve set, we need to add the map to the HTTP context when we initialize it. To do that, we need to update the httpheaders struct first:

type httpHeaders struct {
  // Embed the default http context here,
  // so that we don't need to reimplement all the methods.
  types.DefaultHttpContext
  contextID         uint32
  additionalHeaders map[string]string
}
Then, in the NewHttpContext function, we can instantiate the httpHeaders with the additional headers map coming from the plugin context:

func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
  return &httpHeaders{contextID: contextID, additionalHeaders: ctx.additionalHeaders}
}
Finally, to set the headers we modify the OnHttpResponseHeaders function, iterate through the additionalHeaders map, and call the AddHttpResponseHeader for each item:

func (ctx *httpHeaders) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
  proxywasm.LogInfo("OnHttpResponseHeaders")

  for key, value := range ctx.additionalHeaders {
    if err := proxywasm.AddHttpResponseHeader(key, value); err != nil {
        proxywasm.LogCriticalf("failed to add header: %v", err)
        return types.ActionPause
    }
    proxywasm.LogInfof("header set: %s=%s", key, value)
  }
    
  return types.ActionContinue
}
```

Let’s rebuild the extension again:

tinygo build -o main.wasm -scheduler=none -target=wasi main.go
Also, let’s update the config file to include additional headers in the filter configuration (the configuration field):

```yaml
- name: envoy.filters.http.wasm
  typed_config:
    "@type": type.googleapis.com/udpa.type.v1.TypedStruct
    type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
    value:
      config:
        vm_config:
          runtime: "envoy.wasm.runtime.v8"
          code:
            local:
              filename: "main.wasm"
        # ADD THESE LINES
        configuration:
          "@type": type.googleapis.com/google.protobuf.StringValue
          value: |
            header_1=somevalue
            header_2=secondvalue
```
With the filter updated, we can re-run the proxy. When you send a request, you’ll notice the headers we set in the filter configuration are added as response headers:

```shell
$ curl -v localhost:10000
...
< HTTP/1.1 200 OK
< content-length: 13
< content-type: text/plain
< header_1: somevalue
< header_2: secondvalue
< date: Mon, 22 Jun 2021 17:54:53 GMT
< server: envoy
...
```

Add a metric
Let’s add another feature — a counter that increases each time there’s a request header called hello set.

First, let’s update the pluginContext to include the helloHeaderCounter:

```go
type pluginContext struct {
  // Embed the default plugin context here,
  // so that we don't need to reimplement all the methods.
  types.DefaultPluginContext
  additionalHeaders  map[string]string
  contextID          uint32
  // ADD THIS LINE
  helloHeaderCounter proxywasm.MetricCounter 
}
With the metric counter in the struct, we can now create it in the NewPluginContext function. We’ll call the header hello_header_counter.

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
  return &pluginContext{contextID: contextID, additionalHeaders: map[string]string{}, helloHeaderCounter: proxywasm.DefineCounterMetric("hello_header_counter")}
}
Since we want to check the incoming request headers to decide whether to increment the counter, we need to add the helloHeaderCounter to the httpHeaders struct as well:

type httpHeaders struct {
  // Embed the default http context here,
  // so that we don't need to reimplement all the methods.
  types.DefaultHttpContext
  contextID          uint32
  additionalHeaders  map[string]string
  // ADD THIS LINE
  helloHeaderCounter proxywasm.MetricCounter
}
Also, we need to get the counter from the pluginContext and set it when we’re creating the new HTTP context:

// Override types.DefaultPluginContext.
func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
  return &httpHeaders{contextID: contextID, additionalHeaders: ctx.additionalHeaders, helloHeaderCounter: ctx.helloHeaderCounter}
}
Now that we’ve piped the helloHeaderCounter all the way through to the httpHeaders, we can use it in the OnHttpRequestHeaders function:

func (ctx *httpHeaders) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
  proxywasm.LogInfo("OnHttpRequestHeaders")

  _, err := proxywasm.GetHttpRequestHeader("hello")
  if err != nil {
    // Ignore if header is not set
    return types.ActionContinue
  }

  ctx.helloHeaderCounter.Increment(1)
  proxywasm.LogInfo("hello_header_counter incremented")
  return types.ActionContinue
}
```

Here, we’re checking whether the “hello” request header is defined (note that we don’t care about the header value), and if it’s defined, we call the Increment function on the counter instance. Otherwise, we’ll ignore it and return ActionContinue if we get an error from the GetHttpRequestHeader call.

Let’s rebuild the extension again:

tinygo build -o main.wasm -scheduler=none -target=wasi main.go
And then re-run the Envoy proxy. Make a couple of requests like this:

curl -H "hello: something" localhost:10000
You’ll notice the log Envoy log entry like this one:

"`text wasm log: hello_header_counter incremented


You can also use the admin address on port 9901 to check that the metric is being tracked:

```sh
$ curl localhost:9901/stats/prometheus | grep hello
# TYPE envoy_hello_header_counter counter
envoy_hello_header_counter{} 1
```
## Files

### main.go

```go
package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	// Embed the default VM context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultVMContext
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
}

// Override types.DefaultPluginContext.
func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpHeaders{contextID: contextID}
}

type httpHeaders struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	contextID uint32
}

func (ctx *httpHeaders) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.LogInfo("OnHttpRequestHeaders")
	return types.ActionContinue
}

func (ctx *httpHeaders) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.LogInfo("OnHttpResponseHeaders")
	return types.ActionContinue
}

func (ctx *httpHeaders) OnHttpStreamDone() {
	proxywasm.LogInfof("%d finished", ctx.contextID)
}
```

### 8lab2wasmconfig-221021-125356.yaml

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
            - name: envoy.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                stat_prefix: ingress_http
                codec_type: auto
                route_config:
                  name: local_route
                  virtual_hosts:
                    - name: local_service
                      domains:
                        - "*"
                      routes:
                        - match:
                            prefix: "/"
                          direct_response:
                            status: 200
                            body:
                              inline_string: "hello world\n"
                http_filters:
                  - name: envoy.filters.http.wasm
                    typed_config:
                      "@type": type.googleapis.com/udpa.type.v1.TypedStruct
                      type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
                      value:
                        config:
                          vm_config:
                            runtime: "envoy.wasm.runtime.v8"
                            code:
                              local:
                                filename: "main.wasm"
                  - name: envoy.filters.http.router
                    typed_config: 
                      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
```