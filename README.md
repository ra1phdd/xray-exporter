<div align="center">
<h1>XRay Exporter</h1>
<p>An exporter that collect XRay metrics over its <a href="https://xtls.github.io/en/config/stats.html">Stats API</a> and export them to Prometheus</p>

<p>
    <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://goreportcard.com/badge/github.com/ra1phdd/xray-exporter" alt="Go report">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
    <br>
    <img src="https://app.travis-ci.com/ra1phdd/xray-exporter.svg?branch=master" alt="Build Status">
    <img src="https://coveralls.io/repos/github.com/ra1phdd/xray-exporter/badge.svg?branch=master" alt="Coverage">
</p>

**English** | [Russian](README.ru.md)
</div>

---

![](https://i.loli.net/2020/06/12/KzjOnyu93VEIPiW.png)

## Quick Start

### Binaries

The latest binaries are made available on GitHub [releases](https://github.com/ra1phdd/xray-exporter/releases) page:

```bash
wget -O /tmp/xray-exporter https://github.com/ra1phdd/xray-exporter/releases/latest/download/xray-exporter_linux_amd64
mv /tmp/xray-exporter /usr/local/bin/xray-exporter
chmod +x /usr/local/bin/xray-exporter
```

### Docker (Recommended)

Docker setup lives in [`deployments/docker`](deployments/docker):

```bash
docker build -f deployments/docker/Dockerfile -t xray-exporter:local .
```

### Grafana Dashboard

A simple Grafana dashboard is also available [here](deployments/observability/grafana/dashboard.json). Please refer to the [Grafana docs](https://grafana.com/docs/grafana/latest/reference/export_import/#importing-a-dashboard) to get the steps of importing dashboards from JSON files.

## Tutorial

The project already contains deployment manifests in the [`deployments`](deployments) folder.

1. Prepare XRay config:
Use [`deployments/xray/config.json`](deployments/xray/config.json). It includes a basic configuration with statistics collection and API enabled on `127.0.0.1:54321` inside the XRay container

2. Prepare Docker Compose workspace:
The compose file is [`deployments/docker/docker-compose.yml`](deployments/docker/docker-compose.yml). It starts `xray`, `xray-exporter`, `prometheus`, and `grafana`.

Expected local structure near compose file:
```plain
deployments/docker/
  docker-compose.yml
  xray/config.json
  prometheus/prometheus.yml
  grafana/
```

Example setup from repository root:
```bash
mkdir -p deployments/docker/xray deployments/docker/prometheus deployments/docker/grafana
cp deployments/xray/config.json deployments/docker/xray/config.json
cp deployments/observability/prometheus/prometheus.yml deployments/docker/prometheus/prometheus.yml
```

3. Start stack:
```bash
cd deployments/docker
docker compose up -d
```

4. Verify exporter:
- Open `http://localhost:9550/` for home page.
- Open `http://localhost:9550/scrape` for XRay metrics.
- Basic Auth in compose defaults to:
`username=prometheus`, `password=change-me`.

If scrape is successful, response includes:
```plain
# HELP xray_up Indicate scrape succeeded or not
# TYPE xray_up gauge
xray_up 1
```

5. Verify observability:
- Prometheus config comes from [`deployments/observability/prometheus/prometheus.yml`](deployments/observability/prometheus/prometheus.yml).
- Grafana is available on `http://localhost:3000`.
- Dashboard JSON is [`deployments/observability/grafana/dashboard.json`](deployments/observability/grafana/dashboard.json).

If `xray_up` is missing or equals `0`, check logs of `xray` and `xray-exporter` containers.

## Runtime & Traffic Metrics

The exporter doesn't retain the original metric names from XRay intentionally:

| Runtime Metric   | Exposed Metric                    |
|:-----------------|:----------------------------------|
| `uptime`         | `xray_uptime_seconds`             |
| `num_goroutine`  | `xray_goroutines`                 |
| `alloc`          | `xray_memstats_alloc_bytes`       |
| `total_alloc`    | `xray_memstats_alloc_bytes_total` |
| `sys`            | `xray_memstats_sys_bytes`         |
| `mallocs`        | `xray_memstats_mallocs_total`     |
| `frees`          | `xray_memstats_frees_total`       |
| `live_objects`   | Removed. See the appendix below.  |
| `num_gc`         | `xray_memstats_num_gc`            |
| `pause_total_ns` | `xray_memstats_pause_total_ns`    |

| Statistic Metric                           | Exposed Metric                                                              |
|:-------------------------------------------|:----------------------------------------------------------------------------|
| `inbound>>>tag-name>>>traffic>>>uplink`    | `xray_traffic_uplink_bytes_total{dimension="inbound",target="tag-name"}`    |
| `inbound>>>tag-name>>>traffic>>>downlink`  | `xray_traffic_downlink_bytes_total{dimension="inbound",target="tag-name"}`  |
| `outbound>>>tag-name>>>traffic>>>uplink`   | `xray_traffic_uplink_bytes_total{dimension="outbound",target="tag-name"}`   |
| `outbound>>>tag-name>>>traffic>>>downlink` | `xray_traffic_downlink_bytes_total{dimension="outbound",target="tag-name"}` |
| `user>>>user-email>>traffic>>>uplink`      | `xray_traffic_uplink_bytes_total{dimension="user",target="user-email"}`     |
| `user>>>user-email>>>traffic>>>downlink`   | `xray_traffic_downlink_bytes_total{dimension="user",target="user-email"}`   |
| ...                                        | ...                                                                         |

- The value of `live_objects` can be calculated by `memstats_mallocs_total - memstats_frees_total`.

## 🤝 Contribute & Roadmap

PRs welcome! The codebase is intentionally small and readable.

See our [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Developer group building, join after your first merged PR!
