# Prometheus Exporter for Plex

Expose library playback, storage, and host metrics in a Prometheus format.

# Configuration

The exporter is configured via required environment variables:

- `PLEX_SERVER`: The full URL where your server can be reached, including the scheme and port (if not 80 or 443). For example `http://192.168.0.10:32400` or `https://my.plex.tld`.
- `PLEX_TOKEN`: A [Plex token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/) belonging to the server administrator. 

# Running

The exporter runs via Docker:

```bash
docker run \
  -name prom-plex-exporter \
  -p 9000:9000 \
  -e PLEX_SERVER="<Your Plex server URL>" \
  -e PLEX_TOKEN="<Your Plex server admin token>" \
  ghcr.io/jsclayton/prometheus-plex-exporter
```

Or via Docker Compose:

```yaml
prom-plex-exporter:
  image: ghcr.io/jsclayton/prometheus-plex-exporter
  ports:
    - 9000:9000/tcp
  environment:
    PLEX_SERVER: <Your Plex server URL>
    PLEX_TOKEN: <Your Plex server admin token>
```

A sample dashboard can be found in the [examples](examples/dashboards/Media%20Server.json)

# Exporting Metrics

The simplest way to start visualizaing your metrics is with the Free Forever [Grafana Cloud](https://grafana.com/docs/grafana-cloud/) and [Grafana Agent](https://grafana.com/docs/agent/latest/).

Here's an example config file that will read metrics from the exporter and ship them to [Prometheus](https://grafana.com/docs/grafana-cloud/data-configuration/metrics/metrics-prometheus/) via `remote_write`:


```yaml
metrics:
  configs:
  - name: prom-plex
    scrape_configs:
      - job_name: prom-plex
        static_configs:
          - targets:
            - <IP/address and port of the exporter endpoint>
    remote_write:
      - url: <Your Metrics instance remote_write endpoint>
        basic_auth:
          username: <Your Metrics instance ID>
          password: <Your Grafana.com API Key>
```