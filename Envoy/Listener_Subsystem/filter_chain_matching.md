# Filter Chain Matching
Filter chain matching allows us to specify the criteria for selecting a specific filter chain for a listener.

We can define multiple filter chains in the configuration and then select and execute them based on the destination port, server name, protocol, and other properties. For example, we could check which hostname is connecting and select a different filter chain. If hostname hello.com connects, we could choose a filter chain that presents the certificate for that specific hostname.

Before Envoy can start filter matching, it needs to have some data extracted from the received packet by listener filters. After that, for Envoy to select a specific filter chain, all match criteria must be fulfilled. For example, if we’re matching on hostname and port, both values need to match for Envoy to select that filter chain.

The property matching order is as follows:

1. Destination port (when use_original_dst is used)
2. Destination IP address
3. Server name (SNI for TLS protocol)
4. Transport protocol
5. Application protocols (ALPN for TLS protocol)
6. Directly connected source IP address (this is only different from the source IP address if we’re using a filter that overrides the source address, for example, the proxy protocol listener filter)
7. Source type (e.g., any, local, or external network)
8. Source IP address
9. Source port
10. Specific criteria, such as server name/SNI or IP addresses, also allow ranges or wildcards to be used. If using wildcard criteria in multiple filter chains, the most specific value will be matched.


For example, here’s how the order from most specific to least specific match would look like for www.hello.com:

```
www.hello.com
*.hello.com
*.com
```

Any filter chain without the server name criteria
Here’s an example of how we could configure filter chain matches using different properties:

```yaml
filter_chains:
- filter_chain_match:
    server_names:
      - "*.hello.com"
  filters:
    ...
- filter_chain_match:
    source_prefix_ranges:
      - address_prefix: 192.0.0.1
        prefix_len: 32
  filters:
    ...
- filter_chain_match:
    transport_protocol: tls
  filters:
    ...
```

Let’s assume a TLS request comes in from the IP address 192.0.0.1 and has SNI set to v1.hello.com. Keeping the order in mind, the first filter chain match that satisfies all criteria is the server name match (v1.hello.com). Therefore the Envoy executes the filters under that match.

However, if the request comes in from the IP 192.0.0.1, it wouldn’t be TLS, and the SNI doesn’t match the *.hello.com. Envoy will execute the second filter chain – the one that matches the specific IP address.