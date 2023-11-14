package main

import (
	"flag"
	"fmt"
	"os"
)

func view(args []string) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(2)
	}

	n, err := backend.Get(fs.Arg(0))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", string(n.Data))
}
