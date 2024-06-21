# Get Mesh

Istio is one of the most popular and fast-growing open-source projects. Its release schedule can be very aggressive for enterprise lifecycle and change management practices. GetMesh's versions of Istio are actively supported for security patches and other bug updates and have a much longer support life than provided by upstream Istio. GetMesh addresses this concern by testing all Istio versions against different Kubernetes distributions for functional integrity.

Some service mesh customers need to support elevated security requirements. GetMesh addresses compliance by offering two flavors of the Istio distribution:

* tetrate distro that tracks the upstream Istio and may have additional patches applied
* tetratefips distro that is a FIPS-compliant version of the `tetrate` flavor

## How to get started?
The first step is to download GetMesh CLI. You can install GetMesh on macOS and Linux platforms. We can use the following command to download the latest version of GetMesh and certified Istio:

> curl -sL https://istio.tetratelabs.io/getmesh/install.sh | bash

We can run the version command to ensure GetMesh got successfully installed. For example:

```shell
$ getmesh version
getmesh version: 1.1.5
active istioctl: 1.17.2-tetrate-v0
no running Istio pods in "istio-system"
1.17.2-tetrate-v0
```
The version command outputs the version of GetMesh, the version of active Istio CLI, and versions of Istio installed on the Kubernetes cluster (if installed).

## Installing Istio with GetMesh
GetMesh communicates with the active Kubernetes cluster from the Kubernetes config file.

To install the demo profile of Istio on a currently active Kubernetes cluster, we can use the getmesh istioctl command like this:

> getmesh istioctl install --set profile=demo
The command will check the cluster to make sure it’s ready to install Istio, and once you confirm, the installer will proceed to install Istio using the selected profile.

If we check the version now, you’ll notice that the output shows the control plane and data plane versions.

## Validating the configuration
The config-validate command allows you to validate the current config and any YAML manifests that are not yet applied.

The command invokes validations using external sources such as upstream Istio validations, Kiali libraries, and GetMesh custom configuration checks.

Here’s an example of the command output if there are no namespaces labeled for Istio injection:

```shell
$ getmesh config-validate
Running the config validator. This may take some time...

2021-08-02T19:20:33.873244Z     info    klog    Throttling request took 1.196458809s, request: GET:https://35.185.226.9/api/v1/namespaces/istio-system/configmaps/istio[]
NAMESPACE       NAME    RESOURCE TYPE   ERROR CODE      SEVERITY        MESSAGE                                     
default         default Namespace       IST0102         Info            The namespace is not enabled for Istio injection. Run 'kubectl label namespace default istio-injection=enabled' to enable it, or 'kubectl label namespace default istio-injection=disabled' to explicitly mark it as not needing injection.
```

The error codes of the found issues are prefixed by 'IST' or 'KIA'. For the detailed explanation, please refer to

```shell
- https://istio.io/latest/docs/reference/config/analysis/ for 'IST' error codes
- https://kiali.io/documentation/latest/validations/ for 'KIA' error codes
```

Similarly, you can also pass in a YAML file to validate it, before deploying it to the cluster. For example:

> $ getmesh config-validate my-resources.yaml

Managing multiple Istio CLIs
We can use the show command to list the currently downloaded versions of Istio:

getmesh show
Here’s how the output might look like:

1.17.2-tetrate-v0 (Active)
If the version we’d like to use is not available locally on the computer, we can use the getmesh list command to list all trusted Istio versions:

```shell
$ getmesh list
ISTIO VERSION   FLAVOR     FLAVOR VERSION    K8S VERSIONS
   *1.17.2       tetrate          0        1.23,1.24,1.25,1.26
   1.17.2      tetratefips        0        1.23,1.24,1.25,1.26
   1.17.2         istio           0        1.23,1.24,1.25,1.26
   1.17.1        tetrate          0        1.23,1.24,1.25,1.26
   1.17.1      tetratefips        0        1.23,1.24,1.25,1.26
   1.17.1         istio           0        1.23,1.24,1.25,1.26
   1.17.0        tetrate          0        1.23,1.24,1.25,1.26
   1.17.0         istio           0        1.23,1.24,1.25,1.26
   1.16.4        tetrate          0        1.22,1.23,1.24,1.25
   1.16.4      tetratefips        0        1.22,1.23,1.24,1.25
   1.16.4         istio           0        1.22,1.23,1.24,1.25
   1.16.3        tetrate          0        1.22,1.23,1.24,1.25
   1.16.3      tetratefips        0        1.22,1.23,1.24,1.25
   1.16.3         istio           0        1.22,1.23,1.24,1.25
   1.16.2        tetrate          0        1.22,1.23,1.24,1.25
   1.16.2      tetratefips        0        1.22,1.23,1.24,1.25
   1.16.2         istio           0        1.22,1.23,1.24,1.25
   1.16.1        tetrate          0        1.22,1.23,1.24,1.25
   1.16.1      tetratefips        0        1.22,1.23,1.24,1.25
   1.16.1         istio           0        1.22,1.23,1.24,1.25
   1.16.0        tetrate          0        1.22,1.23,1.24,1.25
   1.16.0      tetratefips        0        1.22,1.23,1.24,1.25
   1.16.0         istio           0        1.22,1.23,1.24,1.25
   1.15.7        tetrate          0        1.22,1.23,1.24,1.25
   1.15.7      tetratefips        0        1.22,1.23,1.24,1.25
   1.15.7         istio           0        1.22,1.23,1.24,1.25
   ...
```

To fetch a specific version (let’s say the 1.16.4 tetratefips flavor), we can use the fetch command:

> getmesh fetch --version 1.16.4 --flavor tetratefips --flavor-version 0

When the above command completes, GetMesh sets the fetched Istio CLI version as the active version of Istio CLI. For example, running the show command now shows that the tetratefips version 1.16.4 is active:

```shell
$ getmesh show
1.16.4-tetratefips-v0 (Active)
1.17.2-tetrate-v0
```

Similarly, if we run getmesh istioctl version, we’ll notice the version of Istio CLI that’s in use:

```shell
$ getmesh istioctl version
client version: 1.16.4-tetratefips-v0
control plane version: 1.17.2
data plane version: 1.17.2-tetrate-v0 (2 proxies)
To switch to a different version of the Istio CLI, we can run the getmesh switch command:

getmesh switch --name 1.17.2-tetrate-v0
```
