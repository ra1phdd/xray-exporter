<div align="center">
<h1>XRay Exporter</h1>
<p>Экспортер, который собирает метрики XRay через <a href="https://xtls.github.io/ru/config/stats.html">Stats API</a> и экспортирует их в Prometheus</p>

<p>
    <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://goreportcard.com/badge/github.com/ra1phdd/xray-exporter" alt="Go report">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
    <br>
    <img src="https://app.travis-ci.com/ra1phdd/xray-exporter.svg?branch=master" alt="Build Status">
    <img src="https://coveralls.io/repos/github.com/ra1phdd/xray-exporter/badge.svg?branch=master" alt="Coverage">
</p>

[English](README.md) | **Russian**
</div>

![](https://i.loli.net/2020/06/12/KzjOnyu93VEIPiW.png)

## Быстрый старт

### Бинарные файлы

Актуальные бинарники доступны на странице [релизов](https://github.com/ra1phdd/xray-exporter/releases) GitHub:

```bash
wget -O /tmp/xray-exporter https://github.com/ra1phdd/xray-exporter/releases/latest/download/xray-exporter_linux_amd64
mv /tmp/xray-exporter /usr/local/bin/xray-exporter
chmod +x /usr/local/bin/xray-exporter
```

### Docker (рекомендуется)

Конфигурация Docker находится в [`deployments/docker`](deployments/docker):

```bash
docker build -f deployments/docker/Dockerfile -t xray-exporter:local .
```

### Grafana Dashboard

Простой дашборд Grafana также доступен [здесь](deployments/observability/grafana/dashboard.json). Импорт из JSON описан в [документации Grafana](https://grafana.com/docs/grafana/latest/reference/export_import/#importing-a-dashboard).

## Руководство

В проекте уже есть манифесты для развёртывания в папке [`deployments`](deployments).

1. Подготовьте конфиг XRay:
Используйте [`deployments/xray/config.json`](deployments/xray/config.json). В нём включены сбор статистики и API на `127.0.0.1:54321` внутри контейнера XRay.

2. Подготовьте рабочую директорию Docker Compose:
Файл compose: [`deployments/docker/docker-compose.yml`](deployments/docker/docker-compose.yml). Он запускает `xray`, `xray-exporter`, `prometheus` и `grafana`.

Ожидаемая локальная структура рядом с compose-файлом:
```plain
deployments/docker/
  docker-compose.yml
  xray/config.json
  prometheus/prometheus.yml
  grafana/
```

Пример настройки из корня репозитория:
```bash
mkdir -p deployments/docker/xray deployments/docker/prometheus deployments/docker/grafana
cp deployments/xray/config.json deployments/docker/xray/config.json
cp deployments/observability/prometheus/prometheus.yml deployments/docker/prometheus/prometheus.yml
```

3. Запустите стек:
```bash
cd deployments/docker
docker compose up -d
```

4. Проверьте экспортёр:
- Откройте `http://localhost:9550/` для главной страницы.
- Откройте `http://localhost:9550/scrape` для метрик XRay.
- Basic Auth в compose по умолчанию:
`username=prometheus`, `password=change-me`.

Если scrape успешен, в ответе будет:
```plain
# HELP xray_up Indicate scrape succeeded or not
# TYPE xray_up gauge
xray_up 1
```

5. Проверьте observability-стек:
- Конфиг Prometheus берётся из [`deployments/observability/prometheus/prometheus.yml`](deployments/observability/prometheus/prometheus.yml).
- Grafana доступна по адресу `http://localhost:3000`.
- JSON дашборда: [`deployments/observability/grafana/dashboard.json`](deployments/observability/grafana/dashboard.json).

Если `xray_up` отсутствует или равен `0`, проверьте логи контейнеров `xray` и `xray-exporter`.

## Runtime и Traffic метрики

Экспортёр намеренно не сохраняет исходные имена метрик XRay:

| Runtime Metric   | Exposed Metric                    |
|:-----------------|:----------------------------------|
| `uptime`         | `xray_uptime_seconds`             |
| `num_goroutine`  | `xray_goroutines`                 |
| `alloc`          | `xray_memstats_alloc_bytes`       |
| `total_alloc`    | `xray_memstats_alloc_bytes_total` |
| `sys`            | `xray_memstats_sys_bytes`         |
| `mallocs`        | `xray_memstats_mallocs_total`     |
| `frees`          | `xray_memstats_frees_total`       |
| `live_objects`   | Удалено. См. примечание ниже.     |
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

- Значение `live_objects` можно вычислить как `memstats_mallocs_total - memstats_frees_total`.

## 🤝 Участие и roadmap

Pull request'ы приветствуются. Кодовая база намеренно небольшая и читаемая.

Подробности смотрите в [CONTRIBUTING.md](CONTRIBUTING.md).

Собираем группу разработчиков, присоединяйтесь после первого влитого PR!
