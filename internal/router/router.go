package router

import (
	"github.com/connorpalermo/url-shortener/internal/endpoint"
	"github.com/go-chi/chi"
)

func New(h endpoint.Provider) *chi.Mux {
	m := chi.NewRouter()

	m.Get(endpoint.HealthCheckEndpoint, h.HealthCheckHandler())
	m.Get(endpoint.RedirectEndpoint, h.RedirectHandler())
	m.Post(endpoint.ShortenURLEndpoint, h.ShortenHandler())
	m.Get("/", h.RedirectHandler())

	return m
}
