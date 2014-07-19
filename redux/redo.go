package main

import (
	"flag"
	"os"
	"strings"

	"github.com/gyepisam/multiflag"
	"github.com/jireva/redux"
)

var cmdRedo = &Command{
	UsageLine: "redux redo [OPTION]... [TARGET]...",
	Short:     "Builds files atomically.",
	Long:      "The redo command builds files atomically by running a do script asssociated with the target.",
	LinkName:  "redo",
}

func init() {
	// avoid initialization loop
	cmdRedo.Run = runRedo
}

var (
	verbosity *multiflag.Value
	debug     *multiflag.Value
)

func init() {
	flg := flag.NewFlagSet("redo", flag.ContinueOnError)

	verbosity = multiflag.BoolSet(flg, "verbose", "false", "Be verbose. Repeat for intensity.", "v")

	debug = multiflag.BoolSet(flg, "debug", "false", "Print debugging output.", "d")

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

	// Set explicit options to avoid clobbering environment inherited options.
	if n := verbosity.NArg(); n > 0 {
		os.Setenv("REDO_VERBOSE", strings.Repeat("x", n))
		redux.Verbosity = n
	}

	if n := debug.NArg(); n > 0 {
		os.Setenv("REDO_DEBUG", "true")
		redux.Debug = true
	}

	// If no arguments are specified print usage and exit.
	if len(targets) == 0 {
		cmdRedo.Flag.Usage()
		os.Exit(1)
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// It *is* slower to reinitialize for each target, but doing
	// so guarantees that a single redo call with multiple targets that
	// potentially have differing roots will work correctly.
	for _, path := range targets {
		if file, err := redux.NewFile(wd, path); err != nil {
			return err
		} else {
			if err := file.Redo(); err != nil {
				return err
			}
		}
	}

	return nil
}
