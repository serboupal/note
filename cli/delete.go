package main

import (
	"flag"
	"os"
)

func delete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(2)
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
