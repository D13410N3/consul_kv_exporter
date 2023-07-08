# Prometheus Exporter for Consul KV Modify Index

This is a simple Prometheus exporter written in Go that fetches the Consul KV Modify Index metrics from multiple data centers and directories in Consul. It exposes these metrics in a format that Prometheus can scrape and store for monitoring and alerting purposes.

## Configuration

The exporter requires a YAML configuration file to specify the data centers and directories to monitor in Consul. The format of the configuration file should be as follows:

```yaml
dc:
  datacenter1:
    directories:
      - "foo/bar"
      - "bar/foo"
  datacenter2:
    directories:
      - "foobar"
```

## Environment Variables

The following environment variables *should be configured*:

- `CONFIG_FILE`: Path to the YAML configuration file specifying the data centers and directories to monitor in Consul.
- `CONSUL_BASE_URI`: Base URI of the Consul server API.
- `LISTEN_ADDR`: Address on which the Prometheus exporter should listen for scraping.

## Metrics

The exporter collects the Consul KV Modify Index values (directories too) and exposes them using the following metric:

- `consul_kv_modify_index`: The Consul KV Modify Index value for each key in Consul KV. The metric includes the following labels:
  - `dc`: Data center name.
  - `key`: Key name in Consul KV

### Metric Examples

Example metric with the `consul_kv_modify_index` metric:

```
consul_kv_modify_index{dc="datacenter1", key="foo/bar/"} 6063284     ### The directory by itself
consul_kv_modify_index{dc="datacenter1", key="foo/bar/key1"} 6131529
consul_kv_modify_index{dc="datacenter2", key="foobar/"} 6131529      ### The directory by itself
```
