package redo

import (
	"os"
	"path/filepath"
	"strings"
)

// Command represents a redux command such as redo, ifchange, etc.
type Command struct {

	// Run runs the command.
	Run func(name string, arg []string)

	// LinkName is the the name used to link to the executable so it can be called as such.
	LinkName string

	// UsageLine shows the usage for the command.
	UsageLine string

	// Short is a short, single line, description.
	Short string

	// Long is a long description.
	Long string

	// Flags is a list of flags that the command handles.
	Flags []string
}

// Name returns the name of the command, which is the second word in UsageLine.
func (cmd *Command) Name() string {
	s := strings.SplitN(cmd.UsageLine, " ", 3)
	if len(s) < 2 {
		return cmd.UsageLine
	}
	return s[1]
}

var (
	commands = []*Command{
		cmdHelp,
		cmdInit,
		/*
			cmdRedo,
			cmdIfChange,
			cmdIfCreate,
		*/
	}
)

func cmdByName(name string) *Command {
	for _, cmd := range commands {
		if name == cmd.Name() {
			return cmd
		}
	}
	return nil
}

func cmdByLinkName(exeName string) *Command {
	for _, cmd := range commands {
		if exeName == cmd.LinkName {
			return cmd
		}
	}
	return nil
}

var exitStatus = 0

func main() {

	var cmd *Command
	var args = os.Args[1:]

	// Called by link?
	cmd = cmdByLinkName(filepath.Base(os.Args[0]))
	if cmd != nil {
		cmd.Run(cmd.LinkName, args)
		return
	}

	var cmdName string

	if len(args) < 1 {
		cmd = cmdHelp
		exitStatus = 2
	} else {

		cmdName = args[1]
		cmd = cmdByName(cmdName)

		if cmd == nil {
			cmd = cmdHelp
			exitStatus = 2
		}

		args = args[2:]
	}

	cmd.Run(cmdName, args)

	os.Exit(exitStatus)
}
