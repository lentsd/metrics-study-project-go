package agent

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMetrics(t *testing.T) {
	agent := CreateAgent()

	agent.getMetrics()

	if agent.pollCount != 1 {
		t.Errorf("pollCount = %d, want 1", agent.pollCount)
	}

	if _, ok := agent.gaugeMetrics["RandomValue"]; !ok {
		t.Error("No RandomValue")
	}

	agent.getMetrics()

	if agent.pollCount != 2 {
		t.Errorf("pollCount = %d, want 2", agent.pollCount)
	}

}

func TestSendMetric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}

			if r.URL.Path != "/update/gauge/Alloc/12.5" {
				t.Errorf("path = %s, want /update/gauge/Alloc/12.5", r.URL.Path)
			}

			if got := r.Header.Get("Content-Type"); got != "text/plain" {
				t.Errorf("Content-Type = %s, want text/plain", got)
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			if len(body) != 0 {
				t.Errorf("body = %q, want empty", body)
			}

			w.WriteHeader(http.StatusOK)
		},
	))

	t.Cleanup(server.Close)

	agent := CreateAgent()
	agent.serverAddress = server.URL

	err := agent.sendMetric("Alloc", 12.5)
	if err != nil {
		t.Fatalf("sendMetric() error = %v", err)
	}
}
