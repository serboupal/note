package note

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidName = errors.New("invalid name for note")
	ErrNoteExist   = errors.New("note name already exist")
)

type Note struct {
	Id     string
	Name   string
	Date   time.Time
	Tags   []string
	Groups []string
	Size   int

	Data []byte
}

func NewNote(name string, title string, data []byte) (*Note, error) {
	n := Note{
		Name: name,
		Date: time.Now(),
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

func InvalidName(name string) bool {
	if strings.ContainsAny(name, " <>:\"|?*") || strings.Contains(name, "..") {
		return true
	}
	return false
}

type Backend interface {
	Init() error
	Add(n *Note) error
	List(name string) ([]Note, error)
	ListAll() ([]Note, error)
	Get(name string) (Note, error)
	// Search(query string) []*Note
	// Remove(n *Note) error
}