# Wasm Plugins
WebAssembly or Wasm is a portable binary format for executable code that relies on an open standard. It allows developers to write code in their preferred programming language and then compile it to WebAssembly.

Wasm is an ideal candidate for a plugin mechanism because of its binary portability and the fact that it is programming-language agnostic.

Wasm plugins are isolated from the host environment and executed in a memory-safe sandbox environment called a virtual machine. They communicate with the host environment through a well-defined API.

[Proxy-Wasm ](https://github.com/proxy-wasm/spec)is a specification for an Application Binary Interface (API) tailored for extending proxies with WebAssembly. Istio’s sidecar proxy, Envoy, is one of Proxy-Wasm’s target host environments.

In a previous lesson, we discussed the EnvoyFilter resource, and used a custom Lua script to inject an HTTP header into the response. In the example, we used the Lua Envoy filter ( type.googleapis.com/envoy.config.filter.http.lua.v2.Lua) that allowed us to specify a Lua script. The custom code we specified in the Lua filter configuration was then called and executed by the Envoy proxy.

Using a Lua filter is a good approach to take if the functionality we’re trying to implement is not too complex.

Another option for extending functionality is Wasm filters. In this option we write the custom functionality in a language such as Go or Rust, compile it into a Wasm plugin and have Envoy load it dynamically at runtime.

For Istio to load the Wasm plugin, we need to specify where to find it in the configuration. We could use the EnvoyFilter resource to do that, however, in Istio 1.12 a new resource called WasmPlugin was added. Moreover, the Istio agent was enhanced to support downloading Wasm plugins from an OCI-compliant registry.

These changes allow us to push the compiled Wasm plugin to an OCI-compliant registry, just like if we’d push a Docker image. Then, we can refer to the Wasm plugin in the registry using the WasmPlugin resource.

## WasmPlugin resource
The WasmPlugin resource allows us to configure Wasm plugins for Envoy proxies running in the Istio service mesh. Here is an example WasmPlugin resource:

```yaml
apiVersion: extensions.istio.io/v1alpha1
kind: WasmPlugin
metadata:
  name: hello-world-wasm
  namespace: default
spec:
  selector:
    labels:
      app: hello-world
  url: oci://my-registry/tetrate/hello-world:v1
  pluginConfig:
    greeting: hello
    something: anything
  vmConfig:
    - name: TRUST_DOMAIN
      value: "cluster.local"
```

There are a couple of key fields in the WasmPlugin resource:

1. selector

The labels are used to select which workloads the Wasm plugin will be applied to.

2. url

The URL of the Wasm plugin. The valid schemes are oci:// (default), file:// and http[s]://. When using the oci:// we can also optionally provide image pull policy (imagePullPolicy) and image pull secret ( imagePullSecret).

3. pluginConfig

The configuration of the Wasm plugin. The configuration can be read from within the plugin code via a [call](https://github.com/tetratelabs/proxy-wasm-go-sdk/blob/main/proxywasm/hostcall.go#L34) exposed by the [Proxy-Wasm](https://github.com/proxy-wasm/spec) ABI specification.

4. vmConfig

The configuration of the VM that will be used to run the Wasm plugin. It allows us to inject environment variables into the VM that will be used by the Wasm plugin.

Other resource settings include:

* priority: allows us to determine the ordering of WasmPlugins, and
* phase: determines where in the filter chain to inject the WasmPlugin.

> To get a better understanding of Envoy, we offer a [free Envoy Fundamentals course](https://academy.tetrate.io/courses/envoy-fundamentals).

