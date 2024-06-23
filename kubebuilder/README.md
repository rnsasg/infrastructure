```shell
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder && mv kubebuilder /usr/local/bin/
```

# Initial Setup

## Install Kubebuilder

```shell
mkdir cronjob-tutorial
cd cronjob-tutorial
kubebuilder init --domain tutorial.kubebuilder.io
```

## Install Kubectl

```shell
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

## Install kind (Kubernetes in Docker):

```shell
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
chmod +x kind
sudo mv kind /usr/local/bin/
```

# Create a Kubernetes Cluster with kind:

```shell
kind create cluster
```

## Step 2: Initialize a New Project with Kubebuilder

### Create a new directory for your project and initialize it:

```shell
mkdir cronjob-tutorial
cd cronjob-tutorial
kubebuilder init --domain tutorial.kubebuilder.io
```

### Create an API:

```shell
kubebuilder create api --group batch --version v1 --kind CronJob
```




## Clean Up

### Delete Kubernetes Resources

kubectl delete -f config/crd/bases
kubectl delete -f config/default

<!-- kubectl delete namespace <your-namespace>
kubectl delete -f <path-to-your-resources> -->
kind delete cluster
make clean

## Optionally, you can manually remove specific directories or files:

rm -rf bin/
rm -rf config/crd/bases
rm -rf config/default
rm -rf hack/

## Remove Go Modules and Dependencies

rm -rf vendor
rm go.sum
rm go.mod


## Reinitialize the Kubebuilder Project

kubebuilder init --domain tutorial.kubebuilder.io
kubebuilder create api --group batch --version v1 --kind CronJob

## Reinstall Dependencies
go mod tidy
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
go install github.com/go-delve/delve/cmd/dlv@latest

## Verify Environment and Tools
kubebuilder version
kubectl version --client
kind version
go version

## Rebuild and Redeploy
make all
kind create cluster
make install
make run




