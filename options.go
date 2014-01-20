// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"flag"
	"os"
)

// Options set by main()
var (
	Verbosity int
	Debug     bool
	ShellArgs string
)

func init() {
	Verbosity = len(os.Getenv("REDO_VERBOSE"))
	Debug = len(os.Getenv("REDO_DEBUG")) > 0
	ShellArgs = os.Getenv("REDO_SHELL_ARGS")
}

func Verbose() bool { return Verbosity > 0 }

// RedoIfX abstracts functionality common to redo-ifchange and redo-ifcreate
func RedoIfX(fn func(*File, *File) error) error {

	dependentPath := os.Getenv(REDO_PARENT_ENV_NAME)
	if len(dependentPath) == 0 {
		Fatal("Missing env variable %s", REDO_PARENT_ENV_NAME)
	}

	wd, err := os.Getwd()
	if err != nil {
	  return err
	}

	// The action is triggered by dependent.
	dependent, err := NewFile(wd, dependentPath)
	if err != nil {
		FatalErr(err)
	}

	for _, path := range flag.Args() {
		if file, err := NewFile(wd, path); err != nil {
			FatalErr(err)
		} else if err := fn(file, dependent); err != nil {
			FatalErr(err)
		}
	}

	return nil
}
