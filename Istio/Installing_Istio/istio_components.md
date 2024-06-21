# Components

In this lesson we discuss:

* the components that make up an Istio installation,
* the concept of configuration profiles, and
* explore the different methods that are available for installing Istio onto a Kubernetes cluster.

## Istio components
Istio consists of multiple components that can be deployed together or separately.

The core components are:

* istiod: the Istio control plane.
* Istio ingress gateway: a deployment of Envoy designed to manage ingress traffic into the mesh.
* Istio egress gateway: a deployment of Envoy designed for managing egress traffic out of the mesh.

The Envoy sidecars are also components of a service mesh, but they do not feature in the installation process. The sidecars are deployed alongside Kubernetes workloads post-installation.

## Istio configuration profiles
[Istio configuration profiles](https://istio.io/latest/docs/setup/additional-setup/config-profiles/) simplify the process of configuring an Istio service mesh.

Each profile configures Istio in a specific way for a particular use case.

The list of profiles includes:

- [ ] minimal: installs only the Istio control plane, no gateway components
- [ ] default: recommended for production deployments, deploys the Istio control plane and an ingress gateway
- [ ] demo: useful for showcasing Istio, for demonstration or learning purposes, and deploys all Istio core components
- [ ] empty: a base profile for custom configurations, often used for [deploying additional, perhaps dedicated gateways](https://istio.io/latest/docs/setup/additional-setup/gateway/)
- [ ] preview: deploys Istio with experimental (preview) features
- [ ] remote: used in the context of installing Istio on a remote cluster (where the control plane resides in another cluster)
Profiles simplify the otherwise tedious task of having to set values for dozens of configuration fields.

The following table shows which Istio core components are included with each configuration profile:

| Profile                | default | demo | minimal | remote | empty | preview |
|------------------------|---------|------|---------|--------|-------|---------|
| Core components        |         |      |         |        |       |         |
| istio-egressgateway    | ✔       |      |         |        |       |         |
| istio-ingressgateway   | ✔       | ✔    |         |        | ✔     |         |
| istiod                 | ✔       | ✔    | ✔       |        |       | ✔       |

## Installation Methods

### The Istio Operator
The [Kubernetes Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) is often used as a mechanism for installing software in Kubernetes.

Installation with the operator works as follows:

Deploy the operator to the cluster with the command istioctl operator init
Apply the IstioOperator resource to the cluster with kubectl
The operator installs Istio on the cluster according to the IstioOperator resource specification
This method has been deprecated because it requires giving the operator controller elevated privileges on the Kubernetes cluster, something that we ought to avoid from a security perspective.

### The Istio CLI
Installation with the Istio CLI is the simplest, and the community-preferred installation method. This method does not have any of the security drawbacks associated with using the Istio Operator.

Installation is performed with the istioctl install subcommand.

This installation method retains the use of the [IstioOperator API](https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/) for configuring Istio.

The simplest way to use this command is together with a named profile, for example:

> istioctl install --set profile=demo

Alternatively, we can supply an IstioOperator custom resource that configures each aspect of Istio in a more fine-grained manner.

The resource is supplied to the CLI command with the -f flag, for example:

> istioctl install -f my-operator-config.yaml

Here is an example operator resource that specifies the default profile, but deviates from it by enabling the egress gateway component, enabling tracing, and configuring proxy logs to standard output:

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: default
  components:
    egressGateways:
    - name: istio-egressgateway
      enabled: true
  meshConfig:
    enableTracing: true
    accessLogFile: /dev/stdout
```

The IstioOperator resource provides many other configuration parameters besides.

In environments that require Kubernetes resources to be audited or otherwise vetted before being applied to a target cluster, Istio provides a mechanism to generate the Kubernetes manifest file that captures all Kubernetes resources that need to be applied to install Istio.

Example:

> istioctl manifest generate -f my-operator-config.yaml

After the audit passes, the manifest file can then be applied with kubectl.

## Helm
Helm charts are a de facto standard for installing software on Kubernetes.

Istio provides three distinct charts for installing Istio in a flexible fashion:

istio/base: Installs shared components such as CRDs. This is necessary for every Istio installation.
istio/istiod: Installs the Istio control plane.
istio/gateway: For installing ingress and egress gateways.

## Which method should I use?
The Istio documentation answers the frequently asked question [Which Istio installation method should I use](https://istio.io/latest/about/faq/#install-method-selection)?, which details the pros and cons of each of the installation methods we described.

The Istio reference documentation devotes an [entire section ](https://istio.io/latest/docs/setup/install/)to the subject of installation, and discusses additional installation-related topics besides what we covered here. We urge you to explore this resource further on your own.
