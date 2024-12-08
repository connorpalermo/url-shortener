package endpoint

import (
	"fmt"
	"net/http"

	"github.com/connorpalermo/url-shortener/constant/logkey"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

const (
	RedirectEndpoint   = "/{" + ShortUrlParam + "}"
	ShortUrlParam      = "shortUrl"
	RedirectError      = "shortUrl mapping invalid or not found in database"
	ShortUrlParamError = "shortUrl parameter is missing"
)

func (h *Handler) RedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortUrl := chi.URLParam(r, ShortUrlParam)
		fmt.Println("Short URL is: " + shortUrl)
		if shortUrl == "" {
			h.Logger.Error("shortUrl parameter is missing in the request")
			http.Error(w, ShortUrlParamError, http.StatusBadRequest)
			return
		}

		h.Logger.Info("redirecting from shortUrl", zap.String(logkey.ShortenedURL, shortUrl))

		originalURL, err := h.UrlShortenerProvider.GetOriginalURL(shortUrl)
		if err != nil {
			h.Logger.Error("failed to retrieve original URL", zap.Error(err))
			http.Error(w, RedirectError, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}
