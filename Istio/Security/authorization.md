## Authorization
Authorization is answering the access control portion of the access control question. Is an (authenticated) principal allowed to perform an action on an object? Can user A send a GET request to the path /hello to service A?

Although we’ve authenticated the principal, the principal might not be allowed to perform an action. Your company ID card might be valid and authentic, but I won’t be able to use it to enter the offices of a different company. If we continue with the customs officer metaphor from before, we could say authorization is similar to a visa stamp in your passport.

This brings us to the next point: authentication without authorization (and vice-versa) doesn’t do much. For proper access control, we need both.

For example, if we only authenticate the principals and we don’t authorize them, they can do whatever they want and perform any actions on any objects. Conversely, if we authorize a request but don’t authenticate it, we can pretend to be someone else and perform any actions on any objects again.

Istio allows us to define access control at the mesh, namespace, and workload level using the AuthorizationPolicy resource. The AuthorizationPolicy supports DENY, ALLOW, AUDIT, and CUSTOM actions. U

Each Envoy proxy instance runs an authorization engine that authorizes requests at runtime. When requests come to the proxy, the engine evaluates the request context against the authorization policies and returns either ALLOW or DENY. The AUDIT action decides whether to log requests that match the rules. Note that the AUDIT policy does not affect whether requests are allowed or denied.

There’s no need to enable the authorization features explicitly. To enforce the access control, we can create an authorization policy to apply to our workloads.

The AuthorizationPolicy resource is where we can use the principal from the PeerAuthentication policies and the RequestAuthentication policy.

There are three parts we need to think of when defining an AuthorizationPolicy:

Select the workloads to apply the policy to
Action to take (deny, allow, or audit)
Rules when to take the action
Let’s see how these parts map to the fields in the AuthorizationPolicy resource using this example:

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
 name: customers-deny
 namespace: default
spec:
 selector:
   matchLabels:
     app: customers
     version: v2
 action: DENY
 rules:
 - from:
   - source:
       notNamespaces: ["default"]
```

Using the selector and matchLabels, we can select the workloads that the policy applies to. In our case, we are selecting all workloads with app: customers and version: v2 labels set. We’ve set the action to DENY with the action field.

Finally, we define all rules in the rules field. The rules in our example are saying to DENY the request (action) sent to the customers v2 workloads when the requests are coming from outside of the default namespace.

In addition to the from field in the rules, we can customize them further using the to and when fields. Let’s look at an example using those fields:

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
 name: customers-deny
 namespace: default
spec:
 selector:
   matchLabels:
     app: customers
     version: v2
 action: DENY
 rules:
 - from:
   - source:
       notNamespaces: ["default"]
 - to:
    - operation:
        methods: ["GET"]
 - when:
    - key: request.headers[User-Agent]
      values: ["Mozilla/*"]
```

We added the to and when fields to the rules section. If we translate the above rule, we can deny the GET requests to customer v2 workloads when they come from outside the default namespace and when the User Agent header value matches regex Mozilla/*.

Together, to defines what actions the policy allows, from defines who can take those actions, when represents the properties each request must have to be permitted by the policy, and selector determines which workloads will enforce the policy.

If there are multiple policies used for a single workload, Istio evaluates the deny policies first. The evaluation follows these rules:

1. If there are DENY policies that match the request, deny the request
2. If there are no ALLOW policies for the workload, allow the request
3. If any of the ALLOW policies match the request, allow the request
4. Deny the request

## Sources
The source we used in the above examples was notNamespaces. We could also use any of the following fields to specify the source of the requests as shown in the table.

| Source               | Example                           | Explanation                                                   |
|----------------------|-----------------------------------|---------------------------------------------------------------|
| principals           | principals: ["my-service-account"] | Any workload using my-service-account                         |
| notPrincipals        | notPrincipals: ["my-service-account"] | Any workload except my-service-account                        |
| requestPrincipals    | requestPrincipals: ["my-issuer/hello"] | Any workload with valid JWT and request principal my-issuer/hello |
| notRequestPrincipals | notRequestPrincipals: ["*"]       | Any workload without request principal (only valid JWT tokens)|
| namespaces           | namespaces: ["default"]           | Any workload from default namespace                           |
| notNamespaces        | notNamespaces: ["prod"]           | Any workload not in the prod namespace                        |
| ipBlocks             | ipBlocks: ["1.2.3.4", "9.8.7.6/15"] | Any workload with IP address of 1.2.3.4 or an IP address from the CIDR block |
| notIpBlock           | ipBlocks: ["1.2.3.4/24"]          | Any IP address that’s outside of the CIDR block               |


## Operations
We can define the operations under the to field. If we set more than one operation, the AND semantic gets used. Just like the sources, the operations come in pairs with a positive and a negative match. The values set to the operations fields are strings:

* hosts and notHosts
* ports and notPorts
* methods and notMethods
* paths and notPath

All these operations apply to the request attributes. For example, to match on a specific request path, we could use paths: ["/api/*", "/admin" ] or a specific port ports: ["8080"] and so on.

### Conditions
To specify the conditions, we have to provide a key field. The key field is the name of an Istio attribute. For example, request.headers , source.ip, destination.port, and so on. Refer to the [Authorization Policy Conditions](https://istio.io/latest/docs/reference/config/security/conditions/).

The second part of a condition is the values or the notValues list of strings. Here’s a snippet of a when condition:

```yaml
...
 - when:
    - key: source.ip
      notValues: ["10.0.1.1"]
```



