package main

import (
	"flag"
	"fmt"
)

func view(args []string) {
	fl := flag.NewFlagSet("view", flag.ContinueOnError)
	usg := "NAME"
	fl.Usage = func() { usage(fl, nil, usg) }
	fl.Parse(args)

	if fl.NArg() == 0 {
		fl.Usage()
	}

	n, err := backend.Get(fl.Arg(0))
	if err != nil {
		errExit(err.Error())
	}
	fmt.Printf("%s", string(n.Data))
}
