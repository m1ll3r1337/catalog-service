package rprocessor

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/m1ll3r1337/catalog-service/internal/pkg/http/httph"
)

func reg(r *mux.Router, method, path string, handler http.HandlerFunc) {
	r.Methods(method).Path(path).Handler(handler)
}

func middlewaresToGorilla(middlewares []httph.Middleware) []mux.MiddlewareFunc {
	out := make([]mux.MiddlewareFunc, 0, len(middlewares))
	for _, mw := range middlewares {
		out = append(out, mux.MiddlewareFunc(mw))
	}
	return out
}
