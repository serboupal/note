package main

import (
	"flag"
)

func search(args []string) {
	fl := flag.NewFlagSet("search", flag.ContinueOnError)
	fl.Usage = func() { usage(fl, nil) }
	fl.Parse(args)

	if fl.NArg() == 0 {
		fl.Usage()
	}

	notes, err := backend.Search(fl.Arg(0))
	if err != nil {
		errExit(err.Error())
	}
	printList(notes)
}
