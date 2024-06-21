# Peer and Request Authentication
Istio provides two types of authentication: **peer authentication** and **request authentication**.

## Peer authentication
Peer authentication is used for service-to-service authentication to verify the client that’s making the connection.

When two services try to communicate, mutual TLS requires both to provide certificates so both parties know who they are talking to. If we want to enable strict mutual TLS between services, we can use the PeerAuthentication resource to set the mTLS mode to STRICT.

Using the PeerAuthentication resource, we can turn on mutual TLS (mTLS) across the mesh without making code changes.

However, Istio also supports a graceful mode where we can opt into mutual TLS one workload or namespace at the time. This mode is called permissive mode.

Permissive mode is enabled by default when you install Istio. With permissive mode enabled, if a client tries to connect to me via mutual TLS, I’ll serve mutual TLS. If the client doesn’t use mutual TLS, I can respond in plain text as well. I am permitting the client to do mTLS or not. Using this mode, you can gradually roll out mutual TLS across your mesh.

To recap, PeerAuthentication talks about how the workloads or services communicate. It isn’t saying anything about end-users. So how could we authenticate users?

## Request authentication
The request authentication (RequestAuthentication resource) verifies the credential attached to the request, and we use it for end-user authentication.

The request-level authentication is done with [JSON Web Token](https://jwt.io/) (JWT) validation. Istio supports any OpenID Connect providers, such as Auth0, Firebase or Google Auth, Keycloak, ORY Hydra. So just like we used SPIFFE identity to authenticate the services, we can use JWT tokens to authenticate users.
