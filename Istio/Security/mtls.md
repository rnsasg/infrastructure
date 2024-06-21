## Mutual TLS
The communication between workloads in services goes through the Envoy proxies. When a workload sends a request to another workload using mTLS, Istio re-routes the traffic to the sidecar Envoy.

Once the mTLS connection is established, the request is forwarded from the client Envoy proxy to the server-side Envoy proxy. Then, the sidecar Envoy starts an mTLS handshake with the server-side Envoy. During the handshake, the caller does a secure naming check to verify the service account in the server certificate is authorized to run the target service. After the authorization on the server-side, the sidecar forwards the traffic to the workload.

We can change the mTLS behavior in the service’s destination rule. The supported TLS modes are DISABLE (no TLS connection), SIMPLE (originate a TLS connection to the upstream endpoint), MUTUAL (uses mTLS by presenting client certificates for authentication), and ISTIO_MUTUAL (similar to MUTUAL, but uses Istio’s automatically generated certificates for mTLS).

### Permissive mode
A permissive mode is a unique option that allows a service to simultaneously accept plaintext traffic and mTLS traffic. The purpose of this feature is to improve the mTLS onboarding experience.

By default, Istio configures the destination workloads using the permissive mode. Istio tracks the workloads that use the Istio proxies and automatically sends the mTLS traffic to them. If workloads don’t have a proxy, Istio sends plain text traffic.

The server accepts plain text traffic and mTLS traffic without breaking anything when using the permissive mode. The permissive mode gives us time to install and configure sidecars to send mTLS traffic gradually.

Once all workloads have the sidecars installed, we can switch to the strict mTLS mode. To do that, we can create a PeerAuthentication resource. We can prevent non-mutual TLS traffic and require that all communication uses mTLS.

We can create the PeerAuthentication resource and enforce strict mode in each namespace separately at first. Then, we can create a policy in the root namespace (istio-system in our case) that implements the policy globally across the mesh:

```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: istio-system
spec:
  mtls:
    mode: STRICT
```

Additionally, we can also specify the selector field and apply the policy only to specific workloads in the mesh. The example below enables STRICT mode for workloads with the specified label:

```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: my-namespace
spec:
  selector:
    matchLabels:
      app: customers
  mtls:
    mode: STRICT
```
