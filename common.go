package redo

import (
	"fmt"
	"strings"
)

const (
	// TASK_PREFIX is a marker for scripts that don't produce content.
	TASK_PREFIX = '@'

	// REDO_DIR names the hidden directory used for data and configuration.
	REDO_DIR = ".redo"

	// REDO_DIR_ENV_NAME names the environment variable for the REDO_DIR hidden directory.
	REDO_DIR_ENV_NAME = "REDO_DIR"

	REDO_PARENT_ENV_NAME = "REDO_PARENT"

	// KEY_SEPARATOR is used to join the parts of the database key.
	KEY_SEPARATOR = "/"

	// AUTO marks system generated event records.
	AUTO = "auto"
)

// Dependency Relations
type Relation string

const (
	SATISFIES Relation = "satisfies"
	REQUIRES  Relation = "requires"
)

// Directory creation permission mode
const DIR_PERM = 0755

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

func RecordRelation(dependent *File, target *File, event Event, prereq Prerequisite) error {
	if err := dependent.PutPrerequisite(event, target.PathHash, prereq); err != nil {
		return err
	}

	if err := target.PutDependency(event, dependent.PathHash, dependent.Path); err != nil {
		return err
	}

	return nil
}
