package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Command represents a redux command such as redo, ifchange, etc.
type Command struct {

	// Run runs the command.
	Run func(args []string)

	// LinkName is the the name used to link to the executable so it can be called as such.
	LinkName string

	// UsageLine shows the usage for the command.
	UsageLine string

	// Short is a short, single line, description.
	Short string

	// Long is a long description.
	Long string

	// Flag is a list of flags that the command handles.
	Flag *flag.FlagSet

	// Denotes whether the help flag has been invoked
	Help bool
}

// Name returns the name of the command, which is the second word in UsageLine.
func (cmd *Command) Name() string {
	s := strings.SplitN(cmd.UsageLine, " ", 3)
	if len(s) < 2 {
		return cmd.UsageLine
	}
	return s[1]
}

var commands = []*Command{
	cmdInit,
	cmdIfChange,
	cmdIfCreate,
	cmdRedo,
}

func cmdByName(name string) *Command {
	for _, cmd := range commands {
		if name == cmd.Name() {
			return cmd
		}
	}
	return nil
}

func cmdByLinkName(linkName string) *Command {
	for _, cmd := range commands {
		if linkName == cmd.LinkName {
			return cmd
		}
	}
	return nil
}

var wantHelp bool

func initFlags() {

	helpFlags := []string{"help", "h", "?"}
	helpUsage := "Show help"

	for _, name := range helpFlags {
		flag.BoolVar(&wantHelp, name, false, helpUsage)
	}

	for _, cmd := range commands {
		name := cmd.Name()
		if cmd.Flag == nil {
			cmd.Flag = flag.NewFlagSet(name, flag.ContinueOnError)
		}
		cmd.Flag.Usage = func() {
			printHelp(os.Stderr, name)
		}

		if f := cmd.Flag.Lookup("help"); f == nil {
			for _, name := range helpFlags {
				cmd.Flag.BoolVar(&cmd.Help, name, false, helpUsage)
			}
		}
	}
}

func runCommand(cmd *Command, args []string) {
	cmd.Flag.Parse(args)

	if cmd.Help {
		printHelp(os.Stderr, cmd.Name())
		os.Exit(0)
	}

	cmd.Run(cmd.Flag.Args())
}

func main() {

	initFlags()

	// Called by link?
	cmd := cmdByLinkName(filepath.Base(os.Args[0]))
	if cmd != nil {
		runCommand(cmd, os.Args[1:])
		return
	}

	flag.Parse()
	if wantHelp {
		printHelp(os.Stderr)
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		printHelpAll(os.Stderr)
		os.Exit(2)
		return
	}

	cmdName := args[0]
	if cmdName == "help" {
		printHelp(os.Stderr, args[1:]...)
		os.Exit(1)
		return
	}

	cmd = cmdByName(cmdName)
	if cmd == nil {
		printHelp(os.Stderr, cmdName)
		os.Exit(2)
		return
	}

	runCommand(cmd, args[1:])
	return
}

func printHelpAll(out io.Writer) {
	const (
		header = `
redux implements a set of redo top down build tools.
usage: redux command [options] [arguments]

Commands:
`
		footer = "See 'redux help [command]' for more information"
	)

	io.WriteString(out, header)

	for _, cmd := range commands {
		fmt.Fprintf(out, "%11s -- %s\n", cmd.Name(), cmd.Short)
	}

	fmt.Fprintf(out, "\n%s\n", footer)

	return
}

func printHelp(out io.Writer, args ...string) {
	if len(args) == 0 {
		printHelpAll(out)
		return
	}

	cmdName := args[0]

	if cmd := cmdByName(cmdName); cmd != nil {
		fmt.Fprintf(out, "%s\nusage: %s\n\nOptions\n\n", cmd.Short, cmd.UsageLine)
		cmd.Flag.SetOutput(out)
		cmd.Flag.PrintDefaults()
		fmt.Fprintf(out, "%s\n", cmd.Long)
		return
	}

	fmt.Fprintf(out, "%s: unknown command %s. See %s --help\n", os.Args[0], cmdName, os.Args[0])
}
