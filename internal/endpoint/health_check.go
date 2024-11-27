package endpoint

import (
	"encoding/json"
	"net/http"

	"github.com/connorpalermo/url-shortener/constant/logkey"
	"go.uber.org/zap"
)

const (
	HealthCheckEndpoint = "/health"
	ApiVersion          = "v1"
	Environment         = "local" // TODO: update this to pull AWS environment
	ErrorResp           = "error creating health check response"
)

func (h *Handler) HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Logger.Info("retrieving details for health check")
		body := &HealthCheck{
			Version: ApiVersion,
			Region:  Environment,
		}
		h.Logger.Info("successfully created health check response")
		b, _ := json.Marshal(body)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(b)
		if err != nil {
			h.Logger.Error("failed to write health check response", zap.String(logkey.Error, err.Error()))
			http.Error(w, ErrorResp, http.StatusInternalServerError)
		}
	}
}

type HealthCheck struct {
	Region  string `json:"region"`
	Version string `json:"version"`
}
