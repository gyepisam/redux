package redux

import (
	"io"
	"io/ioutil"
	"os"
)

// An Output is the output of a .do scripts, either through stdout or $3 (Arg3)
// If the .do script invocation is equivalent to the sh command,
//
//	  sh target.ext.do target.ext target tmp0 > tmp1
//
// tmp0 and tmp1 would be outputs.
type Output struct {
	*os.File
	// if file is used as $3 and it needs to be copied,
	// file attributes such as chmod may need to be adjusted.
	IsArg3 bool
}

func (out *Output) Copy(destDir string) (destPath string, err error) {

	src, err := os.Open(out.Name())
	if err != nil {
		return
	}

	dst, err := ioutil.TempFile(destDir, "-redux-output-")
	if err != nil {
		return
	}

	destPath = dst.Name()

	defer func() {
		src.Close()
		dst.Close()

		if err != nil {
			os.Remove(dst.Name())
		}
	}()

	_, err = io.Copy(dst, src)
	if err != nil {
		return
	}

	if !out.IsArg3 {
		return
	}

	// chmod may have been called on an Arg3 file
	// so it may be necessary to fix up the new file similarly.

	srcInfo, err := src.Stat()
	if err != nil {
		return "", err
	}

	dstInfo, err := dst.Stat()
	if err != nil {
		return "", err
	}

	if perm := srcInfo.Mode() & os.ModePerm; perm != (dstInfo.Mode() & os.ModePerm) {
		err := dst.Chmod(perm)
		if err != nil {
			return "", err
		}
	}

	// Fixup file ownership as necessary.
	// These operations are not portable, but should always succeed where they are supported.
	srcUid, srcGid, srcErr := statUidGid(srcInfo)
	if srcErr != nil {
		return
	}

	dstUid, dstGid, dstErr := statUidGid(dstInfo)
	if dstErr != nil {
		return
	}

	if dstUid != srcUid || dstGid != srcGid {
		err = dst.Chown(int(srcUid), int(srcGid))
	}

	return
}

func (out *Output) Size() (size int64, err error) {
	var finfo os.FileInfo

	if out.IsArg3 {
		// f.Stat() doesn't work for the file on $3 since it was written to by a different process.
		finfo, err = os.Stat(out.Name())
	} else {
		finfo, err = out.Stat()
	}

	if err == nil {
		size = finfo.Size()
	}
	return
}
