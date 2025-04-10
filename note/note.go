package note

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidName   = errors.New("invalid name for note")
	ErrIntegrityFail = errors.New("note integrity check failed")
	ErrNoteExist     = errors.New("note name already exist")
	ErrNotModified   = errors.New("note not modified")
	ErrNotFound      = errors.New("note not found")
)

type Backend interface {
	Init() error
	Create(n *Note) error
	Get(name string) (Note, error)
	Update(name string, data []byte) error
	Delete(n *Note) error
	List(name string) ([]Note, error)
	Search(query string) ([]Note, error)
}

type Note struct {
	Id     string     `json:"id,omitempty"`
	Name   string     `json:"name,omitempty"`
	Date   *time.Time `json:"date,omitempty"`
	Tags   []string   `json:"tags,omitempty"`
	Groups []string   `json:"groups,omitempty"`
	Size   int        `json:"size,omitempty"`
	Data   []byte     `json:"data,omitempty"`
}

func NewNote(name string, title string, data []byte) (*Note, error) {
	ti := time.Now()
	n := Note{
		Name: name,
		Date: &ti,
		Data: data,
		Size: len(data),
	}
	if InvalidName(name) {
		return nil, ErrInvalidName
	}
	hash := sha256.Sum256(data)
	n.Id = fmt.Sprintf("%x", hash)
	return &n, nil
}

func (n *Note) String() string {
	return fmt.Sprintf("%s,%s,%s", n.Id, n.Date.Format(time.DateTime), n.Name)
}

func (n *Note) Parse(s string) error {
	item := strings.Split(s, ",")
	if len(item) != 3 {
		return fmt.Errorf("invalid note string")
	}

	ti, err := time.Parse(time.DateTime, item[1])
	if err != nil {
		return err
	}
	n.Id = item[0]
	n.Name = item[2]
	n.Date = &ti
	return nil
}

func (n *Note) Check() error {
	hash := sha256.Sum256(n.Data)
	Id := fmt.Sprintf("%x", hash)
	if n.Id != Id {
		return ErrIntegrityFail
	}
	if InvalidName(n.Name) {
		return ErrInvalidName
	}
	return nil
}

func InvalidName(name string) bool {
	if strings.ContainsAny(name, " <>:\"|?*") || strings.Contains(name, "..") {
		return true
	}
	return false
}
