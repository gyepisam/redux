package redo

import (
	"strings"
)

const (
	// Prefix marker for scripts that don't produce content.
	TASK_PREFIX = '@'

	// Name of redo hidden directory.
	REDO_DIR = ".redo"

	// Name of environment variable for hidden directory.
	REDO_DIR_ENV_NAME = "REDO_DIR"

	REDO_PARENT_ENV_NAME = "REDO_PARENT"
)

// Redo Events of note
type Event string

const (
	IFCREATE Event = "ifcreate"
	IFCHANGE Event = "ifchange"
)

const AUTO = "auto"

func (e Event) String() string {
	return string(e)
}

func (e Event) Prefix(value string) Event {
	return Event(strings.Join([]string{value, string(e)}, "/"))
}

func (e Event) AutoPrefix() Event {
	return e.Prefix(AUTO)
}

// Dependency Relations
type Relation string

const (
	SATISFIES Relation = "satisfies"
	REQUIRES  Relation = "requires"
)

// Directory creation permission mode
const DIR_PERM = 0755
