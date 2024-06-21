# Multi-cluster deployments
A multi-cluster deployment (two or more clusters) gives us a greater degree of isolation and availability, but the cost we pay is in increased complexity. If the scenarios call for high availability (HA), we will have to deploy clusters across multiple zones and regions.

The next decision we need to make is to decide if we want to run the clusters within one network or if we want to use multiple networks.

The following figure shows a multi-cluster scenario (Cluster A, B, and C), deployed across two networks.

<img src="../images/Advance_1.png"></img>

## Network deployment models
When there are multiple networks involved, the workloads running inside the clusters have to use Istio gateways to reach workloads in other clusters. The use of multiple networks allows for better fault tolerance and scaling of network addresses.

<img src="../images/Advance_2.png"></img>

## Control plane deployment models

Istio service mesh uses the control plane to configure all communications between workloads inside the mesh. The control plane the workloads connect to depends on their configuration.

In the simplest case, we have a service mesh with a single control plane in a single cluster. This is the configuration we’ve been using throughout this course.

The shared control plane model involves multiple clusters where the control plane is running in one cluster only. That cluster is referred to as a primary cluster, while other clusters in the deployment are called remote clusters. These clusters don’t have their own control plane, instead, they are sharing the control plane from the primary cluster.

<img src="../images/Advance_3.png"></img>

Another deployment model is where we treat all clusters as remote clusters that are controlled by an external control plane. This gives us a complete separation between the control plane and the data plane. A typical example of an external control plane is when a cloud vendor is managing it.

For high availability, we should deploy multiple control plane instances across multiple clusters, zones, or regions as shown in the figure below.


<img src="../images/Advance_4.png"></img>


This model offers improved availability and configuration isolation. If one of the control planes becomes unavailable, the outage is limited to that one control plane. To improve that, you can implement failover and configure workload instances to connect to another control plane in case of a failure.

For the highest availability possible, we can deploy a control plane inside each cluster.

## Mesh deployment models
All diagrams and scenarios we look at so far were using a single mesh. In a single mesh model, all services are in one mesh, regardless of how many clusters and networks they are spanning.

A deployment model where multiple meshes are federated together is called a **multi-mesh deployment**. In this model, services can communicate across mesh boundaries. The model gives us a cleaner organizational boundary, stronger isolation and allows us to reuse service names and namespaces.

When federating two meshes, each mesh can expose a set of services and identities, which all participating meshes can recognize. To enable cross-mesh service communication we have to enable trust between the two meshes. Trust can be established by importing a trust bundle to a mesh and configuring local policies for those identities.

## Tenancy models
A tenant is a group of users sharing common access and privileges to a set of workloads. Isolation between the tenants is done through network configuration and policies. Istio supports namespace and cluster tenancies. Note that the tenancy we talk about here is soft multi-tenancy, not hard. There is no guaranteed protection against things like noisy neighbor problems when multiple tenants share the same Istio control plane.

Within a mesh, Istio uses namespaces as a unit of tenancy. If using Kubernetes we can grant permissions for workloads deployments per namespace. By default, services from different namespaces can communicate with each other through fully qualified names.

In the Security module, we have learned how to improve isolation using authorization policies and restrict access to only the appropriate callers.

In the multi-cluster deployment models, the namespaces in each cluster sharing the same name are considered the same namespace. Service Customers from namespace default in cluster A refers to the same service as service Customers from namespace default in cluster B. When traffic is sent to service Customers, load balancing is done across merged endpoints of both services as shown in the following figure.

<img src="../images/Advance_5.png"></img>

To configure cluster tenancy in Istio we need to configure each cluster as an independent service mesh. The meshes can be controlled and operated by separate teams and we can connect the meshes together into a multi-mesh deployment. If we use the same example as before, service Customers running in the default namespace in cluster A does not refer to the same service as service Customers from the default namespace in cluster B.

Another important part of the tenancy is isolating configuration from different tenants. At the moment, Istio does not address this issue, however, it encourages it through configuration that’s scoped at the namespace level.


## Best multi-cluster deployment
The best multi-cluster deployment topology is one where each cluster has its own control plane. For normal service mesh deployments at scale, it is recommended you use multi-mesh deployments and have a separate system that’s orchestrating the meshes externally. It is generally recommended to always use ingress gateways across clusters, even if they span a single network. Direct pod to pod connectivity requires populating endpoint data across multiple clusters which can slow down and complicate things. A simpler solution is to have traffic flow through ingresses across clusters instead.
