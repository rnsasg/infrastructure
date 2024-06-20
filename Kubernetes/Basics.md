Basic Concepts
What is Kubernetes?

Kubernetes is an open-source container orchestration platform for automating the deployment, scaling, and management of containerized applications.
What are Pods in Kubernetes?

Pods are the smallest and simplest Kubernetes objects. A Pod represents a single instance of a running process in your cluster and can contain one or more containers.
What is a Node in Kubernetes?

A Node is a worker machine in Kubernetes. It can be a virtual or physical machine, and it runs Pods.
What is a Cluster in Kubernetes?

A cluster is a set of nodes (machines) where Kubernetes is installed. A Kubernetes cluster is managed by a master node.
Core Components
What are the main components of the Kubernetes control plane?

API Server: Exposes the Kubernetes API.
etcd: Stores all cluster data.
Controller Manager: Runs controller processes.
Scheduler: Assigns Pods to nodes.
What is a Namespace in Kubernetes?

Namespaces are a way to divide cluster resources between multiple users (via resource quota).
What is a Deployment in Kubernetes?

A Deployment provides declarative updates for Pods and ReplicaSets.
What is a Service in Kubernetes?

A Service is an abstraction that defines a logical set of Pods and a policy to access them.
Advanced Topics
What are StatefulSets and when would you use them?

StatefulSets are used for stateful applications and guarantee the ordering and uniqueness of Pods.
What is a DaemonSet?

A DaemonSet ensures that all or some nodes run a copy of a Pod.
What is a ConfigMap and a Secret?

ConfigMap: Used to store non-confidential data in key-value pairs.
Secret: Used to store confidential data like passwords, OAuth tokens, and SSH keys.
What is the purpose of an Ingress in Kubernetes?

Ingress manages external access to services in a cluster, typically HTTP.
Networking
Explain Kubernetes networking.

Kubernetes networking follows certain rules like each Pod gets its own IP address, and containers within a Pod share the Pod IP.
What are Network Policies?

Network Policies are used to control the communication between Pods and between Pods and other network endpoints.
Storage
What is a PersistentVolume (PV) and PersistentVolumeClaim (PVC)?

PV: A piece of storage in the cluster.
PVC: A request for storage by a user.
What are StorageClasses?

StorageClasses provide a way to describe the “classes” of storage available.
Practical Knowledge
How do you deploy a Kubernetes cluster?

Using tools like kubeadm, kops, or managed Kubernetes services like GKE, EKS, or AKS.
How do you monitor a Kubernetes cluster?

Using tools like Prometheus, Grafana, and the Kubernetes dashboard.
How do you perform rolling updates and rollbacks in Kubernetes?

Using Deployments to update and roll back applications.
Common Commands
How do you view logs of a Pod?

kubectl logs <pod-name>
How do you get the details of a Node?

kubectl describe node <node-name>
How do you scale a Deployment?

kubectl scale deployment <deployment-name> --replicas=<number>
Practice Questions
Describe the process of creating a Pod in Kubernetes.

Creating a Pod involves defining a Pod manifest (YAML file) and using the kubectl apply -f <file> command.
How do you secure a Kubernetes cluster?

Implement RBAC (Role-Based Access Control), use network policies, encrypt data at rest and in transit, and ensure the API server is secure.
How does Kubernetes handle resource limits for containers?

Using resource requests and limits defined in the Pod specification.
Hands-on Practice
Set up a local Kubernetes cluster: Use Minikube or kind (Kubernetes in Docker).
Deploy applications: Create simple deployments and services.
Manage configurations: Use ConfigMaps and Secrets.
Implement Ingress: Set up an Ingress controller and define Ingress resources.
Monitor your cluster: Install Prometheus and Grafana, and explore metrics.
Resources for Preparation
Kubernetes Documentation: The official docs are comprehensive and up-to-date.
Practice Labs: Use platforms like Katacoda, Play with Kubernetes, or Civo for interactive learning.
Online Courses: Platforms like Coursera, Udemy, or A Cloud Guru offer Kubernetes courses.
Mock Interview
Mock interviews: Practice with a friend or use online platforms offering mock interview services.
By covering these areas, you'll be well-prepared for most Kubernetes interviews. Good luck!





