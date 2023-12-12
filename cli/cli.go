package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/serboupal/note/internal/local"
	"github.com/serboupal/note/note"
)

const appFolder = "note"

var backend note.Backend

type cmd struct {
	fn   func([]string)
	desc string
}

var commands = map[string]cmd{
	"add":    {fn: add, desc: "add note"},
	"list":   {fn: list, desc: "list notes"},
	"view":   {fn: view, desc: "view note content"},
	"search": {fn: search, desc: "search in note content"},
	"delete": {fn: delete, desc: "delete note"},
	"edit":   {fn: edit, desc: "edit note"},
}

var ErrFileEmpty = errors.New("file is empty")

func main() {
	backend = local.NewLocal(appFolder)
	err := backend.Init()
	if err != nil {
		panic(err)
	}

	debug := flag.Bool("debug", false, "print debug information")
	flag.Parse()

	subcommand := flag.Args()
	if len(subcommand) == 0 {
		usage("", flag.CommandLine, commands)
		os.Exit(1)
	}
	cmd, ok := commands[subcommand[0]]
	if !ok {
		usage("", flag.CommandLine, commands)
		os.Exit(1)
	}

	if *debug {
		log.Println("Debug info on")
	}
	cmd.fn(subcommand[1:])
}

func usage(cmdName string, fs *flag.FlagSet, c map[string]cmd) {
	fmt.Printf("Usage:\n  %s %s [options] [file]\n\n", os.Args[0], cmdName)
	if c != nil {
		fmt.Printf("Commands:\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 0, 2, ' ', 0)
		for k, v := range c {
			fmt.Fprintf(w, "  %s\t%s\n", k, v.desc)
		}
		w.Flush()
		fmt.Println()
	}
	fmt.Printf("Options:\n")
	fs.PrintDefaults()
}

func isPipe(p *os.File) bool {
	sin, _ := p.Stat()
	if (sin.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

func errExit(s string) {
	fmt.Println(s)
	os.Exit(1)
}
