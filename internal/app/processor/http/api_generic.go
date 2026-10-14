package rprocessor

import (
	"net/http"

	"github.com/gorilla/mux"
	rhandler "github.com/m1ll3r1337/catalog-service/internal/app/handler/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func vGenericRegHealthCheck(r *mux.Router, h rhandler.Health) {
	reg(r, http.MethodGet, "/health", h.LastCheck)
	reg(r, http.MethodGet, "/metrics", promhttp.Handler().ServeHTTP)
}

func handlerNotFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
