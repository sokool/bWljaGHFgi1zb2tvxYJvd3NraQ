package fetcher

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/sokool/wpf/internal/platform/log"
)

type Reader struct {
	url      string
	log      log.Printer
	client   *http.Client
	clock    *time.Ticker
	shutdown context.Context
	store    func(Response) error
	cancel   func()
}

func NewReader(url string, options ...ReaderOption) *Reader {
	r := &Reader{
		url:      url,
		client:   &http.Client{Timeout: time.Second * 5},
		clock:    time.NewTicker(time.Hour),
		log:      log.NewLogger(os.Stdout).Tag(url),
		shutdown: context.Background(),
		store:    func(Response) error { return nil },
	}

	for i := range options {
		options[i](r)
	}

	r.shutdown, r.cancel = context.WithCancel(r.shutdown)

	go r.run()

	return r
}

func (r *Reader) Close() error { r.cancel(); return nil }

func (r *Reader) run() {
	r.log("reading...")
	defer r.log("done")

	for {
		var result Response
		var err error

		select {
		case <-r.clock.C:
			if result, err = r.read(r.url); err != nil {
				r.log("fetch failed due %s", err)
			}

			if size := len(result.Body); size > 0 {
				r.log("DBG %.2fKB fetched in %s", float32(size)/1024, result.Duration)
			}

			if err = r.store(result); err != nil {
				r.log("store failed due %s", err)
			}

		case <-r.shutdown.Done():
			r.clock.Stop()
			return
		}
	}
}

func (r *Reader) read(url string) (fr Response, err error) {
	defer func() { fr.Duration = time.Since(fr.CreatedAt) }()
	var hr *http.Response

	fr.CreatedAt = time.Now()

	if hr, err = r.client.Get(url); err != nil {
		return fr, err
	}

	if fr.Body, err = ioutil.ReadAll(hr.Body); err != nil {
		return fr, err
	}

	return fr, nil

}

type ReaderOption func(*Reader)

func ReaderLogger(p log.Printer) ReaderOption {
	return func(o *Reader) { o.log = p }
}

func ReaderInterval(d time.Duration) ReaderOption {
	return func(o *Reader) { o.clock = time.NewTicker(d) }
}

func ReaderTimeout(d time.Duration) ReaderOption {
	return func(o *Reader) { o.client = &http.Client{Timeout: d} }
}

func ReaderShutdown(c context.Context) ReaderOption {
	return func(o *Reader) { o.shutdown = c }
}

func ReaderStorage(s func(Response) error) ReaderOption {
	return func(o *Reader) { o.store = s }
}
