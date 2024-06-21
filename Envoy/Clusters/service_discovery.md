# Service discovery

Clusters can be configured statically in the configuration file or dynamically through the cluster discovery service (CDS) API. Each cluster is a collection of endpoints that Envoy needs to resolve to send the traffic to.

The process of resolving the endpoints is known as service discovery.

# What are endpoints?
A cluster is a collection of endpoints that identify a specific host. Each endpoint has the following properties:

## Address (address)

The address represents the upstream host address. The form of the address depends on the cluster type. The address is expected to be an IP for STATIC or EDS cluster types, and for LOGICAL or STRICT DNS cluster types, the address is expected to be a hostname resolved via DNS.

## Hostname (hostname)

A hostname is associated with the endpoint. Note that the hostname is not used for routing or resolving addresses. It’s associated with the endpoint and can be used for any features requiring a hostname, such as auto host rewrite.

## Health check configuration (health_check_config)

The optional health check configuration is used for the health checker to contact the health-checked host. The configuration contains the hostname and the port where the host can be contacted to perform the health check. Note that this configuration is applicable only for upstream clusters with active health checking enabled.

# Service discovery types
There are five supported service discovery types – let’s look at them more closely.

## Static (STATIC)
The static service discovery type is the simplest. In the configuration, we specify a resolved network name for each host in the cluster. For example:

```yaml
  clusters:
  - name: my_cluster_name
    type: STATIC
    load_assignment:
      cluster_name: my_service_name
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8080
```

Note that if we don’t provide the type, it defaults to STATIC.

## Strict DNS (STRICT_DNS)
With strict DNS, Envoy continuously and asynchronously resolves the DNS endpoints defined in the cluster. If the DNS query returns multiple IP addresses, Envoy assumes they are part of the cluster and load balances between them. Similarly, if the DNS query returns zero hosts, Envoy assumes the cluster doesn’t have any.

A note on health checking – the health checking is not shared if multiple DNS names resolve to the same IP address. This might cause an unnecessary load on the upstream host because Envoy performs the health check on the same IP address multiple times (across different DNS names).

When respect_dns_ttl field is enabled, we can control the continuous resolution of DNS names using the dns_refresh_rate. If not specified, the DNS refresh rate defaults to 5000 ms. Another setting (dns_failure_refresh_rate) controls the refresh frequency during failures. If not provided, Envoy uses the dns_refresh_rate.

Here’s an example of the STRICT_DNS service discovery type:

```yaml
  clusters:
  - name: my_cluster_name
    type: STRICT_DNS
    load_assignment:
      cluster_name: my_service_name
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: my-service
                port_value: 8080
```

## Logical DNS (LOGICAL_DNS)
The logical DNS service discovery is similar to the strict DNS and uses the asynchronous resolution mechanism. However, it only uses the first IP address that’s returned when a new connection needs to be initiated.

Therefore, a single logical connection pool may contain physical connections to various upstream hosts. These connections never drain, even when the DNS resolution returns zero hosts.

What is a connection pool? Each endpoint in the cluster will have one or more connection pools. For example, depending on the supported upstream protocols, there might be one connection pool per protocol allocated. Each worker thread in Envoy also maintains its connection pool for each cluster. For example, if Envoy has two threads and a cluster that supports both HTTP/1 and HTTP/2, there will be at least four connection pools. The way connection pools are is based on the underlying wire protocol. With HTTP/1.1, the connection pool acquires connections to the endpoint as needed (up to the circuit breaking limit). Requests are bound to connections as they become available. When using HTTP/2, the connection pool multiplexes multiple requests over a single connection, up to the limits specified by the max_concurrent_streams and max_requests_per_connections. The HTTP/2 connection pool establishes as many connections as needed to serve the requests.

A typical use case for logical DNS is for large-scale web services. Typically using round-robin DNS returns a different result of multiple IP addresses on each query. If we used strict DNS resolution, Envoy would assume cluster endpoints change on each resolution internally and drain the connection pools. With logical DNS, the connections will stay alive until they get cycled.

Like the strict DNS, the logical DNS also uses the respect_dns_ttl and the dns_refresh_rate fields to configure the DNS refresh rate.

```yaml
  clusters:
  - name: my_cluster_name
    type: LOGICAL_DNS
    load_assignment:
      cluster_name: my_service_name
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: my-service
                port_value: 8080
```

Endpoints discovery service (EDS)
Envoy can use the endpoint discovery services to fetch the cluster endpoints. Typically, this is the preferred service discovery mechanism. Envoy gets explicit knowledge of each upstream host (i.e., no need to route through a DNS-resolved load balancer). Each endpoint can carry extra attributes that can inform Envoy of the load balancing weight, canary status, zone, etc.

```yaml
  clusters:
  - name: my_cluster_name
    type: EDS
    eds_cluster_config:
      eds_config:
        ...
```

In more detail, we explain the dynamic configuration in the Dynamic configuration and xDS chapter.

Original destination (ORIGINAL_DST)
We use the original destination cluster type when connections to Envoy go through iptables REDIRECT or TPROXY target or with the Proxy Protocol.

In this scenario, the requests are forwarded to upstream hosts as addressed by the redirection metadata (for example, using the x-envoy-original-dst-host header) without any configuration or upstream host discovery.

The connections to upstream hosts are pooled and flushed when they’ve been idle longer than specified in the cleanup_interval field (defaults to 5000 ms).

```yaml
clusters:
  - name: original_dst_cluster
    type: ORIGINAL_DST
    lb_policy: ORIGINAL_DST_LB
```

The only load balancing policy that the ORIGINAL_DST cluster type can use is the ORIGINAL_DST_LB policy.

In addition to the above service discovery mechanisms, Envoy also supports custom cluster discovery mechanisms. We can configure a custom discovery mechanism using the cluster_type field.