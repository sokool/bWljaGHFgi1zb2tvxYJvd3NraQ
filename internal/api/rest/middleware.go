package rest

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sokool/wpf/internal/fetcher"
	"github.com/sokool/wpf/internal/platform/log"
	"github.com/sokool/wpf/internal/platform/web"
)

type middleware struct {
	*web.Router
	log      log.Printer
	fetchers *fetcher.Manager
}

func (m *middleware) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := time.Now()
		next.ServeHTTP(w, r)

		var s runtime.MemStats
		runtime.ReadMemStats(&s)

		var ha = float32(s.HeapAlloc) / 1024 / 1024
		var hi = float32(s.HeapInuse) / 1024 / 1024

		m.log("DBG %s %s in %s, memory allocated: %.2f MB, in use: %.2f MB, goroutines: %d",
			r.URL, r.Method, time.Since(n), ha, hi, runtime.NumGoroutine())
	})

}

func (m *middleware) limitSize(mb int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 1024*mb)
			next.ServeHTTP(w, r)
		})
	}
}

func (m *middleware) gzip(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		h.ServeHTTP(&gzipw{gz, w}, r)
	})
}

func (m *middleware) read(src *http.Request, request interface{}) error {
	return json.NewDecoder(src.Body).Decode(&request)
}

func (m *middleware) write(to http.ResponseWriter, response interface{}) {
	if err, ok := response.(error); ok && err != nil {

		switch {
		case err == fetcher.ErrNotFound:
			m.err(to, http.StatusNotFound, nil)
			return

		default:
			m.err(to, http.StatusConflict, err)
		}

		return
	}

	if response == nil {
		to.WriteHeader(http.StatusOK)
		return
	}

	to.Header().Set("content-type", "application/json")
	to.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(to).Encode(response); err != nil {
		to.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (m *middleware) err(to http.ResponseWriter, status int, err error) {
	if err != nil {
		http.Error(to, err.Error(), status)
		return
	}

	to.WriteHeader(status)
}

func (m *middleware) param(from *http.Request, name string) string {
	return mux.Vars(from)[name]
}

type gzipw struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipw) Write(b []byte) (int, error) { return w.Writer.Write(b) }
