package local

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/serboupal/note/note"
)

type Local struct {
	config string
	data   string
	tags   string
	groups string

	dir string
}

// NewLocal returs a Local backend that uses the dir as the folder name to save
// data.
func NewLocal(dir string) *Local {
	return &Local{dir: dir}
}

var _ = (note.Backend)(&Local{})

func (l *Local) Init() error {
	return l.mkDirs()
}

func (dir *Local) Add(n *note.Note) error {
	prefix := n.Id[:2]
	suffix := n.Id[2:]

	err := os.MkdirAll(filepath.Join(dir.data, prefix), 0744)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(dir.data, prefix, suffix))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(n.Data)
	if err != nil {
		return err
	}

	err = dir.addNoteData(n)
	if err != nil {
		return err
	}
	return nil
}

func (dir *Local) List(name string) ([]note.Note, error) {
	if note.InvalidName(name) {
		return nil, note.ErrInvalidName
	}
	var r []note.Note

	all, err := dir.ListAll()
	if err != nil {
		return nil, err
	}

	for _, v := range all {
		if strings.Contains(v.Name, name) {
			r = append(r, v)
		}
	}
	return r, nil
}

func (dir *Local) ListAll() ([]note.Note, error) {
	var r []note.Note

	file, err := os.Open(dir.data + "/index")
	if err != nil && os.IsNotExist(err) {
		return r, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	fs := bufio.NewScanner(file)
	for fs.Scan() {
		data := strings.Split(fs.Text(), ",")
		ti, err := time.Parse(time.Layout, data[1])
		if err != nil {
			return nil, err
		}
		n := note.Note{
			Name: data[0],
			Date: ti,
			Id:   data[2],
		}
		r = append(r, n)
		slices.Reverse(r)
	}
	return r, nil
}

func (dir *Local) loadData(n *note.Note) error {
	data, err := os.ReadFile(dir.data + "/" + n.Id[:2] + "/" + n.Id[2:])
	if err != nil {
		return err
	}
	n.Data = data
	return nil
}

func (dir *Local) Get(name string) (note.Note, error) {
	r, err := dir.List(name)
	if err != nil {
		return note.Note{}, err
	}
	for _, v := range r {
		if v.Name == name {
			err := dir.loadData(&v)
			if err != nil {
				return note.Note{}, err
			}
			return v, nil
		}
	}
	return note.Note{}, errors.New("not found")
}

func (dir *Local) addNoteData(n *note.Note) error {
	f, err := os.OpenFile(dir.data+"/index", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s,%s,%s\n", n.Name, n.Date.Format(time.Layout), n.Id)
	if _, err := f.Write([]byte(line)); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func (dir *Local) mkDirs() error {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	data, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir.config = filepath.Join(cfg, dir.dir)
	err = os.Mkdir(dir.config, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	dir.data = filepath.Join(data, "."+dir.dir)
	err = os.Mkdir(dir.data, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	dir.tags = filepath.Join(dir.data, "tags")
	err = os.Mkdir(dir.tags, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	dir.groups = filepath.Join(dir.data, "groups")
	err = os.Mkdir(dir.groups, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
