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
	fl := flag.NewFlagSet("list", flag.ContinueOnError)
	usg := "[EXPRESSION]"
	fl.Usage = func() { usage(fl, nil, usg) }
	fl.Parse(args)

	if fl.NArg() > 1 {
		fl.Usage()
	}

	var data []note.Note
	var err error
	if fl.NArg() == 0 {
		data, err = backend.List("")
	} else {
		data, err = backend.List(fl.Arg(0))
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(data) == 0 {
		fmt.Println(note.ErrNotFound)
		return
	}
	printList(data)
}

func printList(notes []note.Note) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if len(notes) == 0 {
		fmt.Fprintln(w, "No notes found")
		return
	}
	fmt.Fprintf(w, "NAME\tDATE\n")
	for _, v := range notes {
		fmt.Fprintf(w, "%s\t%s\n", v.Name, v.Date.Format(time.RFC822))
	}
	w.Flush()
}
