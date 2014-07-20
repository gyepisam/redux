package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Command represents a redux command such as redo, ifchange, etc.
type Command struct {

	// Run runs the command.
	Run func(args []string) error

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
	cmdInstall,
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
		printHelpAll(os.Stderr)
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		printHelpAll(os.Stderr)
		os.Exit(2)
		return
	}

	cmdName := args[0]

	if cmdName == "help" || cmdName == "documentation" {

		if len(args) < 2 {
			printHelpAll(os.Stdout)
			return
		}

		cmd := cmdByName(args[1])
		if cmd == nil {
			printUnknown(os.Stderr, args[1])
			os.Exit(2)
		} else {
			cmd.printDoc(os.Stdout, cmdName)
		}
		return
	}

	cmd = cmdByName(cmdName)
	if cmd == nil {
		printUnknown(os.Stderr, cmdName)
		os.Exit(2)
		return
	}

	runCommand(cmd, args[1:])
	return
}

func runCommand(cmd *Command, args []string) {
	err := cmd.Flag.Parse(args)
	if err != nil || cmd.Help {
		printHelp(os.Stderr, cmd.Name())
		os.Exit(0)
		return
	}

	err = cmd.Run(cmd.Flag.Args())
	if err != nil {
		fatalErr(err)
		return
	}
	os.Exit(0)
}

var templates = map[string]string{
	"overview": `redux is an implementation of the redo top down build tools.

Usage: redux command [options] [arguments]

Commands:
{{range .}}
{{.Name | printf "%11s"}} -- {{.Short}}
{{end}}

See 'redux help [command]' for details about each command.
`,

	"help": `{{.Name}} - {{.Short}}

Usage: {{.UsageLine}}

Options

{{.Options}}

{{.Long}}
`,
	"documentation": `
#NAME

{{.Name}} - {{.Short}}

#SYNOPSIS

{{.UsageLine}}

#OPTIONS

{{.Options}}

#NOTES

{{.Long}}
`,
}

func (cmd *Command) printDoc(out io.Writer, docType string) {
	text, ok := templates[docType]
	if !ok {
		panic("unknown docType: " + docType)
	}

	tmpl, err := template.New(docType).Parse(text)
	if err != nil {
		panic(err)
	}

	if docType == "overview" {
		err = tmpl.Execute(out, commands)
		if err != nil {
			panic(err)
		}
		return
	}

	var buf bytes.Buffer
	cmd.Flag.SetOutput(&buf)
	cmd.Flag.PrintDefaults()

	data := map[string]string{
		"Name":      cmd.Name(),
		"UsageLine": cmd.UsageLine,
		"Short":     cmd.Short,
		"Long":      cmd.Long,
		"Options":   buf.String(),
	}

	err = tmpl.Execute(out, data)
	if err != nil {
		panic(err)
	}
}

func printHelpAll(out io.Writer) {
	cmd := &Command{}
	cmd.printDoc(out, "overview")
}

func printUnknown(out io.Writer, name string) {
	fmt.Fprintf(out, "%s: unknown command %s. See %s --help\n", os.Args[0], name, os.Args[0])
}

func printHelp(out io.Writer, args ...string) {

	if len(args) == 0 {
		printHelpAll(out)
		return
	}

	cmdName := args[0]

	cmd := cmdByName(cmdName)
	if cmd == nil {
		printUnknown(out, cmdName)
		return
	}

	cmd.printDoc(out, "help")
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %s: ", os.Args[0])
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func fatalErr(err error) {
	fatal("%s", err)
}
