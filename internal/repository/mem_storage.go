package repository

import (
	"strconv"
	"sync"
)

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func (storage *MemStorage) SetGauge(name string, value float64) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	storage.gauges[name] = value
}

func (storage *MemStorage) AddCounter(name string, value int64) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	storage.counters[name] += value
}

func (storage *MemStorage) GetGauge(name string) (float64, bool) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	value, ok := storage.gauges[name]
	return value, ok
}

func (storage *MemStorage) GetCounter(name string) (int64, bool) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	value, ok := storage.counters[name]
	return value, ok
}

func (storage *MemStorage) GetMetric(metricType string, metricName string) (string, bool) {
	switch metricType {
	case "gauge":
		res, ok := storage.GetGauge(metricName)

		return strconv.FormatFloat(res, 'f', -1, 64), ok
	case "counter":
		res, ok := storage.GetCounter(metricName)

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
	storage.mu.RLock()
	defer storage.mu.RUnlock()

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
