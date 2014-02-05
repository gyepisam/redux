// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gyepisam/fileutils"
)

// File represents a source or target file..
type File struct {
	Target string // file name argument to redo, redo-ifchange, redo-ifcreate, etc

	RootDir string // contains .redo directory
	Path    string // Relative to RootDir

	Dir      string
	Name     string
	Basename string
	Ext      string // File extension. Could be empty. Includes preceeding dot.

	PathHash Hash   // SHA1 hash of Path. Used as database key.
	DoFile   string // Do script used to generate target output.

	Config     Config
	db         DB
	IsTaskFlag bool // If true, target is a task and run for side effects
}

// IsTask denotes when the current target is a task script, either
// implicitly (name begins with @) or explicitly (-task argument to redo).
func (f *File) IsTask() bool {
	return f.IsTaskFlag || len(f.Name) > 0 && f.Name[0] == TASK_PREFIX
}

func splitpath(path string) (string, string) {
	return filepath.Dir(path), filepath.Base(path)
}

// NewFile creates and returns a File instance for the given path.
// If the path is not fully qualified, it is made relative to dir.
// The newly created instance is initialized with the database specified by
// the configuration file found in its root directory or the default database.
// If a file does not have a root directory, it is initialized with a NullDb
// and HasNullDb will return true.
func NewFile(dir, path string) (f *File, err error) {

	if path == "" {
		return nil, errors.New("target path cannot be empty")
	}

	var targetPath string

	if filepath.IsAbs(path) {
		targetPath = path
	} else {
		targetPath = filepath.Clean(filepath.Join(dir, path))
	}

	if isdir, err := fileutils.IsDir(targetPath); err != nil {
		return nil, err
	} else if isdir {
		return nil, fmt.Errorf("target %s is a directory", targetPath)
	}

	f = new(File)

	f.Target = path

	rootDir, filename := splitpath(targetPath)
	relPath := &RelPath{}
	relPath.Add(filename)

	hasRoot := false

	for {
		exists, err := fileutils.DirExists(filepath.Join(rootDir, REDO_DIR))
		if err != nil {
			return nil, err
		}
		if exists {
			hasRoot = true
			break
		}
		if rootDir == "/" || rootDir == "." {
			break
		}
		rootDir, filename = splitpath(rootDir)
		relPath.Add(filename)
	}

	f.RootDir = rootDir

	f.Path = relPath.Join()

	f.PathHash = MakeHash(f.Path)

	f.Debug("@Hash %s: %s -> %s\n", f.RootDir, f.Path, f.PathHash)

	f.Dir, f.Name = splitpath(f.Fullpath())
	f.Ext = filepath.Ext(f.Name)
	f.Basename = f.Name[:len(f.Name)-len(f.Ext)]

	if hasRoot {
		// TODO(gsam): Read config file in rootDir to determine DB type, if any.
		// Default to FileDB if not specified.
		// f.Config = Config{DBType: "file"}
		f.db, err = FileDbOpen(f.RootDir)
		if err != nil {
			return nil, err
		}

		err := os.Mkdir(f.tempDir(), 0755)
		if err != nil && !os.IsExist(err) {
			return nil, err
		}

	} else {
		f.Config = Config{DBType: "null"}
		f.db, err = NullDbOpen("")
		if err != nil {
			return nil, err
		}
		f.Debug("@NullDb\n")
	}

	return
}

// HasNullDb specifies whether the File receiver uses a NullDb.
func (f *File) HasNullDb() bool {
	return f.db.IsNull()
}

// Fullpath returns the fully qualified path to the target file.
func (f *File) Fullpath() string {
	return filepath.Join(f.RootDir, f.Path)
}

// Rel makes path relative to f.RootDir.
func (f *File) Rel(path string) string {
	relpath, err := filepath.Rel(f.RootDir, path)
	if err != nil {
		panic(err)
	}
	return filepath.Clean(relpath)
}

// Abs returns a cleaned up fullpath by joining f.RootDir to path.
func (f *File) Abs(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Clean(filepath.Join(f.RootDir, path))
}

// Exist verifies that the file exists on disk.
func (f *File) Exists() (bool, error) {
	return fileutils.FileExists(f.Fullpath())
}

// HasDoFile returns true if the receiver has been assigned a .do script.
func (f *File) HasDoFile() bool {
	return len(f.DoFile) > 0
}

// IsCurrent returns a boolean denoting whether the target is up to date.

// A target is up to date if the following conditions hold:
//   The file exists
//   The file has not been flagged to be rebuilt
//   The file has not changed since creation. That is; the file has a metadata record
//   	and that record matches the actual file metadata.
//   All the file's prerequisites are also current.

func (f *File) IsCurrent() (bool, error) {
	return f.isCurrent()
}

func (f *File) isCurrent() (bool, error) {

	reason := func(msg string) (bool, error) {
		f.Debug("@Outdated because %s\n", msg)
		return false, nil
	}

	if f.MustRebuild() {
		return reason("REBUILD")
	}

	storedMeta, found, err := f.GetMetadata()
	if err != nil {
		return false, err
	} else if !found {
		return reason("no record metadata")
	}

	fileMeta, err := f.NewMetadata()
	if err != nil {
		return false, err
	} else if fileMeta == nil {
		return reason("no file metadata")
	}

	if !storedMeta.Equal(fileMeta) {
		return reason("record metadata != file metadata")
	}

	// redo-ifcreate dependencies
	created, err := f.PrerequisiteFiles(IFCREATE, AUTO_IFCREATE)
	if err != nil {
		return false, err
	}

	for _, prerequisite := range created {
		if exists, err := prerequisite.Exists(); err != nil {
			return false, err
		} else if exists {
			return reason("ifcreate dependency target exists")
		}
	}

	// redo-ifchange dependencies
	changed, err := f.Prerequisites(IFCHANGE, AUTO_IFCHANGE)
	if err != nil {
		return false, err
	}

	for _, prerequisite := range changed {
		if isCurrent, err := prerequisite.IsCurrent(f.RootDir); err != nil || !isCurrent {
			return isCurrent, err
		}
	}

	return true, nil
}

// NewMetadata computes and returns the file metadata.
func (f *File) NewMetadata() (m *Metadata, err error) {

	m, err = NewMetadata(f.Fullpath(), f.Path)
	if m == nil || err == nil {
		return
	}

	if len(f.DoFile) > 0 {
		if path, err := filepath.Rel(f.RootDir, f.DoFile); err != nil {
			m.DoFile = f.DoFile
		} else {
			m.DoFile = path
		}
	}

	return
}

// ContentHash returns a cryptographic hash of the file contents.
func (f *File) ContentHash() (Hash, error) {
	return ContentHash(f.Fullpath())
}

// Log prints out messages to stderr when the verbosity is greater than N.
func (f *File) Log(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

// Debug prints out messages to stderr when the debug flag is enabled.
func (f *File) Debug(format string, args ...interface{}) {
	if Debug {
		for i, value := range args {
			if value == nil {
				args[i] = "<nil>"
			}
		}
		fmt.Fprintf(os.Stderr, "%s %s: ", os.Args[0], f.Target)
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func (f *File) GenerateNotifications(oldMeta, newMeta *Metadata) error {

	if oldMeta == nil {
		if err := f.NotifyDependents(IFCREATE); err != nil {
			return err
		}
	}

	if !newMeta.Equal(oldMeta) {
		if err := f.NotifyDependents(IFCHANGE); err != nil {
			return err
		}
	}

	return nil
}

// RedoDir returns the path to the .redo directory.
func (f *File) RedoDir() string {
	return filepath.Join(f.RootDir, REDO_DIR)
}

func (f *File) tempDir() string {
	if s := os.Getenv("REDO_TMP_DIR"); len(s) > 0 {
		return s
	}
	return filepath.Join(f.RedoDir(), "tmp")
}

func (f *File) tempFile() (*os.File, error) {
	return ioutil.TempFile(f.tempDir(), strings.Replace(f.Name, ".", "-", -1)+"-redo-tmp-")
}

// NewOutput returns an initialized Output
func (f *File) NewOutput(isArg3 bool) (*Output, error) {
	tmp, err := f.tempFile()
	if err != nil {
		return nil, err
	}
	return &Output{tmp, isArg3}, nil
}

func statUidGid(finfo os.FileInfo) (uint32, uint32, error) {
	sys := finfo.Sys()
	if sys == nil {
		return 0, 0, errors.New("finfo.Sys() is unsupported")
	}
	stat := sys.(*syscall.Stat_t)
	return stat.Uid, stat.Gid, nil
}
