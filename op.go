// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"fileutils"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
				return target.Errorf(".do file not found for %", target.Target)
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

	for dir := f.Dir; ; /* no test */ dir = filepath.Dir(dir) {
		for _, candidate := range candidates {
			path := filepath.Join(dir, candidate)
			var exists bool // avoid rescoping err
			exists, err = fileutils.FileExists(path)
			if err != nil {
				return
			} else if exists {
				f.DoFile = path
				return
			} else {
				missing = append(missing, path)
			}
		}
		if dir == f.RootDir {
			return
		}
	}

	return
}

// RunDoFile executes the do file script, records the metadata for the resulting output, then
// saves the resulting output to the target file, if applicable.
func (target *File) RunDoFile() (err error) {
	/*
		A well behaved .do file writes to stdout or to the $3 file, but not both.

		In order to catch misbehaviour, the .do script is run with stdout going to
		a different temp file from $3.

		In the ideal case where one file has content and the other is empty,
		the former is returned while the latter is deleted.
		If both are non-empty, both are deleted and an error reported.
		If both are empty, the first one is returned and the second deleted.
	*/

	var outputs [2]*os.File

	// If the do file is a task, the first output goes to stdout
	// and the second to a file that will be subsequently deleted.
	for i := 0; i < len(outputs); i++ {
		if i == 0 && target.IsTask() {
			outputs[i] = os.Stdout
		} else {
			outputs[i], err = ioutil.TempFile(target.Dir, target.Basename+"-redo-tmp-")
			if err != nil {
				return
			}
			// TODO: add an option to control temp file deletion on failure for debugging purposes?
			defer func(path string) {
				if err != nil {
					os.Remove(path)
				}
			}(outputs[i].Name())
		}
	}

	redoDepth := os.Getenv("REDO_DEPTH")

	if Verbose() {
		prefix := redoDepth
		if redoParent := os.Getenv(REDO_PARENT_ENV_NAME); redoParent != "" {
			prefix += target.Rel(redoParent) + " => "
		}
		target.Log("%s%s (%s)\n", prefix, target.Rel(target.Fullpath()), target.Rel(target.DoFile))
	}

	args := []string{"-e"}

	if ShellArgs != "" {
		args = append(args, ShellArgs)
	}

	args = append(args, target.DoFile, target.Path, target.Basename, outputs[1].Name())

	const shell = "/bin/sh"
	cmd := exec.Command(shell, args...)
	cmd.Dir = filepath.Dir(target.DoFile) //TODO -- run in target directory instead?
	cmd.Stdout = outputs[0]
	cmd.Stderr = os.Stderr

	// Add environment variables, replacing existing entries if necessary.
	cmdEnv := os.Environ()
	env := map[string]string{REDO_PARENT_ENV_NAME: target.Fullpath(), "REDO_DEPTH": redoDepth + " "}

	// Update environment values, if they exist and append when they dont.
TOP:
	for key, value := range env {
		prefix := key + "="
		for i, entry := range cmdEnv {
			if strings.HasPrefix(entry, prefix) {
				cmdEnv[i] = prefix + value
				continue TOP
			}
		}
		cmdEnv = append(cmdEnv, prefix+value)
	}

	cmd.Env = cmdEnv

	if err := cmd.Run(); err != nil {
		if Verbose() {
			return target.Errorf("%s %s: %s", shell, strings.Join(args, " "), err)
		}
		return target.Errorf("%s", err)
	}

	if target.IsTask() {
		// Task files should not write to the temp file.
		f := outputs[1]

		defer func(f *os.File) {
			f.Close()
			os.Remove(f.Name())
		}(f)

		if finfo, err := f.Stat(); err != nil {
			return err
		} else if finfo.Size() > 0 {
			return target.Errorf("Task do file %s unexpectedly wrote to $3", target.DoFile)
		}

		return nil
	}

	writtenTo := 0 // number of files written to
	idx := 0       // index of correct output, with appropriate default.

	for i, f := range outputs {
		// f.Stat() doesn't work for the file on $3 since it was written to by a different process.
		// Rather than using f.Stat() on one and os.Stat() on the other, use the latter on both.
		if finfo, err := os.Stat(f.Name()); err != nil {
			return err
		} else if finfo.Size() > 0 {
			writtenTo++
			idx = i
		}
	}

	// It is an error to write to both files.
	// Select neither so both will be deleted.
	if writtenTo == len(outputs) {
		idx = -1
	}

	for i, f := range outputs {
		f.Close()
		if i != idx {
			os.Remove(f.Name()) // ignored file.
		}
	}

	// and finally, the reckoning
	if writtenTo < len(outputs) {
		return os.Rename(outputs[idx].Name(), target.Fullpath())
	} else {
		return target.Errorf(".do file %s wrote to stdout and to file $3", target.DoFile)
	}
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
	} else if targetMeta == nil {
		goto REDO
	}

	if isCurrent, err := target.IsCurrent(); err != nil {
		return err
	} else if !isCurrent {
		goto REDO
	} else {

		// dependent's version of the target's state.
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
	if err := target.Redo(); err != nil {
		return err
	}

	if targetMeta, err := target.NewMetadata(); err != nil {
		return err
	} else if targetMeta == nil {
		return fmt.Errorf("Cannot find recently created target: %s", target.Target)
	} else {
		return recordRelation(targetMeta)
	}
}

/* RedoIfCreate records a dependency record on a file that does not yet exist */
func (target *File) RedoIfCreate(dependent *File) error {
	if exists, err := target.Exists(); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("%s. File exists", dependent.Target, target.Target)
	}

	//In case it existed before
	target.DeleteMetadata()

	return RecordRelation(dependent, target, IFCREATE, nil)
}


