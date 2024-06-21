## Install Istio
We will install the Istio 1.10.3 on the cluster using the GetMesh CLI.

### Download GetMesh CLI:
```shell
curl -sL https://istio.tetratelabs.io/getmesh/install.sh | bash
Install Istio:
getmesh istioctl install --set profile=demo
```

When the installation completes, label the default namespace with istio-injection=enabled label:

```shell
kubectl label namespace default istio-injection=enabled
```