# Listeners and listener draining
The /listeners endpoints list all configured listeners. This includes the names as well as the address and ports each listener is listening on.

For example:

```shell
$ curl localhost:9901/listeners
http_8080::0.0.0.0:8080
http_hello_world_9090::0.0.0.0:9090
```

For the JSON output, we can append the ?format=json to the URL:

```shell
$ curl localhost:9901/listeners?format=json
{
 "listener_statuses": [
  {
   "name": "http_8080",
   "local_address": {
    "socket_address": {
     "address": "0.0.0.0",
     "port_value": 8080
    }
   }
  },
  {
   "name": "http_hello_world_9090",
   "local_address": {
    "socket_address": {
     "address": "0.0.0.0",
     "port_value": 9090
    }
   }
  }
 ]
}
```

## Listener draining
A typical scenario when draining occurs is during hot restart draining. It involves reducing the number of open connections by instructing the listeners to stop accepting incoming requests before the Envoy process is shut down.

By default, if we shut Envoy down, all connections are immediately closed. To do a graceful shutdown (i.e., don’t close existing connections), we can use the /drain_listeners endpoint with an optional graceful query parameter.

Envoy drains the connection based on the configuration specified through the --drain-time-s and --drain-strategy.

If not provided, the drain time defaults to 10 minutes (600 seconds). The value specifies how long Envoy will drain the connection — i.e., wait before closing them.

The drain strategy parameter determines the behavior during the drain sequence (e.g., during hot restart) where connections are terminated by sending the “Connection: CLOSE” (HTTP/1.1) or GOAWAY frame (HTTP/2).

There are two supported strategies: gradual (default) and immediate. When using the gradual strategy, the percentage of requests encouraged to drain increases to 100% as the drain time elapses. The immediate strategy will enable all requests to drain as soon as the drain sequence begins.

The draining is done per listener. However, it must be supported at the network filter level. The filters that currently support graceful draining are Redis, Mongo, and the HTTP connection manager.

Another option on the endpoint is the ability to drain all inbound listeners using the inboundonly query parameter (e.g., /drain_listeners?inboundonly). This uses the traffic_direction field on the listener to determine the traffic direction.