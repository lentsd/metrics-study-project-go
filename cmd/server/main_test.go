package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	models "metrics-study-project-go/internal/model"

	"go.uber.org/zap"
)

func postJSON(t *testing.T, router http.Handler, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != status {
		t.Fatalf("POST %s: status = %d, want %d; body = %s", path, res.Code, status, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if !json.Valid(res.Body.Bytes()) {
		t.Fatalf("response is not JSON: %s", res.Body.String())
	}
	return res
}

func assertMetric(t *testing.T, res *httptest.ResponseRecorder, wantJSON string) {
	t.Helper()
	var got, want models.Metrics
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatalf("decode expected metric: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("metric = %s, want %s", res.Body.String(), wantJSON)
	}
}

func TestJSONMetrics(t *testing.T) {
	tests := []struct {
		name    string
		updates []string
		query   string
		want    []string
	}{
		{
			name: "gauge replacement and zero",
			updates: []string{
				`{"id":"LastGC","type":"gauge","value":1744184459}`,
				`{"id":"LastGC","type":"gauge","value":-12.5}`,
				`{"id":"LastGC","type":"gauge","value":0}`,
			},
			query: `{"id":"LastGC","type":"gauge"}`,
			want: []string{
				`{"id":"LastGC","type":"gauge","value":1744184459}`,
				`{"id":"LastGC","type":"gauge","value":-12.5}`,
				`{"id":"LastGC","type":"gauge","value":0}`,
			},
		},
		{
			name: "counter accumulation and zero",
			updates: []string{
				`{"id":"PollCount","type":"counter","delta":0}`,
				`{"id":"PollCount","type":"counter","delta":7}`,
				`{"id":"PollCount","type":"counter","delta":-2}`,
			},
			query: `{"id":"PollCount","type":"counter"}`,
			want: []string{
				`{"id":"PollCount","type":"counter","delta":0}`,
				`{"id":"PollCount","type":"counter","delta":7}`,
				`{"id":"PollCount","type":"counter","delta":5}`,
			},
		},
		{
			name:    "counter precision",
			updates: []string{`{"id":"Total","type":"counter","delta":9223372036854775807}`},
			query:   `{"id":"Total","type":"counter"}`,
			want:    []string{`{"id":"Total","type":"counter","delta":9223372036854775807}`},
		},
		{
			name:    "only return fields for the requested type",
			updates: []string{`{"id":"Alloc","type":"gauge","value":3.5,"delta":9,"hash":"unused"}`},
			query:   `{"id":"Alloc","type":"gauge","value":999,"delta":9}`,
			want:    []string{`{"id":"Alloc","type":"gauge","value":3.5}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newRouter(zap.NewNop())
			for i, update := range tt.updates {
				assertMetric(t, postJSON(t, router, "/update", update, http.StatusOK), tt.want[i])
				assertMetric(t, postJSON(t, router, "/value", tt.query, http.StatusOK), tt.want[i])
			}
		})
	}
}

func TestJSONMetricErrors(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"empty body", "/update", "", http.StatusBadRequest},
		{"malformed JSON", "/update", `{"id":`, http.StatusBadRequest},
		{"null", "/update", `null`, http.StatusBadRequest},
		{"array", "/update", `[]`, http.StatusBadRequest},
		{"missing name", "/update", `{"type":"gauge","value":1}`, http.StatusNotFound},
		{"missing type", "/update", `{"id":"Alloc","value":1}`, http.StatusBadRequest},
		{"unsupported type", "/update", `{"id":"Alloc","type":"histogram","value":1}`, http.StatusBadRequest},
		{"missing gauge value", "/update", `{"id":"Alloc","type":"gauge"}`, http.StatusBadRequest},
		{"null gauge value", "/update", `{"id":"Alloc","type":"gauge","value":null}`, http.StatusBadRequest},
		{"missing counter delta", "/update", `{"id":"PollCount","type":"counter"}`, http.StatusBadRequest},
		{"null counter delta", "/update", `{"id":"PollCount","type":"counter","delta":null}`, http.StatusBadRequest},
		{"wrong value type", "/update", `{"id":"Alloc","type":"gauge","value":"1"}`, http.StatusBadRequest},
		{"fractional delta", "/update", `{"id":"PollCount","type":"counter","delta":1.5}`, http.StatusBadRequest},
		{"overflow delta", "/update", `{"id":"PollCount","type":"counter","delta":9223372036854775808}`, http.StatusBadRequest},
		{"multiple objects", "/update", `{"id":"Alloc","type":"gauge","value":1} {}`, http.StatusBadRequest},
		{"trailing garbage", "/update", `{"id":"Alloc","type":"gauge","value":1} trailing`, http.StatusBadRequest},
		{"malformed value query", "/value", `{`, http.StatusBadRequest},
		{"unknown query type", "/value", `{"id":"Alloc","type":"unknown"}`, http.StatusBadRequest},
		{"missing gauge", "/value", `{"id":"Unknown","type":"gauge"}`, http.StatusNotFound},
		{"missing counter", "/value", `{"id":"Unknown","type":"counter"}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newRouter(zap.NewNop())
			initial := `{"id":"Alloc","type":"gauge","value":12.5}`
			postJSON(t, router, "/update", initial, http.StatusOK)
			postJSON(t, router, tt.path, tt.body, tt.status)
			assertMetric(t, postJSON(t, router, "/value", `{"id":"Alloc","type":"gauge"}`, http.StatusOK), initial)
		})
	}
}

func TestMetricRoutesShareStorage(t *testing.T) {
	router := newRouter(zap.NewNop())
	for _, metricType := range []string{"gauge", "counter"} {
		req := httptest.NewRequest(http.MethodPost, "/update/"+metricType+"/Shared/5", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("legacy update: status = %d", res.Code)
		}
	}
	assertMetric(t, postJSON(t, router, "/value/", `{"id":"Shared","type":"gauge"}`, http.StatusOK), `{"id":"Shared","type":"gauge","value":5}`)
	assertMetric(t, postJSON(t, router, "/value/", `{"id":"Shared","type":"counter"}`, http.StatusOK), `{"id":"Shared","type":"counter","delta":5}`)
	postJSON(t, router, "/update/", `{"id":"Shared","type":"gauge","value":12.5}`, http.StatusOK)
	postJSON(t, router, "/update/", `{"id":"Shared","type":"counter","delta":2}`, http.StatusOK)

	for metricType, want := range map[string]string{"gauge": "12.5", "counter": "7"} {
		req := httptest.NewRequest(http.MethodGet, "/value/"+metricType+"/Shared", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != http.StatusOK || res.Body.String() != want {
			t.Errorf("legacy %s value: status = %d, body = %q, want 200 and %q", metricType, res.Code, res.Body.String(), want)
		}
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), "Shared") {
		t.Errorf("HTML metrics: status = %d, body = %q", res.Code, res.Body.String())
	}
}
