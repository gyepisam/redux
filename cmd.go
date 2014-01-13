//package cmd provides shared flag handling
package cmd

import (
	"flag"
	"os"
)

var (
	verbose bool
	trace   bool
	help    bool
)

func init() {
	flag.BoolVar(&verbose, "verbose", false, "increase verbosity")
	flag.BoolVar(&trace, "trace", false, "trace shell script execution")
	flag.BoolVar(&help, "help", false, "show this message")
}

func Verbose() bool { return verbose }
func Trace() bool { return trace }

func Init() {
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}

	// TODO (gsam): add these as Redo attributes.
	// TODO (gsam): change verbose to boolean and add verbosity as int.
	if Verbose() {
		os.Setenv("REDO_VERBOSE", "1")
	}

	if Trace() {
		os.Setenv("REDO_TRACE", "1")
	}
}
