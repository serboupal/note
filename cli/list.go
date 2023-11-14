package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/serboupal/note/note"
)

func list(args []string) {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.Parse(args)

	if fs.NArg() > 1 {
		fs.Usage()
		os.Exit(1)
	}

	var data []note.Note
	var err error
	if fs.NArg() == 0 {
		data, err = backend.ListAll()
	} else {
		data, err = backend.List(fs.Arg(0))
	}
	if err != nil {
		panic(err)
	}

	if len(data) == 0 {
		fmt.Println("No notes found")
		return
	}
	printList(data)
}

func printList(notes []note.Note) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "NAME\tDATE\n")
	for _, v := range notes {
		fmt.Fprintf(w, "%s\t%s\n", v.Name, v.Date.Format(time.RFC822))
	}
	w.Flush()
}
