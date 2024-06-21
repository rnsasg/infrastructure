# Lab: Installing Istio
In this lab, we will install Istio on a Kubernetes cluster using the Istio CLI.

## Prerequisites
To install Istio, we will need a running instance of a Kubernetes cluster. All cloud providers have a managed Kubernetes cluster offering we can use to install Istio service mesh.

We can also run a Kubernetes cluster locally on our computer using one of the following platforms:

* [Minikube](https://istio.io/latest/docs/setup/platform-setup/minikube/)
* [Docker Desktop](https://istio.io/latest/docs/setup/platform-setup/docker/)
* [kind](https://istio.io/latest/docs/setup/platform-setup/kind/)
* [MicroK8s](https://istio.io/latest/docs/setup/platform-setup/microk8s/)

When using a local Kubernetes cluster, ensure your computer meets the minimum requirements for Istio installation (e.g., 16384 MB RAM and 4 CPUs). Also, ensure the Kubernetes cluster version is v1.23.0 or higher.

## Kubernetes CLI
If you need to install the Kubernetes CLI, follow these instructions.

We can run kubectl version to check if the CLI got installed. You should see the output similar to this one:

```shell
$ kubectl version
Client Version: v1.26.1
Kustomize Version: v4.5.7
Server Version: v1.25.7+k3s1
```

## Download Istio
The first step to installing Istio is downloading the Istio CLI (istioctl), installation manifests, samples, and tools.

The easiest way to install the latest version is to use the downloadIstio script. Open a terminal window and open the folder where you want to download Istio, then run the download script:

```shell
$ curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.17.2 sh -
```

Istio release is downloaded and unpacked to the folder called istio-1.17.2. To access istioctl we should add it to the path:

```shell
$ cd istio-1.17.2
$ export PATH=$PWD/bin:$PATH
```

To check istioctl is on the path, run istioctl version. You should see an output like this:

```shell
$ istioctl version
no running Istio pods in "istio-system"
1.17.2
```

# Install Istio
Istio supports multiple configuration profiles. The difference between the profiles is in components that get installed.

```shell
$ istioctl profile list
Istio configuration profiles:
    default
    demo
    empty
    external
    minimal
    openshift
    preview
    remote
```

The recommended profile for production deployments is the default profile. We will be installing the demo profile as it contains all core components, has a high level of tracing and logging enabled, and is meant for learning about different Istio features.

We can also start with the minimal component and individually install other features, like ingress and egress gateway, later.

We can install Istio using the istioctl install command.

A simple way to set the profile we wish to use is with the --set flag. For example:

> $ istioctl install --set profile=demo
Alternatively, we can draft an IstioOperator resource.

Create a file called demo-profile.yaml with the following contents:

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: demo
```

To install Istio according to the above configuration, we use the -f flag:

> $ istioctl install -f demo-profile.yaml

In both cases we’ll get prompted to proceed the installation and once we confirm, the Istio service mesh will be deployed.

To check the deployed resource, we can look at the status of the Pods in the istio-system namespace:

```shell
$ kubectl get po -n istio-system
NAME                                   READY   STATUS    RESTARTS   AGE
istiod-64848b6c78-r5zpx                1/1     Running   0          27s
istio-ingressgateway-f56888458-g6jfj   1/1     Running   0          19s
istio-egressgateway-85649899f8-rdb44   1/1     Running   0          19s
```

## Updating the Istio installation
To update the installation, we can modify the existing IstioOperator resource and apply it to the cluster. For example, if we wanted to remove the egress gateway, we could update the IstioOperator resource like this:

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: demo
  components:
    egressGateways:
    - name: istio-egressgateway
      enabled: false
```

Here is the reference to the IstioOperator options.

Save the above YAML to iop-demo-no-egress.yaml and apply it using istioctl install -f iop-demo-no-egress.yaml.

Just like before we’ll get prompted to proceed with the installation. You’ll notice that the egress gateway is no longer in the list of installed pods in the istio-system namespace.

Another option for updating the Istio installation is to create separate IstioOperator resources. That way, we can have a resource for the base installation and separately apply different operators using an empty installation profile. For example, here’s how we could create a separate IstioOperator resource that only deploys an internal ingress gateway:

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: internal-gateway-only
spec:
  profile: empty
  components:
    ingressGateways:
      - namespace: some-namespace
        name: ilb-gateway
        enabled: true
        label:
          istio: ilb-gateway
        k8s:
          serviceAnnotations:
            networking.gke.io/load-balancer-type: "Internal"
```

## Enable sidecar injection
As we’ve learned in the previous section, service mesh needs the sidecar proxies running alongside each application.

To inject the sidecar proxy into an existing Kubernetes deployment, we can use kube-inject sub-command in the istioctl command.

However, we can also enable automatic sidecar injection on any Kubernetes namespace. If we label the namespace with the istio-injection=enabled label, Istio control plane will monitor that namespace for new Kubernetes deployments, it will automatically intercept them and inject Envoy sidecars into each pod.

Let’s enable automatic sidecar injection on the default namespace by setting the istio-injection label:

```shell
$ kubectl label namespace default istio-injection=enabled
namespace/default labeled
```

To check the namespace is labeled, run the command below. The default namespace should be the only one with the value enabled.

```shell
$ kubectl get namespace -L istio-injection
NAME              STATUS   AGE    ISTIO-INJECTION
default           Active   114m   enabled
istio-system      Active   29m
kube-node-lease   Active   114m
kube-public       Active   114m
kube-system       Active   114m
```

We can now try creating a Deployment in the default namespace and observe the injected proxy. We will create a deployment called my-nginx with a single container using image nginx:

```shell
$ kubectl create deploy my-nginx --image=nginx
deployment.apps/my-nginx created
```

If we look at the Pods, you will notice there are two containers in the Pod:

```shell
$ kubectl get po
NAME                        READY   STATUS    RESTARTS   AGE
my-nginx-6b74b79f57-gh5fp   2/2     Running   0          62s
```

Similarly, describing the Pod shows Kubernetes created both an nginx container and an istio-proxy container:

```shell
$ kubectl describe po my-nginx-6b74b79f57-gh5fp
...
Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  70s   default-scheduler  Successfully assigned default/my-nginx-6b74b79f57-gh5fp to gke-cluster-1-default-pool-c2743eca-sts7
  Normal  Pulled     69s   kubelet            Container image "docker.io/istio/proxyv2:1.17.2" already present on machine
  Normal  Created    69s   kubelet            Created container istio-init
  Normal  Started    69s   kubelet            Started container istio-init
  Normal  Pulling    68s   kubelet            Pulling image "nginx"
  Normal  Pulled     64s   kubelet            Successfully pulled image "nginx" in 4.334525037s
  Normal  Created    63s   kubelet            Created container nginx
  Normal  Started    63s   kubelet            Started container nginx
  Normal  Pulled     63s   kubelet            Container image "docker.io/istio/proxyv2:1.17.2" already present on machine
  Normal  Created    63s   kubelet            Created container istio-proxy
  Normal  Started    63s   kubelet            Started container istio-proxy
```

To remove the deployment, run the delete command:

```shell
$ kubectl delete deployment my-nginx
deployment.apps "my-nginx" deleted
```

## Uninstalling Istio
To remove the installation, we can use the uninstall command:

```shell
$ istioctl uninstall --purge
```

## iopdemonoegress-230608-114724.yaml

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: demo
  components:
    egressGateways:
    - name: istio-egressgateway
      enabled: false
```

## internalgateway-230608-114724.yaml

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: internal-gateway-only
spec:
  profile: empty
  components:
    ingressGateways:
      - namespace: some-namespace
        name: ilb-gateway
        enabled: true
        label:
          istio: ilb-gateway
        k8s:
          serviceAnnotations:
            networking.gke.io/load-balancer-type: "Internal"
```

## demoprofile-230608-114724.yaml

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: demo
```


