## Advanced Features
This module covers advanced features relating to Istio service mesh:

1. Istio deployment models. When planning how to deploy Istio, we need to decide whether we want our mesh to span a single or multiple Kubernetes clusters, whether services span one or more networks, whether to use a single or multiple control planes, and whether we need to support a single or multiple tenants.

2. Onboarding VM workloads. Not all mesh workloads run on Kubernetes. Here we discuss how to make workloads running on VMs a part of the mesh.

3. Extending the behavior of Envoy sidecars with Wasm plugins. Istio and Envoy support extension of the behavior of sidecars via WebAssembly plugins. We will provide an overview of this technology and how it works with Istio.

## Modules

1. [Multi Cluster Deployments](./multi_cluster_deployment.md)
2. [VM Workloads](./vm_workload.md)
3. [WASM Plugin](./wasm_plugin.md)
4. [WASM Plugin Lab](./wasm_plugin_lab.md)
5. [VM to Istio Service Mesh](./vm_to_istio_service_mesh.md)

