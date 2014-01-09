package redo

import (
	"fmt"
	"os"
)

func Trace(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
}
