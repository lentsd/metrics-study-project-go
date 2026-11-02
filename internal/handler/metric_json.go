package handler

import (
	"encoding/json"
	"io"
	"net/http"

	models "metrics-study-project-go/internal/model"
)

func UpdateJSONHandler(rw http.ResponseWriter, r *http.Request, storage MetricStorage) {
	metric, ok := decodeMetric(rw, r)
	if !ok {
		return
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			writeJSONError(rw, http.StatusBadRequest, "Missing gauge value")
			return
		}
		storage.SetGauge(metric.ID, *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			writeJSONError(rw, http.StatusBadRequest, "Missing counter delta")
			return
		}
		storage.AddCounter(metric.ID, *metric.Delta)
	}

	updated, _ := getJSONMetric(storage, metric)
	writeJSON(rw, http.StatusOK, updated)
}

func ValueJSONHandler(rw http.ResponseWriter, r *http.Request, storage MetricStorage) {
	metric, ok := decodeMetric(rw, r)
	if !ok {
		return
	}

	stored, ok := getJSONMetric(storage, metric)
	if !ok {
		writeJSONError(rw, http.StatusNotFound, "No metric found")
		return
	}

	writeJSON(rw, http.StatusOK, stored)
}

func decodeMetric(rw http.ResponseWriter, r *http.Request) (models.Metrics, bool) {
	var metric *models.Metrics
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&metric); err != nil || metric == nil {
		writeJSONError(rw, http.StatusBadRequest, "Invalid metric JSON")
		return models.Metrics{}, false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeJSONError(rw, http.StatusBadRequest, "Expected a single JSON object")
		return *metric, false
	}
	if metric.ID == "" {
		writeJSONError(rw, http.StatusNotFound, "Empty name")
		return *metric, false
	}
	if metric.MType != models.Gauge && metric.MType != models.Counter {
		writeJSONError(rw, http.StatusBadRequest, "Only gauge and counter type allowed")
		return *metric, false
	}

	return *metric, true
}

func getJSONMetric(storage MetricStorage, metric models.Metrics) (models.Metrics, bool) {
	result := models.Metrics{ID: metric.ID, MType: metric.MType}
	switch metric.MType {
	case models.Gauge:
		value, ok := storage.GetGauge(metric.ID)
		result.Value = &value
		return result, ok
	case models.Counter:
		delta, ok := storage.GetCounter(metric.ID)
		result.Delta = &delta
		return result, ok
	default:
		return result, false
	}
}

func writeJSONError(rw http.ResponseWriter, status int, message string) {
	writeJSON(rw, status, map[string]string{"error": message})
}

func writeJSON(rw http.ResponseWriter, status int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		status = http.StatusInternalServerError
		body = []byte(`{"error":"Failed to encode response"}`)
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	_, _ = rw.Write(body)
}
