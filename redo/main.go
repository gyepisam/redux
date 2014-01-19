// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fileutils"
	"flag"
	"fmt"
	"os"
	"strings"

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

	debug := multiflag.Bool("debug", "false", "Print debugging output.", "d")

	isTask := flag.Bool("task", false, "Run .do script for side effects and ignore output.")

	shArgs := flag.String("sh", "", "Extra arguments for /bin/sh.")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// set options from environment if not provided.
	if verbosity.NArg() == 0 {
		for i := len(os.Getenv("REDO_VERBOSE")); i > 0; i-- {
			verbosity.Set("true")
		}
	}

	if debug.NArg() == 0 {
		if len(os.Getenv("REDO_DEBUG")) > 0 {
			debug.Set("true")
		}
	}

	// Set explicit options to avoid clobbering environment inherited options.
	if n := verbosity.NArg(); n > 0 {
		os.Setenv("REDO_VERBOSE", strings.Repeat("x", n))
		redo.Verbosity = n
	}

	if n := debug.NArg(); n > 0 {
		os.Setenv("REDO_DEBUG", "true")
		redo.Debug = true
	}

	if s := *shArgs; s != "" {
		os.Setenv("REDO_SHELL_ARGS", s)
		redo.ShellArgs = s
	}

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

	wd, err := os.Getwd()
	if err != nil {
	  redo.FatalErr(err)
	}

	// It *is* slower to reinitialize for each target, but doing
	// so guarantees that a single redo call with multiple targets that
	// potentially have differing roots will work correctly.
	for _, path := range targets {
		if file, err := redo.NewFile(wd, path); err != nil {
			redo.FatalErr(err)
		} else {
			file.IsTaskFlag = *isTask
			if err := file.Redo(); err != nil {
				redo.FatalErr(err)
			}
		}
	}
}
