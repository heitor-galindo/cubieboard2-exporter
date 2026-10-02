package main

import (
	"flag"
	"log"
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
	scrapeErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "armbian_exporter_scrape_errors_total",
		Help: "Number of failed sysfs reads, by sensor",
	}, []string{"sensor"})
)

func init() {
	prometheus.MustRegister(cpuTempC, pmicTempC, acVoltage, acCurrent, acConnected, vbusVoltage, coolingState, coolingMaxState, scrapeErrors)
}

// sensor binds a sysfs file to the gauge it feeds, with the divisor needed
// to convert the raw kernel value (milli-units) into the gauge's unit.
type sensor struct {
	name  string
	path  *string
	gauge prometheus.Gauge
	scale float64
}

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

func updateMetrics(sensors []sensor) {
	for _, s := range sensors {
		v, ok := readValue(*s.path)
		if !ok {
			scrapeErrors.WithLabelValues(s.name).Inc()
			continue
		}
		s.gauge.Set(v / s.scale)
	}
}

func metricsHandler(sensors []sensor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateMetrics(sensors)
		promhttp.Handler().ServeHTTP(w, r)
	}
}

func main() {
	listenAddress := flag.String("listen-address", ":9101", "Address to listen on for metrics")
	cpuTempPath := flag.String("cpu-temp-path", "/sys/devices/virtual/thermal/thermal_zone0/temp", "Path to the CPU temperature sysfs file")
	pmicTempPath := flag.String("pmic-temp-path", "/sys/power/axp_pmu/pmu/temp", "Path to the PMIC temperature sysfs file")
	acVoltagePath := flag.String("ac-voltage-path", "/sys/power/axp_pmu/ac/voltage", "Path to the AC voltage sysfs file")
	acCurrentPath := flag.String("ac-current-path", "/sys/power/axp_pmu/ac/amperage", "Path to the AC current sysfs file")
	acConnectedPath := flag.String("ac-connected-path", "/sys/power/axp_pmu/ac/connected", "Path to the AC connected sysfs file")
	vbusVoltagePath := flag.String("vbus-voltage-path", "/sys/power/axp_pmu/vbus/voltage", "Path to the VBUS voltage sysfs file")
	coolingStatePath := flag.String("cooling-state-path", "/sys/class/thermal/cooling_device0/cur_state", "Path to the cooling state sysfs file")
	coolingMaxStatePath := flag.String("cooling-max-state-path", "/sys/class/thermal/cooling_device0/max_state", "Path to the cooling max state sysfs file")
	flag.Parse()

	sensors := []sensor{
		{"cpu_temp", cpuTempPath, cpuTempC, 1000},
		{"pmic_temp", pmicTempPath, pmicTempC, 1000},
		{"ac_voltage", acVoltagePath, acVoltage, 1000000},
		{"ac_current", acCurrentPath, acCurrent, 1000000},
		{"ac_connected", acConnectedPath, acConnected, 1},
		{"vbus_voltage", vbusVoltagePath, vbusVoltage, 1000000},
		{"cooling_state", coolingStatePath, coolingState, 1},
		{"cooling_max_state", coolingMaxStatePath, coolingMaxState, 1},
	}

	http.HandleFunc("/metrics", metricsHandler(sensors))
	log.Printf("listening on %s", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
