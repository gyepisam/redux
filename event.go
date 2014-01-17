// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redo

/*
* An Event denotes a state change upon which dependencies are based.
* The line "redo-ifchange B" in the file A.do creates a prerequisite
* from A to B based on the ifchange event.
 */

// Event names a change or create event
type Event string

// Events of note
const (
	IFCREATE Event = "ifcreate"
	IFCHANGE Event = "ifchange"

	AUTO_IFCREATE Event = AUTO + KEY_SEPARATOR + "ifcreate"
	AUTO_IFCHANGE Event = AUTO + KEY_SEPARATOR + "ifchange"
)

// String representation.
func (event Event) String() string {
	return string(event)
}
