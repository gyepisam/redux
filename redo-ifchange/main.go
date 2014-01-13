package main

import (
	"flag"
	"fmt"
	"os"

	"redo"
)

func init() {
	flag.Usage = func() {
		header := `
Usage: %s [OPTION]... [TARGET]...
A redo build tool..

`
		fmt.Fprintf(os.Stderr, header, os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
  redo.Init()
  redo.RedoIfX(func(file *redo.File, dependent *redo.File) error {
	return file.RedoIfChange(dependent)
  })
}
