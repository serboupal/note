package main

import (
	"flag"
)

func edit(args []string) {
	fl := flag.NewFlagSet("edit", flag.ContinueOnError)
	fl.Usage = func() { usage(fl, nil) }
	fl.Parse(args)

	if fl.NArg() == 0 {
		fl.Usage()
	}

	name := fl.Args()[0]

	note, err := backend.Get(name)
	if err != nil {
		errExit(err.Error())
	}

	var buf []byte

	buf, err = openEditor(note.Data)
	if err != nil {
		errExit(err.Error())
	}
	if len(buf) == 0 {
		errExit("Invalid buffer")
	}

	err = backend.Update(note.Name, buf)
	if err != nil {
		errExit(err.Error())
	}
}
