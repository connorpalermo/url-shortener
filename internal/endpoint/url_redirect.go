package endpoint

import (
	"net/http"

	"github.com/connorpalermo/url-shortener/constant/logkey"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

const (
	RedirectEndpoint = "/{shortUrl}"
)

func (h *Handler) RedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortUrl := chi.URLParam(r, "shortUrl")
		if shortUrl == "" {
			h.Logger.Error("shortUrl parameter is missing in the request")
			http.Error(w, "shortUrl is missing", http.StatusBadRequest)
			return
		}

		h.Logger.Info("redirecting from shortUrl", zap.String(logkey.ShortenedURL, shortUrl))

		originalURL, err := h.UrlShortenerProvider.GetOriginalURL(shortUrl)
		if err != nil {
			h.Logger.Error("failed to retrieve original URL", zap.Error(err))
			http.Error(w, "failed to redirect", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}
