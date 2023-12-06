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
	//title := fs.String("title", "", "title of the note")
	fs.Parse(args)

	if fs.NArg() == 0 {
		//	fs.Usage()
		os.Exit(1)
	}

	name := fs.Args()[0]

	if note.InvalidName(name) {
		errExit("invalid name for note")
	}

	if backend.Exist(name) {
		errExit(note.ErrNoteExist.Error())
	}

	var bf []byte
	var err error

	if isPipe(os.Stdin) {
		bf, err = readPipe()
	} else {
		bf, err = openEditor(nil)
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
