package endpoint

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func Test_ShortenHandler_Success(t *testing.T) {
	mockLogger := zaptest.NewLogger(t)
	mockUrlShortenerProvider := new(MockUrlShortenerProvider)

	mockUrlShortenerProvider.On("ShortenURL", "http://example.com").Return("short.ly/123", nil)

	handler := Handler{
		Logger:               mockLogger,
		UrlShortenerProvider: mockUrlShortenerProvider,
	}

	body := `{"original_url": "http://example.com"}`
	request := httptest.NewRequest("POST", ShortenURLEndpoint, io.NopCloser(strings.NewReader(body)))
	rr := httptest.NewRecorder()

	handler.ShortenHandler().ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	var response ShortenResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "short.ly/123", response.ShortenURL)

	mockUrlShortenerProvider.AssertExpectations(t)
}

func Test_ShortenHandler_InvalidBody(t *testing.T) {
	mockLogger := zaptest.NewLogger(t)
	mockUrlShortenerProvider := new(MockUrlShortenerProvider)

	handler := Handler{
		Logger:               mockLogger,
		UrlShortenerProvider: mockUrlShortenerProvider,
	}

	body := `{"some_key": "some_value"}`
	request := httptest.NewRequest("POST", ShortenURLEndpoint, io.NopCloser(strings.NewReader(body)))
	rr := httptest.NewRecorder()

	handler.ShortenHandler().ServeHTTP(rr, request)

	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)

	mockUrlShortenerProvider.AssertNotCalled(t, "ShortenURL")
}

func Test_ShortenHandler_ErrorInShortenURL(t *testing.T) {
	mockLogger := zaptest.NewLogger(t)
	mockUrlShortenerProvider := new(MockUrlShortenerProvider)

	mockUrlShortenerProvider.On("ShortenURL", "http://example.com").Return("", errors.New("some error"))

	handler := Handler{
		Logger:               mockLogger,
		UrlShortenerProvider: mockUrlShortenerProvider,
	}

	body := `{"original_url": "http://example.com"}`
	request := httptest.NewRequest("POST", ShortenURLEndpoint, io.NopCloser(strings.NewReader(body)))
	rr := httptest.NewRecorder()

	handler.ShortenHandler().ServeHTTP(rr, request)

	assert.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)

	mockUrlShortenerProvider.AssertExpectations(t)
}

func Test_ShortenHandler_ErrorWritingResponse(t *testing.T) {
	mockLogger := zaptest.NewLogger(t)
	mockUrlShortenerProvider := new(MockUrlShortenerProvider)

	mockUrlShortenerProvider.On("ShortenURL", "http://example.com").Return("short.ly/123", nil)

	handler := Handler{
		Logger:               mockLogger,
		UrlShortenerProvider: mockUrlShortenerProvider,
	}

	body := `{"original_url": "http://example.com"}`
	request := httptest.NewRequest("POST", ShortenURLEndpoint, io.NopCloser(strings.NewReader(body)))

	rr := &errorResponseWriter{ResponseWriter: httptest.NewRecorder()}

	handler.ShortenHandler().ServeHTTP(rr, request)

	assert.Equal(t, http.StatusInternalServerError, rr.ResponseWriter.(*httptest.ResponseRecorder).Result().StatusCode)

	mockUrlShortenerProvider.AssertExpectations(t)
}
