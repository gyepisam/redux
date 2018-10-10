package redux

import (
	"fmt"
	"github.com/gyepisam/fileutils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

/*
findDofile searches for the most specific .do file for the target and returns a DoInfo structure.
The structure's Missing field contains paths to more specific .do files, if any, that were not found.
If a file is found the structure's Name and Arg2 fields are also set appropriately.
*/
func (f *File) findDoFile() (*DoInfo, error) {

	relPath := &RelPath{}
	var missing []string

	dir := f.Dir
	candidates := f.DoInfoCandidates()

TOP:
	for {

		for _, do := range candidates {
			path := filepath.Join(dir, do.Name)
			exists, err := fileutils.FileExists(path)
			f.Debug("%s %t %v\n", path, exists, err)
			if err != nil {
				return nil, err
			} else if exists {
				do.Dir = dir
				do.RelDir = relPath.Join()
				do.Missing = missing
				return do, nil
			}

			missing = append(missing, path)
		}

		if dir == f.RootDir {
			break TOP
		}
		relPath.Add(filepath.Base(dir))
		dir = filepath.Dir(dir)
	}

	return &DoInfo{Missing: missing}, nil
}

const shell = "/bin/sh"

// RunDoFile executes the do file script, records the metadata for the resulting output, then
// saves the resulting output to the target file, if applicable.
func (target *File) RunDoFile(doInfo *DoInfo) (err error) {
	/*

			  The execution is equivalent to:

			  sh target.ext.do target.ext target tmp0 > tmp1

			  A well behaved .do file writes to stdout (tmp0) or to the $3 file (tmp1), but not both.

			  We use two temp files so as to detect when the .do script misbehaves,
		      in order to avoid producing incorrect output.
	*/

	var outputs [2]*Output

	// If the do file is a task, the first output goes to stdout
	// and the second to a file that will be subsequently deleted.
	for i := 0; i < len(outputs); i++ {
		if i == 0 && target.IsTask() {
			outputs[i], _ = NewOutput(os.Stdout, false)
			continue
		}

		outputs[i], err = target.NewOutput(i == 1)
		if err != nil {
			return err
		}

		defer func(f *Output) {
			f.Close()           // ignore error
			os.Remove(f.Name()) //ignore error
		}(outputs[i])

		if i == 1 && target.IsTask() {
			outputs[i].Close() // child process will write to it.
		}
	}

	err = target.runCmd(outputs, doInfo)
	if err != nil {
		return err
	}

	if target.IsTask() {
		// Task files should not write to the temp file.
		size, err := outputs[1].Size()
		if err != nil {
			return err
		}

		if size > 0 {
			return target.Errorf("Task do file %s unexpectedly wrote to $3", target.DoFile)
		}

		return nil
	}

	//  Pick an output file...
	//  In the normal case one file has content and the other is empty,
	//  so the former is chosen and the latter is deleted.
	//  If both are none empty but the arg3 file has been
	//  modified, it is chosen. Otherwise an error is reported.
	//  If both are non-empty, an error is reported.

	var out *Output

	// number of files written to
	outCount := 0

	for i, f := range outputs {
		size, err := f.Size()
		if err != nil {
			return err
		}

		if size == 0 {
			if f.IsArg3 {
				modified, err := f.Modified()
				if err != nil {
					return err
				}
				if modified {
					out = f
				}
			}
		} else {
			outCount++
			out = f
		}
		if i == 0 {
			f.Close()
		}
	}

	// It is an error to write to both files.
	if outCount == len(outputs) {
		return target.Errorf(".do file %s wrote to stdout and to file $3", target.DoFile)
	}

	if out == nil {
		return target.Errorf("%s: no output or file activity", target.DoFile)
	}

	err = os.Rename(out.Name(), target.Fullpath())
	if err != nil && strings.Index(err.Error(), "cross-device") > -1 {

		// The rename failed due to a cross-device error because the output file
		// tmp dir is on a different device from the target file.
		// Copy the tmp file across the device to the target directory and try again.
		var path string
		path, err = out.Copy(target.Dir)
		if err != nil {
			return err
		}

		err = os.Rename(path, target.Fullpath())
		if err != nil {
			os.Remove(path)
		}
	}

	return err
}

func (target *File) runCmd(outputs [2]*Output, doInfo *DoInfo) error {

	args := []string{"-e"}

	if ShellArgs != "" {
		if ShellArgs[0] != '-' {
			ShellArgs = "-" + ShellArgs
		}
		args = append(args, ShellArgs)
	}

	pending := os.Getenv("REDO_PENDING")
	pendingID := ";" + string(target.FullPathHash)
	target.Debug("Current: [%s]. Pending: [%s].\n", pendingID, pending)

	if strings.Contains(pending, pendingID) {
		return fmt.Errorf("Loop detected on pending target: %s", target.Target)
	}

	pending += pendingID

	relTarget := doInfo.RelPath(target.Name)
	args = append(args, doInfo.Name, relTarget, doInfo.RelPath(doInfo.Arg2), outputs[1].Name())

	target.Debug("@sh %s $3\n", strings.Join(args[0:len(args)-1], " "))

	cmd := exec.Command(shell, args...)
	cmd.Dir = doInfo.Dir
	cmd.Stdout = outputs[0].File
	cmd.Stderr = os.Stderr

	depth := 0
	if i64, err := strconv.ParseInt(os.Getenv("REDO_DEPTH"), 10, 32); err == nil {
		depth = int(i64)
	}

	parent := os.Getenv("REDO_PARENT")

	// Add environment variables, replacing existing entries if necessary.
	cmdEnv := os.Environ()
	env := map[string]string{
		"REDO_PARENT":  relTarget,
		"REDO_DEPTH":   strconv.Itoa(depth + 1),
		"REDO_PENDING": pending,
	}

	// Update environment values if they exist and append when they dont.
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

	if Verbose() {
		prefix := strings.Repeat(" ", depth)
		if parent != "" {
			prefix += parent + " => "
		}
		target.Log("%s%s (%s)\n", prefix, target.Rel(target.Fullpath()), target.Rel(doInfo.Path()))
	}

	err := cmd.Run()
	if err == nil {
		return nil
	}

	if Verbose() {
		return target.Errorf("%s %s: %s", shell, strings.Join(args, " "), err)
	}

	return target.Errorf("%s", err)
}
