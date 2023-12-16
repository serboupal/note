package main

import (
	"flag"
)

func search(args []string) {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.Usage = func() { usage(fs, nil) }
	fs.Parse(args)

	if fs.NArg() == 0 {
		fs.Usage()
	}

	r, err := backend.Search(fs.Arg(0))
	if err != nil {
		errExit(err.Error())
	}
	printList(r)
}
