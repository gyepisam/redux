package main

import (
  "redux"
)

var cmdIfCreate = &Command{
   Run: runIfCreate,
	UsageLine: "redux ifcreate [TARGET...]",
	Short: "Create dependency on non-existence of targets.",
	Long: `
The ifcreate command creates a dependency on the non-existence of the target files.
The current file will be invalidated if the target comes into existence.
If the target exists, the command returns an error.
`,
}

func runIfCreate(args[]string) {
	redux.RedoIfX(func(file *redux.File, dependent *redux.File) error {
		return file.RedoIfCreate(dependent)
	})
}

var cmdIfChange = &Command{
   Run: runIfChange,
	UsageLine: "redux ifcreate [TARGET...]",
	Short: "Create dependency on targets and ensure that targets are up to date.",
	Long: `
The ifchange command creates a dependency on the target files and ensures that
the target files are up to date, calling the redo command, if necessary.

The current file will be invalidated if a target is rebuilt.
`,
}

func runIfChange(args[]string) {
	redux.RedoIfX(func(file *redux.File, dependent *redux.File) error {
		return file.RedoIfChange(dependent)
	})
}
