// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gyepisam/redo"
)

func init() {
	flag.Usage = func() {
		header := `
Usage: %s [OPTIONS] [DIRECTORY ...]
Initialize a redo root directory, creating it, if necessary.
`
		footer := `
If one or more DIRECTORY arguments are specified, %s initializes each one.
If no arguments are provided, but an environment variable named %s exists, it is initialized.
If the environment variable does not exist, the current directory is initialized.

`

		fmt.Fprintf(os.Stderr, header, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, footer, os.Args[0], os.Args[0], redo.REDO_DIR_ENV_NAME)
	}
}

func main() {

	help := flag.Bool("help", false, "show this message.")

	flag.Parse()

	if *help {
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
			redo.Fatal("cannot initialize directory: %s", err)
		}
	}
}
