package main

import (
	"net/http"
	"os"

	"flag"

	"metrics-study-project-go/internal/handler"
	"metrics-study-project-go/internal/repository"

	"github.com/go-chi/chi/v5"
)

var flagRunAddr string

func main() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")

	flag.Parse()

	setupServerAddress()

	memStorage := repository.InitMemStorage()

	router := chi.NewRouter()

	router.Post("/update/{metricType}/{metricName}/{metricValue}", func(rw http.ResponseWriter, r *http.Request) {
		handler.MetricHandler(rw, r, memStorage)
	})

	router.Get("/value/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
		handler.ValueHandler(rw, r, memStorage)
	})

	router.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		handler.HTMLHandler(rw, r, memStorage)
	})

	http.ListenAndServe(flagRunAddr, router)
}

func setupServerAddress() {
	serverAddress, ok := os.LookupEnv("ADDRESS")

	if ok {
		flagRunAddr = serverAddress
	}
}
