# Resiliency
Resiliency is the ability to provide and maintain an acceptable level of service in the face of faults and challenges to regular operation. It’s not about avoiding failures. It’s responding to them in such a way that there’s no downtime or data loss. The goal for resiliency is to return the service to a fully functioning state after a failure occurs.

A crucial element in making services available is using **timeouts** and **retry policies** when making service requests. We can configure both on Istio’s VirtualService.

Using the timeout field, we can define a timeout for HTTP requests. If the request takes longer than the value specified in the timeout field, Envoy proxy will drop the requests and mark them as timed out (return an HTTP 408 to the application). The connections remain open unless outlier detection is triggered. Here’s an example of setting a timeout for a route:

```yaml
...
- route:
  - destination:
      host: customers.default.svc.cluster.local
      subset: v1
  timeout: 10s
...
```

In addition to timeouts, we can also configure a more granular retry policy. We can control the number of retries for a given request and the timeout per try as well as the specific conditions we want to retry on.

Both retries and timeouts happen on the client-side. For example, we can only retry the requests if the upstream server returns any 5xx response code, or retry only on gateway errors (HTTP 502, 503, or 504), or even specify the retriable status codes in the request headers. When Envoy retries a failed request, the endpoint that initially failed and caused the retry is no longer included in the load balancing pool. Let’s say the Kubernetes service has three endpoints (Pods), and one of them fails with a retriable error code. When Envoy retries the request, it won’t resend the request to the original endpoint anymore. Instead, it will send the request to one of the two endpoints that haven’t failed.

Here’s an example of how to set a retry policy for a particular destination:


```yaml
...
- route:
  - destination:
      host: customers.default.svc.cluster.local
      subset: v1
  retries:
    attempts: 10
    perTryTimeout: 2s
    retryOn: connect-failure,reset
...
```

The above retry policy will attempt to retry any request that fails with a connect timeout (connect-failure) or if the server does not respond at all (reset). We set the per-try attempt timeout to 2 seconds and the number of attempts to 10. Note that if we set both retries and timeouts, the timeout value will be the maximum the request will wait. If we had a 10-second timeout specified in the above example, we would only ever wait 10 seconds maximum, even if there are still attempts left in the retry policy.

For more details on retry policies, see the [x-envoy-retry-on documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-on).