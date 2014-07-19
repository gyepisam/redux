// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"fmt"
	"strings"
)

const (
	// REDO_DIR names the hidden directory used for data and configuration.
	REDO_DIR = ".redo"

	// REDO_DIR_ENV_NAME names the environment variable for the REDO_DIR hidden directory.
	REDO_DIR_ENV_NAME = "REDO_DIR"

	// KEY_SEPARATOR is used to join the parts of the database key.
	KEY_SEPARATOR = "/"

	// AUTO marks system generated event records.
	AUTO = "auto"

	// Directory creation permission mode
	DIR_PERM = 0755
)

// Dependency Relations
type Relation string

const (
	SATISFIES Relation = "satisfies"
	REQUIRES  Relation = "requires"
)

// makeKey returns a database key consisting of provided arguments, joined with KEY_SEPARATOR
// and prefixed with the PathHash.
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

func RecordRelation(dependent *File, target *File, event Event, m *Metadata) error {
	if err := dependent.PutPrerequisite(event, target.PathHash, target.AsPrerequisite(dependent.RootDir, m)); err != nil {
		return err
	}

	if err := target.PutDependency(event, dependent.PathHash, dependent.AsDependent(target.RootDir)); err != nil {
		return err
	}

	return nil
}
