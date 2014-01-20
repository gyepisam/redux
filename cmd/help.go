package main

import (
	"fmt"
	"os"
)

var (
	header = "redux implements a set of top down build tools."
	footer = "See 'redux help' for an overview and 'redux help command' for command specific information"
)

var cmdHelp = &Command{
	UsageLine: "redux help [name]",
	Short:     "Documents commands.",
	Long:      header + "\n" + footer,
}

func init() {
	// break referential loop
	cmdHelp.Run = runHelp
}

func runHelp(args []string) {

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "%s\nusage: redux command [options] [arguments]\n\nCommands:\n", header)
		for _, cmd := range commands {
			fmt.Fprintf(os.Stderr, "%11s -- %s\n", cmd.Name(), cmd.Short)
		}
		fmt.Fprintf(os.Stderr, "\n%s\n", footer)
		return
	}

	cmdName := args[0]

	if cmd := cmdByName(cmdName); cmd != nil {
		fmt.Fprintf(os.Stderr, "usage: %s\n\n", cmd.UsageLine)
		if cmd.Flag != nil {
			cmd.Flag.PrintDefaults()
		}
		fmt.Fprintf(os.Stderr, "%s\n", cmd.Long)
		return
	}

	fmt.Fprintf(os.Stderr, "%s: unknown command %s. See %s --help\n", os.Args[0], cmdName, os.Args[0])
	os.Exit(2)
	return
}
