package main

import (
	"fileutils"
	"flag"
	"fmt"
	"os"

	"github.com/gyepisam/multiflag"
	"github.com/gyepisam/redo"
)

const (
	DEFAULT_TARGET = string(redo.TASK_PREFIX) + "all"
	DEFAULT_DO     = DEFAULT_TARGET + ".do"
)

func init() {
	flag.Usage = func() {
		header := `
Usage: %s [OPTION]... [TARGET]...
Build files incrementally.

`

		footer := `
TARGET defaults to %s iff %s exists.
`
		fmt.Fprintf(os.Stderr, header, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, footer, DEFAULT_TARGET, DEFAULT_DO)
	}
}

var ()

func main() {

	help := flag.Bool("help", false, "Show help.")

	verbosity := multiflag.Bool("verbose", "false", "Be verbose. Repeat for intensity.", "v")

	// trace is a bool, but is here represented by a counter so as
	// to distinguish between a default and a user provided false value
	// when falling back to the environment provided setting.
	trace := multiflag.Bool("trace", "false", "Run /bin/sh with -x option.", "t")

	isTask := flag.Bool("task", false, "Run .do script for side effects and ignore output.")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// set options from environment if not provided.
	if verbosity.NArgs() == 0 {
		for i := len(os.Getenv("REDO_VERBOSE")); i > 0; i-- {
			verbosity.Set("true")
		}
	}

	if trace.NArgs() == 0 {
		if val := os.Getenv("REDO_TRACE"); val != "" {
			trace.Set("true")
		}
	}

	redo.Verbosity = verbosity.NArgs()
	redo.Trace = trace.NArgs() > 0

	targets := flag.Args()

	// If no arguments are specified, use run default target if its .do file exists.
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
		} else {
			file.IsTaskFlag = *isTask
			if err := file.Redo(); err != nil {
				redo.FatalErr(err)
			}
		}
	}
}
