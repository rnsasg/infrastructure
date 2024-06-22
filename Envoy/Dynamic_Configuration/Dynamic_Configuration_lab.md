# Lab 1: Dynamic Configuration from filesystem
We’ll create a dynamic Envoy configuration in this lab and configure listeners, clusters, routes, and endpoints through individual configuration files.

Let’s start with the minimal Envoy configuration:

```yaml
node:
  cluster: cluster-1
  id: envoy-instance-1
dynamic_resources:
  lds_config:
    path: ./lds.yaml
  cds_config:
    path: ./cds.yaml
admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
  access_log:
  - name: envoy.access_loggers.file
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
```

Save the above YAML to envoy-proxy-1.yaml file. We also need to create empty (for now) cds.yaml and lds.yaml files:

touch {cds,lds}.yaml
We can now run the Envoy proxy with this configuration – func-e run -c envoy-proxy-1.yaml. If we look at the generated configuration (e.g. localhost:9901/config_dump) we’ll notice that it’s empty because we haven’t provided any listeners or clusters.

Let’s create the listener and routes configuration next:

```yaml
version_info: "0"
resources:
- "@type": type.googleapis.com/envoy.config.listener.v3.Listener
  name: listener_0
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 10000
  filter_chains:
  - filters:
    - name: envoy.filters.network.http_connection_manager
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
        stat_prefix: listener_http
        http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
        rds:
          route_config_name: route_config_1
          config_source:
            path: ./rds.yaml
```

Create an empty rds.yaml file (touch rds.yaml) and save the above YAML to lds.yaml. Because Envoy will only watch the file path for moves, saving the file won’t trigger the configuration reload. To trigger the reload, let’s overwrite the lds.yaml file:

mv lds.yaml tmp; mv tmp lds.yaml
The above command triggers the reload, and we should get the following log entry from Envoy:

```shell
[2021-09-07 19:04:06.710][2113][info][upstream] [source/server/lds_api.cc:78] lds: add/update listener 'listener_0'
```

Similarly, if we’d send a request to localhost:10000, we’d get an HTTP 404.

Let’s create the content for the rds.yaml next:

```yaml
version_info: "0"
resources:
- "@type": type.googleapis.com/envoy.config.route.v3.RouteConfiguration
  name: route_config_1
  virtual_hosts:
  - name: vh
    domains: ["*"]
    routes:
    - match:
        prefix: "/headers"
      route:
        cluster: instance_1
```

Let’s force the reload:

mv rds.yaml tmp; mv tmp rds.yaml
Finally, we also need to configure the clusters. Before we do that, let’s run an httpbin container:

docker run -d -p 5050:80 kennethreitz/httpbin
Now we can update the clusters (cds.yaml) and force the reload (mv cds.yaml tmp; mv tmp cds.yaml):

```yaml
version_info: "0"
resources:
- "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
  name: instance_1
  connect_timeout: 5s
  load_assignment:
    cluster_name: instance_1
    endpoints:
    - lb_endpoints:
      - endpoint:
          address:
            socket_address:
              address: 127.0.0.1
              port_value: 5050
```
As Envoy updates the configuration, we’ll get the following log entry:

```shell
$ [2021-09-07 19:09:15.582][2113][info][upstream] [source/common/upstream/cds_api_helper.cc:65] cds: added/updated 1 cluster(s), skipped 0 unmodified cluster(s)
```

Now we can make a request and verify that traffic reaches the endpoint defined in the cluster:

```shell
$ curl localhost:10000/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "localhost:10000",
    "User-Agent": "curl/7.64.0",
    "X-Envoy-Expected-Rq-Timeout-Ms": "15000"
  }
}
```

Note how we’ve configured the endpoints in the same file as the cluster. We can separate the two by defining the endpoints separately (eds.yaml).

Let’s start by creating the eds.yaml file:

```shell
version_info: "0"
resources:
- "@type": type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment
  cluster_name: instance_1
  endpoints:
  - lb_endpoints:
    - endpoint:
        address:
          socket_address:
            address: 127.0.0.1
            port_value: 5050
```
Save the above YAML to eds.yaml.

To use this endpoints file, we need to update the clusters (cds.yaml) to read the endpoints from eds.yaml:

```shell
version_info: "0"
resources:
- "@type": type.googleapis.com/envoy.config.cluster.v3.Cluster
  name: instance_1
  connect_timeout: 5s
  type: EDS
  eds_cluster_config:
    eds_config:
      path: ./eds.yaml
```

Force the reload by running mv cds.yaml tmp; mv tmp cds.yaml. Envoy will reload the configuration, and we’ll be able to send the requests to localhost:10000/headers just like before. The difference now is that different pieces of configuration are in separate files and can be updated separately.