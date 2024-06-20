# HTTP routing
The previously mentioned router filter (envoy.filters.http.router) is the one that implements HTTP forwarding. The router filter is used in almost all HTTP proxy scenarios. The main job of the router filter is to look at the routing table and route (forward and redirect) the requests accordingly.

The router uses the information from the incoming request (e.g., host or authority headers) and matches it to an upstream cluster through virtual hosts and routing rules.

All configured HTTP filters use the route configuration (route_config) that contains the routing table. Even though the primary consumer of the routing table will be the router filter, other filters have access to it if they want to make any decisions based on the destination of the request.

A set of virtual hosts makes up the route configuration. Each virtual host has a logical name, a set of domains that can get routed to it based on the request header, and a set of routes that specify how to match a request and indicate what to do next.

Envoy also supports priority routing at the route level. Each priority has its connection pool and circuit-breaking settings. The two currently supported priorities are DEFAULT and HIGH. The priority defaults to DEFAULT if we don’t explicitly provide it.

Here’s a snippet that shows an example of a route configuration:

```yaml
route_config:
  name: my_route_config # Name used for stats, not relevant for routing
  virtual_hosts:
  - name: bar_vhost
    domains: ["bar.io"]
    routes:
      - match:
          prefix: "/"
        route:
          priority: HIGH
          cluster: bar_io
  - name: foo_vhost
    domains: ["foo.io"]
    routes:
      - match:
          prefix: "/"
        route:
          cluster: foo_io
      - match:
          prefix: "/api"
        route:
          cluster: foo_io_api
```

When an HTTP request comes in, the virtual host, domain, and route matching happen in order:

The host or authority header gets matched to a value specified in each virtual host’s domains field. For example, virtual host foo_vhost matches if the host header is set to foo.io.

The entries under routes within the matched virtual host are checked next. No further checks are made if a match is found and a cluster is selected. For example, if we matched the foo.io virtual host, and the request prefix is /api, the cluster foo_io_api is selected.

If provided, each virtual cluster (virtual_clusters) in a virtual host is checked for a match. A virtual cluster is used if there’s a match and no further virtual cluster checks are made.

Virtual cluster is a way of specifying regex matching rules against specific endpoints and explicitly generating stats for the matched request.

The order of virtual hosts, as well as routes within each host, matters. Consider the following route configuration:

```yaml
route_config:
  virtual_hosts:
  - name: hello_vhost
    domains: ["hello.io"]
    routes:
      - match:
          prefix: "/api"
        route:
          cluster: hello_io_api
      - match:
          prefix: "/api/v1"
        route:
          cluster: hello_io_api_v1
```

Which route/cluster is selected if we send the following request?

```shell
curl hello.io/api/v1
```

The first route that sets the cluster hello_io_api is matched. That’s because matches are evaluated in order by the prefix. However, we might have wrongly expected the route with the prefix /api/v1 to be matched. To work around this issue, we could swap the order of routes or use a different matching rule.