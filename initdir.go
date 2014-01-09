package redo

import (
	"path"
	"fileutils"
	"fmt"
	"os"
)

// InitDir initializes a redo directory in the specified project root directory.
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
		dirname = path.Join(wd, dirname)
	}

	if isdir, err := fileutils.IsDir(dirname); err != nil {
		return err
	} else if !isdir {
		return fmt.Errorf("Error: %s is not a directory", dirname)
	}

	dirname = path.Join(dirname, REDO_DIR)

	if err := os.MkdirAll(dirname, DIR_PERM); err != nil {
		return err
	}

	return nil
}
