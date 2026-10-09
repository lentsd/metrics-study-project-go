package main

import (
	"net/http"

	"metrics-study-project-go/internal/handler"
	"metrics-study-project-go/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	memStorage := repository.InitMemStorage()

	router := chi.NewRouter()

	router.Post("/update/{metricType}/{metricName}/{metricValue}", func(rw http.ResponseWriter, r *http.Request) {
		handler.MetricHandler(rw, r, memStorage)
	})

	router.Get("/value/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
		handler.ValueHandler(rw, r, memStorage)
	})

	router.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		handler.HtmlHandler(rw, r, memStorage)
	})

	http.ListenAndServe(":8080", router)
}
