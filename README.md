# mackerel-plugin-jolokia

[![Build Status](https://travis-ci.org/y-kuno/mackerel-plugin-jolokia.svg?branch=master)](https://travis-ci.org/y-kuno/mackerel-plugin-jolokia)
![License](https://img.shields.io/github/license/y-kuno/mackerel-plugin-jolokia.svg)
![Release](https://img.shields.io/github/release/y-kuno/mackerel-plugin-jolokia.svg)

[Jolokia](https://jolokia.org/) custom metrics plugin for mackerel.io agent  
This repository releases an artifact to Github Releases, which satisfy the format for mkr plugin installer.

## Synopsis

```shell
mackerel-plugin-jolokia [--host=<host>] [--port=<port>] [--metric-key-prefix=<prefix>] [--exclude-jvm-metrics] [--integration=<integration>] [--custom-metrics-file=<custom-metrics-file>] [--tempfile=<tempfile>]
```

- `host` - The host name or IP address. By default this is `localhost`.
- `port` - The access port. By default this is `8773`.
- `metric-key-prefix` - The metrics key prefix.
- `exclude-jvm-metrics` - Exclude default JVM metrics. By default is collect JVM metrics.
- `integration` - The integration name.  Here [supported integrations list](#integrations).
- `custom-metrics-file` - The custom metrics file. The supported file extension is yaml or yml. [How to custom metrics](#how-to-custom-metrics).
- `tempfile` - The temp file name.

## Integrations

Supported integrations list.

- [Tomcat](doc/tomcat/README.md)

## How to custom metrics

Edit the custom metrics file, in the folder.  
Here [example custom metrics file](custom/example.yml).

### Graph Definition

A custom metrics can specify `graphs` and `metrics`.  
`graphs` represents one graph and includes some `metrics`s which represent each line.

`graphs` includes followings:

- `key`: Key for the graph.
- `label`: Label for the graph.
- `unit`: Unit for lines, `float`, `integer`, `percentage`, `bytes`, `bytes/sec`, `iops` can be specified.
- `metrics`: Array of `metrics` which represents each line.

`metics` includes followings:

- `name`: Key of the line
- `label`: Label of the line
- `diff`: If `diff` is true, differential is used as value.
- `stacked`: If `stacked` is true, the line is stacked.
- `scale`: Each value is multiplied by `scale`.
- `match`: If `match` is true, match with a combination of graph definitions and metric name.

### JMX Metrics

`jmx` includes followings:

- `mbean`: MBean name.
- `attribute`: Array of `attribute` which represents each line.
- `scope`: Array of the scope list.

`attribute` includes followings:

- `name`: Attribute name.
- `prefix`: When it is not empty character, Add prefix for metrics name.


## Installation

* `mkr plugin install y-kuno/mackerel-plugin-jolokia`
* Or download binary in Github Releases

## Example of mackerel-agent.conf

```
[plugin.metrics.jolokia]
command = "/path/to/mackerel-plugin-jolokia"
```