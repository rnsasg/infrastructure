# Failure Injection
To help us with service resiliency, we can use the **fault injection** feature. We can apply the fault injection policies on HTTP traffic and specify one or more faults to inject when forwarding the destination’s request.

There are two types of fault injection. We can **delay** the requests before forwarding and emulate slow network or overloaded service, and we can **abort** the HTTP request and return a specific HTTP error code to the caller. With the abort, we can simulate a faulty upstream service.

Here’s an example of aborting HTTP requests and returning HTTP 404, for 30% of the incoming requests:

```yaml
- route:
  - destination:
      host: customers.default.svc.cluster.local
      subset: v1
  fault:
    abort:
      percentage:
        value: 30
      httpStatus: 404
```

If we don’t specify the percentage, the Envoy proxy will abort all requests. Note that the fault injection affects services that use that VirtualService. It does not affect all consumers of the service.

Similarly, we can apply an optional delay to the requests using the fixedDelay field:

```yaml
- route:
  - destination:
      host: customers.default.svc.cluster.local
      subset: v1
  fault:
    delay:
      percentage:
        value: 5
      fixedDelay: 3s
```

The above setting will apply 3 seconds of delay to 5% of the incoming requests.

Note that the fault injection will **not** trigger any retry policies we have set on the routes. For example, if we injected an HTTP 500 error, the retry policy configured to retry on the HTTP 500 will not be triggered.