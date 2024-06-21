# Certificate creation and rotation
For every workload in the mesh, Istio provisions an X.509 certificate. An agent called pilot-agent runs next to each Envoy proxy and works together with the control plane (istiod) to automate key and certificate rotation.

<img src="../images/security_1.png"></img>

There are three parts in play when creating identities at runtime:

* Citadel (part of the control plane)
* Istio Agent
* Envoy’s Secret Discovery Service (SDS)

**Istio Agent** works together with Envoy sidecars and helps them connect to the mesh by securely passing them configuration and secrets. Even though Istio agent runs in each pod, we consider it a part of the control plane.

Secret Discovery Service (SDS) simplifies certificate management. Without SDS, the certificates must be created as secrets and then mounted into the proxy container’s filesystem. We must update the secrets and re-deploy the proxies when certificates expire because Envoy doesn’t reload certificates dynamically from the disk. When using SDS, an SDS server pushes certificates to Envoy instances. Whenever certificates expire, SDS pushes renewed certificates, and Envoy can use them right away. There is no need to re-deploy the proxies or interrupt traffic. In Istio, the Istio Agent acts as an SDS server and implements the secret discovery service interface.

Every time we create a new service account, Citadel creates a SPIFFE identity for it. Whenever we schedule a workload, the Pilot configures its sidecar with initialization information that includes the workload’s service account.

When the Envoy proxy next to the workload starts, it contacts the Istio Agent and tells the workload’s service account. The agent validates the instance, generates a CSR (Certificate Signing Request), sends the CSR and proof of the workload’s service account (in Kubernetes, the pod’s service account JWT) to Citadel. Citadel will perform authentication and authorization and respond with a signed X.509 certificate. The Istio agent takes the response from Citadel, caches the key and certificate in memory, and serves it to Envoy via SDS over a Unix Domain Socket. Storing the key in memory is more secure than storing it on disks; Istio never writes any key to disk as part of its operation when using SDS. Istio agent also periodically refreshes the credential by retrieving any new SVIDs (SPIFFE Verifiable Identity Document) from Citadel before the current credentials expire.

<img src="../images/security_2.png"></img>

> SVID is a document a workload can use to prove its identity to a resource or caller. It has to be signed by an authority and contain a SPIFFE ID, which represents the service’s identity presenting it, for example, spiffe://clusterlocal/ns/my-namespace/sa/my-sa.

This solution is scalable as each component in the flow is only responsible for a portion of the work. For example, Envoy to expire the certificates, Istio agent to generate private key and CSR, and Citadel to authorize and sign the certificates.