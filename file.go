// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redo

import (
	"errors"
	"fileutils"
	"fmt"
	"os"
	"path/filepath"
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

	PathHash Hash
	DoFile   string

	Config     Config
	db         DB
	IsTaskFlag bool // If true, target is a task and run for side effects
}

// IsTask denotes when the current target is a task script, either
// implicitly (name begins with @) or explicitly (-task argument to redo).
func (f *File) IsTask() bool {
	return f.IsTaskFlag || len(f.Basename) > 0 && f.Basename[0] == TASK_PREFIX
}

// NewFile creates and returns a File instance for the given path.
// The newly created instance is initialized with the database specified by
// the configuration file found in its root directory or the default database.
// If a file does not have a root directory, it is initialized with a NullDb
// and HasNullDb will return true.
func NewFile(targetPath string) (f *File, err error) {

	if targetPath == "" {
		return nil, errors.New("target path cannot be empty")
	}

	if isdir, err := fileutils.IsDir(targetPath); err != nil {
		return nil, err
	} else if isdir {
		return nil, fmt.Errorf("target %s is a directory", targetPath)
	}

	f = new(File)

	f.Target = targetPath

	path, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, err
	}

	components := []string{filepath.Base(path)}
	rootDir := filepath.Dir(path)
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

		components = append(components, filepath.Base(rootDir))
		rootDir = filepath.Dir(rootDir)
	}

	f.RootDir = rootDir

	//components are in reverse order...
	for i, j := 0, len(components)-1; i < j; i, j = i+1, j-1 {
		components[i], components[j] = components[j], components[i]
	}
	f.Path = filepath.Join(components...)

	f.PathHash = MakeHash(f.Path)

	f.Dir, f.Name = filepath.Split(f.Fullpath())
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
	} else {
		f.Config = Config{DBType: "null"}
		f.db, err = NullDbOpen("")
		if err != nil {
			return nil, err
		}
		f.Log("@NullDb for %s\n", f.Target)
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

// Exist verifies that the file exists on disk.
func (f *File) Exists() (bool, error) {
	return fileutils.FileExists(f.Fullpath())
}

// HasDoFile returns true if the receiver has been assigned a .do script.
func (f *File) HasDoFile() bool {
	return len(f.DoFile) > 0
}

/*
 IsCurrent returns a boolean denoting whether the target is up to date.

 A target is up to date if the following conditions hold:
   The file exists
   The file has not been flagged to be rebuilt
   The file has not changed since creation. That is; the file has a metadata record
   	and that record matches the actual file metadata.
   All the file's immediate prerequisites are also current.
   FIXME: May need to remove the limit and check all prerequsites down to the leaves.
*/
func (f *File) IsCurrent() (bool, error) {
	return f.isCurrent(1)
}

func (f *File) isCurrent(depth int) (bool, error) {

	reason := func(msg string) (bool, error) {
		f.Log("@isCurrent %s. %s\n", f.Name, msg)
		return false, nil
	}

	if f.MustRebuild() {
		return reason("REBUILD")
	}

	storedMeta, found, err := f.GetMetadata()
	if err != nil {
		return false, err
	} else if !found {
		return reason("No record metadata")
	}

	fileMeta, err := f.NewMetadata()
	if err != nil {
		return false, err
	} else if fileMeta == nil {
		return reason("No file metadata")
	}

	if !storedMeta.Equal(fileMeta) {
		return reason("record metadata != file metadata")
	}

	if depth > 0 {

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
		depth--
		changed, err := f.PrerequisiteFiles(IFCHANGE, AUTO_IFCHANGE)
		if err != nil {
			return false, err
		}

		for _, prerequisite := range changed {
			if isCurrent, err := prerequisite.isCurrent(depth); err != nil || !isCurrent {
				return isCurrent, err
			}
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
	if Verbosity > 2 {
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
