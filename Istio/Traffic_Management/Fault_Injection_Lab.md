# Observing failure injection and delays in Grafana, Zipkin, and Kiali
In this lab, we will deploy the Web Frontend and Customers v1 service. We will then inject a failure, a delay, and observe both in Zipkin, Kiali, and Grafana.

<img src="../images/fault_injection_1.png"></img>

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

Save the above YAML to gateway.yaml and create the Gateway using kubectl apply -f gateway.yaml.
Next, we will deploy the Web Frontend, Service, and VirtualService.

```yaml
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

Save the above YAML to web-frontend.yaml and create the resources using kubectl apply -f web-frontend.yaml.
Lastly, we will deploy the Customers v1 and related resources.

```yaml
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
kind: DestinationRule
metadata:
  name: customers
spec:
  host: customers.default.svc.cluster.local
  subsets:
    - name: v1
      labels:
        version: v1
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
            subset: v1
```

Save the above YAML to customers.yaml and create the resources using kubectl apply -f customers.yaml.

With the applications deployed, let’s inject a 5-second delay to the Customers service for 50% of all requests. We will inject the delay in the customers VirtualService:

```yaml
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
            subset: v1
      fault:
        delay:
          percent: 50
          fixedDelay: 5s
```

Save above YAML to customers-delay.yaml and update the VirtualService using kubectl apply -f customers-delay.yaml.

To generate some traffic, let’s open a separate terminal window and start making requests to the GATEWAY_URL in an endless loop:

```shell
while true; do curl http://$GATEWAY_URL/; done
```

We should start noticing some of the requests taking longer than usual. Let’s open Grafana and observe these delays.

```shell
$ getmesh istioctl dash grafana
```

When Grafana opens, click Home and the Istio Service Dashboards. On the dashboard, make sure to select the customers.default.svc.cluster.local in the Service dropdown.

If you expand the Client Workloads panel, you’ll notice the increased duration on the Client Request Duration graph, as shown in the figure below.


<img src="../images/fault_injection_2.png"></img>

You can notice the same delay is reported on the web-frontend.default.svc.cluster.local service side.

Let’s see how this delay shows up in Zipkin. Open Zipkin with getmesh istioctl dash zipkin. On the main screen, select the serviceName and web-frontend.default, then add the minDuration criteria and enter 5s and click the search button to find traces.

Click on one of the traces to open the details page. On the details page, we will notice the duration is 5 seconds.

The single trace has four spans - click on the third span that represents the request made from the web-frontend to the customers service.

You’ll notice in the details that the response_flags tag gets set to DI. “DI” stands for “delay injection” and indicates that the request got delayed.

<img src="../images/fault_injection_3.png"></img>

Let’s update the VirtualService again, and this time, we will inject a fault and return HTTP 500 for 50% of the requests.

```yaml
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
            subset: v1
      fault:
        abort:
          httpStatus: 500
          percentage: 
            value: 50
```

Save the above YAML to customers-fault.yaml and update the VirtualService with kubectl apply -f customers-fault.yaml.

Just like before, we will start noticing failures from the request loop. If we go back to Grafana and open the Istio Service Dashboard, we will notice the client success rate dropping and the increase in the 500 responses on the Incoming Requests by Source and Response Code graph, as shown in the figure.

<img src="../images/fault_injection_4.png"></img>

There’s a similar story in Zipkin. If we search for traces again (we can remove the min duration), we will notice the traces with errors will show up in red color, as shown below.

<img src="../images/fault_injection_5.png"></img>

Let’s also open Kiali (getmesh istioctl dash kiali) and look at the service graph by clicking the Graph item. You will notice how the web-frontend service has a red border, as shown below.

<img src="../images/fault_injection_6.png"></img>

If we click on the web-frontend service and look at the sidebar on the right, you will notice the details of the HTTP requests. The graph shows the percentage of success and failures. Both numbers are around 50%, which corresponds to the percentage value we set in the VirtualService.

## Cleanup
Delete the Deployments, Services, VirtualServices, DestinationRule, and the Gateway:

```shell
kubectl delete deploy web-frontend customers-v1
kubectl delete svc customers web-frontend
kubectl delete vs customers web-frontend
kubectl delete gateway gateway
```

## Files

### customers-fault.yaml
```yaml
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
            subset: v1
      fault:
        abort:
          httpStatus: 500
          percentage: 
            value: 50
```

### gateway.yaml
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

### customers.yaml

```yaml
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
kind: DestinationRule
metadata:
  name: customers
spec:
  host: customers.default.svc.cluster.local
  subsets:
    - name: v1
      labels:
        version: v1
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
            subset: v1
```

### customers-delay.yaml
```yaml
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
            subset: v1
      fault:
        delay:
          percent: 50
          fixedDelay: 5s
```

### web-frontend.yaml
```yaml
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
      containers:
        - image: pj3677/web-frontend:1.0.0
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


