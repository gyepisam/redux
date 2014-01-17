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
Usage: %s [OPTION]... [TARGET]...
A redo build tool.

`
		fmt.Fprintf(os.Stderr, header, os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	redo.RedoIfX(func(file *redo.File, dependent *redo.File) error {
		return file.RedoIfCreate(dependent)
	})
}
