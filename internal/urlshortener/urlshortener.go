package urlshortener

import (
	"sync"

	"github.com/connorpalermo/url-shortener/constant/logkey"
	"go.uber.org/zap"
)

type (
	UrlShortener struct {
		Logger             *zap.Logger
		Mu                 sync.Mutex
		Counter            int64
		ShortUrlMapping    map[string]string // TODO: use DB instead
		OriginalUrlMapping map[string]string
	}

	UrlShortenerProvider interface {
		ShortenURL(url string) string
		GetOriginalURL(shortened string) (string, bool)
	}
)

const base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func New(logger *zap.Logger) *UrlShortener {
	return &UrlShortener{
		Logger:             logger,
		Counter:            1,
		ShortUrlMapping:    make(map[string]string),
		OriginalUrlMapping: make(map[string]string),
	}
}

func (u *UrlShortener) ShortenURL(url string) string {
	u.Mu.Lock()
	defer u.Mu.Unlock()

	short, ok := u.OriginalUrlMapping[url]
	if ok {
		// we have already seen this URL
		return short
	}

	u.Logger.Info("shortening original URL: ", zap.String(logkey.OriginalURL, url))

	id := u.Counter
	u.Counter++
	shortened := encodeBase62(id)

	u.Logger.Info("generated shortened URL: ", zap.String(logkey.ShortenedURL, shortened))

	u.ShortUrlMapping[shortened] = url
	u.OriginalUrlMapping[url] = shortened
	return shortened
}

func encodeBase62(id int64) string {
	result := ""
	for id > 0 {
		rem := id % 62
		result = string(base62Chars[rem]) + result
		id /= 62
	}
	return result
}

func (u *UrlShortener) GetOriginalURL(shortened string) (string, bool) {
	u.Mu.Lock()
	defer u.Mu.Unlock()

	u.Logger.Info("getting original URL from shortened URL: ", zap.String(logkey.ShortenedURL, shortened))

	original, exists := u.ShortUrlMapping[shortened]

	u.Logger.Info("retrieved original URL: ", zap.String(logkey.OriginalURL, original))

	return original, exists
}
