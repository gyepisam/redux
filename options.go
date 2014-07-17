// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"os"
)

// Options default to env values, to be overriden by main() if necessary.
var (
	Verbosity = len(os.Getenv("REDO_VERBOSE"))
	Debug     = len(os.Getenv("REDO_DEBUG")) > 0
)

func Verbose() bool { return Verbosity > 0 }
