package redux

// This file exists to setup the binary for tests and to cleanup.
// Please don't create any test files that sort lower than z
// and don't add more tests to this file.

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binDir string

// build the binary, just for testing.
// 1. Ensures that we are using current code
// 2. Not inadvertenly running a version installed somewhere else in the path.
func init() {
	var err error

	binDir, err = ioutil.TempDir("", "redux-test-bin-")
	if err != nil {
		panic(err)
	}

	bin := filepath.Join(binDir, "/redux")

	err = exec.Command("go", "build", "-o", bin, "github.com/gyepisam/redux/redux").Run()
	if err != nil {
		panic(err)
	}

	err = exec.Command(bin, "install", "links").Run()
	if err != nil {
		panic(err)
	}

	//redux should be first in path.
	path := []string{binDir}

	// add all other path entries, except .../go/bin entries.
	for _, slot := range strings.Split(os.Getenv("PATH"), ":") {
		if len(slot) > 0 && strings.Index(slot, "/go/bin") == -1 {
			path = append(path, slot)
		}
	}

	err = os.Setenv("PATH", strings.Join(path, ":"))
	if err != nil {
		panic(err)
	}
}

// This is not a test and must always be the last function in the file!
func TestAtExit(t *testing.T) {
	os.RemoveAll(binDir)
}
