// +build windows

package redux

import (
	"errors"
	"os"
)

func statUidGid(finfo os.FileInfo) (uint32, uint32, error) {
	return 0, 0, errors.New("finfo.Sys() is unsupported")
}
