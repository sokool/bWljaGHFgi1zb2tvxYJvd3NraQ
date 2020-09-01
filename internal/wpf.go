package wpf

import (
	"context"
	"net/http"
	"os"

	"github.com/sokool/wpf/internal/api"
	"github.com/sokool/wpf/internal/fetcher"
	"github.com/sokool/wpf/internal/platform/log"
)

type Service struct {
	Logger   log.Logger
	Fetchers *fetcher.Manager
	RestAPI  http.Handler
	Shutdown context.Context
}

func New(shutdown context.Context) *Service {
	var (
		logger   = log.NewLogger(os.Stdout)
		fetchers = fetcher.NewManager(shutdown,
			fetcher.WithLogger(logger),
			fetcher.WithStorage(fetcher.Memory))

		rest = api.Rest(fetchers, logger.Tag("HTTP"))
	)

	return &Service{
		Logger:   logger,
		Fetchers: fetchers,
		RestAPI:  rest,
		Shutdown: shutdown,
	}
}

func (s *Service) Run(addr string) error {
	s.Logger.Print("runs on %s", addr)
	hs := &http.Server{Addr: addr, Handler: s.RestAPI}

	go func() {
		if err := hs.ListenAndServe(); err != nil {
			s.Logger.Print("shutdown with %s", err)
		}
	}()

	select {
	case <-s.Shutdown.Done():
		return hs.Shutdown(s.Shutdown)
	}
}
