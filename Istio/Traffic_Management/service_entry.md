# Bringing external services to the mesh
With the ServiceEntry resource, we can add additional entries to Istio’s internal service registry and make external services or internal services that are not part of our mesh look like part of our service mesh.

When a service is in the service registry, we can use the traffic routing, failure injection, and other mesh features, just like we would with other services.

Here’s an example of a ServiceEntry resource that declares an external API (api.external-svc.com) we can access over HTTPS.

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: external-svc
spec:
  hosts:
    - api.external-svc.com
  ports:
    - number: 443
      name: https
      protocol: TLS
  resolution: DNS
  location: MESH_EXTERNAL
```

The hosts field can contain multiple external APIs, and in that case, the Envoy sidecar will do the checks based on the hierarchy below. If Envoy cannot inspect any of the items, it moves to the next item in the order.

- HTTP Authority header (in HTTP/2) and Host header in HTTP/1.1),
- SNI,
- IP address and port

Envoy will either blindly forward the request or drop it if none of the above values can be inspected, depending on the Istio installation configuration.

Together with the WorkloadEntry resource, we can handle the migration of VM workloads to Kubernetes. In the WorkloadEntry, we can specify the details of the workload running on a VM (name, address, labels) and then use the workloadSelector field in the ServiceEntry to make the VMs part of Istio’s internal service registry.

For example, let’s say the customers workload is running on two VMs. Additionally, we already have Pods running in Kubernetes with the app: customers label.

Let’s define the WorkloadEntry resources like this:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: WorkloadEntry
metadata:
  name: customers-vm-1
spec:
  serviceAccount: customers
  address: 1.0.0.0
  labels:
    app: customers
    instance-id: vm1
---
apiVersion: networking.istio.io/v1alpha3
kind: WorkloadEntry
metadata:
  name: customers-vm-2
spec:
  serviceAccount: customers
  address: 2.0.0.0
  labels:
    app: customers
    instance-id: vm2
```

We can now create a ServiceEntry resource that spans both the workloads running in Kubernetes as well as the VMs:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: customers-svc
spec:
  hosts:
  - customers.com
  location: MESH_INTERNAL
  ports:
  - number: 80
    name: http
    protocol: HTTP
  resolution: STATIC
  workloadSelector:
    labels:
      app: customers
```

With MESH_INTERNAL setting in the location field, we say that this service is part of the mesh. This value is typically used in cases when we include workloads on unmanaged infrastructure (VMs). The other value for this field, MESH_EXTERNAL, is used for external services consumed through APIs. The MESH_INTERNAL and MESH_EXTERNAL settings control how sidecars in the mesh attempt to communicate with the workload, including whether they’ll use Istio mutual TLS by default.