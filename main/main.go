package main

import (
	"net/http"

	"github.com/connorpalermo/url-shortener/internal/endpoint"
	"github.com/connorpalermo/url-shortener/internal/router"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	mux := router.New(&endpoint.Handler{
		Logger: logger,
	})

	// Start the HTTP server on port 8080
	logger.Info("Starting server on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}

}
