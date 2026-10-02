package main

import (
	"net/http"

	"metrics-study-project-go/internal/handler"
	"metrics-study-project-go/internal/repository"
)

func main() {
	memStorage := repository.InitMemStorage()

	http.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		handler.MetricHandler(w, r, memStorage)
	})

	http.ListenAndServe(":8080", nil)
}
