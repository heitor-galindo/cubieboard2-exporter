package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cpuTempC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "armbian_cpu_temp_celsius",
		Help: "SoC temperature",
	})
	pmicTempC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "armbian_pmic_temp_celsius",
		Help: "PMIC (AXP209) temperature",
	})
	acVoltage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "armbian_ac_voltage_volts",
		Help: "DC-IN voltage from AC input",
	})
	acCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "armbian_ac_current_amps",
		Help: "AC input current draw",
	})
	acConnected = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "armbian_ac_connected",
		Help: "1 if AC power is connected, 0 otherwise",
	})
	vbusVoltage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "armbian_vbus_voltage_volts",
		Help: "USB VBUS voltage",
	})
	coolingState = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "armbian_cooling_state",
		Help: "Current thermal cooling/throttle state",
	})
	coolingMaxState = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "armbian_cooling_max_state",
		Help: "Maximum thermal cooling/throttle state",
	})
)

func readValue(path string) (float64, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		return 0, false
	}
	return val, true
}

func updateMetrics() {
	if v, ok := readValue("/sys/devices/virtual/thermal/thermal_zone0/temp"); ok {
		cpuTempC.Set(v / 1000)
	}
	if v, ok := readValue("/sys/power/axp_pmu/pmu/temp"); ok {
		pmicTempC.Set(v / 1000)
	}
	if v, ok := readValue("/sys/power/axp_pmu/ac/voltage"); ok {
		acVoltage.Set(v / 1000000)
	}
	if v, ok := readValue("/sys/power/axp_pmu/ac/amperage"); ok {
		acCurrent.Set(v / 1000000)
	}
	if v, ok := readValue("/sys/power/axp_pmu/ac/connected"); ok {
		acConnected.Set(v)
	}
	if v, ok := readValue("/sys/power/axp_pmu/vbus/voltage"); ok {
		vbusVoltage.Set(v / 1000000)
	}
	if v, ok := readValue("/sys/class/thermal/cooling_device0/cur_state"); ok {
		coolingState.Set(v)
	}
	if v, ok := readValue("/sys/class/thermal/cooling_device0/max_state"); ok {
		coolingMaxState.Set(v)
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	updateMetrics()
	promhttp.Handler().ServeHTTP(w, r)
}

func init() {
	prometheus.MustRegister(cpuTempC, pmicTempC, acVoltage, acCurrent, acConnected, vbusVoltage, coolingState, coolingMaxState)
}

func main() {
	http.HandleFunc("/metrics", metricsHandler)
	http.ListenAndServe(":9101", nil)
}