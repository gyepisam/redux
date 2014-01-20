package redo

import (
	"fmt"
	"os"
)

// Separate declaration from initialization to break reference loop.
var cmdHelp *Command

func init() {
  cmdHelp = &Command{
	Run:       runHelp,
	UsageLine: "redux help [name]",
	Short:     "Documents redux commands",
	Long: `
	redux implements the redo top down software build method.
	The command redux help prints out documentation for the redux commands.
	Type 'redux help' For an overview and 'redux help command' for command specific information'
	`,
  }
}

func runHelp(name string, args []string) {
	if name == "" {
	  fmt.Fprintf(os.Stderr, "%s\n\nCommands:\n", cmdHelp.Long)
		for _, cmd := range commands {
			fmt.Fprintf(os.Stderr, "%s %s\n", cmd.Name(), cmd.Short)
		}
		return
	}

	cmd := cmdByName(name)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "%s: unknown command %s. Try %s help\n", os.Args[0], name, os.Args[0])
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "%s\n%s\n%s\n", cmd.UsageLine, cmd.Long)

	return
}
