# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0]

### Changed — ⚠️ BREAKING: Prometheus metric-naming compliance pass

This release brings every metric exposed by the exporter in line with the
[Prometheus metric and label naming best practices](https://prometheus.io/docs/practices/naming/).
**Every dashboard, alert rule, and recording rule that references the metrics
below must be updated.** There is no compatibility shim — metric names and,
in several cases, metric *values* have changed.

- **Renamed metrics to `snake_case`**: `bigip_f5os_Memory_available_bytes` →
  `bigip_f5os_memory_available_bytes` (and 3 sibling `Memory_*` metrics).
- **Converted percentages to base-unit ratios (0–1 instead of 0–100)**, per
  the Prometheus convention that `ratio` values must be 0–1:
  `bigip_f5os_Memory_used_percent` → `bigip_f5os_memory_used_ratio`,
  `bigip_f5os_cpu_utilization_percent` and its 5sec/1min/5min-average and
  per-thread siblings (8 metrics total) → `..._ratio`. Values are now
  divided by 100.
  - **Action required:** any alert/dashboard threshold comparing these
    metrics against `85`, `90`, etc. must be rewritten against `0.85`, `0.90`.
- **Converted non-base physical units**:
  - `bigip_f5os_cpu_cache_size_kb` → `bigip_f5os_cpu_cache_size_bytes`
    (×1024).
  - `bigip_f5os_storage_disk_size_gb` → `bigip_f5os_storage_disk_size_bytes`
    (×1e9).
  - `bigip_f5os_psu_current_{in,out}_amp` → `..._amperes`.
  - `bigip_f5os_psu_voltage_{in,out}_volt` → `..._volts`.
- **Renamed disk I/O metrics to remove misleading rate-style units.** These
  metrics are reported by F5OS as cumulative counters (not instantaneous
  rates), despite field names like `total-iops` and `read-latency-ms`:
  - `bigip_f5os_storage_disk_{total,read,write}_iops` →
    `bigip_f5os_storage_disk_{,read_,write_}operations_total`.
  - `bigip_f5os_storage_disk_{read,write}_latency_ms` →
    `bigip_f5os_storage_disk_{read,write}_latency_seconds_total` (values
    divided by 1000 to convert ms→seconds).
  - `bigip_f5os_storage_disk_{read,write}_merged` →
    `bigip_f5os_storage_disk_{read,write}_merged_total` — these were also
    fixed from `GaugeValue` to `CounterValue` to match their documented
    cumulative-counter semantics.
- **Fixed a `tmp_status`/`tpm_status` label-name mismatch**: `README.md`,
  `alerts/README.md`, and the Grafana dashboard referenced a `tmp_status`
  label that did not exist on `bigip_f5os_info` (the code has always used
  `tpm_status`), silently breaking the documented "TPM Integrity Not Valid"
  alert. Docs and dashboard now use the correct `tpm_status` label.
- Updated `alerts/README.md` and `grafana/bigip_f5os_dashboard.json` to use
  every renamed metric/label and rescaled threshold (ratio panels now use
  Grafana's `percentunit` field type with 0–1 thresholds instead of
  `percent` with 0–100 thresholds; disk size/cache size/latency panel units
  updated to bytes/bytes/seconds).

### Migration notes
- Grep your rule files / dashboards for the old names listed above and
  update them before upgrading the exporter in production, or you will
  silently lose alerting/graphing coverage for the affected metrics.
- If you depend on absolute values in recording rules for the rescaled
  metrics (CPU/memory ratios, disk/cache sizes, disk latency), update the
  downstream math — the divide-by-100/1000/multiply-by-1024/1e9 conversions
  are now applied by the exporter itself.
- If you had a working "TPM Integrity Not Valid" alert before this release,
  it was never actually firing due to the `tmp_status`/`tpm_status` typo.
  Re-deploy the corrected alert from `alerts/README.md`.

## [0.1.0] - Initial

### Added
- Initial release of the F5OS Prometheus exporter.
- Collectors for Hardware (CPU, memory, temperature, fans, PSU, storage
  disks), Interfaces, LACP, and License.
- Multi-target `/probe` endpoint with concurrent scraping.
- Token-based authentication against the F5OS REST API.

[Unreleased]: https://github.com/Haameed/f5_f5os_exporter/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Haameed/f5_f5os_exporter/compare/v0.1.0...v1.0.0
[0.1.0]: https://github.com/Haameed/f5_f5os_exporter/releases/tag/v0.1.0
