package main

import (
	"bytes"
	"flag"
	"fmt"
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
		fs.Usage()
		os.Exit(1)
	}

	if note.InvalidName(fs.Args()[0]) {
		fmt.Println("invalid name for note")
		os.Exit(1)
	}

	var bf []byte
	var err error

	if isPipe(os.Stdin) {
		bf, err = readPipe()
	} else {
		bf, err = openEditor()
	}
	if err != nil {
		panic(err)
	}

	n, err := note.NewNote(fs.Args()[0], "", bf)
	if err != nil {
		panic(err)
	}

	err = backend.Add(n)
	if err != nil {
		panic(err)
	}
}

func readPipe() ([]byte, error) {
	return io.ReadAll(os.Stdin)
}

func openEditor() ([]byte, error) {
	tmp, err := os.CreateTemp("", "*.md")
	if err != nil {
		return nil, err
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
		fmt.Println(err)
		return nil, err
	}

	bf := new(bytes.Buffer)
	_, err = io.Copy(bf, tmp)
	if err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
