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
	fl := flag.NewFlagSet("add", flag.ContinueOnError)
	edit := fl.Bool("edit", false, "open editor to modify before adding note")
	usg := "[options] NAME"

	fl.Usage = func() {
		usage(fl, nil, usg)
	}
	fl.Parse(args)

	name := fl.Arg(0)

	if name == "" {
		fl.Usage()
	}

	if note.InvalidName(name) {
		errExit("invalid name for note")
	}

	if _, err := backend.Get(name); err == nil {
		errExit(note.ErrNoteExist.Error())
	}

	var buf []byte
	var err error
	filename := fl.Arg(1)

	if isPipe(os.Stdin) {
		if *edit {
			errExit("you can't use --edit on a pipe")
		}
		buf, err = readPipe()
	} else if filename != "" {
		buf, err = os.ReadFile(filename)
		if err != nil {
			errExit(err.Error())
		}
	} else {
		*edit = true
	}

	if *edit {
		buf, err = openEditor(buf)
	}

	if err != nil {
		errExit(err.Error())
	}

	n, err := note.NewNote(name, "", buf)
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

	buf := new(bytes.Buffer)

	_, err = tmp.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(buf, tmp)
	if err != nil {
		return nil, err
	}

	bufB := buf.Bytes()
	if bytes.Compare(data, bufB) == 0 {
		return bufB, note.ErrNotModified
	}
	return bufB, nil
}
