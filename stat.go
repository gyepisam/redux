// +build !windows

package redux

import (
	"errors"
	"os"
	"syscall"
)

func statUidGid(finfo os.FileInfo) (uint32, uint32, error) {
	sys := finfo.Sys()
	if sys == nil {
		return 0, 0, errors.New("finfo.Sys() is unsupported")
	}
	stat := sys.(*syscall.Stat_t)
	return stat.Uid, stat.Gid, nil
}
