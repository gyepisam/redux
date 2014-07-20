// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/gyepisam/redux"
)

var cmdIfCreate = &Command{
	Run:       runIfCreate,
	UsageLine: "redux ifcreate [TARGET...]",
	LinkName:  "redo-ifcreate",
	Short:     "Creates dependency on non-existence of targets.",
	Long: `
The ifcreate command creates a dependency on the non-existence of the target files.
The current file will be invalidated if the target comes into existence.
If the target exists, the command returns an error.
`,
}

func runIfCreate(args []string) error {
	return redoIfX(args, func(file *redux.File, dependent *redux.File) error {
		return file.RedoIfCreate(dependent)
	})
}

var cmdIfChange = &Command{
	Run:       runIfChange,
	UsageLine: "redux ifchange [TARGET...]",
	LinkName:  "redo-ifchange",
	Short:     "Creates dependency on targets and ensure that targets are up to date.",
	Long: `
The ifchange command creates a dependency on the target files and ensures that
the target files are up to date, calling the redo command, if necessary.

The current file will be invalidated if a target is rebuilt.
`,
}

func runIfChange(args []string) error {
	return redoIfX(args, func(file *redux.File, dependent *redux.File) error {
		return file.RedoIfChange(dependent)
	})
}

func redoIfX(args []string, fn func(*redux.File, *redux.File) error) error {

	dependentPath := os.Getenv("REDO_PARENT")
	if len(dependentPath) == 0 {
		return fmt.Errorf("Missing env variable REDO_PARENT")
	}

	wd := os.Getenv("REDO_PARENT_DIR")

	// The action is triggered by dependent.
	dependent, err := redux.NewFile(wd, dependentPath)
	if err != nil {
		return err
	}

	for _, path := range args {
		if file, err := redux.NewFile(wd, path); err != nil {
			return err
		} else if err := fn(file, dependent); err != nil {
			return err
		}
	}

	return nil
}
