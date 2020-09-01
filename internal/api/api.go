package api

import (
	"net/http"

	"github.com/sokool/wpf/internal/api/rest"
	"github.com/sokool/wpf/internal/fetcher"
	"github.com/sokool/wpf/internal/platform/log"
)

func Rest(fs *fetcher.Manager, lp log.Printer) http.Handler { return rest.New(fs, lp) }
