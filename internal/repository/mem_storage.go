package repository

import "strconv"

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (storage *MemStorage) SetGauge(name string, value float64) {
	storage.gauges[name] = value
}

func (storage *MemStorage) AddCounter(name string, value int64) {
	storage.counters[name] += value
}

func (storage *MemStorage) GetMetric(metricType string, metricName string) (string, bool) {
	switch metricType {
	case "gauge":
		res, ok := storage.gauges[metricName]

		return strconv.FormatFloat(res, 'f', -1, 64), ok
	case "counter":
		res, ok := storage.counters[metricName]

		return strconv.FormatInt(res, 10), ok
	default:
		return "", false
	}
}

type Metric struct {
	MetricName string
	MetricType string
	Value      string
}

func (storage *MemStorage) GetAllMetrics() []Metric {
	var res []Metric

	for gaugeName, gaugeValue := range storage.gauges {
		res = append(res, Metric{
			MetricName: gaugeName,
			MetricType: "gauge",
			Value:      strconv.FormatFloat(gaugeValue, 'f', -1, 64),
		})
	}

	for counterName, counterValue := range storage.counters {
		res = append(res, Metric{
			MetricName: counterName,
			MetricType: "counter",
			Value:      strconv.FormatInt(counterValue, 10),
		})
	}

	return res
}

func InitMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}
