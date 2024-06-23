## References

1. https://medium.com/developingnodes/mastering-kubernetes-operators-your-definitive-guide-to-starting-strong-70ff43579eb9

## Steps

```shell
kubebuilder init --domain example.com --repo=github.com/rnsasg/infrastructure/kubebuilder/ConfigmapSync
kubebuilder create api --group apps --version v1 --kind ConfigMapSync
kind create cluster --name config
make install
make run

kubectl create namespace source 
kubectl create namespace destination
kubectl apply -f configmap.yaml
```


```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sourcecfg
  namespace: source
data:
  # Configuration values
  app.properties: |
    database.url=jdbc:mysql://localhost:3306/mydb
    database.username=root
    database.password=secret
  log.properties: |
    logging.level=INFO
    logging.file=/var/log/app.log
```

```yaml
apiVersion: apps.example.com/v1
kind: ConfigMapSync
metadata:
  labels:
    app.kubernetes.io/name: configmapsync
    app.kubernetes.io/instance: configmapsync-sample
    app.kubernetes.io/part-of: configmapsync
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: configmapsync
  name: configmapsyncs
spec:
  sourceNamespace: source
  destinationNamespace: destination
  configMapName: sourcecfg
```

