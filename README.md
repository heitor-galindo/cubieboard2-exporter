# cubieboard2-exporter

[![CI](https://github.com/heitor-galindo/cubieboard2-exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/heitor-galindo/cubieboard2-exporter/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Custom Prometheus exporter that surfaces the same metrics shown by `armbianmonitor` (CPU temp, PMIC temp, DC-IN voltage, cooling state, etc.) for boards not natively covered by `node_exporter`.

Built for the Cubieboard2 (Allwinner A20 + AXP209 PMIC) running Armbian's legacy sunxi kernel, where these sensors are exposed through non-standard sysfs paths that `node_exporter`'s `hwmon` collector doesn't pick up.

## Metrics

| Metric | Description |
| --- | --- |
| `armbian_cpu_temp_celsius` | SoC temperature |
| `armbian_pmic_temp_celsius` | PMIC (AXP209) temperature |
| `armbian_ac_voltage_volts` | DC-IN voltage from AC input |
| `armbian_ac_current_amps` | AC input current draw |
| `armbian_ac_connected` | 1 if AC power is connected, 0 otherwise |
| `armbian_vbus_voltage_volts` | USB VBUS voltage |
| `armbian_cooling_state` | Current thermal cooling/throttle state |
| `armbian_cooling_max_state` | Maximum thermal cooling/throttle state |
| `armbian_exporter_scrape_errors_total{sensor}` | Count of failed sysfs reads, per sensor |

## Usage

```
./cubieboard2-exporter [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `-listen-address` | `:9101` | Address to listen on for metrics |
| `-cpu-temp-path` | `/sys/devices/virtual/thermal/thermal_zone0/temp` | Path to the CPU temperature sysfs file |
| `-pmic-temp-path` | `/sys/power/axp_pmu/pmu/temp` | Path to the PMIC temperature sysfs file |
| `-ac-voltage-path` | `/sys/power/axp_pmu/ac/voltage` | Path to the AC voltage sysfs file |
| `-ac-current-path` | `/sys/power/axp_pmu/ac/amperage` | Path to the AC current sysfs file |
| `-ac-connected-path` | `/sys/power/axp_pmu/ac/connected` | Path to the AC connected sysfs file |
| `-vbus-voltage-path` | `/sys/power/axp_pmu/vbus/voltage` | Path to the VBUS voltage sysfs file |
| `-cooling-state-path` | `/sys/class/thermal/cooling_device0/cur_state` | Path to the cooling state sysfs file |
| `-cooling-max-state-path` | `/sys/class/thermal/cooling_device0/max_state` | Path to the cooling max state sysfs file |

The flags let the exporter run on other Allwinner/AXP209 boards or kernels where these sysfs paths differ.

Metrics are served at `/metrics`.

## Build

Requires Go (see `go.mod` for the version). Cross-compiles for the Cubieboard2's ARMv7 target:

```
make build   # runs tests, then builds ./cubieboard2-exporter for GOOS=linux GOARCH=arm GOARM=7
make test    # go test -v
```

## License

MIT — see [LICENSE](LICENSE).
