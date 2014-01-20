package main

import (
	"os"
	"fmt"

	"redux"

)

var cmdInit = &Command{
	Run:       runInit,
	LinkName:  "redo-init",
	UsageLine: "redux init [OPTIONS] [DIRECTORY ...]",
	Short:     "Creates or reinitializes one or more redo root directories.",
}

func init() {
  text := `
If one or more DIRECTORY arguments are specified, the command initializes each one.
If no arguments are provided, but an environment variable named %s exists, it is initialized.
If neither arguments nor an environment variable is provided, the current directory is initialized.
`
	cmdInit.Long = fmt.Sprintf(text, redux.REDO_DIR_ENV_NAME)
}

func runInit(args []string) {
	if len(args) == 0 {
		if value := os.Getenv(redux.REDO_DIR_ENV_NAME); value != "" {
			args = append(args, value)
		}
	}

	if len(args) == 0 {
		args = append(args, ".")
	}

	for _, dir := range args {
		if err := redux.InitDir(dir); err != nil {
			redux.Fatal("cannot initialize directory: %s", err)
		}
	}
}
