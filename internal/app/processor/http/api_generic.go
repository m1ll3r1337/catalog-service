package rprocessor

import (
	"net/http"

	"github.com/gorilla/mux"
	rhandler "github.com/m1ll3r1337/catalog-service/internal/app/handler/http"
)

func vGenericRegHealthCheck(r *mux.Router, h rhandler.Health) {
	reg(r, http.MethodGet, "/health", h.LastCheck)
}

func handlerNotFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
