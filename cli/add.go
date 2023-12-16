package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"

	"github.com/serboupal/note/note"
)

func add(args []string) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	edit := fs.Bool("edit", false, "open editor to modify before adding note")

	fs.Usage = func() {
		usage(fs, nil)
	}
	fs.Parse(args)

	name := fs.Arg(0)

	if name == "" {
		fs.Usage()
	}

	if note.InvalidName(name) {
		errExit("invalid name for note")
	}

	if backend.Exist(name) {
		errExit(note.ErrNoteExist.Error())
	}

	var bf []byte
	var err error
	filename := fs.Arg(1)

	if isPipe(os.Stdin) {
		if *edit {
			errExit("you can't use --edit on a pipe")
		}
		bf, err = readPipe()
	} else if filename != "" {
		bf, err = os.ReadFile(filename)
		if err != nil {
			errExit(err.Error())
		}
	} else {
		*edit = true
	}

	if *edit {
		bf, err = openEditor(bf)
	}

	if err != nil {
		errExit(err.Error())
	}

	n, err := note.NewNote(name, "", bf)
	if err != nil {
		errExit(err.Error())
	}

	err = backend.Create(n)
	if err != nil {
		errExit(err.Error())
	}
}

func readPipe() ([]byte, error) {
	return io.ReadAll(os.Stdin)
}

func openEditor(data []byte) ([]byte, error) {
	tmp, err := os.CreateTemp("", "*.md")
	if err != nil {
		return nil, err
	}

	if data != nil {
		_, err := tmp.Write(data)
		if err != nil {
			return nil, err
		}
		tmp.Sync()
	}
	defer func() {
		tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	cmd := exec.Command("vim", tmp.Name())

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	fi, err := tmp.Stat()
	if err != nil {
		return nil, err
	}

	if fi.Size() == 0 {
		err = ErrFileEmpty
		return nil, err
	}

	bf := new(bytes.Buffer)

	_, err = tmp.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(bf, tmp)
	if err != nil {
		return nil, err
	}

	return bf.Bytes(), nil
}
