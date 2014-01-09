package redo

import (
	"errors"
	"fileutils"
	"fmt"
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

	Config Config
	db     DB
}

func (f *File) IsTask() bool {
	return len(f.Basename) > 0 && f.Basename[0] == TASK_PREFIX
}

func (f *File) Errorf(format string, args ...interface{}) error {
  return errors.New(fmt.Sprintf("Error: [Target: %s]: ", f.Target) + fmt.Sprintf(format, args...))
}


func NewFile(targetPath string) (f *File, err error) {

	if targetPath == "" {
		return nil, errors.New("NewFile: target path cannot be empty")
	}

	if isdir, err := fileutils.IsDir(targetPath); err != nil {
		return nil, err
	} else if isdir {
		return nil, fmt.Errorf("NewFile: target %s is a directory", targetPath)
	}

	f = new(File)

	f.Target = targetPath

	path, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, err
	}

	components := []string{filepath.Base(path)}
	rootDir := filepath.Dir(path)

	for {
		exists, err := fileutils.DirExists(filepath.Join(rootDir, REDO_DIR))
		if err != nil {
			return nil, err
		} else if exists {
			break
		} else if rootDir == "/" || rootDir == "." {
			return nil, f.Errorf("cannot find redo root directory")
		} else {
			components = append(components, filepath.Base(rootDir))
		}
		rootDir = filepath.Dir(rootDir)
	}

	f.RootDir = rootDir

	//components are in reverse order...
	for i, j := 0, len(components) - 1; i < j; {
		components[i], components[j] = components[j], components[i]
		i++
		j--
	}
	f.Path = filepath.Join(components...)
	f.PathHash = MakeHash(f.Path)

	f.Config = Config{DBType: "file"}

	// TODO(gsam): Read config file in rootDir to determine DB type, if any.
	// Default to FileDB if not specified.

	if f.Config.DBType == "file" {
		f.db, err = FileDbOpen(f.RootDir)
		if err != nil {
			return nil, err
		}
	}

	f.Dir, f.Name = filepath.Split(f.Fullpath())
	f.Ext = filepath.Ext(f.Name)
	f.Basename = f.Name[:len(f.Name)-len(f.Ext)]


	return
}

func (f *File) Fullpath() string {
	return filepath.Join(f.RootDir, f.Path)
}

func (f *File) Exists() (bool, error) {
	return fileutils.FileExists(f.Fullpath())
}

func (f *File) HasDoFile() bool {
	return len(f.DoFile) > 0
}

func (f *File) relatedFiles(prefix string) ([]*File, error) {

	paths, err := f.db.GetValues(prefix)
	if err != nil {
		return nil, err
	}

	files := make([]*File, len(paths))

	for i, path := range paths {
		if item, err := NewFile(string(path)); err != nil {
			return nil, err
		} else {
			files[i] = item
		}
	}

	return files, nil
}

/*
 IsCurrent returns a boolean denoting whether the target is up to date.

 A target is up to date if the following conditions hold:
   The file exists
   The file has not been flagged to be rebuilt
   The file has not changed since creation. That is; the file has a metadata record
   	and that record matches the actual file metadata.
   All the file's prerequisites are also current.
*/
func (f *File) IsCurrent() (bool, error) {

	if f.MustRebuild() {
		return true, nil
	}

	storedMetadata, found, err := f.GetMetadata()
	if err != nil || !found {
		return found, err
	}

	fileMetadata, found, err := f.NewMetadata()
	if err != nil || !found {
		return found, err
	}

	if storedMetadata != fileMetadata {
		return false, nil
	}

	prerequisites, err := f.PrerequisiteFiles()
	if err != nil {
		return false, err
	}

	for _, prerequisite := range prerequisites {
		if isCurrent, err := prerequisite.IsCurrent(); err != nil || !isCurrent {
			return isCurrent, err
		}
	}

	return true, nil
}

// NewMetadata computes and returns the file metadata.
func (f *File) NewMetadata() (Metadata, bool, error) {
	m, found, err := NewMetadata(f.Fullpath())
	if err == nil && found {
		m.DoFile = f.DoFile
	}
	return m, found, err
}

// ContentHash returns a cryptographic hash of the file contents.
func (f *File) ContentHash() (Hash, error) {
	return ContentHash(f.Fullpath())
}
