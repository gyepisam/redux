package redo

import (
	"flag"
	"os"
)

// allow repeated -v flags to increase verbosity
type BoolCounter struct {
	Name  string
	Value int
}

func (v *BoolCounter) String() string {
	return "false. Repeat for intensity"
}

func (v *BoolCounter) Set(s string) error {
	v.Value++
	return nil
}

func (v *BoolCounter) IsBoolFlag() bool { return true }

var (
	verbose BoolCounter = BoolCounter{Name: "verbosity"}
	trace   BoolCounter = BoolCounter{Name: "trace"}
	help    bool
)

func init() {
	verbose_desc := "increase verbosity"
	flag.Var(&verbose, "verbose", verbose_desc)
	flag.Var(&verbose, "v", verbose_desc)

	trace_desc := "trace shell script execution"
	flag.Var(&trace, "trace", trace_desc)
	flag.Var(&trace, "t", trace_desc)
	flag.Var(&trace, "x", trace_desc)

	flag.BoolVar(&help, "help", false, "show this message")

}

// Allow some flags to be set by environment variables
func init() {
	if verbose.Value == 0 {
		verbose.Value = len(os.Getenv("REDO_VERBOSE"))
	}
	if trace.Value == 0 {
		trace.Value = len(os.Getenv("REDO_TRACE"))
	}
}

func Verbosity(n int) bool { return verbose.Value >= n }
func Verbose() bool        { return Verbosity(1) }
func Trace() bool          { return trace.Value > 0 }

func Init() {
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}
}

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
