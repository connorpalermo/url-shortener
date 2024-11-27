package urlshortener

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_ShortenURL(t *testing.T) {
	tests := map[string]struct {
		orignalURLs []string
		shortURLs   []string
		count       int64
		checkCount  bool
	}{
		"URL shorten valid": {
			orignalURLs: []string{"http://www.example.com"},
			shortURLs:   []string{"b"},
		},
		"URL shorten multiple same URL counter does not increase": {
			orignalURLs: []string{"http://www.example.com", "http://www.example.com", "http://www.example.com", "http://www.example.com", "http://www.example.com"},
			shortURLs:   []string{"b", "b", "b", "b", "b"},
			checkCount:  true,
			count:       int64(2),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, err := zap.NewProduction()
			if err != nil {
				return
			}
			u := New(logger)

			for i, orig := range tc.orignalURLs {
				shortened := u.ShortenURL(orig)

				assert.Equal(t, shortened, tc.shortURLs[i])
			}

			if tc.checkCount {
				assert.Equal(t, tc.count, u.Counter)
			}
		})
	}
}

func Test_GetOriginalURL(t *testing.T) {
	tests := map[string]struct {
		orignalURLs []string
		shortURLs   []string
	}{
		"URL shorten valid": {
			orignalURLs: []string{"http://www.example.com"},
			shortURLs:   []string{"b"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, err := zap.NewProduction()
			if err != nil {
				return
			}
			u := New(logger)

			for _, orig := range tc.orignalURLs {
				shortened := u.ShortenURL(orig)

				original, _ := u.GetOriginalURL(shortened)
				assert.Equal(t, original, orig)
			}
		})
	}
}
