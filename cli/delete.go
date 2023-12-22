package main

import (
	"errors"
	"flag"

	"github.com/serboupal/note/note"
)

func delete(args []string) {
	fl := flag.NewFlagSet("delete", flag.ContinueOnError)
	usg := "NAME"
	fl.Usage = func() { usage(fl, nil, usg) }
	fl.Parse(args)

	if fl.NArg() == 0 {
		fl.Usage()
	}

	n, err := backend.Get(fl.Arg(0))
	if err != nil {
		// delete skip integrity check
		if !errors.Is(err, note.ErrIntegrityFail) {
			errExit(err.Error())
		}
	}

	err = backend.Delete(&n)
	if err != nil {
		errExit(err.Error())
	}
}
