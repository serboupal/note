package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/serboupal/note/internal/https"
	"github.com/serboupal/note/internal/local"
	"github.com/serboupal/note/note"
)

const appFolder = "note"

var cmdOut = os.Stderr
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
	remoteURL := os.Getenv("NOTE_HTTPS_URL")
	remoteToken := os.Getenv("NOTE_HTTPS_TOKEN")

	if remoteURL != "" {
		if remoteToken == "" {
			errExit("To use remote service, you need to provide an auth token")
		}
		backend = https.NewBackendHTTPS(remoteURL, remoteToken)
	} else {
		backend = local.NewBackendLocal(appFolder)
	}

	err := backend.Init()
	if err != nil {
		panic(err)
	}

	flag.Usage = func() {
		usage(flag.CommandLine, commands)
	}
	flag.Parse()

	subcommand := flag.Args()
	if len(subcommand) == 0 {
		flag.Usage()
	}

	cmd, ok := commands[subcommand[0]]
	if !ok {
		flag.Usage()
	}

	cmd.fn(subcommand[1:])
}

func usage(fs *flag.FlagSet, c map[string]cmd) {
	cmdName := fs.Name()
	if cmdName == os.Args[0] {
		cmdName = "command"
	}
	fmt.Fprintf(cmdOut, "Usage:\n  %s [options] %s [file]\n\n", os.Args[0], cmdName)
	if c != nil {
		fmt.Fprintf(cmdOut, "Commands:\n")
		w := tabwriter.NewWriter(cmdOut, 0, 0, 2, ' ', 0)
		for k, v := range c {
			fmt.Fprintf(w, "  %s\t%s\n", k, v.desc)
		}
		w.Flush()
		fmt.Fprintln(cmdOut)
	}
	printOpt(fs)
	os.Exit(1)
}

func printOpt(fs *flag.FlagSet) {
	print := false
	w := tabwriter.NewWriter(cmdOut, 0, 0, 2, ' ', 0)
	fs.VisitAll(func(f *flag.Flag) {
		print = true
		fmt.Fprintf(w, "  --%s\t%s\n", f.Name, f.Usage)
	})
	if print {
		fmt.Fprintf(cmdOut, "Options:\n")
		w.Flush()
	}
}

func isPipe(p *os.File) bool {
	sin, _ := p.Stat()
	if (sin.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

func errExit(s string) {
	fmt.Fprintln(os.Stderr, s)
	os.Exit(1)
}
