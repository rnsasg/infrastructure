## Envoy Example

Let’s take the Web frontend and customer service as an example and see how Envoy determines where to send the request from the web frontend to the customer service (customers.default.svc.cluster.local). Using the istioctl proxy-config command, we can list all listeners of the web frontend pod.

```shell
$ istioctl proxy-config listeners web-frontend-64455cd4c6-p6ft2
ADDRESS      PORT  MATCH   DESTINATION
10.124.0.10  53    ALL     Cluster: outbound|53||kube-dns.kube-system.svc.cluster.local
0.0.0.0      80    ALL     PassthroughCluster
10.124.0.1   443   ALL     Cluster: outbound|443||kubernetes.default.svc.cluster.local
10.124.3.113 443   ALL     Cluster: outbound|443||istiod.istio-system.svc.cluster.local
10.124.7.154 443   ALL     Cluster: outbound|443||metrics-server.kube-system.svc.cluster.local
10.124.7.237 443   ALL     Cluster: outbound|443||istio-egressgateway.istio-system.svc.cluster.local
10.124.8.250 443   ALL     Cluster: outbound|443||istio-ingressgateway.istio-system.svc.cluster.local
10.124.3.113 853   ALL     Cluster: outbound|853||istiod.istio-system.svc.cluster.local
0.0.0.0      8383  ALL     PassthroughCluster
0.0.0.0      15001 ALL     PassthroughCluster
0.0.0.0      15006 ALL     Inline Route: /*
0.0.0.0      15010 ALL     PassthroughCluster
10.124.3.113 15012 ALL     Cluster: outbound|15012||istiod.istio-system.svc.cluster.local
0.0.0.0      15014 ALL     PassthroughCluster
0.0.0.0      15021 ALL     Non-HTTP/Non-TCP
10.124.8.250 15021 ALL     Cluster: outbound|15021||istio-ingressgateway.istio-system.svc.cluster.local
0.0.0.0      15090 ALL     Non-HTTP/Non-TCP
10.124.7.237 15443 ALL     Cluster: outbound|15443||istio-egressgateway.istio-system.svc.cluster.local
10.124.8.250 15443 ALL     Cluster: outbound|15443||istio-ingressgateway.istio-system.svc.cluster.local
10.124.8.250 31400 ALL     Cluster: outbound|31400||istio-ingressgateway.istio-system.svc.cluster.local
```

The request from the web frontend to customers is an outbound HTTP request to port 80. This means that it gets handed off to the 0.0.0.0:80 virtual listener. We can use Istio CLI to filter the listeners by address and port. You can add the -o json to get a JSON representation of the listener:

```shell
$ istioctl proxy-config listeners web-frontend-58d497b6f8-lwqkg --address 0.0.0.0 --port 80 -o json
...
"rds": {
   "configSource": {
      "ads": {},
      "resourceApiVersion": "V3"
   },
   "routeConfigName": "80"
},
...
```

The listener uses RDS (Route Discovery Service) to find the route configuration (80 in our case). Routes are attached to listeners and contain rules that map **virtual hosts** to clusters. This allows us to create traffic routing rules because Envoy can look at headers or paths (the request metadata) and route traffic.

A route selects a **cluster**. A cluster is a group of similar upstream hosts that accept traffic - it’s a collection of endpoints. For example, the collection of all instances of the Web Frontend service is a cluster. We can configure resiliency features within a cluster, such as circuit breakers, outlier detection, and TLS config.

Using the routes command, we can get the route details by filtering all routes by the name:

```shell
$ istioctl proxy-config routes  web-frontend-58d497b6f8-lwqkg --name 80 -o json

[
    {
        "name": "80",
        "virtualHosts": [
            {
                "name": "customers.default.svc.cluster.local:80",
                "domains": [
                    "customers.default.svc.cluster.local",
                    "customers.default.svc.cluster.local:80",
                    "customers",
                    "customers:80",
                    "customers.default.svc.cluster",
                    "customers.default.svc.cluster:80",
                    "customers.default.svc",
                    "customers.default.svc:80",
                    "customers.default",
                    "customers.default:80",
                    "10.124.4.23",
                    "10.124.4.23:80"
                ],
                ],
                "routes": [
                    {
                        "match": {
                            "prefix": "/"
                        },
                        "route": {
                            "cluster": "outbound|80|v1|customers.default.svc.cluster.local",
                            "timeout": "0s",
                            "retryPolicy": {
                                "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                "numRetries": 2,
                                "retryHostPredicate": [
                                    {
                                        "name": "envoy.retry_host_predicates.previous_hosts"
                                    }
                                ],
                                "hostSelectionRetryMaxAttempts": "5",
                                "retriableStatusCodes": [
                                    503
                                ]
                            },
                            "maxGrpcTimeout": "0s"
                        },
...
```

The route 80 configuration has a virtual host for each service. However, because our request is being sent to customers.default.svc.cluster.local, Envoy selects the virtual host (customers.default.svc.cluster.local:80) that matches one of the domains.

Once the domain is matched, Envoy looks at the routes and picks the first one that matches the request. Since we don’t have any special routing rules defined, it matches the first (and only) route that’s defined and instructs Envoy to send the request to the cluster named outbound|80|v1|customers.default.svc.cluster.local.

Note the v1 in the cluster’s name is because we have a DestinationRule deployed that creates the v1 subset. If there are no subsets for a service, that part if left blank: outbound|80||customers.default.svc.cluster.local.

Now that we have the cluster name, we can look up more details. To get an output that clearly shows the FQDN, port, subset and other information, you can omit the -o json flag:

```shell
$ istioctl proxy-config cluster web-frontend-58d497b6f8-lwqkg --fqdn customers.default.svc.cluster.local
SERVICE FQDN                            PORT     SUBSET     DIRECTION     TYPE     DESTINATION RULE
customers.default.svc.cluster.local     80       -          outbound      EDS      customers.default
customers.default.svc.cluster.local     80       v1         outbound      EDS      customers.default
```

Finally, using the cluster name, we can look up the actual endpoints the request will end up at:

```shell
$ istioctl proxy-config endpoints  web-frontend-58d497b6f8-lwqkg --cluster "outbound|80|v1|customers.default.svc.cluster.local"
ENDPOINT            STATUS      OUTLIER CHECK     CLUSTER
10.120.0.4:3000     HEALTHY     OK                outbound|80|v1|customers.default.svc.cluster.local
```

The endpoint address equals the pod IP where the Customer application is running. If we scale the customers deployment, additional endpoints show up in the output, like this:

```shell
$ istioctl proxy-config endpoints web-frontend-58d497b6f8-lwqkg --cluster "outbound|80|v1|customers.default.svc.cluster.local"
ENDPOINT            STATUS      OUTLIER CHECK     CLUSTER
10.120.0.4:3000     HEALTHY     OK                outbound|80|v1|customers.default.svc.cluster.local
10.120.3.2:3000     HEALTHY     OK                outbound|80|v1|customers.default.svc.cluster.local
```

We can also visualize the above flow using the figure below.
<img src=""></img>

## files

### listeners-210119-085816.json

```json
[
    {
        "name": "0.0.0.0_80",
        "address": {
            "socketAddress": {
                "address": "0.0.0.0",
                "portValue": 80
            }
        },
        "filterChains": [
            {
                "filterChainMatch": {
                    "applicationProtocols": [
                        "http/1.0",
                        "http/1.1",
                        "h2c"
                    ]
                },
                "filters": [
                    {
                        "name": "envoy.filters.network.http_connection_manager",
                        "typedConfig": {
                            "@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
                            "statPrefix": "outbound_0.0.0.0_80",
                            "rds": {
                                "configSource": {
                                    "ads": {},
                                    "resourceApiVersion": "V3"
                                },
                                "routeConfigName": "80"
                            },
                            "httpFilters": [
                                {
                                    "name": "istio.metadata_exchange",
                                    "typedConfig": {
                                        "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
                                        "typeUrl": "type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm",
                                        "value": {
                                            "config": {
                                                "configuration": {
                                                    "@type": "type.googleapis.com/google.protobuf.StringValue",
                                                    "value": "{}\n"
                                                },
                                                "vm_config": {
                                                    "code": {
                                                        "local": {
                                                            "inline_string": "envoy.wasm.metadata_exchange"
                                                        }
                                                    },
                                                    "runtime": "envoy.wasm.runtime.null"
                                                }
                                            }
                                        }
                                    }
                                },
                                {
                                    "name": "istio.alpn",
                                    "typedConfig": {
                                        "@type": "type.googleapis.com/istio.envoy.config.filter.http.alpn.v2alpha1.FilterConfig",
                                        "alpnOverride": [
                                            {
                                                "alpnOverride": [
                                                    "istio-http/1.0",
                                                    "istio"
                                                ]
                                            },
                                            {
                                                "upstreamProtocol": "HTTP11",
                                                "alpnOverride": [
                                                    "istio-http/1.1",
                                                    "istio"
                                                ]
                                            },
                                            {
                                                "upstreamProtocol": "HTTP2",
                                                "alpnOverride": [
                                                    "istio-h2",
                                                    "istio"
                                                ]
                                            }
                                        ]
                                    }
                                },
                                {
                                    "name": "envoy.filters.http.cors",
                                    "typedConfig": {
                                        "@type": "type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors"
                                    }
                                },
                                {
                                    "name": "envoy.filters.http.fault",
                                    "typedConfig": {
                                        "@type": "type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault"
                                    }
                                },
                                {
                                    "name": "istio.stats",
                                    "typedConfig": {
                                        "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
                                        "typeUrl": "type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm",
                                        "value": {
                                            "config": {
                                                "configuration": {
                                                    "@type": "type.googleapis.com/google.protobuf.StringValue",
                                                    "value": "{\n  \"debug\": \"false\",\n  \"stat_prefix\": \"istio\"\n}\n"
                                                },
                                                "root_id": "stats_outbound",
                                                "vm_config": {
                                                    "code": {
                                                        "local": {
                                                            "inline_string": "envoy.wasm.stats"
                                                        }
                                                    },
                                                    "runtime": "envoy.wasm.runtime.null",
                                                    "vm_id": "stats_outbound"
                                                }
                                            }
                                        }
                                    }
                                },
                                {
                                    "name": "envoy.filters.http.router",
                                    "typedConfig": {
                                        "@type": "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router"
                                    }
                                }
                            ],
                            "tracing": {
                                "clientSampling": {
                                    "value": 100
                                },
                                "randomSampling": {
                                    "value": 100
                                },
                                "overallSampling": {
                                    "value": 100
                                }
                            },
                            "streamIdleTimeout": "0s",
                            "accessLog": [
                                {
                                    "name": "envoy.access_loggers.file",
                                    "typedConfig": {
                                        "@type": "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog",
                                        "path": "/dev/stdout",
                                        "format": "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \"%DYNAMIC_METADATA(istio.mixer:status)%\" \"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n"
                                    }
                                }
                            ],
                            "useRemoteAddress": false,
                            "generateRequestId": true,
                            "upgradeConfigs": [
                                {
                                    "upgradeType": "websocket"
                                }
                            ],
                            "normalizePath": true
                        }
                    }
                ]
            },
            {
                "filterChainMatch": {},
                "filters": [
                    {
                        "name": "istio.stats",
                        "typedConfig": {
                            "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
                            "typeUrl": "type.googleapis.com/envoy.extensions.filters.network.wasm.v3.Wasm",
                            "value": {
                                "config": {
                                    "configuration": {
                                        "@type": "type.googleapis.com/google.protobuf.StringValue",
                                        "value": "{\n  \"debug\": \"false\",\n  \"stat_prefix\": \"istio\"\n}\n"
                                    },
                                    "root_id": "stats_outbound",
                                    "vm_config": {
                                        "code": {
                                            "local": {
                                                "inline_string": "envoy.wasm.stats"
                                            }
                                        },
                                        "runtime": "envoy.wasm.runtime.null",
                                        "vm_id": "tcp_stats_outbound"
                                    }
                                }
                            }
                        }
                    },
                    {
                        "name": "envoy.filters.network.tcp_proxy",
                        "typedConfig": {
                            "@type": "type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy",
                            "statPrefix": "PassthroughCluster",
                            "cluster": "PassthroughCluster",
                            "accessLog": [
                                {
                                    "name": "envoy.access_loggers.file",
                                    "typedConfig": {
                                        "@type": "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog",
                                        "path": "/dev/stdout",
                                        "format": "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \"%DYNAMIC_METADATA(istio.mixer:status)%\" \"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n"
                                    }
                                }
                            ]
                        }
                    }
                ],
                "name": "PassthroughFilterChain"
            }
        ],
        "deprecatedV1": {
            "bindToPort": false
        },
        "listenerFilters": [
            {
                "name": "envoy.filters.listener.tls_inspector",
                "typedConfig": {
                    "@type": "type.googleapis.com/envoy.extensions.filters.listener.tls_inspector.v3.TlsInspector"
                }
            },
            {
                "name": "envoy.filters.listener.http_inspector",
                "typedConfig": {
                    "@type": "type.googleapis.com/envoy.extensions.filters.listener.http_inspector.v3.HttpInspector"
                }
            }
        ],
        "listenerFiltersTimeout": "5s",
        "continueOnListenerFiltersTimeout": true,
        "trafficDirection": "OUTBOUND"
    }
]
```


### routes-210119-085816.json

```json
[
    {
        "name": "80",
        "virtualHosts": [
            {
                "name": "allow_any",
                "domains": [
                    "*"
                ],
                "routes": [
                    {
                        "name": "allow_any",
                        "match": {
                            "prefix": "/"
                        },
                        "route": {
                            "cluster": "PassthroughCluster",
                            "timeout": "0s",
                            "maxGrpcTimeout": "0s"
                        }
                    }
                ],
                "includeRequestAttemptCount": true
            },
            {
                "name": "customers.default.svc.cluster.local:80",
                "domains": [
                    "customers.default.svc.cluster.local",
                    "customers.default.svc.cluster.local:80",
                    "customers",
                    "customers:80",
                    "customers.default.svc.cluster",
                    "customers.default.svc.cluster:80",
                    "customers.default.svc",
                    "customers.default.svc:80",
                    "customers.default",
                    "customers.default:80",
                    "10.72.11.194",
                    "10.72.11.194:80"
                ],
                "routes": [
                    {
                        "match": {
                            "prefix": "/"
                        },
                        "route": {
                            "cluster": "outbound|80|v1|customers.default.svc.cluster.local",
                            "timeout": "0s",
                            "retryPolicy": {
                                "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                "numRetries": 2,
                                "retryHostPredicate": [
                                    {
                                        "name": "envoy.retry_host_predicates.previous_hosts"
                                    }
                                ],
                                "hostSelectionRetryMaxAttempts": "5",
                                "retriableStatusCodes": [
                                    503
                                ]
                            },
                            "maxGrpcTimeout": "0s"
                        },
                        "metadata": {
                            "filterMetadata": {
                                "istio": {
                                    "config": "/apis/networking.istio.io/v1alpha3/namespaces/default/virtual-service/customers"
                                }
                            }
                        },
                        "decorator": {
                            "operation": "customers.default.svc.cluster.local:80/*"
                        }
                    }
                ],
                "includeRequestAttemptCount": true
            },
            {
                "name": "default-http-backend.kube-system.svc.cluster.local:80",
                "domains": [
                    "default-http-backend.kube-system.svc.cluster.local",
                    "default-http-backend.kube-system.svc.cluster.local:80",
                    "default-http-backend.kube-system",
                    "default-http-backend.kube-system:80",
                    "default-http-backend.kube-system.svc.cluster",
                    "default-http-backend.kube-system.svc.cluster:80",
                    "default-http-backend.kube-system.svc",
                    "default-http-backend.kube-system.svc:80",
                    "10.72.13.106",
                    "10.72.13.106:80"
                ],
                "routes": [
                    {
                        "name": "default",
                        "match": {
                            "prefix": "/"
                        },
                        "route": {
                            "cluster": "outbound|80||default-http-backend.kube-system.svc.cluster.local",
                            "timeout": "0s",
                            "retryPolicy": {
                                "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                "numRetries": 2,
                                "retryHostPredicate": [
                                    {
                                        "name": "envoy.retry_host_predicates.previous_hosts"
                                    }
                                ],
                                "hostSelectionRetryMaxAttempts": "5",
                                "retriableStatusCodes": [
                                    503
                                ]
                            },
                            "maxGrpcTimeout": "0s"
                        },
                        "decorator": {
                            "operation": "default-http-backend.kube-system.svc.cluster.local:80/*"
                        }
                    }
                ],
                "includeRequestAttemptCount": true
            },
            {
                "name": "istio-egressgateway.istio-system.svc.cluster.local:80",
                "domains": [
                    "istio-egressgateway.istio-system.svc.cluster.local",
                    "istio-egressgateway.istio-system.svc.cluster.local:80",
                    "istio-egressgateway.istio-system",
                    "istio-egressgateway.istio-system:80",
                    "istio-egressgateway.istio-system.svc.cluster",
                    "istio-egressgateway.istio-system.svc.cluster:80",
                    "istio-egressgateway.istio-system.svc",
                    "istio-egressgateway.istio-system.svc:80",
                    "10.72.13.110",
                    "10.72.13.110:80"
                ],
                "routes": [
                    {
                        "name": "default",
                        "match": {
                            "prefix": "/"
                        },
                        "route": {
                            "cluster": "outbound|80||istio-egressgateway.istio-system.svc.cluster.local",
                            "timeout": "0s",
                            "retryPolicy": {
                                "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                "numRetries": 2,
                                "retryHostPredicate": [
                                    {
                                        "name": "envoy.retry_host_predicates.previous_hosts"
                                    }
                                ],
                                "hostSelectionRetryMaxAttempts": "5",
                                "retriableStatusCodes": [
                                    503
                                ]
                            },
                            "maxGrpcTimeout": "0s"
                        },
                        "decorator": {
                            "operation": "istio-egressgateway.istio-system.svc.cluster.local:80/*"
                        }
                    }
                ],
                "includeRequestAttemptCount": true
            },
            {
                "name": "istio-ingressgateway.istio-system.svc.cluster.local:80",
                "domains": [
                    "istio-ingressgateway.istio-system.svc.cluster.local",
                    "istio-ingressgateway.istio-system.svc.cluster.local:80",
                    "istio-ingressgateway.istio-system",
                    "istio-ingressgateway.istio-system:80",
                    "istio-ingressgateway.istio-system.svc.cluster",
                    "istio-ingressgateway.istio-system.svc.cluster:80",
                    "istio-ingressgateway.istio-system.svc",
                    "istio-ingressgateway.istio-system.svc:80",
                    "10.72.7.166",
                    "10.72.7.166:80"
                ],
                "routes": [
                    {
                        "name": "default",
                        "match": {
                            "prefix": "/"
                        },
                        "route": {
                            "cluster": "outbound|80||istio-ingressgateway.istio-system.svc.cluster.local",
                            "timeout": "0s",
                            "retryPolicy": {
                                "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                "numRetries": 2,
                                "retryHostPredicate": [
                                    {
                                        "name": "envoy.retry_host_predicates.previous_hosts"
                                    }
                                ],
                                "hostSelectionRetryMaxAttempts": "5",
                                "retriableStatusCodes": [
                                    503
                                ]
                            },
                            "maxGrpcTimeout": "0s"
                        },
                        "decorator": {
                            "operation": "istio-ingressgateway.istio-system.svc.cluster.local:80/*"
                        }
                    }
                ],
                "includeRequestAttemptCount": true
            },
            {
                "name": "web-frontend.default.svc.cluster.local:80",
                "domains": [
                    "web-frontend.default.svc.cluster.local",
                    "web-frontend.default.svc.cluster.local:80",
                    "web-frontend",
                    "web-frontend:80",
                    "web-frontend.default.svc.cluster",
                    "web-frontend.default.svc.cluster:80",
                    "web-frontend.default.svc",
                    "web-frontend.default.svc:80",
                    "web-frontend.default",
                    "web-frontend.default:80",
                    "10.72.11.212",
                    "10.72.11.212:80"
                ],
                "routes": [
                    {
                        "name": "default",
                        "match": {
                            "prefix": "/"
                        },
                        "route": {
                            "cluster": "outbound|80||web-frontend.default.svc.cluster.local",
                            "timeout": "0s",
                            "retryPolicy": {
                                "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                "numRetries": 2,
                                "retryHostPredicate": [
                                    {
                                        "name": "envoy.retry_host_predicates.previous_hosts"
                                    }
                                ],
                                "hostSelectionRetryMaxAttempts": "5",
                                "retriableStatusCodes": [
                                    503
                                ]
                            },
                            "maxGrpcTimeout": "0s"
                        },
                        "decorator": {
                            "operation": "web-frontend.default.svc.cluster.local:80/*"
                        }
                    }
                ],
                "includeRequestAttemptCount": true
            }
        ],
        "validateClusters": false
    }
]
```
