# Lab 3: Match Transport and Application Protocols
In this lab, we’ll learn how to use the TLS inspector filter to select a specific filter chain. Individual filter chains will distribute traffic to different upstream clusters based on the transport_protocol and application_protocol.

We’ll use the mendhak/http-https-echo Docker image for our upstream hosts. These containers can be configured to listens for HTTP/HTTPS and echo the responses back.

We’ll run three instances of the image to represent non-TLS HTTP, TLS HTTP/1.1, and TLS HTTP/2 protocols:

## non-TLS HTTP
docker run -dit -p 8080:8080 -t mendhak/http-https-echo:18

## TLS HTTP1.1
docker run -dit -e HTTPS_PORT=443 -p 443:443 -t mendhak/http-https-echo:18

## TLS HTTP2
docker run -dit -e HTTPS_PORT=8443 -p 8443:8443 -t mendhak/http-https-echo:18
To ensure all three containers are up and running, we can use curl to send a couple of requests and see if we get back the responses. From the output, we can also check that the hostname matches the actual Docker container ID.

## non-TLS HTTP
```shell
$ curl http://localhost:8080
{
  "path": "/",
  "headers": {
    "host": "localhost:8080",
    "user-agent": "curl/7.64.0",
    "accept": "*/*"
  },
  "method": "GET",
  "body": "",
  "fresh": false,
  "hostname": "localhost",
  "ip": "::ffff:172.18.0.1",
  "ips": [],
  "protocol": "http",
  "query": {},
  "subdomains": [],
  "xhr": false,
  "os": {
    "hostname": "100db0dce742"
  },
  "connection": {}
}
```

Note when sending the requests to the other two containers, we’ll use the -k flag to tell curl to skip the verification of the server’s TLS certificate. Additionally, we can use the --http1.1 and --http2 flags to send an HTTP1.1 or HTTP2 requests:

# HTTP1.1
```shell
$ curl -k --http1.1 https://localhost:443
{
  "path": "/",
  "headers": {
    "host": "localhost",
    "user-agent": "curl/7.64.0",
    "accept": "*/*"
  },
  "method": "GET",
  "body": "",
  "fresh": false,
  "hostname": "localhost",
  "ip": "::ffff:172.18.0.1",
  "ips": [],
  "protocol": "https",
  "query": {},
  "subdomains": [],
  "xhr": false,
  "os": {
    "hostname": "51afc40f7506"
  },
  "connection": {
    "servername": "localhost"
  }
}
```

```shell
$ curl -k --http2 https://localhost:8443
{
  "path": "/",
  "headers": {
    "host": "localhost:8443",
    "user-agent": "curl/7.64.0",
    "accept": "*/*"
  },
  "method": "GET",
  "body": "",
  "fresh": false,
  "hostname": "localhost",
  "ip": "::ffff:172.18.0.1",
  "ips": [],
  "protocol": "https",
  "query": {},
  "subdomains": [],
  "xhr": false,
  "os": {
    "hostname": "40e7143e6a55"
  },
  "connection": {
    "servername": "localhost"
  }
}
```

Once we’ve verified the containers are running correctly, we can create the Envoy configuration. We’ll use the tls_inspector and filter_chain_match fields to check if the transport protocol is TLS and if the application protocol is either HTTP1.1 (http/1.1) or HTTP2 (h2). Based on that information, we’ll have different clusters that will forward the traffic to upstream hosts (Docker containers). Remember the HTTP is running on port 8080, TLS HTTP/1.1 on port 443, and TLS HTTP2 on port 8443:

```yaml
static_resources:
  listeners:
  - address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    listener_filters:
    - name: "envoy.filters.listener.tls_inspector"
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.filters.listener.tls_inspector.v3.TlsInspector
    filter_chains:
    - filter_chain_match:
        # Match TLS and HTTP2
        transport_protocol: tls
        application_protocols: [h2]
      filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: service-tls-http2
          stat_prefix: https_passthrough
    - filter_chain_match:
        # Match TLS and HTTP1.1
        transport_protocol: tls
        application_protocols: [http/1.1]
      filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: service-tls-http1.1
          stat_prefix: https_passthrough
    - filter_chain_match:
      # No matches here, go to HTTP upstream
      filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: service-http
          stat_prefix: ingress_http
  clusters:
  - name: service-tls-http2
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: service-tls-http2
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8443
  - name: service-tls-http1.1
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: service-tls-http1.1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 443
  - name: service-http
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: service-http
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8080
admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901
```

Save the above YAML to tls.yaml and run it using func-e run -c tls.yaml.

To test this out, we can make similar curl requests as before and check if the hostnames match the Docker containers that are running:

```shell
$ curl http://localhost:10000 | jq '.os.hostname'
"100db0dce742"

$ curl -k --http1.1 https://localhost:10000 | jq '.os.hostname'
"51afc40f7506"

$ curl -k --http2 https://localhost:10000 | jq '.os.hostname'
"40e7143e6a55"
```

Alternatively, we could check the logs from the individual containers to see that requests are being sent correctly.

## tls-210927-165942.yaml

```yaml
static_resources:
  listeners:
  - address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    listener_filters:
    - name: "envoy.filters.listener.tls_inspector"
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.filters.listener.tls_inspector.v3.TlsInspector
    filter_chains:
    - filter_chain_match:
        # Match TLS and HTTP2
        transport_protocol: tls
        application_protocols: [h2]
      filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: service-tls-http2
          stat_prefix: https_passthrough
    - filter_chain_match:
        # Match TLS and HTTP1.1
        transport_protocol: tls
        application_protocols: [http/1.1]
      filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: service-tls-http1.1
          stat_prefix: https_passthrough
    - filter_chain_match:
      # No matches here, go to HTTP upstream
      filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: service-http
          stat_prefix: ingress_http
  clusters:
  - name: service-tls-http2
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: service-tls-http2
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8443
  - name: service-tls-http1.1
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: service-tls-http1.1
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 443
  - name: service-http
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: service-http
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8080
admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901
```