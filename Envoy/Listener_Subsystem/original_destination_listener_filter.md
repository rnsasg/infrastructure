# Original Destination Listener Filter
The [Original destination filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/original_dst_filter) (envoy.filters.listener.original_dst) reads the SO_ORIGINAL_DST socket option. This option is set when a connection has been redirected by an iptables REDIRECT or TPROXY target (if transparent option is set). The filter can be used in connection with a cluster with the ORIGINAL_DST type.

When using the ORIGINAL_DST cluster type, the requests get forwarded to upstream hosts as addressed by the redirection metadata without making any host discovery. Therefore defining any endpoints in the cluster doesn’t make sense because the endpoint is taken from the original packet and isn’t selected by a load balancer.

We can use Envoy as a generic proxy that forwards all requests to the original destination using this cluster type.

To use the ORIGINAL_DST cluster, the traffic needs to reach Envoy through an iptables REDIRECT or TPROXY target.

```yaml
...
listener_filters:
- name: envoy.filters.listener.original_dst
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.filters.listener.original_dst.v3.OriginalDst
...
clusters:
  - name: original_dst_cluster
    connect_timeout: 5s
    type: ORIGNAL_DST
    lb_policy: CLUSTER_PROVIDED
```

