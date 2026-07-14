# F5OS Exporter

[![Go Reference](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Prometheus](https://img.shields.io/badge/Prometheus-Exporter-E6522C?logo=prometheus&logoColor=white)](https://prometheus.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#contributing)

A **Prometheus exporter** for [F5 F5OS](https://www.f5.com/products/big-ip-services/f5os) platforms (rSeries / VELOS), built on top of the **F5OS REST API** (OpenConfig-based).

It follows the standard [multi-target exporter pattern](https://prometheus.io/docs/guides/multi-target-exporter/) (`/probe?target=...`), allowing a single exporter instance to monitor **many F5OS platforms concurrently** — without running an agent on each device.

---

## Table of Contents

- [Features](#features)
- [How It Works](#how-it-works)
- [Metrics](#metrics)
- [Installation](#installation)
- [Configuration](#configuration)
- [Command-Line Flags](#command-line-flags)
- [Running the Exporter](#running-the-exporter)
- [Prometheus Configuration](#prometheus-configuration)
- [HTTP Endpoints](#http-endpoints)
- [Grafana Dashboard](#grafana-dashboard)
- [Security Considerations](#security-considerations)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- 🔌 **Multi-target & agentless** — monitor multiple F5OS platforms from one exporter (blackbox-style `/probe`).
- ⚡ **Concurrent scraping** — each subsystem is collected in parallel via goroutines.
- 🔑 **Token-based authentication** — authenticates against the F5OS REST API; credentials never appear in URLs.
- 📊 **Rich metric coverage**:
  - **Platform / Hardware** — CPU cores/threads/frequency/cache, per-thread CPU utilization (current + 5s/1m/5m averages), platform memory (total/used/free/available), platform temperature (current/avg/min/max), chassis fan speeds, PSU voltage/current/temperature/fan, and a platform-info metric including TPM integrity status.
  - **Storage / Disks** — per-disk inventory (model, serial, type, size) and cumulative I/O counters (IOPS, bytes, latency-ms, merged) for every installed disk.
  - **Interfaces** — operational/admin state, MTU, FEC mode, LACP state, in/out octets, unicast/broadcast/multicast packets, discards, errors, FCS errors, plus L2 ethernet counters (CRC, oversize, jabber, fragment, MAC pause/control, 802.1q), speed/duplex negotiation, MAC address, LAG membership, auto-negotiation and flow-control state.
  - **LACP** — per-LAG member state (in-sync, collecting, distributing, aggregatable) and LACP PDU counters (in/out packets, RX errors), enriched with activity/timeout/system-id/partner-id labels.
  - **License** — licensed version, platform ID, licensed date and service-check date (as Unix timestamps for age-based alerting).

- 🏷️ Consistent `target` label on every metric for easy multi-device dashboards.
- 🩺 Built-in `probe_success` and `probe_duration_seconds` metrics per scrape.
- 🧩 **Graceful degradation** — subsystems that are unavailable on a given platform are logged and skipped rather than failing the whole probe.
- 🔢 **Correct metric types** — cumulative I/O and packet counters are exposed as Prometheus counters so `rate()` works as expected.

---

## How It Works

```
                          ┌──────────────────────────┐
 Prometheus  ── /probe ──▶│       F5OS Exporter       │
 (target=...)             │                           │
                          │  1. Resolve target+creds  │
                          │  2. Obtain auth token     │
                          │  3. Fan-out probes (HTTP) │──┐ F5OS REST API
                          │  4. Aggregate metrics     │  │ (HTTPS / OpenConfig)
                          └──────────────────────────┘  │
                                       ▲                 ▼
                                       │        ┌──────────────────┐
                                       └────────│    F5 F5OS(s)     │
                                                └──────────────────┘
```

On each `/probe` request:

1. The `target` query parameter is parsed and matched against the configured credentials.
2. The exporter authenticates against the F5OS platform to obtain an auth token.
3. All collectors (Hardware, Storage, Interfaces, LACP, License) run **concurrently**.
4. Metrics are merged and returned in the Prometheus exposition format.

---

## Metrics

All metrics are prefixed with `bigip_f5os_` and carry a `target` label.

### Probe (per scrape)

| Metric | Type | Description |
|--------|------|-------------|
| `probe_success` | gauge | `1` if the probe succeeded, `0` otherwise |
| `probe_duration_seconds` | gauge | Time taken to complete the probe |

### Platform Info (`bigip_f5os_info`)

Constant `1`; platform metadata carried in labels.

Labels: `target`, `description`, `tmp_status`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_info` | gauge | Platform identity + TPM integrity status (`tmp_status`, e.g. `Valid`) |

### Memory (`bigip_f5os_Memory_*`)

Label: `target`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_Memory_available_bytes` | gauge | Memory available for allocation (bytes) |
| `bigip_f5os_Memory_free_bytes` | gauge | Free (unused) memory (bytes) |
| `bigip_f5os_Memory_used_percent` | gauge | Memory used (%) |
| `bigip_f5os_Memory_platform_used_bytes` | gauge | Platform memory used (bytes) |
| `bigip_f5os_Memory_platform_total_bytes` | gauge | Platform memory total (bytes) |

### CPU (`bigip_f5os_cpu_*`)

Labels: `target` (thread metrics add `thread_index`, `thread_name`).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_cpu_count` | gauge | Physical CPU core count |
| `bigip_f5os_cpu_thread_count` | gauge | Logical CPU thread count |
| `bigip_f5os_cpu_frequency_hz` | gauge | Rated CPU frequency (Hz) |
| `bigip_f5os_cpu_cache_size_kb` | gauge | CPU cache size (KiB) |
| `bigip_f5os_cpu_utilization_percent` | gauge | Overall CPU utilization (current) |
| `bigip_f5os_cpu_utilization_5sec_avg_percent` | gauge | Overall CPU utilization (5s avg) |
| `bigip_f5os_cpu_utilization_1min_avg_percent` | gauge | Overall CPU utilization (1m avg) |
| `bigip_f5os_cpu_utilization_5min_avg_percent` | gauge | Overall CPU utilization (5m avg) |
| `bigip_f5os_cpu_thread_utilization_percent` | gauge | Per-thread utilization (current) |
| `bigip_f5os_cpu_thread_utilization_5sec_avg_percent` | gauge | Per-thread utilization (5s avg) |
| `bigip_f5os_cpu_thread_utilization_1min_avg_percent` | gauge | Per-thread utilization (1m avg) |
| `bigip_f5os_cpu_thread_utilization_5min_avg_percent` | gauge | Per-thread utilization (5m avg) |

### Temperature (`bigip_f5os_temperature_*`)

Label: `target`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_temperature_current_celsius` | gauge | Current platform temperature (°C) |
| `bigip_f5os_temperature_average_celsius` | gauge | Average platform temperature (°C) |
| `bigip_f5os_temperature_minimum_celsius` | gauge | Minimum platform temperature (°C) |
| `bigip_f5os_temperature_maximum_celsius` | gauge | Maximum platform temperature (°C) |

### Storage / Disks (`bigip_f5os_storage_disk_*`)

Labels: `target`, `disk_name` (inventory adds `model`, `serial`, `type`).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_storage_disk_size_gb` | gauge | Disk capacity (GB); metadata in labels |
| `bigip_f5os_storage_disk_total_iops` | counter | Cumulative total IOPS |
| `bigip_f5os_storage_disk_read_iops` | counter | Cumulative read IOPS |
| `bigip_f5os_storage_disk_write_iops` | counter | Cumulative write IOPS |
| `bigip_f5os_storage_disk_read_bytes` | counter | Cumulative bytes read |
| `bigip_f5os_storage_disk_write_bytes` | counter | Cumulative bytes written |
| `bigip_f5os_storage_disk_read_latency_ms` | counter | Cumulative read latency (ms) |
| `bigip_f5os_storage_disk_write_latency_ms` | counter | Cumulative write latency (ms) |
| `bigip_f5os_storage_disk_read_merged` | counter | Cumulative merged read requests |
| `bigip_f5os_storage_disk_write_merged` | counter | Cumulative merged write requests |

### Fans (`bigip_f5os_fan_speed_rpm`)

Labels: `target`, `fan_name`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_fan_speed_rpm` | gauge | Chassis fan-tray fan speed (RPM) |

### PSU (`bigip_f5os_psu_*`)

Labels: `target`, `psu_name`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_psu_current_in_amp` | gauge | PSU input current (A) |
| `bigip_f5os_psu_current_out_amp` | gauge | PSU output current (A) |
| `bigip_f5os_psu_voltage_in_volt` | gauge | PSU input voltage (V) |
| `bigip_f5os_psu_voltage_out_volt` | gauge | PSU output voltage (V) |
| `bigip_f5os_psu_temperature_1_celsius` | gauge | PSU temperature sensor 1 (°C) |
| `bigip_f5os_psu_temperature_2_celsius` | gauge | PSU temperature sensor 2 (°C) |
| `bigip_f5os_psu_temperature_3_celsius` | gauge | PSU temperature sensor 3 (°C) |
| `bigip_f5os_psu_fan_1_speed_rpm` | gauge | PSU internal fan speed (RPM) |

### Interfaces (`bigip_f5os_interface_*`)

Labels: `target`, `interface` (info metrics add descriptive labels).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_interface_info` | gauge | Interface info (labels: `type`, `oper_status`, `fec`, `lacp_state`) |
| `bigip_f5os_interface_enabled` | gauge | Admin enabled `1`/`0` |
| `bigip_f5os_interface_oper_status_up` | gauge | Operational status `1`=UP, `0`=otherwise |
| `bigip_f5os_interface_mtu_bytes` | gauge | Interface MTU (bytes) |
| `bigip_f5os_interface_in_octets_total` | counter | Received octets |
| `bigip_f5os_interface_in_unicast_packets_total` | counter | Received unicast packets |
| `bigip_f5os_interface_in_broadcast_packets_total` | counter | Received broadcast packets |
| `bigip_f5os_interface_in_multicast_packets_total` | counter | Received multicast packets |
| `bigip_f5os_interface_in_discards_total` | counter | Received discards |
| `bigip_f5os_interface_in_errors_total` | counter | Received errors |
| `bigip_f5os_interface_in_fcs_errors_total` | counter | Received FCS errors |
| `bigip_f5os_interface_out_octets_total` | counter | Transmitted octets |
| `bigip_f5os_interface_out_unicast_packets_total` | counter | Transmitted unicast packets |
| `bigip_f5os_interface_out_broadcast_packets_total` | counter | Transmitted broadcast packets |
| `bigip_f5os_interface_out_multicast_packets_total` | counter | Transmitted multicast packets |
| `bigip_f5os_interface_out_discards_total` | counter | Transmitted discards |
| `bigip_f5os_interface_out_errors_total` | counter | Transmitted errors |

#### Ethernet sub-metrics (`bigip_f5os_interface_ethernet_*`)

Labels: `target`, `interface` (info metric adds speed/duplex/MAC/LAG labels).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_interface_ethernet_info` | gauge | Ethernet link info (labels: `port_speed`, `negotiated_port_speed`, `duplex_mode`, `negotiated_duplex_mode`, `hw_mac_address`, `aggregate_id`) |
| `bigip_f5os_interface_ethernet_auto_negotiate` | gauge | Auto-negotiate `1`/`0` |
| `bigip_f5os_interface_ethernet_flow_control_rx` | gauge | RX flow control `1`=on, `0`=off |
| `bigip_f5os_interface_ethernet_flow_control_tx` | gauge | TX flow control `1`=on, `0`=off |
| `bigip_f5os_interface_ethernet_in_mac_control_frames_total` | counter | Received MAC control frames |
| `bigip_f5os_interface_ethernet_in_mac_pause_frames_total` | counter | Received MAC pause frames |
| `bigip_f5os_interface_ethernet_in_oversize_frames_total` | counter | Received oversize frames |
| `bigip_f5os_interface_ethernet_in_jabber_frames_total` | counter | Received jabber frames |
| `bigip_f5os_interface_ethernet_in_fragment_frames_total` | counter | Received fragment frames |
| `bigip_f5os_interface_ethernet_in_8021q_frames_total` | counter | Received 802.1q frames |
| `bigip_f5os_interface_ethernet_in_crc_errors_total` | counter | Received CRC errors |
| `bigip_f5os_interface_ethernet_out_mac_control_frames_total` | counter | Transmitted MAC control frames |
| `bigip_f5os_interface_ethernet_out_mac_pause_frames_total` | counter | Transmitted MAC pause frames |
| `bigip_f5os_interface_ethernet_out_8021q_frames_total` | counter | Transmitted 802.1q frames |

### LACP (`bigip_f5os_lacp_*`)

Labels: `target`, `lag_name`, `interface` (info metrics add more labels).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_lacp_interface_info` | gauge | LAG info (labels: `interval`, `lacp_mode`, `system_id_mac`) |
| `bigip_f5os_lacp_member_info` | gauge | Member info (labels: `activity`, `timeout`, `synchronization`, `system_id`, `partner_id`) |
| `bigip_f5os_lacp_member_aggregatable` | gauge | Member aggregatable `1`/`0` |
| `bigip_f5os_lacp_member_collecting` | gauge | Member collecting `1`/`0` |
| `bigip_f5os_lacp_member_distributing` | gauge | Member distributing `1`/`0` |
| `bigip_f5os_lacp_member_in_sync` | gauge | Member synchronization `1`=IN_SYNC, `0`=otherwise |
| `bigip_f5os_lacp_in_packets_total` | counter | LACP PDUs received |
| `bigip_f5os_lacp_out_packets_total` | counter | LACP PDUs transmitted |
| `bigip_f5os_lacp_rx_errors_total` | counter | LACP receive errors |

### License (`bigip_f5os_license_*`)

Labels: `target` (info metric adds `licensed_version`, `platform_id`).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_f5os_license_info` | gauge | License info (labels: `licensed_version`, `platform_id`) |
| `bigip_f5os_license_licensed_date_seconds` | gauge | Licensed date (Unix timestamp) |
| `bigip_f5os_license_service_check_date_seconds` | gauge | Service check date (Unix timestamp) |

#### Useful queries

```promql
# Disk average write latency (ms per operation)
rate(bigip_f5os_storage_disk_write_latency_ms[5m])
  / clamp_min(rate(bigip_f5os_storage_disk_write_iops[5m]), 1)

# Interface inbound throughput (bits/s)
rate(bigip_f5os_interface_in_octets_total[5m]) * 8

# License service-check age in days
(time() - bigip_f5os_license_service_check_date_seconds) / 86400

# Alert on LACP member out of sync
bigip_f5os_lacp_member_in_sync == 0
```

---

## Installation

### Build from Source

Requires **Go 1.25+**.

```bash
git clone https://github.com/Haameed/f5_f5os_exporter.git
cd f5_f5os_exporter
go build -o f5_f5os_exporter ./cmd/f5_f5os_exporter
```

### Docker

#### Pull from GitHub Container Registry

Pre-built multi-arch images (amd64 / arm64) are published on every release:

```bash
docker pull ghcr.io/haameed/f5_f5os_exporter:latest
# or a specific version:
docker pull ghcr.io/haameed/f5_f5os_exporter:v0.1.0
# Run it:
docker run -d --name f5_f5os_exporter \
  -p 11001:11001 \
  -v "$(pwd)/f5os-config.yaml:/etc/f5_f5os_exporter/f5os-config.yaml:ro" \
  ghcr.io/haameed/f5_f5os_exporter:latest -insecure
```

#### Build locally

A minimal multi-stage `Dockerfile`:

```dockerfile
# ---- build ----
FROM golang:1.25 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/f5_f5os_exporter ./cmd/f5_f5os_exporter

# ---- runtime ----
FROM gcr.io/distroless/static-debian12
COPY --from=build /bin/f5_f5os_exporter /bin/f5_f5os_exporter
EXPOSE 11001
ENTRYPOINT ["/bin/f5_f5os_exporter"]
CMD ["-config", "/etc/f5_f5os_exporter/f5os-config.yaml"]
```

```bash
docker build -t f5_f5os_exporter .
docker run -d --name f5_f5os_exporter \
  -p 11001:11001 \
  -v "$(pwd)/f5os-config.yaml:/etc/f5_f5os_exporter/f5os-config.yaml:ro" \
  f5_f5os_exporter -insecure
```

---

## Configuration

Credentials are provided via a YAML file. Each key is the **target URL** (scheme + host), mapped to its `username` / `password`.

Create `f5os-config.yaml`:

```yaml
https://192.168.100.10:
  username: yourusername
  password: yourpassword

https://192.168.100.11:
  username: yourusername
  password: yourpassword
```

> ℹ️ A sample file is provided as [`config-example.yml`](config-example.yml).
> The `target` you pass to `/probe?target=...` **must exactly match** a key in this file
> (same scheme, host, and port).

> ⚠️ Token authentication requires the **`https`** scheme.

---

## Command-Line Flags

| Flag             | Default              | Description |
|------------------|----------------------|-------------|
| `-config`        | `f5os-config.yaml`   | Path to the credentials YAML file |
| `-listen`        | `:11001`             | Address the HTTP server listens on |
| `-scrape-timeout`| `30`                 | Maximum seconds allowed for a single scrape |
| `-https-timeout` | `10`                 | TLS handshake timeout in seconds |
| `-insecure`      | `false`              | Skip TLS certificate verification (useful for self-signed F5OS certs) |

---

## Running the Exporter

```bash
./f5_f5os_exporter -config f5os-config.yaml -insecure
```

Test a probe manually:

```bash
curl 'http://localhost:11001/probe?target=https://192.168.1.1'
```

You should see Prometheus-formatted metrics, ending with `probe_success 1`.

---

## Prometheus Configuration

Use the multi-target relabeling pattern so the `target` becomes a query parameter and the device address is preserved in the `instance` label:

```yaml
scrape_configs:
  - job_name: 'f5os'
    metrics_path: /probe
    static_configs:
      - targets:
          - https://192.168.100.10
          - https://192.168.100.11
    relabel_configs:
      # Pass the target as a query parameter to the exporter
      - source_labels: [__address__]
        target_label: __param_target
      # Preserve the real device address as the instance label
      - source_labels: [__param_target]
        target_label: instance
      # Point the actual scrape at the exporter
      - target_label: __address__
        replacement: localhost:11001
```

---

## HTTP Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /probe?target=<url>` | Scrape metrics for a single F5OS target |
| `GET /metrics` | The exporter's own (process/Go) metrics |
| `GET /health` | Liveness/health check — returns `200 OK` |

---

## Grafana Dashboard

A ready-to-import Grafana dashboard is provided in [`dashboards/bigip_f5os_dashboard.json`](dashboards/bigip_f5os_dashboard.json).

It is organized into collapsible rows — **Device**, **Compute**, **Disks**, **Hardware**, **Interfaces**, and **LACP** — and includes:

- A platform-info table (with color-coded TPM integrity status) and a license summary with service-check-age thresholds.
- CPU / memory / per-thread utilization panels.
- Disk IOPS, throughput, average latency (ms/op) and merged-I/O panels.
- Temperature, fan, and PSU (voltage / current / temperature / fan) panels.
- Interface throughput, packet-rate, error/discard, and L2 frame-anomaly panels, plus a joined "Ethernet / Link Details" table.
- LACP member-state and PDU-rate panels.

The dashboard uses `target` and `interface` template variables and expects a Prometheus datasource containing metrics prefixed `bigip_f5os_`.

---

## Security Considerations

- **Credentials at rest** — the config file contains plaintext credentials. Restrict its permissions (`chmod 600`) and consider mounting it as a read-only secret in containerized environments.
- **Least privilege** — use a dedicated F5OS user with read-only permissions.
- **TLS verification** — `-insecure` disables certificate validation. Prefer trusting the F5OS CA and leaving verification enabled in production.
- **Network exposure** — the `/probe` endpoint accepts arbitrary `target` values that match the config map. Keep the exporter on a trusted management network.
- **Sensitive labels** — the exporter deliberately omits sensitive license fields (appliance serial number, registration key) from exported metrics.

---

## Development

### Project Layout

```
.
├── cmd/f5_f5os_exporter       # main entrypoint
├── internal
│   ├── config                 # flag parsing + YAML credentials loading
│   └── utils                  # F5OS token authentication
└── pkg
    ├── http                   # F5OS REST client (token-based)
    └── probe                  # collectors: hardware, interface, lacp, license
```

### Run Tests

```bash
go test ./...
```

### Adding a New Collector

1. Create a new file under `pkg/probe/` exposing a function with the signature:
   ```go
   func GetMyProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool)
   ```
2. Register it in the `allProbes` slice in `pkg/probe/probe.go`.
3. The framework runs it concurrently and aggregates its metrics automatically.

> 💡 The shared `stringToFloat` helper strips unit suffixes (e.g. `2401.000(MHz)`, `700.00GB`) and safely returns `0` for empty values. Reuse it for any string-valued numeric fields, and prefer `prometheus.CounterValue` for cumulative counters so `rate()` works correctly.

---

## Contributing

**Everyone is welcome to participate and contribute!** 🎉

- 🐛 Found a bug? [Open an issue](https://github.com/Haameed/f5_f5os_exporter/issues).
- 💡 Have a feature idea or a new metric to expose? Open an issue to discuss it.
- 🔧 Submit Pull Requests — please run `go fmt ./...`, `go vet ./...`, and `go test ./...` before opening.

When contributing metrics, follow the
[Prometheus metric naming best practices](https://prometheus.io/docs/practices/naming/)
(base units, `_total` for counters, descriptive `HELP` text).

---

## Alerting

A complete set of recommended alerts (for both Prometheus and Grafana) is
documented in [`alerts/README.md`](alerts/README.md), covering availability,
system load (CPU / memory), storage latency, hardware sensors (temperature /
fans / PSU), TPM integrity, interfaces, LACP, and license service-check age.

---

## License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 Hamed Maleki