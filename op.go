package redo

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

	cachedMetadata, recordFound, err := target.GetMetadata()
	if err != nil {
		return err
	}

	targetMetadata, targetExists, err := target.NewMetadata()
	if err != nil {
		return err
	}

	if targetExists {
		if recordFound {
			if target.HasDoFile() {
				return target.redoTarget(doFilesNotFound, IFCHANGE)
			} else if cachedMetadata.HasDoFile() {
				return target.Errorf("Missing .do file")
			} else if cachedMetadata != targetMetadata {
				return target.redoStatic(IFCHANGE)
			}
		} else {
			if target.HasDoFile() {
				return target.redoTarget(doFilesNotFound, IFCHANGE, IFCREATE)
			} else {
				return target.redoStatic(IFCREATE)
			}
		}
	} else {
		if recordFound {
			// The file existed at one point but was deleted...
			if target.HasDoFile() {
				return target.redoTarget(doFilesNotFound, IFCHANGE, IFCREATE)
			} else if cachedMetadata.HasDoFile() {
				return target.Errorf("Missing .do file")
			} else { // target is a deleted source file
				if err = target.NotifyDependents(IFCHANGE); err != nil {
					return err
				} else if err = target.DeleteMetadata(); err != nil {
					return err
				}
				return fmt.Errorf("Source file [%s] does not exist", target.Target)
			}
		} else {
			if target.HasDoFile() {
				return target.redoTarget(doFilesNotFound, IFCHANGE, IFCREATE)
			} else {
				return target.Errorf(".do file not found")
			}
		}
	}

	return nil
}

// redoTarget records a target's .do file dependencies, runs the target's do file and notifies dependents.
func (f *File) redoTarget(doFilesNotFound []string, events ...Event) error {

	if f.HasNullDb() {
		return f.ErrUninitialized()
	}

	// auto dependencies on .do files
	if err := f.DeleteDoPrerequisites(); err != nil {
		return err
	}

	for _, path := range doFilesNotFound {
		if relpath, err := filepath.Rel(f.RootDir, path); err != nil {
			return err
		} else if err := f.PutPrerequisite(AUTO_IFCREATE, MakeHash(relpath), Prerequisite{Path: relpath}); err != nil {
			return err
		}
	}

	doFile, err := NewFile(f.DoFile)
	if err != nil {
		return err
	}

	// metadata needs to be stored twice.
	doMetadata, found, err := doFile.NewMetadata()

	if err != nil {
		return err
	} else if !found {
		return doFile.Errorf("Cannot create metadata")
	} else if err := doFile.PutMetadata(&doMetadata); err != nil {
		return err
	}

	doPrerequisite := doFile.AsPrerequisiteMetadata(doMetadata)

	if err := f.PutPrerequisite(AUTO_IFCHANGE, doFile.PathHash, doPrerequisite); err != nil {
		return err
	}

	actions := []func() error{f.RunDoFile}

	// A task script's output is not saved so there's no metadata to store.
	if !f.IsTask() {
		actions = append(actions, func() error { return f.PutMetadata(nil) })
	}

	actions = append(actions, f.DeleteMustRebuild)

	// capture event values individually into sequential calls.
	for _, event := range events {
		actions = append(actions, func(event Event) func() error {
			return func() error {
				return f.NotifyDependents(event)
			}
		}(event))
	}

	for _, fn := range actions {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// redoStatic tracks changes and dependencies for static files, which are edited manually and do not have a do script.
func (f *File) redoStatic(event Event) error {

	// A file with a NullDb exists outside this (or any) redo project directory
	// and has no database in which to store metadata or dependencies.
	// Such a file is still useful it can serve as a prerequisite for files in the redo project directory..
	if f.HasNullDb() {
		return nil
	}

	if err := f.PutMetadata(nil); err != nil {
		return err
	}

	return f.NotifyDependents(event)
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

	verbosity := verbose.Value

	redoDepth := os.Getenv("REDO_DEPTH")

	if verbosity > 1 {
		if redoParent := os.Getenv(REDO_PARENT_ENV_NAME); redoParent != "" && verbosity > 2 {
			fmt.Fprintf(os.Stderr, "%s%s => %s\n", redoDepth, redoParent, target.Path)
		}
		fmt.Fprintf(os.Stderr, "%sredo %s\n", redoDepth, target.Name)
	}

	args := []string{"-e"}

	if verbosity > 2 {
		args = append(args, "-v")
	}

	if Trace() {
		args = append(args, "-x")
	}

	args = append(args, target.DoFile, target.Path, target.Basename, outputs[1].Name())

	const shell = "/bin/sh"
	cmd := exec.Command(shell, args...)
	cmd.Dir = filepath.Dir(target.DoFile)
	cmd.Stdout = outputs[0]
	cmd.Stderr = os.Stderr

	// Add environment variables, replacing existing entries if necessary.
	cmdEnv := os.Environ()

	env := map[string]string{REDO_PARENT_ENV_NAME: target.Path, "REDO_DEPTH": redoDepth + " "}

	if verbosity > 0 {
		env["REDO_VERBOSE"] = strings.Repeat("x", verbosity)
	}

	if Trace() {
		env["REDO_TRACE"] = "1"
	}

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
		if verbosity > 0 {
			return target.Errorf("[%s %s]: %s", shell, strings.Join(args, " "), err)
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
		if finfo, err := f.Stat(); err != nil {
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

	recordRelation := func(m Metadata) error {
		p := target.AsPrerequisiteMetadata(m)
		return RecordRelation(dependent, target, IFCHANGE, p)
	}

	isCurrent, err := target.IsCurrent()
	if err != nil {
		return err
	} else if isCurrent {
		targetMetadata, exists, err := target.NewMetadata()
		if err != nil {
			return err
		} else if exists {

			// dependent's version of the target's state.
			prereq, found, err := dependent.GetPrerequisite(IFCHANGE, target.PathHash)
			if err != nil {
				return err
			} else if found {
				if prereq.Metadata.Equal(targetMetadata) {
					// target is up to date and its current state agrees with dependent's version.
					// Nothing to do here.
					return nil
				}
			} else {
				// There is no record of the dependency so this is the first time through.
				// Since the target is up to date, use its metadata for the dependency.
				return recordRelation(targetMetadata)
			}
		} else {
			/*
				Technically, this branch should be an error: a target just deemed to be current should not
				subsequently fail to exist. However, it is certainly possible for a file to be deleted
				between the two actions. Fortuitously, the file will be recreated, if possible.
			*/
		}
	}

	if err := target.Redo(); err != nil {
		return err
	}

	if targetMetadata, found, err := target.NewMetadata(); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("Cannot find recently created target: %s", target.Target)
	} else {
		return recordRelation(targetMetadata)
	}
}

/* RedoIfCreate records a dependency record on a file that does not yet exist */
func (target *File) RedoIfCreate(dependent *File) error {
	if exists, err := target.Exists(); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("[%s] redo-ifcreate %s. Target exists", dependent.Target, target.Target)
	}

	//In case it existed before
	target.DeleteMetadata()

	return RecordRelation(dependent, target, IFCREATE, target.AsPrerequisite())
}


// InitDir initializes a redo directory in the specified project root directory, creating it if necessary.
func InitDir(dirname string) error {

	if len(dirname) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dirname = wd
	} else if c := dirname[0]; c != '.' && c != '/' {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dirname = path.Join(wd, dirname)
	}

	return os.MkdirAll(path.Join(dirname, REDO_DIR), DIR_PERM)
}
