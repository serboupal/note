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
var ErrInvalidQuery = errors.New("invalid query")

type api struct {
	backend note.Backend
	token   string
}

func main() {
	tkn := os.Getenv("NOTE_HTTPS_TOKEN")
	if tkn == "" {
		fmt.Fprintf(os.Stderr, "please set NOTE_HTTPS_TOKEN\n")
		os.Exit(1)
		return
	}
	api := api{
		token: tkn,
	}
	api.backend = local.NewBackend(".note")

	err := api.backend.Init()
	if err != nil {
		fmt.Fprint(os.Stderr, "Error initializing backend\n")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", api.auth(api.tmpRouter))

	http.ListenAndServe("0.0.0.0:48374", mux)
}

// when go 1.22 releases, change this to new http.muxer
func (a *api) tmpRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch r.Method {
	case http.MethodGet:
		if path == "/" {
			a.listHandler(w, r)
			return
		} else if path == "/search" {
			a.searchHandler(w, r)
			return
		}
		a.getHandler(w, r)
		return
	case http.MethodPost:
		if path == "/" {
			a.createHandler(w, r)
			return
		}
	case http.MethodPut:
		if path != "/" {
			a.updateHandler(w, r)
			return
		}
	case http.MethodDelete:
		if path != "/" {
			a.deleteHandler(w, r)
			return
		}
	case http.MethodOptions:
		a.response(w, r, nil)
	default:
		a.error(w, r, http.StatusNotImplemented, ErrNotImplemented)
	}
}

func (a *api) listHandler(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	var list []note.Note
	var err error
	if filter == "" {
		list, err = a.backend.ListAll()
	} else {
		list, err = a.backend.List(filter)
	}
	if err != nil {
		if errors.Is(err, note.ErrNotFound) {
			a.error(w, r, http.StatusNotFound, err)
		} else {
			a.error(w, r, http.StatusInternalServerError, err)
		}
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

func (a *api) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	var list []note.Note
	var err error
	if query == "" {
		a.error(w, r, http.StatusBadRequest, ErrInvalidQuery)
		return
	}
	list, err = a.backend.Search(query)
	if err != nil {
		a.error(w, r, http.StatusInternalServerError, err)
		return
	}
	a.response(w, r, list)

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
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
	w.Header().Add("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE, PATCH")
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

func (a *api) auth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		token := strings.TrimPrefix(bearer, "Bearer ")
		if a.token != token && r.Method != http.MethodOptions {
			fmt.Println("auth failed", r.RemoteAddr)
			a.error(w, r, http.StatusUnauthorized, nil)
			return
		}
		f(w, r)
	}
}
