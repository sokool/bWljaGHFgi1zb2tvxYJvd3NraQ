package fetcher

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sokool/wpf/internal/platform/log"
)

func NewManager(shutdown context.Context, options ...Option) *Manager {
	f := &Manager{
		logger:   log.NewLogger(os.Stdout),
		urls:     Memory,
		readers:  map[string]*Reader{},
		shutdown: shutdown,
	}

	for i := range options {
		options[i](f)
	}

	return f
}

type Manager struct {
	mu       sync.Mutex
	urls     Storage
	logger   log.Logger
	readers  map[string]*Reader
	shutdown context.Context
}

func (m *Manager) Fetch(u *URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.urls.Store(u); err != nil {
		return err
	}

	return m.fetch(*u)
}

func (m *Manager) Remove(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var u URL
	var err error

	if u, err = m.urls.One(id); err != nil {
		return err
	}

	if err = m.stop(u); err != nil {
		return err
	}

	return m.urls.Remove(u.ID)
}

func (m *Manager) Pause(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, err := m.urls.One(id)
	if err != nil {
		return err
	}

	return m.stop(u)
}

func (m *Manager) Resume(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, err := m.urls.One(id)
	if err != nil {
		return err
	}

	return m.fetch(u)
}

func (m *Manager) One(id int) (URL, error) { return m.urls.One(id) }

func (m *Manager) All() ([]URL, error) { return m.urls.All() }

func (m *Manager) fetch(u URL) error {
	if _, ok := m.readers[u.Resource]; ok {
		return fmt.Errorf("%s already fetching", u.Resource)
	}

	m.readers[u.Resource] = NewReader(u.Resource,
		ReaderShutdown(m.shutdown),
		ReaderInterval(u.Interval),
		ReaderTimeout(time.Second*5),
		ReaderLogger(m.logger.Tag(u.String())),
		ReaderStorage(func(r Response) error { return m.urls.Append(u.ID, r) }),
	)

	return nil
}

func (m *Manager) stop(u URL) error {
	if r, reading := m.readers[u.Resource]; reading {
		if err := r.Close(); err != nil {
			return err
		}
	}

	delete(m.readers, u.Resource)

	return nil
}

type Option func(*Manager)

func WithLogger(l log.Logger) Option {
	return func(f *Manager) { f.logger = l }
}

func WithStorage(s Storage) Option {
	return func(f *Manager) { f.urls = s }
}
