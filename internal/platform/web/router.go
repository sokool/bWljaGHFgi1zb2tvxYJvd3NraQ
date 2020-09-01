package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct{ mr *mux.Router }

func NewRouter() *Router {
	return &Router{mr: mux.NewRouter()}
}

func (r *Router) Prefix(prefix string, ms ...Middleware) *Router {
	pr := r.mr.PathPrefix(prefix).Name(prefix + "prefix").Subrouter()

	for _, m := range ms {
		pr.Use(mux.MiddlewareFunc(m))
	}

	return &Router{mr: pr}
}

func (r *Router) Handle(h http.HandlerFunc, path string, method string, ms ...Middleware) *Router {
	handler := http.Handler(h)
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}

	rh := r.mr.Handle(path, handler)
	rh.Methods(method)

	return r
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.mr.ServeHTTP(res, req)
}

type Middleware func(http.Handler) http.Handler
