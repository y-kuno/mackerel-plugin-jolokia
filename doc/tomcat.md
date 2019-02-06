# Integration: Tomcat

## Overview

This collects Tomcat metrics.

* Overall activity metrics: error count, request count, processing times
* Thread pool metrics: thread count, number of threads busy

And more.

## Setup

### Configuration

1. Edit the `mackerel-agent.conf` file.
2. Restart the Agent

### Example of mackerel-agent.conf

```
[plugin.metrics.tomcat]
command = "/path/to/mackerel-plugin-jolokia --integration=tomcat"
```