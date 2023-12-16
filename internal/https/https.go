package https

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/serboupal/note/note"
)

var _ = (note.Backend)(&https{})

type https struct {
	client *http.Client
	token  string
	url    string
}

func NewBackendHTTPS(url, token string) *https {
	return &https{url: url, token: token}
}

func (h *https) Init() error {
	h.client = &http.Client{Timeout: 5 * time.Second}
	return nil
}

func (h *https) Create(n *note.Note) error {
	return nil
}

func (h *https) Get(name string) (note.Note, error) {
	return note.Note{}, nil
}

func (h *https) Update(name string, data []byte) error {
	return nil
}

func (h *https) Delete(n *note.Note) error {
	return nil
}

func (h *https) Exist(name string) bool {
	return false
}

func (h *https) List(name string) ([]note.Note, error) {
	return nil, nil
}

func (h *https) ListAll() ([]note.Note, error) {
	req, err := h.newRequest("GET", "", nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
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
	return nil, nil
}

func (h *https) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	join, err := url.JoinPath(h.url, path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, join, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+h.token)
	return req, err
}
