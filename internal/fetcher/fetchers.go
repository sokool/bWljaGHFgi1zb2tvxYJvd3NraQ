package fetcher

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sokool/wpf/internal/platform/log"
)

func NewManager(options ...Option) *Fetchers {
	f := &Fetchers{
		log:      log.NewLogger(os.Stdout),
		urls:     Memory,
		readers:  map[string]*Reader{},
		shutdown: context.Background(),
	}

	for i := range options {
		options[i](f)
	}

	return f
}

type Fetchers struct {
	mu       sync.Mutex
	log      log.Logger
	urls     Storage
	readers  map[string]*Reader
	shutdown context.Context
}

func (m *Fetchers) Fetch(u *URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.urls.Store(u); err != nil {
		return err
	}

	return m.read(*u)
}

func (m *Fetchers) Get(id int) (URL, error) {
	return m.urls.Get(id)
}

func (m *Fetchers) Remove(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var u URL
	var err error

	if u, err = m.urls.Get(id); err != nil {
		return err
	}

	if err = m.stop(u); err != nil {
		return err
	}

	return m.urls.Remove(u.ID)
}

func (m *Fetchers) Pause(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, err := m.urls.Get(id)
	if err != nil {
		return err
	}

	return m.stop(u)
}

func (m *Fetchers) Resume(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, err := m.urls.Get(id)
	if err != nil {
		return err
	}

	return m.read(u)
}

func (m *Fetchers) read(u URL) error {
	if _, ok := m.readers[u.Resource]; ok {
		return fmt.Errorf("%s already fetching", u.Resource)
	}

	m.readers[u.Resource] = NewReader(u.Resource,
		ReaderShutdown(m.shutdown),
		ReaderInterval(u.Interval),
		ReaderTimeout(time.Second*5),
		ReaderLogger(m.log.Tag(u.String())),
		ReaderStorage(func(r Response) error { return m.urls.Append(u.ID, r) }),
	)

	return nil
}

func (m *Fetchers) stop(u URL) error {
	if r, reading := m.readers[u.Resource]; reading {
		if err := r.Close(); err != nil {
			return err
		}
	}

	delete(m.readers, u.Resource)

	return nil
}

type Option func(*Fetchers)

func WithLogger(l log.Logger) Option {
	return func(f *Fetchers) { f.log = l }
}

func WithStorage(s Storage) Option {
	return func(f *Fetchers) { f.urls = s }
}

func WithShutdown(c context.Context) Option {
	return func(f *Fetchers) { f.shutdown = c }
}
