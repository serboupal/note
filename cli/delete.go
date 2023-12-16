package main

import (
	"flag"
)

func delete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.Usage = func() { usage(fs, nil) }
	fs.Parse(args)

	if fs.NArg() == 0 {
		fs.Usage()
	}

	note, err := backend.Get(fs.Arg(0))
	if err != nil {
		errExit(err.Error())
	}

	err = backend.Delete(&note)
	if err != nil {
		errExit(err.Error())
	}
}
