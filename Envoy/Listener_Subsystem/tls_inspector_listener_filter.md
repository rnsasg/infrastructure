# TLS Inspector Listener Filter
The [TLS inspector](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/tls_inspector) listener filter allows us to detect whether the transport appears to be TLS or plaintext. If the transport is TLS, it detects the server name indication (SNI) and/or the client’s application-layer protocol negotiation (ALPN).

## What is SNI?

SNI or server name indication is an extension to the TLS protocol, and it tells us which hostname is connecting at the start of the TLS handshake process. We can serve multiple HTTPS services (with different certificates) on the same IP address and port using SNI. If a client connects with hostname “hello.com”, the server can present the certificate for that hostname. Similarly, if the client connects with “example.com” the server offers that certificate.

## What is ALPN?

ALPN or application-layer protocol negotiation is an extension to the TLS protocol that allows the application layer to negotiate which protocol should be performed over a secure connection without making additional round trips. Using ALPN we can determine whether the client is speaking HTTP/1.1 or HTTP/2.

We can use SNI and ALPN values to match filter chains using the server_names (for SNI) and/or application_protocols (for ALPN) fields.

The snippet below shows how we could use the application_protocols and server_names to execute different filter chains.

```yaml
...
    listener_filters:
      - name: "envoy.filters.listener.tls_inspector"
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.listener.tls_inspector.v3.TlsInspector
    filter_chains:
    - filter_chain_match:
        application_protocols: ["h2c"]
      filters:
      - name: some_filter
        ... 
    - filter_chain_match:
        server_names: "something.hello.com"
      transport_socket:
      ...
      filters:
      - name: another_filter
...
```
