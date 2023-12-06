package main

import (
	"flag"
	"fmt"
	"os"
)

func view(args []string) {
	fs := flag.NewFlagSet("view", flag.ContinueOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(2)
	}

	n, err := backend.Get(fs.Arg(0))
	if err != nil {
		errExit(err.Error())
	}
	fmt.Printf("%s", string(n.Data))
}
