package router

import (
	"github.com/connorpalermo/url-shortener/internal/endpoint"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func New(h endpoint.Provider) *chi.Mux {
	m := chi.NewRouter()

	m.Use(middleware.Logger)
	m.Use(middleware.Recoverer)

	m.Get(endpoint.HealthCheckEndpoint, h.HealthCheckHandler())
	m.Get(endpoint.RedirectEndpoint, h.RedirectHandler())
	m.Post(endpoint.ShortenURLEndpoint, h.ShortenHandler())
	m.Get("/", h.RedirectHandler())

	return m
}
