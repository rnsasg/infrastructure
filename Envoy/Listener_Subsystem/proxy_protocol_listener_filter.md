# Proxy Protocol Listener Filter
The [Proxy protocol](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/proxy_protocol) listener filter (envoy.filters.listener.proxy_protocol) adds support for the[ HAProxy proxy protocol](https://www.haproxy.org/download/1.9/doc/proxy-protocol.txt).

Proxies use their IP stack to connect to remote servers and lose the source and destination information from the initial connection. The PROXY protocol allows us to chain proxies without losing client information. The protocol defines a way for communicating metadata about a connection over TCP before the main TCP stream. The metadata includes the source IP address.

Using this filter, Envoy can consume the metadata from the PROXY protocol and propagate it into an x-forwarded-for header, for example.