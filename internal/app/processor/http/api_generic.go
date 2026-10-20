package rprocessor

import (
	"net/http"
	"net/http/pprof"

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

func vGenericRegPprof(r *mux.Router) {
	reg(r, http.MethodGet, "/debug/pprof", pprofIndex)
	reg(r, http.MethodGet, "/debug/pprof/", pprof.Index)

	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	r.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	r.Handle("/debug/pprof/block", pprof.Handler("block"))
	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
}

func pprofIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
}
