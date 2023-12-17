package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/serboupal/note/internal/local"
	"github.com/serboupal/note/note"
)

var ErrNotImplemented = errors.New("not implemented")

type api struct {
	backend note.Backend
}

func main() {
	api := api{}
	api.backend = local.NewBackend(".note")

	err := api.backend.Init()
	if err != nil {
		fmt.Fprint(os.Stderr, "Error initializing backend\n")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", api.tmpRouter)

	http.ListenAndServe("localhost:48374", mux)
}

// when go 1.22 releases, change this to new http.muxer
func (a *api) tmpRouter(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s, %s\n", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/" {
			a.listHandler(w, r)
			return
		}
		a.getHandler(w, r)
		return
	case http.MethodPost:
		if r.URL.Path == "/" {
			a.createHandler(w, r)
			return
		}

	case http.MethodPut:
		if r.URL.Path != "/" {
			a.updateHandler(w, r)
			return
		}
	case http.MethodDelete:
		if r.URL.Path != "/" {
			a.deleteHandler(w, r)
			return
		}
	}
	a.error(w, r, http.StatusNotImplemented, ErrNotImplemented)
}

func (a *api) listHandler(w http.ResponseWriter, r *http.Request) {
	list, err := a.backend.ListAll()
	if err != nil {
		a.error(w, r, http.StatusInternalServerError, err)
		return
	}
	a.response(w, r, list)
}

func (a *api) createHandler(w http.ResponseWriter, r *http.Request) {
	n := note.Note{}
	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		a.error(w, r, http.StatusBadRequest, err)
		return
	}
	err = a.backend.Create(&n)
	if err != nil {
		if errors.Is(err, note.ErrNoteExist) {
			a.error(w, r, http.StatusConflict, err)
			return
		}
		a.error(w, r, http.StatusBadRequest, err)
		return
	}
	a.response(w, r, nil)
}

func (a *api) getHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/")
	n, err := a.backend.Get(name)
	if err != nil {
		if errors.Is(err, note.ErrNotFound) {
			a.error(w, r, http.StatusNotFound, err)
			return
		} else if errors.Is(err, note.ErrIntegrityFail) {
			a.raw_response(w, r, http.StatusUnprocessableEntity, n)
			return
		}
		a.error(w, r, http.StatusInternalServerError, err)
		return
	}
	a.response(w, r, n)
}

func (a *api) updateHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/")
	if _, err := a.backend.Get(name); err != nil {
		a.error(w, r, http.StatusBadRequest, note.ErrInvalidName)
		return
	}

	n := note.Note{}
	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		a.error(w, r, http.StatusBadRequest, err)
		return
	}
	err = a.backend.Update(name, n.Data)
	if err != nil {
		a.error(w, r, http.StatusInternalServerError, err)
		return
	}
	a.response(w, r, nil)
}

func (a *api) deleteHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	name := strings.TrimPrefix(r.URL.Path, "/")
	n := note.Note{}

	if n, err = a.backend.Get(name); err != nil {
		if !errors.Is(err, note.ErrIntegrityFail) {
			a.error(w, r, http.StatusBadRequest, err)
			return
		}
	}

	err = a.backend.Delete(&n)
	if err != nil {
		a.error(w, r, http.StatusInternalServerError, err)
		return
	}
	a.response(w, r, nil)
}

func (a *api) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	fmt.Printf("%d, %s, %v\n", code, r.URL.Path, err)

	w.WriteHeader(code)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func (a *api) response(w http.ResponseWriter, r *http.Request, data any) {
	a.raw_response(w, r, http.StatusOK, data)
}

func (a *api) raw_response(w http.ResponseWriter, r *http.Request, code int, data any) {
	if data == nil {
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	ret, err := json.Marshal(data)
	if err != nil {
		a.error(w, r, http.StatusInternalServerError, err)
		return
	}
	w.Write(ret)
}
