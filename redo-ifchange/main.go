package main

import (
	"flag"
	"fmt"
	"os"

	"redo"
	"redo/cmd"
)

func init() {
	flag.Usage = func() {
		header := `
Usage: %s [OPTION]... [TARGET]...
A redo build tool.

`
		fmt.Fprintf(os.Stderr, header, os.Args[0])

		flag.PrintDefaults()

		//footer := ` `
		//		fmt.Fprintf(os.Stderr, footer)
	}
}

func main() {

	cmd.Init()

	dependentPath := os.Getenv(redo.REDO_PARENT_ENV_NAME)
	if len(dependentPath) == 0 {
	  redo.Fatal("Missing env variable %s", redo.REDO_PARENT_ENV_NAME)
	}

	// The action is triggered by dependent.
	dependent, err := redo.NewFile(dependentPath)
	if err != nil {
			redo.FatalErr(err)
	}

	for _, path := range flag.Args() {

		if file, err := redo.NewFile(path); err != nil {
			redo.FatalErr(err)
		} else if err := file.RedoIfChange(dependent); err != nil {
			redo.FatalErr(err)
		}
	}
}
