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
	IsArg3 bool
}

func NewOutput(file *os.File) *Output {
	return &Output{File: file}
}

func (out *Output) SetupArg3() error {
	if err := out.Close(); err != nil {
		return err
	}

	if err := os.Remove(out.Name()); err != nil {
		return err
	}

	out.IsArg3 = true

	return nil
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

	if out.IsArg3 {
		err = out.copyAttribs(src, dst)
	}

	return
}

func (out *Output) copyAttribs(src, dst *os.File) (err error) {

	srcInfo, err := src.Stat()
	if err != nil {
		return
	}

	dstInfo, err := dst.Stat()
	if err != nil {
		return
	}

	if perm := srcInfo.Mode() & os.ModePerm; perm != (dstInfo.Mode() & os.ModePerm) {
		err = dst.Chmod(perm)
		if err != nil {
			return
		}
	}

	// Fixup file ownership as necessary.
	// These operations are not portable, but should always succeed where they are supported.
	srcUid, srcGid, err := statUidGid(srcInfo)
	if err != nil {
		return
	}

	dstUid, dstGid, err := statUidGid(dstInfo)
	if err != nil {
		return
	}

	if dstUid != srcUid || dstGid != srcGid {
		err = dst.Chown(int(srcUid), int(srcGid))
	}

	return
}

func (out *Output) Cleanup() {
	_ = out.Close()           // ignore error
	_ = os.Remove(out.Name()) //ignore error
}
