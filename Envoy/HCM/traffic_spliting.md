# Traffic splitting

Envoy supports traffic splitting to different routes within the same virtual host. We can split the traffic between two or more upstream clusters.
There are two different approaches. The first one uses the percentages specified in the runtime object, and the second one uses weighted clusters.

## Traffic splitting using runtime percentages

Using the percentages from the runtime object lends itself well to the canary release or progressive delivery scenarios. In this scenario, we want to gradually shift traffic from one upstream cluster to another.
The way to achieve this is by providing a runtime_fraction configuration. Let’s use an example to explain how traffic splitting using runtime percentages works.

```yaml
route_config:
  virtual_hosts:
  - name: hello_vhost
    domains: ["hello.io"]
    routes:
      - match:
          prefix: "/"
          runtime_fraction:
            default_value:
              numerator: 90
              denominator: HUNDRED
        route:
          cluster: hello_v1
      - match:
          prefix: "/"
        route:
          cluster: hello_v2
```

The above configuration declares two versions of the hello service: hello_v1 and hello_v2.

In the first match, we’re configuring the runtime_fraction field by specifying a numerator (90) and a denominator (HUNDRED). Envoy calculates the final fractional value using the numerator and the denominator. In this case, the final value is 90% (90/100 = 0.9 = 90%).

Envoy generates a random number within the range [0, denominator) (e.g. [0, 100] in our case). If the random number is less than the numerator value, the router matches the route and sends the traffic to the cluster hello_v1 in our case.

If the random number is greater than the numerator, Envoy continues to evaluate the remaining match criteria. Since we have the exact prefix match for the second route, it matches, and Envoy sends the traffic to cluster hello_v2. Once we set the numerator value to 0, no random numbers will be greater than the numerator value. Hence all traffic will go to the second route.

We can also set the denominator value in a runtime key. For example:

```yaml
route_config:
  virtual_hosts:
  - name: hello_vhost
    domains: ["hello.io"]
    routes:
      - match:
          prefix: "/"
          runtime_fraction:
            default_value:
              numerator: 0
              denominator: HUNDRED
            runtime_key: routing.hello_io
        route:
          cluster: hello_v1
      - match:
          prefix: "/"
        route:
          cluster: hello_v2
...
layered_runtime:
  layers:
  - name: static_layer
    static_layer:
      routing.hello_io: 90
```

In this example, we’re specifying a runtime key called routing.hello_io. We can set the value for that key under the layered runtime field in the configuration – this could also be read and updated dynamically either from a file or through a runtime discovery service (RTDS). For simplicity’s sake, we’re setting it directly in the config file.

When Envoy does the matching, it will see that the runtime_key is provided and use that value instead of the numerator value. With the runtime key, we don’t have to hard-code the value in the configuration, and we can have Envoy read it from a separate file or RTDS.

The approach with runtime percentages works well when you have two clusters. Still, it becomes complicated when you want to split traffic to more than two clusters or if you’re running A/B testing or multivariate testing scenarios.

## Traffic splitting using weighted clusters
The weighted clusters approach is ideal when splitting traffic between two or more versions of the service. In this approach, we assign different weights for multiple upstream clusters. Whereas the method with runtime percentages uses numerous routes, we only need a single route for the weighted clusters.

We’ll talk more about upstream clusters in the next module. To explain the traffic splitting with weighted clusters, we can think of an upstream cluster as a collection of endpoints traffic can be sent to.

We specify multiple weighted clusters (weighted_clusters) within the route instead of setting a single cluster (cluster).

Continuing with the previous example, this is how we could re-write the configuration to use weighted clusters instead:

```yaml
route_config:
  virtual_hosts:
  - name: hello_vhost
    domains: ["hello.io"]
    routes:
      - match:
          prefix: "/"
        route:
          weighted_clusters:
            clusters:
              - name: hello_v1
                weight: 90
              - name: hello_v2
                weight: 10
```

Under the weighted clusters, we could also set the runtime_key_prefix that will read the weights from the runtime key configuration. Envoy uses the weights next to each cluster if the runtime key configuration is not there.

```yaml
route_config:
  virtual_hosts:
  - name: hello_vhost
    domains: ["hello.io"]
    routes:
      - match:
          prefix: "/"
        route:
          weighted_clusters:
            runtime_key_prefix: routing.hello_io
            clusters:
              - name: hello_v1
                weight: 90
              - name: hello_v2
                weight: 10
...
layered_runtime:
  layers:
  - name: static_layer
    static_layer:
      routing.hello_io.hello_v1: 90
      routing.hello_io.hello_v2: 10
```

The weight represents the percentage of the traffic Envoy sends to the upstream cluster. The sum of all weights has to be 100. However, using the total_weight field, we can control the value the sum of all weights has to equal to. For example, the following snippet sets the total_weight to 15:

```yaml
route_config:
  virtual_hosts:
  - name: hello_vhost
    domains: ["hello.io"]
    routes:
      - match:
          prefix: "/"
        route:
          weighted_clusters:
            runtime_key_prefix: routing.hello_io
            total_weight: 15
            clusters:
              - name: hello_v1
                weight: 5
              - name: hello_v2
                weight: 5
              - name: hello_v3
                weight: 5
```

To dynamically control the weights, we can set the runtime_key_prefix. The router uses the runtime key prefix value to construct the runtime keys associated with each cluster. If we provide the runtime key prefix, the router will check the runtime_key_prefix + "." + cluster_name value, where cluster_name denotes the entry in the clusters array (e.g. hello_v1, hello_v2). If Envoy doesn’t find the runtime key, it will use the value specified in the configuration as the default value.