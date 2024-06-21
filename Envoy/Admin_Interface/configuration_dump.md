# Configuration dump
The /config_dump endpoint is a quick way to show the currently loaded Envoy configuration as JSON-serialized proto messages.

Envoy outputs the configuration for the following components, and in the order presented below:

* bootstrap
* clusters
* endpoints
* listeners
* scoped routes
* routes
* secrets


## Including the EDS config
To output the endpoint discovery service (EDS) configuration, we can add the ?include_eds parameter to the query.

## Filtering the output
Similarly, we can filter the output by providing the resources we want to include and a mask to return a subset of fields.

For example, to output only a static cluster configuration, we can use the static_clusters field from [ClustersConfigDump proto](https://www.envoyproxy.io/docs/envoy/latest/api-v3/admin/v3/config_dump.proto#envoy-v3-api-msg-admin-v3-clustersconfigdump) in the resource query parameter:

```shell
$ curl localhost:9901/config_dump?resource=static_clusters
{
 "configs": [
  {
   "@type": "type.googleapis.com/envoy.admin.v3.ClustersConfigDump.StaticCluster",
   "cluster": {
    "@type": "type.googleapis.com/envoy.config.cluster.v3.Cluster",
    "name": "instance_1",
  },
  ...
  {
   "@type": "type.googleapis.com/envoy.admin.v3.ClustersConfigDump.StaticCluster",
   "cluster": {
    "@type": "type.googleapis.com/envoy.config.cluster.v3.Cluster",
    "name": "instance_2",
...
```

## Using the mask parameter
To narrow the output further, we can specify the field in the mask parameter. For example, to show only the connect_timeout values for every cluster:

```shell
$ curl localhost:9901/config_dump?resource=static_clusters&mask=cluster.connect_timeout
{
 "configs": [
  {
   "@type": "type.googleapis.com/envoy.admin.v3.ClustersConfigDump.StaticCluster",
   "cluster": {
    "@type": "type.googleapis.com/envoy.config.cluster.v3.Cluster",
    "connect_timeout": "5s"
   }
  },
  {
   "@type": "type.googleapis.com/envoy.admin.v3.ClustersConfigDump.StaticCluster",
   "cluster": {
    "@type": "type.googleapis.com/envoy.config.cluster.v3.Cluster",
    "connect_timeout": "5s"
   }
  },
  {
   "@type": "type.googleapis.com/envoy.admin.v3.ClustersConfigDump.StaticCluster",
   "cluster": {
    "@type": "type.googleapis.com/envoy.config.cluster.v3.Cluster",
    "connect_timeout": "1s"
   }
  }
 ]
}
```

## Using regular expressions
Another filtering option is to specify a regular expression that matches the names of loaded configurations. For example, to output all listeners whose name field matches the regular expression .*listener.*, we could write the following:

```shell
$ curl localhost:9901/config_dump?resource=static_clusters&name_regex=.*listener.*

{
 "configs": [
  {
   "@type": "type.googleapis.com/envoy.admin.v3.ListenersConfigDump.StaticListener",
   "listener": {
    "@type": "type.googleapis.com/envoy.config.listener.v3.Listener",
    "name": "listener_0",
    "address": {
     "socket_address": {
      "address": "0.0.0.0",
      "port_value": 10000
     }
    },
    "filter_chains": [
     {}
    ]
   },
   "last_updated": "2021-11-15T20:06:51.208Z"
  }
 ]
}
```

Similarly, the /init_dump endpoint lists current information of unready targets of various Envoy components. Like the configuration dump, we can use the mask query parameter to filter for particular fields.

## Certificates

The /certs outputs all loaded TLS certificates. The data includes the certificate file name, serial number, subject alternate names, and days until expiration. The result is in JSON format, and it follows the [admin.v3.Certificates proto](https://www.envoyproxy.io/docs/envoy/latest/api-v3/admin/v3/certs.proto#envoy-v3-api-msg-admin-v3-certificates).
