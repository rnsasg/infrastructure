# Sidecar Resource
Sidecar resource describes the configuration of sidecar proxies. By default, all proxies in the mesh have the configuration required to reach every workload in the mesh and accept traffic on all ports.

In addition to configuring the set of ports/protocols proxy accepts when forwarding the traffic, we can restrict the collection of services the proxy can reach when forwarding outbound traffic.

Note that this restriction here is in the configuration only. It’s not a security boundary. You can still reach the services, but Istio will not propagate the configuration for that service to the proxy.

Below is an example of a sidecar proxy resource in the foo namespace that configures all workloads in that namespace to only see the workloads in the same namespace and workloads in the istio-system namespace.

```yaml
apiVersion:
 networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: default
  namespace: foo
spec:
  egress:
    - hosts:
      - "./*"
      - "istio-system/*"
```

We can deploy a sidecar resource to one or more namespaces inside the Kubernetes cluster. Still, there can only be one sidecar resource per namespace if there’s not workload selector defined.

Three parts make up the sidecar resource, a workload selector, an ingress listener, and an egress listener.

## Workload Selector
The workload selector determines which workloads are affected by the sidecar configuration. You can decide to control all sidecars in a namespace, regardless of the workload, or provide a workload selector to apply the configuration only to specific workloads.

For example, this YAML applies to all proxies inside the default namespace, because there’s no selector defined:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: default-sidecar
  namespace: default
spec:
  egress:
  - hosts:
    - "default/*"
    - "istio-system/*"
    - "staging/*"
```

The egress section specifies that the proxies can access services running in default, istio-system, and staging namespaces. To apply the resource only to specific workloads, we can use the workloadSelector field. For example, setting the selector to version: v1 will only apply to the workloads with that label set:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: default-sidecar
  namespace: default
spec:
  workloadSelector:
    labels:
      version: v1
  egress:
  - hosts:
    - "default/*"
    - "istio-system/*"
    - "staging/*"
```

## Ingress and Egress Listener
The ingress listener section of the resource defines which inbound traffic is accepted. Similarly, with the egress listener, you can specify the properties for outbound traffic.

Each ingress listener needs a port set where the traffic will be received (for example, 3000 in the example below) and a default endpoint. The default endpoint can either be a loopback IP endpoint or a Unix domain socket. The endpoint configures where Envoy forwards the traffic.

```yaml
...
  ingress:
  - port:
      number: 3000
      protocol: HTTP
      name: somename
    defaultEndpoint: 127.0.0.1:8080
...
```

The above snippet configures the ingress listener to listen on the port 3000 and forward traffic to the loopback IP on the port 8080 where your service is listening to. Additionally, we could set the bind field to specify an IP address or domain socket where we want the proxy to listen for the incoming traffic. Finally, we can use the field captureMode to configure how and if traffic even gets captured.

The egress listener has similar fields, with the addition of the hosts field. With the hosts field, you can specify the service hosts with namespace/dnsName format. For example, myservice.default or default/*. Services specified in the hosts field can be services from the mesh registry, external services (defined with ServiceEntry), or virtual services.

```yaml
  egress:
  - port:
      number: 8080
      protocol: HTTP
    hosts:
    - "staging/*"
```

With the YAML above, the sidecar proxies the traffic that’s bound for port 8080 for services running in the staging namespace.