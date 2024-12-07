package endpoint

import (
	"net/http"

	"github.com/connorpalermo/url-shortener/internal/urlshortener"
	"go.uber.org/zap"
)

type (
	Provider interface {
		HealthCheckHandler() http.HandlerFunc
		RedirectHandler() http.HandlerFunc
	}

	Handler struct {
		Logger               *zap.Logger
		UrlShortenerProvider urlshortener.UrlShortenerProvider
	}
)
