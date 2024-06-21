# Sending envoy logs to google log server

In this lab, we’ll learn how to send Envoy application logs to Google Cloud Logging. We’ll run Envoy on a virtual machine (VM) instance that’s running in the GCP., configuring the Envoy instance to send application logs to cloud logging in GCP. Configuring the Envoy instance will allow us to view Envoy logs in logs explorer, get log analytics, and use other Google cloud features.

To use the log collection, analysis, and other tools in Google Cloud, we need to install the Cloud Logging agent (Ops Agent).

For this demo, we’ll install the Ops Agent on an individual VM. Also, note that other cloud providers might use different logging tools and services.

## Installing Ops agent
In Google Cloud, create a new VM instance in your region. Once VM is created, we can SSH into the instance and install the Ops Agent.

From the VM instance, run the following command to install the Ops Agent:

```shell
curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
sudo bash add-google-cloud-ops-agent-repo.sh --also-install
```

Another option to install the Ops Agent is by following these steps:

1. From the navigation page in GCP, select Monitoring.
2. In the Monitoring navigation page, select Dashboards.
3. In the dashboard table, find and click the VM Instances.
4. Select the checkbox next to an instance that doesn’t have an agent installed (e.g., the Agent column shows Not Detected).
5. Click the Install agents button, and from the window that opens, click the Run in cloud shell button to start the installation.

The command to install the agent will open in the cloud shell. The last thing you need to do is press Enter to start the installation.

Here’s how the command and output of successful installation look in the cloud shell:

```shell
$ :> agents_to_install.csv && \
→ echo '"projects/envoy-project/zones/us-west1-a/instances/envoy-instance","[{""type"":""ops-agent""}]"' >> agents_to_install.csv && \
→ curl -sSO https://dl.google.com/cloudagents/mass-provision-google-cloud-ops-agents.py && \
→ python3 mass-provision-google-cloud-ops-agents.py --file agents_to_install.csv
2021-11-03T19:04:31.577710Z Processing instance: projects/peterjs-project/zones/us-west1-a/instances/some-instance.
---------------------Getting output-------------------------
Progress: |==================================================| 100.0% [1/1] (100.0%) completed; [1/1] (100.0%) succeeded; [0/1] (0.0%) failed;
Instance: projects/envoy-project/zones/us-west1-a/instances/envoy-instance successfully runs ops-agent. See log file in: ./google_cloud_ops_agent_provisioning/20211103-190431_576419/envoy-project_us-west1-a_envoy-instance.log

SUCCEEDED: [1/1] (100.0%)
FAILED: [0/1] (0.0%)
COMPLETED: [1/1] (100.0%)
```

See script log file: ./google_cloud_ops_agent_provisioning/20211103-190431_576419/wrapper_script.log
As installation progresses, the Agent column in the VM instances dashboard will show Pending. Once the agent installation completes, the value changes to Ops Agent, which indicates the Ops Agent is successfully installed.

We can now SSH into the VM instance and install func-e (to run Envoy), create a basic Envoy configuration, and run it so that Envoy application logs will be sent to the cloud logs in GCP.

Installing func-e
To install func-e on the VM, run:

```shell
curl https://func-e.io/install.sh | sudo bash -s -- -b /usr/local/bin
```

We can run func-e --version to check that the installation was successful.

## Sending Envoy application logs to cloud logging
Let’s create a barebones Envoy configuration we’ll use in this lab:

```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          http_filters:
          - name: envoy.filters.http.router
            typed_config: 
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          route_config:
            name: my_first_route
            virtual_hosts:
            - name: direct_response_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                direct_response:
                  status: 200
                  body:
                    inline_string: "200"
```
Save the above YAML to 6-lab-3-gcp-logging.yaml.

Let’s configure the Ops Agent and create a new receiver for Envoy that will describe how to retrieve the logs.

```yaml
logging:
  receivers:
    envoy:
      type: files
      include_paths:
        - /var/log/envoy.log
  service:
    pipelines:
      default_pipeline:
        receivers: [envoy]
```

Save the above contents to /etc/google-cloud-ops-agent/config.yaml file on the VM instance. To restart the Ops Agent, run sudo service google-cloud-ops-agent restart.

With the Ops Agent using the new configuration, we can run Envoy and tell it to write the logs to /var/log/envoy.log file, where the agent will pick it up:

sudo func-e run -c 6-lab-3-gcp-logging.yaml --log-path /var/log/envoy.log
Next, we can click Logging and then the Logs explorer in GCP to see the logs from the Envoy instance running on the VM.

Envoy logs in Log Explorer in GCP
Envoy logs in Log Explorer in GCP