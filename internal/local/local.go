package local

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/serboupal/note/dbline"
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

	return dir.addNoteData(n)
}

func (dir *Local) List(name string) ([]note.Note, error) {
	if note.InvalidName(name) {
		return nil, note.ErrInvalidName
	}
	var r []note.Note

	all, err := dir.loadIndex()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return all, nil
	}
	for _, v := range all {
		if strings.Contains(v.Name, name) {
			r = append(r, v)
		}
	}
	return r, nil
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
	return dir.Delete(&n)
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
	err := dbline.DeleteEntry(dir.data+"/index", n.Id)
	if err != nil {
		return err
	}

	path, err := dir.newPathFromId(n.Id)
	if err != nil {
		return err
	}

	return os.Remove(path.full)
}

func (dir *Local) loadIndex() ([]note.Note, error) {
	r, err := dbline.Open[*note.Note](dir.data + "/index")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, note.ErrNotFound
		}
		return nil, err
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
	return dbline.AppendEntry[*note.Note](dir.data+"/index", n)
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
