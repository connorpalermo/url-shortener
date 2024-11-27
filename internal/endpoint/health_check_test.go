package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	local   = "local"
	version = "v1"
)

type errorResponseWriter struct {
	http.ResponseWriter
}

func (e *errorResponseWriter) Write(b []byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func Test_HealthCheckHandelr(t *testing.T) {
	wantResp := HealthCheck{
		Region:  local,
		Version: version,
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	h := Handler{
		Logger: logger,
	}

	rr := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "http://example.com", nil)
	h.HealthCheckHandler().ServeHTTP(rr, request)

	assert.EqualValues(t, http.StatusOK, rr.Result().StatusCode)

	gotResp := HealthCheck{}

	bytes, _ := io.ReadAll(rr.Body)
	_ = json.Unmarshal(bytes, &gotResp)

	assert.EqualValues(t, wantResp, gotResp)
}

func Test_HealthCheckHandler_ErrorCase(t *testing.T) {
	mockLogger := zaptest.NewLogger(t)
	h := Handler{Logger: mockLogger}

	rr := &errorResponseWriter{httptest.NewRecorder()}
	request := httptest.NewRequest("GET", "http://example.com", nil)
	h.HealthCheckHandler().ServeHTTP(rr, request)

	assert.NotNil(t, rr)

	assert.Equal(t, 0, rr.ResponseWriter.(*httptest.ResponseRecorder).Body.Len())
}
