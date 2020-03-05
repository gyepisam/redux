package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/gyepisam/fileutils"
	"github.com/gyepisam/multiflag"
	"github.com/gyepisam/redux"
)

var cmdRedo = &Command{
	UsageLine: "redux redo [OPTION]... [TARGET]...",
	Short:     "Builds files atomically.",
	LinkName:  "redo",
}

func init() {
	// break loop
	cmdRedo.Run = runRedo

	text := `
The redo command builds files atomically by running a do script associated with the target.

redo normally requires one or more target arguments.
If no target argument is provided, redo runs the default do script %s if it exists in the current
directory.

For compatibility, if %s does not exist, but %s exists, it is used instead.
`
	it := redux.TASK_PREFIX + redux.DEFAULT_DO
	cmdRedo.Long = fmt.Sprintf(text, it, it, redux.DEFAULT_DO)
}

var (
	verbosity   *multiflag.Value
	debug       *multiflag.Value
	isTask      bool
	shArgs      string
	ignored     bool // like /dev/null for variables
	concurrency int
)

func init() {
	flg := flag.NewFlagSet("redo", flag.ContinueOnError)

	verbosity = multiflag.BoolSet(flg, "verbose", "false", "Be verbose. Repeat for intensity.", "v")

	debug = multiflag.BoolSet(flg, "debug", "false", "Print debugging output.", "d")

	flg.BoolVar(&isTask, "task", false, "Run .do script for side effects and ignore output.")

	flg.StringVar(&shArgs, "sh", "", "Extra arguments for /bin/sh.")

	flg.BoolVar(&ignored, "old-args", false, "Ignored apenwarr redo compatibility flag")

	flg.IntVar(&concurrency, "j", runtime.GOMAXPROCS(-1), "number of concurrent jobs")

	cmdRedo.Flag = flg
}

func runRedo(targets []string) error {

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

	if s := shArgs; s != "" {
		os.Setenv("REDO_SHELL_ARGS", s)
		redux.ShellArgs = s
	}

	// if shell args are set, ensure that at least minimal verbosity is also set.
	if redux.ShellArgs != "" && (verbosity.NArg() == 0) {
		verbosity.Set("true")
	}

	// Set explicit options to avoid clobbering environment inherited options.
	if n := verbosity.NArg(); n > 0 {
		os.Setenv("REDO_VERBOSE", strings.Repeat("x", n))
		redux.Verbosity = n
	}

	if n := debug.NArg(); n > 0 {
		os.Setenv("REDO_DEBUG", "true")
		redux.Debug = true
	}

	// If no argument is specified, use default target if its .do file exists.
	// Otherwise, print usage and exit.
	if len(targets) == 0 {
		for _, prefix := range []string{redux.TASK_PREFIX, ""} {
			doFile := prefix + redux.DEFAULT_DO
			if found, err := fileutils.FileExists(doFile); err != nil {
				return err
			} else if found {
				targets = append(targets, prefix+redux.DEFAULT_TARGET)
				break
			}
		}

		if len(targets) == 0 {
			cmdRedo.Flag.Usage()
			os.Exit(1)
			return nil
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	concurrency := concurrency
	if concurrency <= 0 {
		concurrency = runtime.GOMAXPROCS(-1)
	}

	// It *is* slower to reinitialize for each target, but doing
	// so guarantees that a single redo call with multiple targets that
	// potentially have differing roots will work correctly.
	for _, path := range targets {
		file, err := redux.NewFile(wd, path)
		if err != nil {
			return err
		}
		file.SetTaskFlag(isTask)
		if err := file.Redo(); err != nil {
			return err
		}
	}

	return nil
}
