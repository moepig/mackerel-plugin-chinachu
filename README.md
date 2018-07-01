mackerel-plugin-chinachu
=======================

Chinachu WUI custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-chinachu -host=<hostname or ip> -port=<port> [-metric-key-prefix=<metric-key-prefix> [-tempfile=<tempfile>]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.chinachu]
command = "/path/to/mackerel-plugin-chinachu -host=localhost -port=20772"
```