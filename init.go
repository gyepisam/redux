package redo

import (
  "os"
  "path/filepath"
)

var cmdInit = &Command{
	Run:       runInit,
	LinkName:  "redo-init",
	UsageLine: "redux init [OPTIONS] [DIRECTORY ...]",
	Short:     "Creates or reinitializes one or more redo root directories.",
	Long: `
If one or more DIRECTORY arguments are specified, the command initializes each one.
If no arguments are provided, but an environment variable named REDO_DIR exists, it is initialized.
If neither arguments nor an environment variable is provided, the current directory is initialized.
`,
}

func runInit(name string, args []string) {
	if len(args) == 0 {
		if value := os.Getenv(REDO_DIR_ENV_NAME); value != "" {
			args = append(args, value)
		}
	}

	if len(args) == 0 {
		args = append(args, ".")
	}

	for _, dir := range args {
		if err := InitDir(dir); err != nil {
			Fatal("cannot initialize directory: %s", err)
		}
	}
}

// InitDir initializes a redo directory in the specified project root directory, creating it if necessary.
func InitDir(dirname string) error {

	if len(dirname) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dirname = wd
	} else if c := dirname[0]; c != '.' && c != '/' {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dirname = filepath.Join(wd, dirname)
	}

	return os.MkdirAll(filepath.Join(dirname, REDO_DIR), DIR_PERM)
}
