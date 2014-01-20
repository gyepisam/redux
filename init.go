package redux

import (
  "os"
  "path/filepath"
)

// InitDir creates a redo directory in the specified project root directory.
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
