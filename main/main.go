package main

import (
	"net/http"

	"github.com/connorpalermo/url-shortener/internal/endpoint"
	"github.com/connorpalermo/url-shortener/internal/persistence"
	"github.com/connorpalermo/url-shortener/internal/router"
	"github.com/connorpalermo/url-shortener/internal/urlshortener"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	db, err := persistence.New(logger)
	if err != nil {
		logger.Error("failed to initialize db client")
		return
	}

	u := &urlshortener.UrlShortener{
		Logger:   logger,
		DBClient: db,
	}

	mux := router.New(&endpoint.Handler{
		Logger:               logger,
		UrlShortenerProvider: u,
	})

	logger.Info("Starting server on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}

}
