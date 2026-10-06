package agent

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"
)

type Agent struct {
	pollCount     int64
	gaugeMetrics  map[string]float64
	serverAddress string
}

func (a *Agent) getMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	a.gaugeMetrics["Alloc"] = float64(memStats.Alloc)
	a.gaugeMetrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	a.gaugeMetrics["Frees"] = float64(memStats.Frees)
	a.gaugeMetrics["GCCPUFraction"] = memStats.GCCPUFraction
	a.gaugeMetrics["GCSys"] = float64(memStats.GCSys)
	a.gaugeMetrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	a.gaugeMetrics["HeapIdle"] = float64(memStats.HeapIdle)
	a.gaugeMetrics["HeapInuse"] = float64(memStats.HeapInuse)
	a.gaugeMetrics["HeapObjects"] = float64(memStats.HeapObjects)
	a.gaugeMetrics["HeapReleased"] = float64(memStats.HeapReleased)
	a.gaugeMetrics["HeapSys"] = float64(memStats.HeapSys)
	a.gaugeMetrics["LastGC"] = float64(memStats.LastGC)
	a.gaugeMetrics["Lookups"] = float64(memStats.Lookups)
	a.gaugeMetrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	a.gaugeMetrics["MCacheSys"] = float64(memStats.MCacheSys)
	a.gaugeMetrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	a.gaugeMetrics["MSpanSys"] = float64(memStats.MSpanSys)
	a.gaugeMetrics["Mallocs"] = float64(memStats.Mallocs)
	a.gaugeMetrics["NextGC"] = float64(memStats.NextGC)
	a.gaugeMetrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	a.gaugeMetrics["NumGC"] = float64(memStats.NumGC)
	a.gaugeMetrics["OtherSys"] = float64(memStats.OtherSys)
	a.gaugeMetrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	a.gaugeMetrics["StackInuse"] = float64(memStats.StackInuse)
	a.gaugeMetrics["StackSys"] = float64(memStats.StackSys)
	a.gaugeMetrics["Sys"] = float64(memStats.Sys)
	a.gaugeMetrics["TotalAlloc"] = float64(memStats.TotalAlloc)

	a.pollCount++
	a.gaugeMetrics["RandomValue"] = rand.Float64()
}

func (a *Agent) sendMetric(name string, value float64) error {
	address := fmt.Sprintf("%s/update/gauge/%s/%g", a.serverAddress, name, value)

	response, err := http.Post(address, "text/plain", nil)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("send metric: %s status %d", name, response.StatusCode)
	}

	return nil
}

func (a *Agent) sendCounter(name string, value int64) error {
	address := fmt.Sprintf("%s/update/counter/%s/%d", a.serverAddress, name, value)

	response, err := http.Post(address, "text/plain", nil)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("send counter: %s status %d", name, response.StatusCode)
	}

	return nil
}

func (a *Agent) sendMetrics() error {
	for name, value := range a.gaugeMetrics {
		err := a.sendMetric(name, value)
		if err != nil {
			return err
		}
	}

	err := a.sendCounter("PollCount", a.pollCount)

	if err != nil {
		return err
	}

	a.pollCount = 0
	return nil
}

func (a *Agent) Run() {
	lastUpdate := time.Now()

	for {
		a.getMetrics()

		if time.Since(lastUpdate) >= time.Second*10 {
			err := a.sendMetrics()

			if err != nil {
				fmt.Println("send metrics:", err)
			}

			lastUpdate = time.Now()
		}

		time.Sleep(time.Second * 2)
	}
}

func CreateAgent() *Agent {
	var agent Agent

	agent.serverAddress = "http://localhost:8080"
	agent.gaugeMetrics = make(map[string]float64)

	return &agent
}
