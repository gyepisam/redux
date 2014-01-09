package redo

import (
  "fmt"
  "path/filepath"
)

// makeKey returns a database key consisting of provided arguments, prefixed
// with the path hash.
func (f *File) makeKey(subkeys ...interface{}) string {
	keys := make([]string, len(subkeys)+1)

	keys[0] = string(f.PathHash)

	for i, value := range subkeys {
		keys[i+1] = fmt.Sprintf("%s", value)
	}

	// Could also use string.Join(keys, ".") ...
	return filepath.Join(keys...)
}


func (f *File) metadataKey() string {
	return f.makeKey("METADATA")
}

func (f *File) mustRebuildKey() string {
	return f.makeKey("REBUILD")
}
