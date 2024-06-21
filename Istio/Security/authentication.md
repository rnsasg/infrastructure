## Authentication
To explain what authentication or authn is, we’ll start with the question access control tries to answer: can a subject perform an action on an object?

If we translate the above question to the Istio and Kubernetes world, it would be “Can Service X perform an action on Service Y?”

The three key components from this question are principal, action, and object.

Both the principal and object are services in Kubernetes. Assuming we are talking about HTTP, the action is a GET request, a POST, a PUT, and so on.

Authentication is all about the principal (or the service’s identity in our case). Authentication is the act of validating some credential and ensuring that the credential is both valid and trusted. Once the authentication gets performed, we have an authenticated principal. Next time you travel and present your passport or an ID to the customs officer, they authenticate it, ensuring your credential (passport or ID) is valid and trusted.

In Kubernetes, each workload gets assigned a unique identity that it uses to communicate with every other workload – the identity gets provided to workloads in the form of service accounts. A service account is the identity Pods present in the runtime.

Istio uses the X.509 certificate from the service account, and it creates a new identity according to the spec called SPIFFE (Secure Production Identity Framework for Everyone).

The identity in the certificate gets encoded in the Subject alternate name field of the certificate. It looks like this:

> spiffe://cluster.local/ns/<pod namespace>/sa/<pod service account>

The Envoy proxies are modified so when they do the TLS handshake, they’ll also do the portion required by the SPIFFE validation (check the SAN field) to get a valid SPIFFE identity. After this process, we can use the authenticated principals for policy.
