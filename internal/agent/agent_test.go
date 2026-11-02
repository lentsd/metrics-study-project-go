package agent

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	models "metrics-study-project-go/internal/model"
)

func TestGetMetrics(t *testing.T) {
	agent := CreateAgent("localhost:8080", 10, 2)

	agent.getMetrics()

	if agent.pollCount != 1 {
		t.Errorf("pollCount = %d, want 1", agent.pollCount)
	}

	names := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
		"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
		"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
		"StackSys", "Sys", "TotalAlloc", "RandomValue",
	}
	if len(agent.gaugeMetrics) != len(names) {
		t.Errorf("gauge count = %d, want %d", len(agent.gaugeMetrics), len(names))
	}
	for _, name := range names {
		if _, ok := agent.gaugeMetrics[name]; !ok {
			t.Errorf("missing gauge %s", name)
		}
	}

	agent.getMetrics()

	if agent.pollCount != 2 {
		t.Errorf("pollCount = %d, want 2", agent.pollCount)
	}

}

func TestSendMetric(t *testing.T) {
	tests := []struct {
		name string
		send func(*Agent) error
		want string
	}{
		{"gauge", func(a *Agent) error { return a.sendMetric("Alloc", 12.5) }, `{"id":"Alloc","type":"gauge","value":12.5}`},
		{"zero gauge", func(a *Agent) error { return a.sendMetric("Alloc", 0) }, `{"id":"Alloc","type":"gauge","value":0}`},
		{"counter", func(a *Agent) error { return a.sendCounter("PollCount", 3) }, `{"id":"PollCount","type":"counter","delta":3}`},
		{"zero counter", func(a *Agent) error { return a.sendCounter("PollCount", 0) }, `{"id":"PollCount","type":"counter","delta":0}`},
		{"counter precision", func(a *Agent) error { return a.sendCounter("PollCount", math.MaxInt64) }, `{"id":"PollCount","type":"counter","delta":9223372036854775807}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got models.Metrics
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodPost || r.URL.Path != "/update" {
					t.Errorf("request = %s %s, want POST /update", r.Method, r.URL.Path)
				}
				if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
					t.Errorf("Content-Type = %s, want application/json", contentType)
				}
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Errorf("decode request: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(got)
			}))
			t.Cleanup(server.Close)
			agent := CreateAgent(server.Listener.Addr().String(), 10, 2)
			if err := tt.send(agent); err != nil {
				t.Fatalf("send metric: %v", err)
			}
			var want models.Metrics
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			if requests != 1 || !reflect.DeepEqual(got, want) {
				t.Errorf("requests = %d, metric = %+v, want one request with %s", requests, got, tt.want)
			}
		})
	}
}

func TestSendAllMetrics(t *testing.T) {
	received := make(map[string]models.Metrics)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		var metric models.Metrics
		if r.Method != http.MethodPost || r.URL.Path != "/update" || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected metric request: %s %s, Content-Type %q", r.Method, r.URL.Path, r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			t.Errorf("decode metric: %v", err)
		}
		received[metric.ID] = metric
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(metric)
	}))
	t.Cleanup(server.Close)
	agent := CreateAgent(server.Listener.Addr().String(), 10, 2)
	agent.getMetrics()
	agent.getMetrics()
	if err := agent.sendMetrics(); err != nil {
		t.Fatalf("sendMetrics: %v", err)
	}
	if requests != len(agent.gaugeMetrics)+1 || len(received) != len(agent.gaugeMetrics)+1 {
		t.Fatalf("requests = %d, received %d metrics, want %d", requests, len(received), len(agent.gaugeMetrics)+1)
	}
	for name, value := range agent.gaugeMetrics {
		metric := received[name]
		if metric.MType != models.Gauge || metric.Value == nil || *metric.Value != value || metric.Delta != nil {
			t.Errorf("gauge %s = %+v, want value %g", name, metric, value)
		}
	}
	counter := received["PollCount"]
	if counter.MType != models.Counter || counter.Delta == nil || *counter.Delta != 2 || counter.Value != nil {
		t.Errorf("PollCount = %+v, want counter delta 2", counter)
	}
	if agent.pollCount != 0 {
		t.Errorf("pollCount = %d after successful send, want 0", agent.pollCount)
	}
}

func TestSendMetricsFailurePreservesPollCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)
	agent := CreateAgent(server.Listener.Addr().String(), 10, 2)
	agent.getMetrics()
	if err := agent.sendMetrics(); err == nil {
		t.Error("sendMetrics succeeded for status 503")
	}
	if agent.pollCount != 1 {
		t.Errorf("pollCount = %d after failed send, want 1", agent.pollCount)
	}
}
