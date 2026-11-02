package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"metrics-study-project-go/internal/handler"
	"metrics-study-project-go/internal/middleware"
	"metrics-study-project-go/internal/repository"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var flagRunAddr string

func main() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")

	flag.Parse()

	setupServerAddress()

	loggerConfig := zap.NewProductionConfig()

	logger, err := loggerConfig.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize logger: %v\n", err)
		os.Exit(1)
	}

	defer logger.Sync()

	if err := http.ListenAndServe(flagRunAddr, newRouter(logger)); err != nil {
		logger.Info("server stopped", zap.Error(err))
	}
}

func newRouter(logger *zap.Logger) http.Handler {
	memStorage := repository.InitMemStorage()

	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(logger))

	updateJSON := func(rw http.ResponseWriter, r *http.Request) {
		handler.UpdateJSONHandler(rw, r, memStorage)
	}
	router.Post("/update", updateJSON)
	router.Post("/update/", updateJSON)

	valueJSON := func(rw http.ResponseWriter, r *http.Request) {
		handler.ValueJSONHandler(rw, r, memStorage)
	}
	router.Post("/value", valueJSON)
	router.Post("/value/", valueJSON)

	router.Post("/update/{metricType}/{metricName}/{metricValue}", func(rw http.ResponseWriter, r *http.Request) {
		handler.MetricHandler(rw, r, memStorage)
	})

	router.Get("/value/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
		handler.ValueHandler(rw, r, memStorage)
	})

	router.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		handler.HTMLHandler(rw, r, memStorage)
	})

	return router
}

func setupServerAddress() {
	serverAddress, ok := os.LookupEnv("ADDRESS")

	if ok {
		flagRunAddr = serverAddress
	}
}
