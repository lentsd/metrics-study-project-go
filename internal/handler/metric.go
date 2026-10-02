package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	models "metrics-study-project-go/internal/model"
)

// URL format: /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>

type MetricStorage interface {
	SetGauge(string, float64)
	AddCounter(string, int64)
}

func MetricHandler(res http.ResponseWriter, req *http.Request, metricStorage MetricStorage) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST method is allowed!", http.StatusMethodNotAllowed)

		return
	}

	pathParts := strings.Split(req.URL.Path, "/")

	if len(pathParts) < 4 || pathParts[3] == "" {
		http.Error(res, "Empty name", http.StatusNotFound)

		return
	}

	if len(pathParts) != 5 {
		http.Error(res, "Wrong format!", http.StatusBadRequest)

		return
	}

	metricType := pathParts[2]
	metricName := pathParts[3]
	metricValue := pathParts[4]

	fmt.Println(metricType, metricName, metricValue)

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
