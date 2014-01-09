package main

import (
	"fileutils"
	"flag"
	"fmt"
	"os"
	"redo"
	"redo/cmd"
)

const DEFAULT_TARGET = string(redo.TASK_PREFIX) + "all"
const DEFAULT_DO = DEFAULT_TARGET + ".do"

func init() {
	flag.Usage = func() {
		header := `
Usage: %s [OPTION]... [TARGET]...
Build files incrementally

`

		footer := `
TARGET defaults to %s iff %s exists.
`
		fmt.Fprintf(os.Stderr, header, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, footer, DEFAULT_TARGET, DEFAULT_DO)
	}
}

func main() {

	cmd.Init()

	targets := flag.Args()

	// If no arguments are specified, use run default target if it exists.
	// Otherwise, print usage and exit.
	if len(targets) == 0 {
		if found, err := fileutils.FileExists(DEFAULT_DO); err != nil {
			redo.FatalErr(err)
		} else if found {
			targets = append(targets, DEFAULT_TARGET)
		} else {
			flag.Usage()
			os.Exit(1)
		}
	}

	// It *is* slower to reinitialize for each target, but doing
	// so guarantees that a single redo call with multiple targets that
	// potentially have differing roots will work correctly.
	for _, path := range targets {
		if file, err := redo.NewFile(path); err != nil {
			redo.FatalErr(err)
		} else if err := file.Redo(); err != nil {
			redo.FatalErr(err)
		}
	}
}
