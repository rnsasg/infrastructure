# Lab 1: Original Destination Filter
In this lab, we’ll learn how to configure the original destination filter. To do this, we’ll need to enable IP forwarding and then update the iptables rules to capture all traffic and redirect it to the port Envoy is listening on.

We’ll be using a Linux virtual machine instead of the Google Cloud Shell.

Let’s start by enabling IP forwarding:

##  Enable IP forwarding
sudo sysctl -w net.ipv4.ip_forward=1
Next, we need to configure iptables to capture all traffic sent to port 80 and redirect it to port 10000. The Envoy proxy will be listening on port 10000.

First, we need to determine the network interface name we’ll use in the iptables command. We can list the network interfaces using the ip link show command. For example:

```shell
peter@instance-1:~$ ip link show
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
2: ens4: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1460 qdisc mq state UP mode DEFAULT group default qlen 1000    link/ether 42:01:0a:8a:00:2e brd ff:ff:ff:ff:ff:ff
```

The output tells us we have two network interfaces – the loopback interface and an interface called ens4. This is the interface name we’ll use in the iptables command:

## Capture all traffic from outside to port 80 and redirect it to port 10000
> sudo iptables -t nat -A PREROUTING -i ens4 -p tcp --dport 80 -j REDIRECT --to-port 10000
Finally, we’ll run another iptables to command that prevents routing loops when requests are made from the virtual machine. Setting this rule will allow us to run curl tetrate.io from the VM and still get redirected to port 10000:

## Enables us to run `curl` from the same instance (i.e. prevents routing loops)
> sudo iptables -t nat -A OUTPUT -p tcp -m owner ! --uid-owner root --dport 80 -j REDIRECT --to-port 10000
With iptables rules modified, we can create the following Envoy configuration:

```yaml
static_resources:
  listeners:
    - name: inbound
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 10000
      listener_filters:
        - name: envoy.filters.listener.original_dst
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.listener.original_dst.v3.OriginalDst
      filter_chains:
        - filters:
          - name: envoy.filters.network.http_connection_manager
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
              stat_prefix: ingress_http
              access_log:
              - name: envoy.access_loggers.file
                typed_config:
                  "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                  path: ./envoy.log
              http_filters:
              - name: envoy.filters.http.router
                typed_config: 
                   "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
              route_config:
                virtual_hosts:
                - name: proxy
                  domains: ["*"]
                  routes:
                  - match:
                      prefix: "/"
                    route:
                      cluster: original_dst_cluster
  clusters:
    - name: original_dst_cluster
      type: ORIGINAL_DST
      connect_timeout: 5s
      lb_policy: CLUSTER_PROVIDED
      original_dst_lb_config:
        use_http_header: true
```

The configuration looks similar to the one we’ve already seen. We’re adding the original_dst filter to the listener_filters, enabling access logging to a file and routing all traffic to a cluster called original_dst_cluster. This cluster has the type set to ORIGINAL_DST, sending the request to the original destination.

Additionally, we’ve set the use_http_header field to true. When set to true, we can use the x-envoy-original-dst-host header to override the destination address. Note that this header is not sanitized by default, so enabling it allows routing traffic to arbitrary hosts, which might have security consequences. We’re using it here only as an example.

Original DST filter
Original DST filter
For the transparent proxy scenario, this is all we need. We don’t want to do any resolving. We want to proxy the request to the original destination.

Save the above YAML to 5-lab-1-originaldst.yaml.

To run it, we’ll use [func-e CLI](https://func-e.io/). Let’s install the CLI on the VM:

> curl https://func-e.io/install.sh | sudo bash -s -- -b /usr/local/bin
Now we can run the Envoy proxy with the configuration we created:

```shell 
sudo func-e run -c 5-lab-1-originaldst.yaml 
```

> Note that we’re running func-e with sudo in this scenario, so we can use the same machine to test the proxy out and prevent routing loops (see the second iptables rule).

We can send a request to tetrate.io, and if we look in the envoy.log file, we’ll see the following entry:

```shell
[2021-07-07T21:22:57.294Z] "GET / HTTP/1.1" 301 - 0 227 34 34 "-" "curl/7.64.0" "5fd04969-27b0-4d37-b56c-c273a410da46" "tetrate.io" "75.119.195.116:80"
```

The log entry shows that iptables captured the request and redirected it to the port 10000 where Envoy is listening. Then, Envoy proxied the request to the original destination.

We can also make requests from outside of the virtual machine. From a second terminal – this time, we’re using the Google Cloud Shell, and we’re not in the virtual machine – we can send a request to the virtual machine IP address and provide the x-envoy-original-dst-host header that with an IP address we want Envoy to send the request to.

> I am using google.com in this example. To get the IP address, you can run nslookup google.com and use the IP address from that command.

```shell
$ curl -H "x-envoy-original-dst-host: 74.125.199.139:80" [vm-ip-address]
<HTML><HEAD><meta http-equiv="content-type" content="text/html;charset=utf-8">
<TITLE>301 Moved</TITLE></HEAD><BODY>
<H1>301 Moved</H1>
The document has moved
<A HREF="http://www.google.com/">here</A>.
</BODY></HTML>
```

You’ll notice the response is proxied to google.com. We can also check the envoy.log on the VM to see the log entry.

To clean up the iptables rules and disable IP forwarding, run:

## Disable IP forwarding
sudo sysctl -w net.ipv4.ip_forward=0

##  Delete all rules from the nat table
sudo iptables -t nat -F

## 5lab1originaldst-221021-124150.yaml

```yaml
static_resources:
  listeners:
    - name: inbound
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 10000
      listener_filters:
        - name: envoy.filters.listener.original_dst
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.listener.original_dst.v3.OriginalDst
      filter_chains:
        - filters:
          - name: envoy.filters.network.http_connection_manager
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
              stat_prefix: ingress_http
              access_log:
              - name: envoy.access_loggers.file
                typed_config:
                  "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                  path: ./envoy.log
              http_filters:
              - name: envoy.filters.http.router
                typed_config: 
                  "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
              route_config:
                virtual_hosts:
                - name: proxy
                  domains: ["*"]
                  routes:
                  - match:
                      prefix: "/"
                    route:
                      cluster: original_dst_cluster
  clusters:
    - name: original_dst_cluster
      type: ORIGINAL_DST
      connect_timeout: 5s
      lb_policy: CLUSTER_PROVIDED
      original_dst_lb_config:
        use_http_header: true
```
