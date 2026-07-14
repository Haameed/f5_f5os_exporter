# Alerting Guide — F5 F5OS Exporter

This document provides a complete set of recommended alerts for the
`f5_f5os_exporter`, suitable for both **Prometheus alert rules** and
**Grafana Alerting** (the queries are identical — only the configuration UI
differs).

> **Thresholds are sensible starting points.** Tune them to your environment,
> especially platform-rated limits (PSU output voltage, temperature ceilings)
> and traffic volume (see the Interfaces section for a volume-independent
> approach to error alerting).

## Table of Contents

- [Severity Levels](#severity-levels)
- [How to Use](#how-to-use)
- [Availability & Probe](#availability--probe)
- [System (CPU / Memory)](#system-cpu--memory)
- [Storage / Disks](#storage--disks)
- [Hardware (Temperature / Fans / PSU)](#hardware-temperature--fans--psu)
- [Security (TPM)](#security-tpm)
- [Interfaces](#interfaces)
- [LACP](#lacp)
- [License](#license)

## Severity Levels

| Severity   | Meaning                                            |
|------------|----------------------------------------------------|
| `critical` | Immediate action required (outage / hardware risk) |
| `warning`  | Needs investigation soon (degradation / capacity)  |

## How to Use

### Prometheus

Copy the `expr` of any alert into a rule file:

```yaml
groups:
  - name: f5_f5os
    rules:
      - alert: F5osCpuHigh
        expr: bigip_f5os_cpu_utilization_1min_avg_percent > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CPU high on {{ $labels.target }}"
          description: "1-minute average CPU utilization is above 85% for 5 minutes."
```

### Grafana Alerting

Use the same `expr` as the alert query, set the **reduce/threshold** to match
the condition, and set the pending period to the `for` value.

---

## Availability & Probe

### Probe Down — `critical`
The exporter cannot scrape the target.

```promql
up{job="f5os"} == 0
```
**For:** `2m`
**Description:** `F5OS exporter/target {{ $labels.target }} is unreachable.`

---

## System (CPU / Memory)

### CPU High — `warning`
```promql
bigip_f5os_cpu_utilization_1min_avg_percent > 85
```
**For:** `5m`
**Description:** `1-minute average CPU utilization on {{ $labels.target }} is above 85%.`

### Memory High — `critical`
```promql
bigip_f5os_Memory_used_percent > 90
```
**For:** `5m`
**Description:** `Memory usage on {{ $labels.target }} is above 90%.`

---

## Storage / Disks

> Disk I/O metrics are cumulative counters. Latency is derived as
> `rate(latency-ms) / rate(iops)` to yield **average milliseconds per
> operation** — a volume-independent value that stays meaningful under any load.

### Disk Write Latency High — `warning`
```promql
rate(bigip_f5os_storage_disk_write_latency_ms[5m])
/
clamp_min(rate(bigip_f5os_storage_disk_write_iops[5m]), 1)
> 50
```
**For:** `5m`
**Description:** `Disk {{ $labels.disk_name }} on {{ $labels.target }} write latency is above 50ms per operation.`

### Disk Read Latency High — `warning`
```promql
rate(bigip_f5os_storage_disk_read_latency_ms[5m])
/
clamp_min(rate(bigip_f5os_storage_disk_read_iops[5m]), 1)
> 50
```
**For:** `5m`
**Description:** `Disk {{ $labels.disk_name }} on {{ $labels.target }} read latency is above 50ms per operation.`

---

## Hardware (Temperature / Fans / PSU)

> These alerts only fire on physical appliances. Tune thresholds to your
> platform's rated limits.

### Platform Temperature High — `critical`
```promql
bigip_f5os_temperature_current_celsius > 70
```
**For:** `5m`
**Description:** `Platform temperature on {{ $labels.target }} is above 70°C.`

### Platform Temperature Near Max — `warning`
> Relative to the platform's own reported maximum (model-independent).
```promql
bigip_f5os_temperature_current_celsius >= bigip_f5os_temperature_maximum_celsius
```
**For:** `5m`
**Description:** `Platform temperature on {{ $labels.target }} has reached its reported maximum.`

### Chassis Fan Stopped — `critical`
```promql
bigip_f5os_fan_speed_rpm < 1000
```
**For:** `2m`
**Description:** `Fan {{ $labels.fan_name }} on {{ $labels.target }} is reporting near-zero RPM.`

### PSU Fan Stopped — `critical`
```promql
bigip_f5os_psu_fan_1_speed_rpm < 1000
```
**For:** `2m`
**Description:** `PSU {{ $labels.psu_name }} fan on {{ $labels.target }} is reporting near-zero RPM.`

### PSU Output Voltage Loss — `critical`
> The r10800 PSU nominal output is ~12V; a drop below 11V catches a real fault
> or a removed/failed supply.
```promql
bigip_f5os_psu_voltage_out_volt < 11
```
**For:** `2m`
**Description:** `PSU {{ $labels.psu_name }} on {{ $labels.target }} output voltage has dropped below 11V.`

### PSU Overtemperature — `warning`
```promql
max by (target, psu_name) (
  bigip_f5os_psu_temperature_1_celsius
  or bigip_f5os_psu_temperature_2_celsius
  or bigip_f5os_psu_temperature_3_celsius
) > 55
```
**For:** `5m`
**Description:** `PSU {{ $labels.psu_name }} on {{ $labels.target }} has a sensor above 55°C.`

---

## Security (TPM)

### TPM Integrity Not Valid — `critical`
> The `tmp_status` label carries the TPM integrity state. Anything other than
> `Valid` indicates a platform-integrity concern.
```promql
bigip_f5os_info{tmp_status!="Valid"}
```
**For:** `5m`
**Description:** `TPM integrity status on {{ $labels.target }} is '{{ $labels.tmp_status }}' (expected 'Valid').`

---

## Interfaces

> Interface counters are cumulative. All rate-based alerts use `rate()` so they
> remain volume-independent and comparable across links of different speeds.

### Interface Down — `critical`
> `oper_status` is exposed both as a label on the info metric and as the numeric
> gauge below. The gauge is easiest to alert on.
```promql
bigip_f5os_interface_oper_status_up == 0
```
**For:** `2m`
**Description:** `Interface {{ $labels.interface }} on {{ $labels.target }} is operationally down.`

### Interface Input Errors — `warning`
```promql
rate(bigip_f5os_interface_in_errors_total[5m]) > 1
```
**For:** `10m`
**Description:** `Interface {{ $labels.interface }} on {{ $labels.target }} is receiving errors.`

### Interface Output Errors — `warning`
```promql
rate(bigip_f5os_interface_out_errors_total[5m]) > 1
```
**For:** `10m`
**Description:** `Interface {{ $labels.interface }} on {{ $labels.target }} is transmitting errors.`

### Interface Discards — `warning`
> Discards indicate congestion or buffer exhaustion.
```promql
rate(bigip_f5os_interface_in_discards_total[5m])
+ rate(bigip_f5os_interface_out_discards_total[5m])
> 10
```
**For:** `10m`
**Description:** `Interface {{ $labels.interface }} on {{ $labels.target }} is discarding packets.`

### Interface CRC / FCS Errors — `warning`
> Physical-layer problems (cabling, optics, duplex mismatch).
```promql
rate(bigip_f5os_interface_ethernet_in_crc_errors_total[5m])
+ rate(bigip_f5os_interface_in_fcs_errors_total[5m])
> 0
```
**For:** `10m`
**Description:** `Interface {{ $labels.interface }} on {{ $labels.target }} has CRC/FCS errors — check cabling/optics.`

---

## LACP

> LACP bundling health. A member that is not in-sync, collecting and
> distributing is not carrying traffic in the LAG.

### LACP Member Out of Sync — `critical`
```promql
bigip_f5os_lacp_member_in_sync == 0
```
**For:** `2m`
**Description:** `LACP member {{ $labels.interface }} in LAG {{ $labels.lag_name }} on {{ $labels.target }} is out of sync.`

### LACP Member Not Distributing — `warning`
```promql
bigip_f5os_lacp_member_distributing == 0
```
**For:** `5m`
**Description:** `LACP member {{ $labels.interface }} in LAG {{ $labels.lag_name }} on {{ $labels.target }} is not distributing traffic.`

### LACP RX Errors — `warning`
```promql
rate(bigip_f5os_lacp_rx_errors_total[5m]) > 0
```
**For:** `10m`
**Description:** `LACP member {{ $labels.interface }} in LAG {{ $labels.lag_name }} on {{ $labels.target }} is receiving LACP errors.`

---

## License

> The service-check date is exposed as a Unix timestamp, so age-based alerting
> is done at query time with `time()`.

### License Service Check Overdue — `warning`
```promql
(time() - bigip_f5os_license_service_check_date_seconds) / 86400 > 365
```
**For:** `1h`
**Description:** `License service check on {{ $labels.target }} is more than 365 days old — verify license/entitlement status.`

### License Service Check Aging — `warning`
```promql
(time() - bigip_f5os_license_service_check_date_seconds) / 86400 > 300
and
(time() - bigip_f5os_license_service_check_date_seconds) / 86400 <= 365
```
**For:** `1h`
**Description:** `License service check on {{ $labels.target }} is over 300 days old — plan a re-check before it lapses.`

---

## Notes

- All alerts include `{{ $labels.target }}` in their description to identify the
  affected device.
- The `for` durations are tuned to avoid alerting on brief spikes. Lower them
  for faster detection, raise them to reduce noise.
- In high-throughput environments, prefer **ratio-based** or **rate-based**
  alerts (like the disk latency ratio and interface error rates) over
  absolute-count thresholds.
- Counter-based metrics (`*_total`, disk IOPS/bytes/latency) must be wrapped in
  `rate()` — never alert on their raw cumulative value.
- Hardware thresholds (70°C, 11V, 1000 RPM, 55°C) should be aligned with your
  specific platform's datasheet limits.