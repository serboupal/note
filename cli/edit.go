package main

import (
	"flag"
)

func edit(args []string) {
	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
	fs.Usage = func() { usage(fs, nil) }
	fs.Parse(args)

	if fs.NArg() == 0 {
		fs.Usage()
	}

	name := fs.Args()[0]

	note, err := backend.Get(name)
	if err != nil {
		errExit(err.Error())
	}

	var bf []byte

	bf, err = openEditor(note.Data)
	if err != nil {
		errExit(err.Error())
	}
	if len(bf) == 0 {
		errExit("Invalid buffer")
	}

	err = backend.Update(note.Name, bf)
	if err != nil {
		errExit(err.Error())
	}
}
