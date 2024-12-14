package endpoint

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockUrlShortenerProvider struct {
	mock.Mock
}

func (m *MockUrlShortenerProvider) GetOriginalURL(shortened string) (string, error) {
	args := m.Called(shortened)
	return args.String(0), args.Error(1)
}

func (m *MockUrlShortenerProvider) ShortenURL(originalUrl string) (string, error) {
	args := m.Called(originalUrl)
	return args.String(0), args.Error(1)
}

func Test_RedirectHandler(t *testing.T) {
	logger, _ := zap.NewProduction()

	tests := []struct {
		name                string
		shortUrl            string
		getOriginalURL      string
		getOriginalURLError error
		expectedStatus      int
		expectedLocation    string
	}{
		{
			name:             "Sad Path Missing shortUrl",
			shortUrl:         "",
			expectedStatus:   http.StatusBadRequest,
			expectedLocation: "",
		},
		{
			name:                "Sad Path GetOriginalURL fails",
			shortUrl:            "b",
			getOriginalURL:      "",
			getOriginalURLError: errors.New("failed to fetch URL"),
			expectedStatus:      http.StatusInternalServerError,
			expectedLocation:    "",
		},
		{
			name:             "Happy Path Successful Redirect",
			shortUrl:         "b",
			getOriginalURL:   "http://www.example.com",
			expectedStatus:   http.StatusFound,
			expectedLocation: "http://www.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockProvider := new(MockUrlShortenerProvider)
			mockProvider.On("GetOriginalURL", tt.shortUrl).Return(tt.getOriginalURL, tt.getOriginalURLError)

			handler := &Handler{
				Logger:               logger,
				UrlShortenerProvider: mockProvider,
			}

			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortUrl, nil)
			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get(RedirectEndpoint, handler.RedirectHandler())
			r.Get("/", handler.RedirectHandler())

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusFound {
				assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"))
			}

			if tt.shortUrl != "" {
				mockProvider.AssertExpectations(t)
			}
		})
	}
}
