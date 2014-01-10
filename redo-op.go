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
				return target.redoTarget(IFCHANGE, doFilesNotFound)
			} else if cachedMetadata.HasDoFile() {
				return target.Errorf("Missing .do file")
			} else if cachedMetadata != targetMetadata {
				return target.redoStatic(IFCHANGE)
			}
		} else {
			if target.HasDoFile() {
				return target.redoTarget(IFCREATE, doFilesNotFound)
			} else {
				return target.redoStatic(IFCREATE)
			}
		}
	} else {
		if recordFound {
			// The file existed at one point but was deleted...
			_ = target.NotifyDependents(IFCREATE)

			if target.HasDoFile() {
				return target.redoTarget(IFCHANGE, doFilesNotFound)
			} else if cachedMetadata.HasDoFile() {
				return target.Errorf("Missing .do file")
			} else { // target is a deleted source file
				if err = target.NotifyDependents(IFCHANGE); err != nil {
					return err
				} else if err = target.DeleteMetadata(); err != nil {
					return err
				}
				return target.Errorf("Static file does not exist")
			}
		} else {
			if target.HasDoFile() {
				return target.redoTarget(IFCREATE, doFilesNotFound)
			} else {
				return target.Errorf(".do file not found")
			}
		}
	}

	return nil
}

// redoTarget records a target's do file dependencies and runs the target's do file
func (f *File) redoTarget(event Event, doFilesNotFound []string) error {

   // auto dependencies on .do files
	if err := f.DeleteDoPrerequisites(); err != nil {
		return err
	}

	for _, file := range doFilesNotFound {
		if err := f.PutDoPrerequisite(IFCREATE, file); err != nil {
			return err
		}
	}

	if doFile, err := NewFile(f.DoFile); err != nil {
		return err
	} else if err = doFile.PutMetadata(); err != nil {
		return err
	}

	if err := f.PutDoPrerequisite(IFCHANGE, f.DoFile); err != nil {
		return err
	}

	notifyDependents :=  func() error {
		return f.NotifyDependents(event)
	}

	actions := []func() error{f.RunDoFile, f.PutMetadata, f.DeleteMustRebuild, notifyDependents}

	for _, fn := range actions {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// redoStatic tracks changes and dependencies for static files, which are edited manually and do not have a do script.
func (f *File) redoStatic(event Event) error {
	if err := f.PutMetadata(); err != nil {
		return err
	}
	return f.NotifyDependents(event)
}

// FindDoFile returns the path to the most specific .do file for the target
// as well as an array of paths to more specific do files, if any, that were not found.
func (f *File) findDoFile() (doFilesNotFound []string, err error) {

	// First, try to find a target specific do file in the same directory
	f.DoFile, doFilesNotFound, err = findDoFile(f.Dir, f.Basename)
	if err != nil || f.HasDoFile() { // error or found do file.
		return
	}

	// No target specific do file found.
	// Search for defaults in target's directory or above.
	var defaults []string

	// Prefer default with same extension as target.
	// Ensure that a target lacking an extension does not
	// cause two searches for a non-existent do file.
	if len(f.Ext) > 0 {
		defaults = append(defaults, "default"+f.Ext)
	}

	defaults = append(defaults, "default")

	var defaultsNotFound []string

	for dir := f.Dir; ; /* no test */ dir = filepath.Dir(dir) {
		f.DoFile, defaultsNotFound, err = findDoFile(dir, defaults...)

		if err != nil {
			return
		}

		if len(defaultsNotFound) > 0 {
			doFilesNotFound = append(doFilesNotFound, defaultsNotFound...)
		}

		if f.HasDoFile() {
			break
		}

		if dir == f.RootDir {
			break
		}
	}

	return
}

// findDoFile searches for each of a list of do files in the specified directory
// and returns a string denoting the first existing file found and an array
// of any files that were not found.
func findDoFile(dir string, names ...string) (fileFound string, filesNotFound []string, err error) {
	for _, name := range names {
		path := filepath.Join(dir, name+".do")
		if exists, lerr := fileutils.FileExists(path); lerr != nil {
			err = lerr
			return
		} else if exists {
			fileFound = path
			return
		} else {
			filesNotFound = append(filesNotFound, path)
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
	// and the to a file that will subsequently be deleted.
	for i := 0; i < len(outputs); i++ {
		if i == 0 && target.IsTask() {
			outputs[i] = os.Stdout
		} else {
			outputs[i], err = ioutil.TempFile(target.Dir, target.Basename+"-redo-tmp-")
			if err != nil {
				return
			}
		}
	}

	//TODO -- add options to Config data structure and set them in main.go
	redoDepth := os.Getenv("REDO_DEPTH")
	verbose := len(os.Getenv("REDO_VERBOSE"))
	trace := os.Getenv("REDO_TRACE")

	if verbose > 0 {
		if redoParent := os.Getenv(REDO_PARENT_ENV_NAME); redoParent != "" && verbose > 1 {
			fmt.Fprintf(os.Stderr, "%s%s => %s\n", redoDepth, redoParent, target.Path)
		}
		fmt.Fprintf(os.Stderr, "%sredo %s\n", redoDepth, target.Name)
	}

	env := os.Environ()

	for k, v := range map[string]string{REDO_PARENT_ENV_NAME: target.Path, "REDO_DEPTH": redoDepth + " "} {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	const shell = "/bin/sh"
	args := []string{"-e", target.DoFile, target.Path, target.Basename, outputs[1].Name()}

	if verbose > 0 {
		args[0] += "v"
	}

	if trace != "" {
		args[0] += "x"
	}

	cmd := exec.Command(shell, args...)
	cmd.Dir = filepath.Dir(target.DoFile)
	cmd.Env = env
	cmd.Stdout = outputs[0]
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return target.Errorf("Error: command [%s %s] failed with error: %s", shell, strings.Join(args, " "), err)
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
			return target.Errorf("Error: Task do file %s unexpectedly wrote to $3", target.DoFile)
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
