## Envoy Basics
For troubleshooting issues with Istio, it is helpful to have a basic understanding of how Envoy works. Envoy configuration is a JSON file that’s divided into multiple sections. The basics concepts we need to understand in Envoy are listeners, routes, clusters, and endpoints.

These concepts map to the Istio and Kubernetes resources, as shown in the following figure.

<img src="../images/debugging_1.png"></img>

**Listeners** are named network locations, typically an IP and port. Envoy listens to these locations and this is where it receives the connections and requests.

There are multiple listeners generated for each sidecar. Every sidecar has a listener that’s bound to **0.0.0.0:15006**. This is the address where the IP tables route all inbound Pod traffic to. The second listener is bound to **0.0.0.0:15001**, and this is where all outbound Pod traffic goes to.

When a request is redirected (using IP tables configuration) to port 15001, the listener hands it off to a virtual listener that best matches the original destination of the request. If it can’t find the destination, it sends traffic according to the OutboundTrafficPolicy that’s configured. By default, the request is sent to the **PassthroughCluster** which connects to the destination chosen by the application, without any load balancing by Envoy.

