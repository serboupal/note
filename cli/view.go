package main

import (
	"flag"
	"fmt"
)

func view(args []string) {
	fs := flag.NewFlagSet("view", flag.ContinueOnError)
	fs.Usage = func() { usage(fs, nil) }
	fs.Parse(args)

	if fs.NArg() == 0 {
		fs.Usage()
	}

	n, err := backend.Get(fs.Arg(0))
	if err != nil {
		errExit(err.Error())
	}
	fmt.Printf("%s", string(n.Data))
}
