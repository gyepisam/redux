package redo

import (
	"path"
	"os"
)

// InitDir initializes a redo directory in the specified project root directory, which is created if necessary.
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

	dirname = path.Join(dirname, REDO_DIR)

	if err := os.MkdirAll(dirname, DIR_PERM); err != nil {
		return err
	}

	return nil
}
