// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"errors"
	"fmt"
	"os"
)

// Fatal
func Fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %s: ", os.Args[0])
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// FatalErr is a global error handler. It prints the error argument and exits the program.
func FatalErr(err error) {
	Fatal("%s", err)
}

// Errorf formats errors for the current file.
func (f *File) Errorf(format string, args ...interface{}) error {
	return errors.New(fmt.Sprintf("%s: ", f.Target) + fmt.Sprintf(format, args...))
}

// ErrUninitialized denotes an uninitialized directory.
func (f *File) ErrUninitialized() error {
	return f.Errorf("cannot find redo root directory")
}

// ErrNotFound is used when the current file is expected to exists and does not.
func (f *File) ErrNotFound(m string) error {
	return f.Errorf("file not found at %s", m)
}
