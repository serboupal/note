package main

import (
	"flag"

	"github.com/serboupal/note/rest"
)

func serve(args []string) {
	fl := flag.NewFlagSet("serve", flag.ContinueOnError)
	usg := ""
	fl.Usage = func() { usage(fl, nil, usg) }
	fl.Parse(args)

	rest.Serve()
}
