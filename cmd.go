package redo

import (
	"flag"
	"os"
)

// Options set by main()
var (
	Verbosity int
	Trace     bool
)

// RedoIfX abstracts functionality common to redo-ifchange and redo-ifcreate
func RedoIfX(fn func(*File, *File) error) error {

	dependentPath := os.Getenv(REDO_PARENT_ENV_NAME)
	if len(dependentPath) == 0 {
		Fatal("Missing env variable %s", REDO_PARENT_ENV_NAME)
	}

	// The action is triggered by dependent.
	dependent, err := NewFile(dependentPath)
	if err != nil {
		FatalErr(err)
	}

	for _, path := range flag.Args() {

		if file, err := NewFile(path); err != nil {
			FatalErr(err)
		} else if err := fn(file, dependent); err != nil {
			FatalErr(err)
		}
	}

	return nil
}
