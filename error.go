package redo

import (
	"fmt"
	"os"
)

func Fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %s: ", os.Args[0])
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func FatalErr(err error) {
	Fatal("%s", err)
}
