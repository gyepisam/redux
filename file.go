package redo


import (
	"path/filepath"
	"crypto/sha1"
	"io"
	"os"
	"time"
	"encoding/hex"
	"io/ioutil"
)

type File struct {
	// Relative to redo root directory
	// Path for /home/user/project/orange.go would be orange.go
	Path string

	// cached
	basename string	
	name string
	dir string
	ext string
	haveNames bool
	
	// cached
	Size int64
	ModTime time.Time
	haveStats bool
	
	// cached
	ContentHash string
}

func NewFile(path string) (f File, err error) {
	f.Path = path
	return
}

// Dir returns directory part of name
func (f File) Dir() string {
	if f.dir == "" {
		f.dir = filepath.Dir(f.Path) 
	}
	return f.dir
}

// splitName separates a filename into a base and an extension
func (f File) splitName() {
	
	f.ext  = filepath.Ext(f.Path)
	f.name = filepath.Base(f.Path)
	f.basename = f.name[:len(f.name) - len(f.ext)]
	
	f.haveNames = true
}

// BaseName returns file name without a suffix.
func (f File) Basename() string {
	if !f.haveNames {
		f.splitName()
	}
	return f.basename
}

// Name returns the file name.
func (f File) Name() string {
	if !f.haveNames {
		f.splitName()
	}
	return f.name
}


// Ext returns file extension or an empty string.
// Note: extension includes the preceding dot.
func (f File) Ext() string {
	if !f.haveNames {
		f.splitName()
	}
	return f.ext
}

func (f File) updateStat() (err error) {
	if finfo, err := os.Stat(f.Path); err == nil {
		f.Size = finfo.Size()
		f.ModTime = finfo.ModTime()
		f.haveStats = true
	}
	return
}


func (f File) updateContentHash() (err error) {

	f, err := os.Open(f.Path)
	if err != nil {
		return
	}
	defer f.Close()

	if b, err := ioutil.ReadAll(f); err != nil {
		return
	} else if hash, err := makeHash(b); err != nil {
		return
	} else {
		f.ContentHash = hash
		return
	}
}

func (f File) Update() error {
	
	if !f.haveStats {
		if err := f.updateStats(); err != nil {
			return err
		}
	}

	if err := f.updateContentHash(); err != nil {
		return err
	}
	
	return nil
}

func (f File) NameHash() (string, error) {
	return makeHash([]byte(f.Path))
}
