package https

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/serboupal/note/note"
)

var _ = (note.Backend)(&https{})
var ErrInvalidResponse = errors.New("invalid response form server")

type https struct {
	client *http.Client
	token  string
	url    string
}

func NewBackend(url, token string) *https {
	return &https{url: url, token: token}
}

func (h *https) Init() error {
	h.client = &http.Client{Timeout: 5 * time.Second}
	return nil
}

func (h *https) Create(n *note.Note) error {
	resp, err := h.newRequestDo("POST", "", n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return note.ErrNoteExist
	} else if resp.StatusCode != http.StatusOK {
		return ErrInvalidResponse
	}
	return nil
}

func (h *https) Get(name string) (n note.Note, err error) {
	resp, err := h.newRequestDo("GET", name, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return n, note.ErrNotFound
	} else if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusUnprocessableEntity {
		return n, ErrInvalidResponse
	}

	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (h *https) Update(name string, data []byte) error {
	n := note.Note{Data: data}

	resp, err := h.newRequestDo("PUT", name, n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return note.ErrNotFound
	} else if resp.StatusCode != http.StatusOK {
		return ErrInvalidResponse
	}
	return nil
}

func (h *https) Delete(n *note.Note) error {
	resp, err := h.newRequestDo("DELETE", n.Name, n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return note.ErrNotFound
	} else if resp.StatusCode != http.StatusOK {
		return ErrInvalidResponse
	}
	return nil
}

func (h *https) List(name string) ([]note.Note, error) {
	panic("implement")
	return nil, nil
}

func (h *https) ListAll() ([]note.Note, error) {
	resp, err := h.newRequestDo("GET", "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, note.ErrNotFound
	}

	notes := []note.Note{}

	err = json.NewDecoder(resp.Body).Decode(&notes)
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (h *https) Search(query string) ([]note.Note, error) {
	panic("implement")
	return nil, nil
}

func (h *https) newRequestDo(method, path string, a any) (*http.Response, error) {
	var buf bytes.Buffer
	if a != nil {
		data, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	join, err := url.JoinPath(h.url, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, join, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+h.token)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}
