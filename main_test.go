package main

import (
	"testing"
	"os"
	"strings"
	"net/http"
	"net/http/httptest"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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