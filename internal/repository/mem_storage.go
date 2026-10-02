package repository

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

func InitMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}
