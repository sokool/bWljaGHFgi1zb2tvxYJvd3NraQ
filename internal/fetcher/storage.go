package fetcher

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
)

var Memory Storage = &memory{urls: map[int]*URL{}, unique: map[string]int{}}

type Storage interface {
	Get(id int) (URL, error)
	Store(u *URL) error
	Append(id int, r Response) error
	Remove(id int) error
}

type URL struct {
	ID       int
	Resource string
	Interval time.Duration
	History  []Response
}

func (u URL) String() string { return fmt.Sprintf("#%d:%s", u.ID, u.Resource) }

type Response struct {
	Body      []byte
	Duration  time.Duration
	CreatedAt time.Time
}

func (r Response) MarshalJSON() ([]byte, error) {
	var jsn struct {
		Data      *string `json:"response"`
		Seconds   float32 `json:"duration"`
		CreatedAt int64   `json:"created_at"`
	}

	if s := string(r.Body); s != "" {
		jsn.Data = &s
	}

	jsn.Seconds, jsn.CreatedAt = float32(r.Duration)/float32(time.Second), r.CreatedAt.Unix()

	return json.Marshal(jsn)
}

var ErrNotFound = fmt.Errorf("fetcher not found")

type memory struct {
	mu     sync.RWMutex
	urls   map[int]*URL
	unique map[string]int
	id     int
}

func (m *memory) Get(id int) (URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.urls[id]
	if !ok {
		return URL{}, ErrNotFound
	}

	return *u, nil
}

func (m *memory) Store(u *URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := url.Parse(u.Resource); err != nil {
		return fmt.Errorf("wrong resource %s", u.Resource)
	}

	if (u.Interval / time.Second) < 1 {
		return fmt.Errorf("interval must be greater than 1 second")
	}

	if _, not := m.unique[u.Resource]; not {
		return fmt.Errorf("url %s is not unique", u.Resource)
	}

	m.unique[u.Resource] = u.ID

	m.id++
	u.ID = m.id
	m.urls[m.id] = u

	return nil
}

func (m *memory) Append(id int, r Response) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.urls[id]
	if !ok {
		return ErrNotFound
	}

	u.History = append(u.History, r)

	return nil
}

func (m *memory) Remove(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.urls[id]
	if !ok {
		return ErrNotFound
	}

	delete(m.unique, u.Resource)
	delete(m.urls, id)

	return nil
}
