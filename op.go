// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"fmt"
	"path/filepath"

	"github.com/gyepisam/fileutils"
)

// Redo finds and executes the .do file for the given target.
func (target *File) Redo() error {

	doFilesNotFound, err := target.findDoFile()
	if err != nil {
		return err
	}

	cachedMeta, recordFound, err := target.GetMetadata()
	if err != nil {
		return err
	}

	targetMeta, err := target.NewMetadata()
	if err != nil {
		return err
	}
	targetExists := targetMeta != nil

	if targetExists {
		if recordFound {
			if target.HasDoFile() {
				return target.redoTarget(doFilesNotFound, targetMeta)
			} else if cachedMeta.HasDoFile() {
				return target.Errorf("Missing .do file")
			} else if !targetMeta.Equal(&cachedMeta) {
				return target.redoStatic(IFCHANGE, targetMeta)
			}
		} else {
			if target.HasDoFile() {
				return target.redoTarget(doFilesNotFound, targetMeta)
			} else {
				return target.redoStatic(IFCREATE, targetMeta)
			}
		}
	} else {
		if recordFound {
			// target existed at one point but was deleted...
			if target.HasDoFile() {
				return target.redoTarget(doFilesNotFound, targetMeta)
			} else if cachedMeta.HasDoFile() {
				return target.Errorf("Missing .do file")
			} else {
				// target is a deleted source file. Clean up and fail.
				if err = target.NotifyDependents(IFCHANGE); err != nil {
					return err
				} else if err = target.DeleteMetadata(); err != nil {
					return err
				}
				return fmt.Errorf("Source file %s does not exist", target.Target)
			}
		} else {
			if target.HasDoFile() {
				return target.redoTarget(doFilesNotFound, targetMeta)
			} else {
				return target.Errorf(".do file not found")
			}
		}
	}

	return nil
}

// redoTarget records a target's .do file dependencies, runs the target's do file and notifies dependents.
func (f *File) redoTarget(doFilesNotFound []string, oldMeta *Metadata) error {

	// can't build without a database...
	if f.HasNullDb() {
		return f.ErrUninitialized()
	}

	// Prerequisites will be recreated...
	// Ideally, this could be done within a transaction to allow for rollback
	// in the event of failure.
	if err := f.DeleteAutoPrerequisites(); err != nil {
		return err
	}

	for _, path := range doFilesNotFound {
		relpath := f.Rel(path)
		err := f.PutPrerequisite(AUTO_IFCREATE, MakeHash(relpath), Prerequisite{Path: relpath})
		if err != nil {
			return err
		}
	}

	doFile, err := NewFile(f.RootDir, f.DoFile)
	if err != nil {
		return err
	}

	// metadata needs to be stored twice and is relatively expensive to acquire.
	doMeta, err := doFile.NewMetadata()

	if err != nil {
		return err
	} else if doMeta == nil {
		return doFile.ErrNotFound("redoTarget: doFile.NewMetadata")
	} else if err := doFile.PutMetadata(doMeta); err != nil {
		return err
	}

	relpath := f.Rel(f.DoFile)
	if err := f.PutPrerequisite(AUTO_IFCHANGE, MakeHash(relpath), Prerequisite{relpath, doMeta}); err != nil {
		return err
	}

	if err := f.RunDoFile(); err != nil {
		return err
	}

	// A task script does not produce output and has no dependencies...
	if f.IsTask() {
		return nil
	}

	newMeta, err := f.NewMetadata()
	if err != nil {
		return err
	} else if newMeta == nil {
		return f.ErrNotFound("redoTarget: f.NewMetadata")
	}

	if err := f.PutMetadata(newMeta); err != nil {
		return err
	}

	if err := f.DeleteMustRebuild(); err != nil {
		return err
	}

	// Notify dependents if a content change has occurred.
	return f.GenerateNotifications(oldMeta, newMeta)
}

// redoStatic tracks changes and dependencies for static files, which are edited manually and do not have a do script.
func (f *File) redoStatic(event Event, oldMeta *Metadata) error {

	// A file that exists outside this (or any) redo project directory
	// and has no database in which to store metadata or dependencies is assigned a NullDb.
	// Such a file is still useful it can serve as a prerequisite for files inside a redo project directory.
	// However, it cannot store metadata or notify dependents of changes.
	if f.HasNullDb() {
		return nil
	}

	newMeta, err := f.NewMetadata()
	if err != nil {
		return err
	} else if newMeta == nil {
		return f.ErrNotFound("redoStatic")
	}

	if err := f.PutMetadata(newMeta); err != nil {
		return err
	}

	return f.GenerateNotifications(oldMeta, newMeta)
}

/* FindDoFile searches for the most specific .do file for the target and, if found, stores its path in f.DoFile.
It returns an array of paths to more specific .do files, if any, that were not found.
Target with extension searches for: target.ext.do, default.ext.do, default.do
 Target without extension searches for: target.do, default.do
*/
func (f *File) findDoFile() (missing []string, err error) {

	var candidates []string

	candidates = append(candidates, f.Name+".do")
	if len(f.Ext) > 0 {
		candidates = append(candidates, "default"+f.Ext+".do")
	}
	candidates = append(candidates, "default.do")

TOP:
	for dir := f.Dir; ; /* no test */ dir = filepath.Dir(dir) {
		for _, candidate := range candidates {
			path := filepath.Join(dir, candidate)
			var exists bool // avoid rescoping err
			exists, err = fileutils.FileExists(path)
			if err != nil {
				break TOP
			} else if exists {
				f.DoFile = path
				break TOP
			} else {
				missing = append(missing, path)
			}
		}
		if dir == f.RootDir {
			break TOP
		}
	}

	return
}

// RedoIfChange runs redo on the target if it is out of date or its current state
// disagrees with its dependent's version of its state.
func (target *File) RedoIfChange(dependent *File) error {

	recordRelation := func(m *Metadata) error {
		return RecordRelation(dependent, target, IFCHANGE, m)
	}

	targetMeta, err := target.NewMetadata()
	if err != nil {
		return err
	}

	// No metadata means the target has not been seen before.
	// Redo will sort that out.
	if targetMeta == nil {
		goto REDO
	}

	if isCurrent, err := target.IsCurrent(); err != nil {
		return err
	} else if !isCurrent {
		goto REDO
	} else {

		// Compare dependent's version of the target's state to its current state.
		// Target is self consistent, but may have changed since the prerequisite record was created.
		prereq, found, err := dependent.GetPrerequisite(IFCHANGE, target.PathHash)
		if err != nil {
			return err
		} else if !found {
			// There is no record of the dependency so this is the first time through.
			// Since the target is up to date, use its metadata for the dependency.
			return recordRelation(targetMeta)
		}

		if prereq.Equal(targetMeta) {
			// target is up to date and its current state agrees with dependent's version.
			// Nothing to do here.
			return nil
		}
	}

REDO:
	err = target.Redo()
	if  err != nil {
		return err
	}

	targetMeta, err = target.NewMetadata()
	if err != nil {
		return err
	}
	if targetMeta == nil {
		return fmt.Errorf("Cannot find recently created target: %s", target.Target)
	}

	return recordRelation(targetMeta)
}

/* RedoIfCreate records a dependency record on a file that does not yet exist */
func (target *File) RedoIfCreate(dependent *File) error {
	if exists, err := target.Exists(); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("%s. File exists", dependent.Target)
	}

	//In case it existed before
	target.DeleteMetadata()

	return RecordRelation(dependent, target, IFCREATE, nil)
}
