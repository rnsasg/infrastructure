# Health checks

The /healthcheck/fail can be used to fail inbound health checks. The endpoint /healthcheck/ok is used to revert the effects of the fail endpoint.

Both endpoints require the use of the HTTP health check filter. We might use this for draining the server before shutting it down or when doing complete restarts. When the fail health check option is invoked, all health checks will fail, regardless of their configuration.

