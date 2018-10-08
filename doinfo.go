package redux

import (
	"path/filepath"
)

// A DoInfo represents an active do file.
type DoInfo struct {
	Dir     string
	Name    string
	Arg2    string   //do file arg2. Depends on target and do file names.
	RelDir  string   //relative directory to target from do script.
	Missing []string //more specific do scripts that were not found.
}

func (do *DoInfo) Path() string {
	return filepath.Join(do.Dir, do.Name)
}

func (do *DoInfo) RelPath(path string) string {
	return filepath.Join(do.RelDir, path)
}
