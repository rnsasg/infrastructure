# Subsets and DestinationRule
The destinations also refer to different **subsets** (or service versions). With subsets, we can identify different variants of our application. In our example, we have two subsets, v1 and v2, which correspond to the two different versions of our customer service. Each subset uses a combination of key/value pairs (labels) to determine which Pods to include. We can declare subsets in a resource type called **DestinationRule**.

Here’s how the DestinationRule resource looks like with two subsets defined:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: customers-destination
spec:
  host: customers.default.svc.cluster.local
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
```

Let’s look at the traffic policies we can set in the DestinationRule.

# Traffic Policies in DestinationRule
With the DestinationRule, we can define load balancing configuration, connection pool size, outlier detection, etc., to apply to the traffic after the routing has occurred. We can set the traffic policy settings under the trafficPolicy field. Here are the settings:

* Load balancer settings
* Connection pool settings
* Outlier detection
* Client TLS settings
* Port traffic policy

## Load Balancer Settings
With the load balancer settings, we can control which load balancer algorithm is used for the destination. Here’s an example of the DestinationRule with the traffic policy that sets the load balancing algorithm for the destination to round-robin:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: customers-destination
spec:
  host: customers.default.svc.cluster.local
  trafficPolicy:
    loadBalancer:
      simple: ROUND_ROBIN
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
```

We can also set up hash-based load balancing and provide session affinity based on the HTTP headers, cookies, or other request properties. Here’s a snippet of the traffic policy that sets the hash-based load balancing and uses a cookie called ’location` for affinity:

```yaml
trafficPolicy:
  loadBalancer:
    consistentHash:
      httpCookie:
        name: location
        ttl: 4s
```

## Connection Pool Settings
These settings can be applied to each host in the upstream service at the TCP and HTTP level, and we can use them to control the volume of connections.

Here’s a snippet that shows how we can set a limit of concurrent requests to the service:

```yaml
spec:
  host: myredissrv.prod.svc.cluster.local
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 50
```

## Outlier Detection
Outlier detection is a circuit breaker implementation that tracks the status of each host (Pod) in the upstream service. If a host starts returning 5xx HTTP errors, it gets ejected from the load balancing pool for a predefined time. For the TCP services, Envoy counts connection timeouts or failures as errors.

Here’s an example that sets a limit of 500 concurrent HTTP2 requests (http2MaxRequests), with not more than ten requests per connection (maxRequestsPerConnection) to the service. The upstream hosts (Pods) get scanned every 5 minutes (interval), and if any of them fails ten consecutive times (consecutiveErrors), Envoy will eject it for 10 minutes (baseEjectionTime).

```yaml
trafficPolicy:
  connectionPool:
    http:
      http2MaxRequests: 500
      maxRequestsPerConnection: 10
  outlierDetection:
    consecutiveErrors: 10
    interval: 5m
    baseEjectionTime: 10m
```

## Client TLS Settings
Contains any TLS related settings for connections to the upstream service. Here’s an example of configuring mutual TLS using the provided certificates:

```yaml
trafficPolicy:
  tls:
    mode: MUTUAL
    clientCertificate: /etc/certs/cert.pem
    privateKey: /etc/certs/key.pem
    caCertificates: /etc/certs/ca.pem
```

Other supported TLS modes are DISABLE (no TLS connection), SIMPLE (originate a TLS connection the upstream endpoint), and ISTIO_MUTUAL (similar to MUTUAL, which uses Istio’s certificates for mTLS).

## Port Traffic Policy
Using the portLevelSettings field we can apply traffic policies to individual ports. For example:

```yaml
trafficPolicy:
  portLevelSettings:
  - port:
      number: 80
    loadBalancer:
      simple: LEAST_CONN
  - port:
      number: 8000
    loadBalancer:
      simple: ROUND_ROBIN
```
