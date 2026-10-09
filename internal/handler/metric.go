package handler

import (
	"io"
	"net/http"
	"strconv"

	models "metrics-study-project-go/internal/model"
	"metrics-study-project-go/internal/repository"

	"github.com/go-chi/chi/v5"
)

// URL format: /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>

type MetricStorage interface {
	SetGauge(string, float64)
	AddCounter(string, int64)
	GetMetric(string, string) (string, bool)
	GetAllMetrics() []repository.Metric
}

func MetricHandler(res http.ResponseWriter, req *http.Request, metricStorage MetricStorage) {

	metricType := chi.URLParam(req, "metricType")
	metricName := chi.URLParam(req, "metricName")
	metricValue := chi.URLParam(req, "metricValue")

	if metricName == "" {
		http.Error(res, "Empty name", http.StatusNotFound)

		return
	}

	if metricType == "" || metricValue == "" {
		http.Error(res, "Wrong format!", http.StatusBadRequest)

		return
	}

	switch metricType {
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(res, "Not allowed metric value type", http.StatusBadRequest)

			return
		}

		metricStorage.SetGauge(metricName, value)

	case models.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(res, "Not allowed metric value type", http.StatusBadRequest)

			return
		}

		metricStorage.AddCounter(metricName, value)

	default:
		http.Error(res, "Only gauge and counter type allowed!", http.StatusBadRequest)

		return
	}
}

func ValueHandler(rw http.ResponseWriter, r *http.Request, metricStorage MetricStorage) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if metricType == "" || metricName == "" {
		http.Error(rw, "Wrong format (no name or type)", http.StatusBadRequest)

		return
	}

	res, ok := metricStorage.GetMetric(metricType, metricName)

	if !ok {
		http.Error(rw, "No metric found", http.StatusNotFound)

		return
	}

	io.WriteString(rw, res)
}
