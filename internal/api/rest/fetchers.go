package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sokool/wpf/internal/fetcher"
)

type fetchers struct{ *middleware }

func (a *fetchers) new(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL      string `json:"url"`
		Interval int    `json:"interval"`
	}

	if err := a.read(r, &request); err != nil {
		a.err(w, http.StatusBadRequest, nil)
		return
	}

	u := fetcher.URL{
		Resource: request.URL,
		Interval: time.Second * time.Duration(request.Interval),
	}

	if err := a.fetchers.Fetch(&u); err != nil {
		a.write(w, err)
		return
	}

	type response struct {
		ID int `json:"id"`
	}

	a.write(w, response{u.ID})
}

func (a *fetchers) get(w http.ResponseWriter, r *http.Request) {
	var u fetcher.URL
	var id int
	var err error

	if id = a.id(r); id <= 0 {
		a.err(w, http.StatusBadRequest, nil)
		return
	}

	if u, err = a.fetchers.One(id); err != nil {
		a.write(w, err)
		return
	}

	a.write(w, u)
}

func (a *fetchers) remove(w http.ResponseWriter, r *http.Request) {
	var id int

	if id = a.id(r); id <= 0 {
		a.err(w, http.StatusBadRequest, nil)
		return
	}

	if err := a.fetchers.Remove(id); err != nil {
		a.write(w, err)
		return
	}
}

func (a *fetchers) pause(w http.ResponseWriter, r *http.Request) {
	var id int

	if id = a.id(r); id <= 0 {
		a.err(w, http.StatusBadRequest, nil)
		return
	}

	if err := a.fetchers.Pause(id); err != nil {
		a.write(w, err)
		return
	}
}

func (a *fetchers) resume(w http.ResponseWriter, r *http.Request) {
	var id int

	if id = a.id(r); id <= 0 {
		a.err(w, http.StatusBadRequest, nil)
		return
	}

	if err := a.fetchers.Resume(id); err != nil {
		a.write(w, err)
		return
	}
}

func (a *fetchers) history(w http.ResponseWriter, r *http.Request) {
	var u fetcher.URL
	var id int
	var err error

	if id = a.id(r); id <= 0 {
		a.err(w, http.StatusBadRequest, nil)
		return
	}

	if u, err = a.fetchers.One(id); err != nil {
		a.write(w, err)
		return
	}

	a.write(w, u.History)
}

func (a *fetchers) list(w http.ResponseWriter, r *http.Request) {
	u, err := a.fetchers.All()
	if err != nil {
		a.write(w, err)
		return
	}

	a.write(w, u)
}

func (a *fetchers) id(r *http.Request) int {
	id, err := strconv.Atoi(a.param(r, "fetcher-id"))
	if err != nil {
		return -1
	}

	return id
}
