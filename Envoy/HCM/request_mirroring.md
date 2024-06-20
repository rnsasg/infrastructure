# Request Mirroring
Using a request mirroring policy (request_mirroring_policies) on the route level, we can configure Envoy to shadow traffic from one cluster to another.

Traffic shadowing or request mirroring is when incoming requests destined for one cluster are duplicated and sent to a second cluster. The mirrored requests are “fire and forget”, meaning that Envoy doesn’t wait for the shadow cluster to respond before sending the response from the primary cluster.

The request mirroring pattern doesn’t impact the traffic sent to the primary cluster, and because Envoy will collect all statistics for the shadow cluster, it’s a helpful technique for testing.

In addition to the “fire and forget”, make sure that the requests you’re mirroring are idempotent. Otherwise, mirrored requests can mess up the backends your services talk to.

The authority/host headers on the shadowed request will have the -shadow string appended.

To configure the mirroring policy, we use the request_mirror_policies field on the route we want to mirror the traffic. We can specify one or more mirroring policies and the fraction of traffic we want to mirror:

```yaml
  route_config:
    name: my_route
    virtual_hosts:
    - name: httpbin
      domains: ["*"]
      routes:
      - match:
          prefix: /
        route:
          cluster: httpbin
          request_mirror_policies:
            cluster: mirror_httpbin
            runtime_fraction:
              default_value:
                numerator: 100
      ...
```

The above configuration will take 100% of incoming requests sent to the cluster httpbin and mirror them to mirror_httpbin.

