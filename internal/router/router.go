package router

import (
	"github.com/connorpalermo/url-shortener/internal/endpoint"
	"github.com/go-chi/chi/v5"
)

func New(h endpoint.Provider) *chi.Mux {
	m := chi.NewRouter()

	m.Get(endpoint.HealthCheckEndpoint, h.HealthCheckHandler())
	m.Get(endpoint.RedirectEndpoint, h.RedirectHandler())

	return m
}
