# Use the official Envoy Docker image
FROM envoyproxy/envoy:v1.21.0

# Copy the Envoy configuration file
COPY envoy.yaml /etc/envoy/envoy.yaml

# Expose ports
EXPOSE 10000 9901

# Run Envoy with the specified configuration file
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml", "--log-level", "info"]


