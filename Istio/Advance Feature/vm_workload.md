# Working with VM workloads
If we have workloads running on virtual machines, we can connect them to the Istio service mesh and make them part of the mesh.

There are two architectures of an Istio service mesh with virtual machines: the single-network architecture and multi-network architecture.

## Single-network architecture
In this scenario, there’s a single network. The Kubernetes cluster and workloads running on virtual machines are in the same network, and they can communicate directly with each other.

<img src="../images/vm_workload_1.png"></img>

The control plane traffic (configuration updates, certificate signing) gets sent through the Gateway.

The VMs are configured with the gateway address to connect to the control plane when they are bootstrapping.

## Multi-network architecture
The multi-network architecture spans multiple networks. The Kubernetes cluster is inside one network, while the VMs are inside a different network. This prevents Pods in the Kubernetes cluster and workloads on the virtual machines from communicating directly.

<img src="../images/vm_workload_2.png"></img>

All traffic, the control plane, and pods-to-workloads flow through the gateway that acts as a bridge between the two networks.

## How are VM workloads represented in Istio?
There are two ways to represent VM workloads inside the Istio service mesh.

A workload group (WorkloadGroup resource) is similar to a Deployment in Kubernetes, and it represents a logical group of virtual machine workloads that share common properties.

The second approach to describe a VM workload is using a workload entry (WorkloadEntry resource). The workload entry is similar to a Pod, and it represents a single instance of a virtual machine workload.

Note that creating the above resource will not provision or run any VM workload instances. These resources are to reference or point to the VM workloads. Istio uses them to know how to configure the mesh appropriately, which services to add to the internal service registry, and so on.

To add a VM to the mesh, we’ll need to create a WorkloadGroup that acts as a template. Then, when we configure and add the VM to the mesh, the control plane automatically creates a corresponding WorkloadEntry.

We’ve mentioned that the WorkloadEntry acts similarly to a Pod. Whenever a VM gets added, the WorkloadEntry resource gets created. Similarly, whenever the VM workload gets removed from the mesh, the resource gets automatically removed.

In addition to the WorkloadEntry resources, we’ll also need to create a Kubernetes service. Creating a Kubernetes service gives us a stable hostname and IP address to access the VM workloads and pods using the selector fields. This also enables us to use Istio’s routing features through the DestinationRule and VirtualService resources.

