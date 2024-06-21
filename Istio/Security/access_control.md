## Access control
In this lab, we will learn how to use an authorization policy to control access between workloads.

Let’s start by deploying the Gateway:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: gateway
spec:
  selector:
    istio: ingressgateway
  servers:
    - port:
        number: 80
        name: http
        protocol: HTTP
      hosts:
        - '*'
```

Save the above YAML to gateway.yaml and deploy the Gateway using kubectl apply -f gateway.yaml
Next, we will create the Web frontend deployment, service account, service, and a VirtualService.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: web-frontend
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-frontend
  labels:
    app: web-frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web-frontend
  template:
    metadata:
      labels:
        app: web-frontend
        version: v1
    spec:
      serviceAccountName: web-frontend
      containers:
        - image: gcr.io/tetratelabs/web-frontend:1.0.0
          imagePullPolicy: Always
          name: web
          ports:
            - containerPort: 8080
          env:
            - name: CUSTOMER_SERVICE_URL
              value: 'http://customers.default.svc.cluster.local'
---
kind: Service
apiVersion: v1
metadata:
  name: web-frontend
  labels:
    app: web-frontend
spec:
  selector:
    app: web-frontend
  ports:
    - port: 80
      name: http
      targetPort: 8080
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: web-frontend
spec:
  hosts:
    - '*'
  gateways:
    - gateway
  http:
    - route:
        - destination:
            host: web-frontend.default.svc.cluster.local
            port:
              number: 80
```

Save the above YAML to web-frontend.yaml and create the resource using kubectl apply -f web-frontend.yaml.

Finally, we will deploy the customers v1 service.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: customers-v1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: customers-v1
  labels:
    app: customers
    version: v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: customers
      version: v1
  template:
    metadata:
      labels:
        app: customers
        version: v1
    spec:
      serviceAccountName: customers-v1
      containers:
        - image: gcr.io/tetratelabs/customers:1.0.0
          imagePullPolicy: Always
          name: svc
          ports:
            - containerPort: 3000
---
kind: Service
apiVersion: v1
metadata:
  name: customers
  labels:
    app: customers
spec:
  selector:
    app: customers
  ports:
    - port: 80
      name: http
      targetPort: 3000
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: customers
spec:
  hosts:
    - 'customers.default.svc.cluster.local'
  http:
    - route:
        - destination:
            host: customers.default.svc.cluster.local
            port:
              number: 80
```

Save the above to customers-v1.yaml and create the deployment and service using kubectl apply -f customers-v1.yaml. If we open the GATEWAY_URL the web frontend page with the data from the customers v1 service should be displayed.

Let’s start by creating an authorization policy that denies all requests in the default namespace.

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
 name: deny-all
 namespace: default
spec:
  {}
```

Save the above to deny-all.yaml and create the policy using kubectl apply -f deny-all.yaml.

If we try to access GATEWAY_URL we will get back the following response:

RBAC: access denied
Similarly, if we run a Pod inside the cluster and make a request from within the default namespace to either the web frontend or the customer service, we’ll get the same error.

Let’s try that:

```shell
$ kubectl run curl --image=radial/busyboxplus:curl -i --tty
If you don't see a command prompt, try pressing enter.
[ root@curl:/ ]$ curl customers
RBAC: access denied
[ root@curl:/ ]$ curl web-frontend
RBAC: access denied
[ root@curl:/ ]$
In both cases, we get back the access denied error.
```

The first thing we will do is to allow requests being sent from the ingress gateway to the web-frontend application using an ALLOW action. In the rules, we specify the source namespace (istio-system) where the ingress gateway is running and the ingress gateway’s service account name in the principals field.

```shell
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-ingress-frontend
  namespace: default
spec:
  selector:
    matchLabels:
      app: web-frontend
  action: ALLOW
  rules:
    - from:
        - source:
            namespaces: ["istio-system"]
        - source:
            principals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
```

Save the above to allow-ingress-frontend.yaml and create the policy using kubectl apply -f allow-ingress-frontend.yaml.

If we try to make a request from our host to the GATEWAY_URL, we will get a different error this time:

```shell
$ curl http://$GATEWAY_URL
"Request failed with status code 403"
```

Note that it takes a couple of seconds for Istio to distribute the policy to all proxies, so you might still see the RBAC: access denied message for a couple of seconds.

This error is coming from the customers service - remember we allowed calls to the web frontend. However, web-frontend still can’t make calls to the customers service.

If we go back to the curl Pod we are running inside the cluster and try to request http://web-frontend we will get an RBAC error. The DENY policy is in effect, and we are only allowing calls to be made from the ingress gateway.

When we deployed the web frontend, we also created a service account for the Pod (otherwise, all Pods in the namespace are assigned the default service account). We can now use that service account to specify where the customer service calls can come from.

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-web-frontend-customers
  namespace: default
spec:
  selector:
    matchLabels:
        app: customers
        version: v1
  action: ALLOW
  rules:
  - from:
    - source:
        namespaces: ["default"]
      source:
        principals: ["cluster.local/ns/default/sa/web-frontend"]
```

Save the above YAML to allow-web-frontend-customers.yaml and create the policy using kubectl apply -f allow-web-frontend-customers.yaml.

As soon as the policy gets created, we will see the web frontend working again - it will get the customer service responses.

We have used multiple authorization policies to explicitly allow calls from the ingress to the front end and the front end to the customer service. Using a deny-all approach is an excellent way to start because we can control, manage, and then explicitly allow the communication we want to happen between services.

Cleanup
Delete the Deployments, Services, VirtualServices, and the Gateway:

```shell
kubectl delete sa customers-v1 web-frontend
kubectl delete deploy web-frontend customers-v1
kubectl delete svc customers web-frontend
kubectl delete vs customers web-frontend
kubectl delete gateway gateway
kubectl delete authorizationpolicy allow-ingress-frontend allow-web-frontend-customers deny-all
kubectl delete pod curl
```
## files

### customersv1-210817-145729.yaml

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: customers-v1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: customers-v1
  labels:
    app: customers
    version: v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: customers
      version: v1
  template:
    metadata:
      labels:
        app: customers
        version: v1
    spec:
      serviceAccountName: customers-v1
      containers:
        - image: gcr.io/tetratelabs/customers:1.0.0
          imagePullPolicy: Always
          name: svc
          ports:
            - containerPort: 3000
---
kind: Service
apiVersion: v1
metadata:
  name: customers
  labels:
    app: customers
spec:
  selector:
    app: customers
  ports:
    - port: 80
      name: http
      targetPort: 3000
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: customers
spec:
  hosts:
    - 'customers.default.svc.cluster.local'
  http:
    - route:
        - destination:
            host: customers.default.svc.cluster.local
            port:
              number: 80
```

### webfrontend-210817-145729.yaml

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: web-frontend
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-frontend
  labels:
    app: web-frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web-frontend
  template:
    metadata:
      labels:
        app: web-frontend
        version: v1
    spec:
      serviceAccountName: web-frontend
      containers:
        - image: gcr.io/tetratelabs/web-frontend:1.0.0
          imagePullPolicy: Always
          name: web
          ports:
            - containerPort: 8080
          env:
            - name: CUSTOMER_SERVICE_URL
              value: 'http://customers.default.svc.cluster.local'
---
kind: Service
apiVersion: v1
metadata:
  name: web-frontend
  labels:
    app: web-frontend
spec:
  selector:
    app: web-frontend
  ports:
    - port: 80
      name: http
      targetPort: 8080
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: web-frontend
spec:
  hosts:
    - '*'
  gateways:
    - gateway
  http:
    - route:
        - destination:
            host: web-frontend.default.svc.cluster.local
            port:
              number: 80
```

### allowingressfrontend-210817-145729.yaml

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-ingress-frontend
  namespace: default
spec:
  selector:
    matchLabels:
      app: web-frontend
  action: ALLOW
  rules:
    - from:
        - source:
            namespaces: ["istio-system"]
        - source:
            principals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
```

### allowwebfrontendcustomers-210817-145729.yaml

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-web-frontend-customers
  namespace: default
spec:
  selector:
    matchLabels:
        app: customers
        version: v1
  action: ALLOW
  rules:
  - from:
    - source:
        namespaces: ["default"]
      source:
        principals: ["cluster.local/ns/default/sa/web-frontend"]
```

### gateway-210119-085711.yaml

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: gateway
spec:
  selector:
    istio: ingressgateway
  servers:
    - port:
        number: 80
        name: http
        protocol: HTTP
      hosts:
        - '*'
```

### denyall-210119-085711.yaml

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
 name: deny-all
 namespace: default
spec:
  {}
```

