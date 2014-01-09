package main

import (
	"flag"
	"fmt"
	"os"
	"redo"
)

var help bool

func init() {
	const usage = "Show help"
	const default_value = false
	flag.BoolVar(&help, "help", default_value, usage)
	flag.BoolVar(&help, "h", default_value, usage+"(short form)")
}

func init() {
	flag.Usage = func() {
		header := `
Usage: %s [OPTIONS] [DIRECTORY ...]
Initializes a project root directory for the redo build tools.
`
		footer := `
If one or more DIRECTORY arguments are specified, %s initializes each one.
If no arguments are provided, but an environment variable named %s exists, it is initialized.
If the environment variable does not exist, %s initializes the current directory.

Relative directories are assumed to be in the current directory.
`

		fmt.Fprintf(os.Stderr, header, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, footer, os.Args[0], os.Args[0], redo.REDO_DIR_ENV_NAME)
	}
}

func main() {

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) == 0 {
		if value := os.Getenv(redo.REDO_DIR_ENV_NAME); value != "" {
			args = append(args, value)
		}
	}

	if len(args) == 0 {
		args = append(args, ".")
	}

	for _, dir := range args {
		if err := redo.InitDir(dir); err != nil {
			redo.Fatal("Cannot initialize directory. Error: %s", err)
		}
	}
}
