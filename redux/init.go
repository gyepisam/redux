package main

import (
	"fmt"
	"os"

	"github.com/jireva/redux"
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

func runInit(args []string) error {
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
			return fmt.Errorf("cannot initialize directory: %s", err)
		}
	}

	return nil
}
