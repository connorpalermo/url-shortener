package endpoint

import (
	"encoding/json"
	"net/http"

	"github.com/connorpalermo/url-shortener/constant/logkey"
	"go.uber.org/zap"
)

const (
	ShortenURLEndpoint = "/shorten"
	InvalidBodyError   = "invalid URL shorten request"
	ShortenURLError    = "failed to generate shortenedURL"
)

func (h *Handler) ShortenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body ShortenRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil || body.OriginalURL == "" {
			http.Error(w, InvalidBodyError, http.StatusBadRequest)
			return
		}
		originalURL := body.OriginalURL

		h.Logger.Info("creating shortenedURL from originalURL", zap.String(logkey.OriginalURL, originalURL))
		shortenedURL, err := h.UrlShortenerProvider.ShortenURL(originalURL)
		if err != nil {
			h.Logger.Error("failed to create shortened URL", zap.Error(err))
			http.Error(w, ShortenURLError, http.StatusInternalServerError)
			return
		}
		shortenResponse := &ShortenResponse{
			ShortenURL: shortenedURL,
		}
		b, _ := json.Marshal(shortenResponse)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(b)
		if err != nil {
			h.Logger.Error("failed to write shortened URL response", zap.String(logkey.Error, err.Error()))
			http.Error(w, ShortenURLError, http.StatusInternalServerError)
		}
	}
}

type ShortenRequest struct {
	OriginalURL string `json:"original_url"`
}

type ShortenResponse struct {
	ShortenURL string `json:"shortened_url"`
}
