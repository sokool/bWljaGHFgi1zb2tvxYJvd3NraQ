package rest

import (
	"net/http"

	"github.com/sokool/wpf/internal/fetcher"
	"github.com/sokool/wpf/internal/platform/log"
	"github.com/sokool/wpf/internal/platform/web"
)

type rest struct {
	*fetchers
	*middleware
}

func New(fm *fetcher.Manager, lp log.Printer) http.Handler {
	m := &middleware{web.NewRouter(), lp, fm}
	r := &rest{&fetchers{m}, m}

	api := r.
		Prefix("/fetchers", m.limitSize(1), m.logger, m.gzip).
		Handle(r.fetchers.new, "", "POST").
		Handle(r.fetchers.list, "", "GET")

	api.
		Prefix("/{fetcher-id}").
		Handle(r.fetchers.get, "", "GET").
		Handle(r.fetchers.remove, "", "DELETE").
		Handle(r.fetchers.history, "/history", "GET").
		Handle(r.fetchers.pause, "/pause", "PATCH").
		Handle(r.fetchers.resume, "/resume", "PATCH")

	return r
}
