package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestReadValue(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    float64
		wantOk  bool
	}{
		{"valid value", "37600\n", 37600, true},
		{"without newline", "4936800", 4936800, true},
		{"empty", "", 0, false},
		{"not a number", "abc\n", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.CreateTemp("", "test")
			defer os.Remove(f.Name())
			f.WriteString(tt.content)
			f.Close()

			got, ok := readValue(f.Name())
			if ok != tt.wantOk || got != tt.want {
				t.Errorf("got (%v, %v), want (%v, %v)", got, ok, tt.want, tt.wantOk)
			}
		})
	}
}
func TestReadValue_FileNotFound(t *testing.T) {
	_, ok := readValue("/path/that/do/not/exists")
	if ok {
		t.Error("expects false for file that dont exists")
	}
}
func TestUpdateMetrics(t *testing.T) {
	f, _ := os.CreateTemp("", "test")
	defer os.Remove(f.Name())
	f.WriteString("25000\n")
	f.Close()

	path := f.Name()
	missing := "/path/that/do/not/exists"

	before := testutil.ToFloat64(scrapeErrors.WithLabelValues("missing_sensor"))

	updateMetrics([]sensor{
		{"ok_sensor", &path, cpuTempC, 1000},
		{"missing_sensor", &missing, pmicTempC, 1000},
	})

	if cpuTempC.Desc() == nil {
		t.Fatal("cpuTempC not registered")
	}
	if got := testutil.ToFloat64(cpuTempC); got != 25 {
		t.Errorf("cpuTempC = %v, want 25", got)
	}
	if after := testutil.ToFloat64(scrapeErrors.WithLabelValues("missing_sensor")); after != before+1 {
		t.Errorf("scrapeErrors[missing_sensor] = %v, want %v", after, before+1)
	}
}

func TestMetricsHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	cpuTempC.Set(42.0)

	promhttp.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expects 200, received %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "armbian_cpu_temp_celsius") {
		t.Error("expected metrics not found in output")
	}
}
