# Advanced Routing
Earlier, we learned how to route traffic between multiple subsets using the proportion of the traffic (weight field). In some cases, pure weight-based traffic routing or splitting is enough. However, there are scenarios and cases where we might need more granular control over how the traffic is split and forwarded to destination services.

Istio allows us to use parts of the incoming requests and match them to the defined values. For example, we can check the **URI prefix** of the incoming request and route the traffic based on that.

| **Property**  | **Description**                                                                                                 |
|-----------|-------------------------------------------------------------------------------------------------------------|
| uri       | Match the request URI to the specified value                                                                |
| schema    | Match the request schema (HTTP, HTTPS, …)                                                                   |
| method    | Match the request method (GET, POST, …)                                                                     |
| authority | Match the request authority header                                                                          |
| headers   | Match the request headers. Headers have to be lower-case and separated by hyphens (e.g. x-my-request-id). Note, if we use headers for matching, other properties get ignored (uri, schema, method, authority) |

Each of the above properties can get matched using one of these methods:

* Exact match: e.g. exact: "value" matches the exact string
* Prefix match: e.g. prefix: "value" matches the prefix only
* Regex match: e.g. regex: "value" matches based on the ECMAscript style regex

For example, let’s say the request URI looks like this: https://dev.example.com/v1/api. To match the request the URI, we’d write it like this:

```yaml
http:
- match:
  - uri:
      prefix: /v1
```

The above snippet would match the incoming request, and the request would get routed to the destination defined in that route.

Another example would be using Regex and matching on a header:

```yaml
http:
- match:
  - headers:
      user-agent:
        regex: '.*Firefox.*'
```

The above match will match any requests where the User Agent header matches the Regex.

## Redirecting and Rewriting Requests
Matching headers and other request properties are helpful, but sometimes we might need to match the requests by the values in the request URI.

For example, let’s consider a scenario where the incoming requests use the /v1/api path, and we want to route the requests to the /v2/api endpoint instead.

The way to do that is to rewrite all incoming requests and authority/host headers that match the /v1/api to /v2/api.

For example:

```yaml
...
http:
  - match:
    - uri:
        prefix: /v1/api
    rewrite:
      uri: /v2/api
    route:
      - destination:
          host: customers.default.svc.cluster.local
...
```

Even though the destination service doesn’t listen on the /v1/api endpoint, Envoy will rewrite the request to /v2/api.

We also have the option of redirecting or forwarding the request to a completely different service. Here’s how we could match on a header and then redirect the request to another service:

```yaml
...
http:
  - match:
    - headers:
        my-header:
          exact: hello
    redirect:
      uri: /hello
      authority: my-service.default.svc.cluster.local:8000
...
```
The redirect and destination fields are mutually exclusive. If we use the redirect, there’s no need to set the destination.

## AND and OR semantics
When doing matching, we can use both AND and OR semantics. Let’s take a look at the following snippet:

```yaml
...
http:
  - match:
    - uri:
        prefix: /v1
      headers:
        my-header:
          exact: hello
...
```
The above snippet uses the AND semantics. It means that both the URI prefix needs to match /v1 AND the header my-header has an exact value hello.

To use the OR semantic, we can add another match entry, like this:

```yaml
...
http:
  - match:
    - uri:
        prefix: /v1
    ...
  - match:
    - headers:
        my-header:
          exact: hello
...
```

In the above example, the matching will be done on the URI prefix first, and if it matches, the request gets routed to the destination. If the first one doesn’t match, the algorithm moves to the second one and tries to match the header. If we omit the match field on the route, it will continually evaluate true.