# Lab 2: TLS Inspector Filter
In this lab, we’ll look at an example of how we could set up Envoy to serve multiple websites with different certificates on a single IP address.

We’ll use self-signed certificates for testing, but the process is identical if using real-signed certificates.

Using openssl we’ll create self-signed certificates for www.hello.com and www.example.com in the certs folder:

```shell
$ mkdir certs && cd certs

$ openssl req -nodes -new -x509 -keyout www_hello_com.key -out www_hello_com.cert -subj "/C=US/ST=Washington/L=Seattle/O=Hello LLC/OU=Org/CN=www.hello.com"

$ openssl req -nodes -new -x509 -keyout www_example_com.key -out www_example_com.cert -subj "/C=US/ST=Washington/L=Seattle/O=Example LLC/OU=Org/CN=www.example.com"
```

For each common name we’ll end up with two files - the private key and the certificate (e.g. www_example_com.key and www_example_com.cert).

We’ll configure the transport_socket field for each filter chain match separately. Here’s a snippet of how to define the TLS transport socket and provide the key and certificate:

```yaml
...
      transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
          common_tls_context:
            tls_certificates:
            - certificate_chain:
                filename: certs/www_hello_com.cert
              private_key:
                filename: certs/www_hello_com.key
...
```

Because we want to use different certificates based on the SNI, we’ll add a TLS inspector filter to the TLS listeners to use the filter_chain_match and the server_names field to match based on the SNI.

Here are the two filter chain match sections – note how each filter chain match has its own transport_socket with pointers to the certificate and key files:

```yaml
    listener_filters:
      - name: "envoy.filters.listener.tls_inspector"
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.listener.tls_inspector.v3.TlsInspector
    filter_chains:
    - filter_chain_match:
        server_names: "www.example.com"
      filters:
        transport_socket:
          name: envoy.transport_sockets.tls
          ...
        http:filters:
          ...
    - filter_chain_match:
        server_names: "www.hello.com"
      filters:
        transport_socket:
          name: envoy.transport_sockets.tls
          ...
        http:filters:
          ...
```

You can find the complete configuration in the 5-lab-2-tls_match.yaml file and run it using lfunc-e run -c 5-lab-2-tls_match.yaml. Since we'll be only using openssl to connect, we don't need the cluster endpoints up and running.

To check the SNI matching works correctly, we can use the openssl and connect to the Envoy listener providing the server name. For example:

```shell
$ openssl s_client -connect 0.0.0.0:443 -servername www.example.com
CONNECTED(00000003)
depth=0 C = US, ST = Washington, L = Seattle, O = Example LLC, OU = Org, CN = www.example.com
verify error:num=18:self signed certificate
verify return:1
depth=0 C = US, ST = Washington, L = Seattle, O = Example LLC, OU = Org, CN = www.example.com
verify return:1
---
Certificate chain
 0 s:C = US, ST = Washington, L = Seattle, O = Example LLC, OU = Org, CN = www.example.com
   i:C = US, ST = Washington, L = Seattle, O = Example LLC, OU = Org, CN = www.example.com
...
```

The command will return the correct peer certificate based on the provided server name.

