package local

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/serboupal/note/note"
)

var (
	ErrInvalidPath = errors.New("invalid path")
)

type Local struct {
	config string
	data   string
	tags   string
	groups string
	dir    string
}

type path struct {
	prefix string
	name   string
	full   string
	dir    string
	id     string
}

var _ = (note.Backend)(&Local{})

func (dir *Local) newPathFromId(id string) (*path, error) {
	if len(id) != 64 {
		return nil, ErrInvalidPath
	}
	p := path{}
	p.id = id
	p.prefix = id[:2]
	p.name = id[2:]
	p.full = filepath.Join(dir.data, p.prefix, p.name)
	p.dir = filepath.Join(dir.data, p.prefix)

	return &p, nil
}

func (dir *Local) newPathFromPath(path string) (*path, error) {
	frag := strings.Split(path, "/")
	if len(frag) < 2 {
		return nil, ErrInvalidPath
	}
	id := frag[len(frag)-2] + frag[len(frag)-1]
	return dir.newPathFromId(id)
}

// NewBackend returs a Local backend that uses the dir as the folder name to save
// data.
func NewBackend(dir string) *Local {
	return &Local{dir: dir}
}

func (l *Local) Init() error {
	return l.mkDirs()
}

func (dir *Local) Create(n *note.Note) error {
	path, err := dir.newPathFromId(n.Id)
	if err != nil {
		return err
	}
	err = os.MkdirAll(path.dir, 0744)
	if err != nil {
		return err
	}

	file, err := os.Create(path.full)
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
	return dir.loadIndex()
}

func (dir *Local) Get(name string) (note.Note, error) {
	r, err := dir.List(name)
	if err != nil {
		return note.Note{}, err
	}
	for _, v := range r {
		if v.Name == name {
			err := dir.loadNoteData(&v)
			if err != nil {
				return note.Note{}, err
			}
			if err = v.Check(); err != nil {
				return v, err
			}
			return v, nil
		}
	}
	return note.Note{}, note.ErrNotFound
}

func (dir *Local) Update(name string, data []byte) error {
	newNote, err := note.NewNote(name, "", data)
	if err != nil {
		return err
	}

	n, err := dir.Get(name)
	if err != nil {
		return err
	}

	if n.Id == newNote.Id {
		return note.ErrNotModified
	}

	newNote.Tags = n.Tags
	newNote.Groups = n.Groups

	err = dir.Create(newNote)
	if err != nil {
		return err
	}

	// is ok to delete after creation because we use Id and not Name to find
	// note
	err = dir.Delete(&n)
	if err != nil {
		return err
	}

	return nil
}

func (dir *Local) Search(s string) ([]note.Note, error) {
	r := []note.Note{}
	notes, err := dir.loadIndex()
	if err != nil {
		return nil, err
	}

	for _, n := range notes {
		err := dir.loadNoteData(&n)
		if err != nil {
			return nil, err
		}
		if strings.Contains(strings.ToLower(string(n.Data)), strings.ToLower(s)) {
			r = append(r, n)
		}
	}
	return r, nil
}

func (dir *Local) Delete(n *note.Note) error {
	f, err := os.Open(dir.data + "/index")
	if err != nil {
		return err
	}
	defer f.Close()

	var bs []byte
	buf := bytes.NewBuffer(bs)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.Contains(text, n.Id) {
			_, err := buf.Write(scanner.Bytes())
			if err != nil {
				return err
			}
			_, err = buf.WriteString("\n")
			if err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	path, err := dir.newPathFromId(n.Id)
	if err != nil {
		return err
	}

	err = os.Remove(path.full)
	if err != nil {
		return err
	}
	return os.WriteFile(dir.data+"/index", buf.Bytes(), 0644)
}

func (dir *Local) loadIndex() ([]note.Note, error) {
	r := []note.Note{}
	data, err := os.ReadFile(dir.data + "/index")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			break
		}
		item := strings.Split(line, ",")
		ti, err := time.Parse(time.DateTime, item[1])
		if err != nil {
			return nil, err
		}
		n := note.Note{
			Id:   item[0],
			Name: item[2],
			Date: &ti,
		}
		r = append(r, n)
	}
	slices.Reverse(r)
	return r, nil
}

func (dir *Local) loadNodeMetadata(n *note.Note) error {
	data, err := dir.loadIndex()
	if err != nil {
		return err
	}
	for _, note := range data {
		if note.Id == n.Id {
			n.Name = note.Name
			n.Date = note.Date
			n.Groups = note.Groups
			n.Tags = note.Tags
			return nil
		}
	}
	return note.ErrNotFound
}

func (dir *Local) loadNoteData(n *note.Note) error {
	path, err := dir.newPathFromId(n.Id)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path.full)
	if err != nil {
		return err
	}
	n.Data = data
	return nil
}

func (dir *Local) addNoteData(n *note.Note) error {
	f, err := os.OpenFile(dir.data+"/index", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s,%s,%s\n", n.Id, n.Date.Format(time.DateTime), n.Name)
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

	dir.data = filepath.Join(data, dir.dir)
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
