package main

import (
	"flag"
	"os"
)

func search(args []string) {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(2)
	}

	r, err := backend.Search(fs.Arg(0))
	if err != nil {
		errExit(err.Error())
	}
	printList(r)
}
