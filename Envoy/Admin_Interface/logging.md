# Logging
The /logging endpoint enables or disables different logging levels for a particular component or all loggers.

To list all loggers, we can send a POST request to the /logging endpoint:

```shell
$ curl -X POST localhost:9901/logging
active loggers:
  admin: info
  alternate_protocols_cache: info
  aws: info
  assert: info
  backtrace: info
  cache_filter: info
  client: info
  config: info
...
```

The output will contain the names of the loggers and the logging level for each logger. To change the logging level for all active loggers, we can use the level parameter. For example, we could run the following to change the logging level of all loggers to debug:

```shell
$ curl -X POST localhost:9901/logging?level=debug
active loggers:
  admin: debug
  alternate_protocols_cache: debug
  aws: debug
  assert: debug
  backtrace: debug
  cache_filter: debug
  client: debug
  config: debug
...
```

To change a particular logger’s level, we can replace the level query parameter name with the logger’s name. For example, to change the admin logger level to warning, we can run the following:

```shell
$ curl -X POST localhost:9901/logging?admin=warning
active loggers:
  admin: warning
  alternate_protocols_cache: info
  aws: info
  assert: info
  backtrace: info
  cache_filter: info
  client: info
  config: info
```

To trigger the reopening of all access logs, we can send a POST request to the /reopen_logs endpoint.
