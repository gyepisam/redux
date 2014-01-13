package redo

import (
  "fmt"
  "strings"
)

// makeKey returns a database key consisting of provided arguments, prefixed
// with the path hash.
func (f *File) makeKey(subkeys ...interface{}) (val string) {

	keys := make([]string, len(subkeys)+1)

	keys[0] = string(f.PathHash)

	for i, value := range subkeys {
		keys[i+1] = fmt.Sprintf("%s", value)
	}

	return strings.Join(keys, KEY_SEPARATOR)
}


func (f *File) metadataKey() string {
	return f.makeKey("METADATA")
}

func (f *File) mustRebuildKey() string {
	return f.makeKey("REBUILD")
}
