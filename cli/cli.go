package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/serboupal/note/internal/local"
	"github.com/serboupal/note/note"
)

const appFolder = "note"

var backend note.Backend

var commands = map[string]func([]string){
	"add":    add,
	"list":   list,
	"view":   view,
	"search": search,
	"delete": delete,
	"edit":   edit,
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
		usage()
		os.Exit(1)
	}
	cmd, ok := commands[subcommand[0]]
	if !ok {
		usage()
		os.Exit(1)
	}

	if *debug {
		log.Println("Debug info on")
	}
	cmd(subcommand[1:])
}

func usage() {
	fmt.Println("Usage")
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
