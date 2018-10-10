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
// The execution is equivalent to:
// sh target.ext.do target.ext target outfn > out0
// A well behaved .do file writes to stdout (out0) or to the $3 file (outfn), but not both.
func (target *File) RunDoFile(doInfo *DoInfo) (err error) {

	// out0 is an open file connected to subprocess stdout
	// However, a task subprocess, meaning it is run for side effects,
	// emits output to stdout.
	var out0 *Output

	if target.IsTask() {
		out0 = NewOutput(os.Stdout)
	} else {
		out0, err = target.NewOutput()
		if err != nil {
			return
		}
		defer out0.Cleanup()
	}

	// outfn is the arg3 filename argument to the do script.
	var outfn *Output

	outfn, err = target.NewOutput()
	if err != nil {
		return
	}
	defer outfn.Cleanup()

	// Arg3 file should not exist prior to script execution
	// so its subsequent existence can be significant.
	if err = outfn.SetupArg3(); err != nil {
		return
	}

	if err = target.runCmd(out0.File, outfn.Name(), doInfo); err != nil {
		return
	}

	file, err := os.Open(outfn.Name())
	if err != nil {
		if os.IsNotExist(err) {
			if target.IsTask() {
				return nil
			}
		} else {
			return
		}
	}

	if target.IsTask() {
		// Task files should not write to the temp file.
		return target.Errorf("Task do file %s unexpectedly wrote to $3", target.DoFile)
	}

	if err = out0.Close(); err != nil {
		return
	}

	outputs := make([]*Output, 0)

	finfo, err := os.Stat(out0.Name())
	if err != nil {
		return
	}
	if finfo.Size() > 0 {
		outputs = append(outputs, out0)
	}

	if file != nil {
		outfn.File = file // for consistency
		if err = outfn.Close(); err != nil {
			return
		}
		outputs = append(outputs, outfn)
	}

	if n := len(outputs); n == 0 {
		return target.Errorf("Do file %s generated no output or file activity", target.DoFile)
	} else if n == 2 {
		return target.Errorf("Do file %s wrote to stdout and to file $3", target.DoFile)
	}

	out := outputs[0]
	err = os.Rename(out.Name(), target.Fullpath())
	if err != nil && strings.Index(err.Error(), "cross-device") > -1 {

		// The rename failed due to a cross-device error because the output file
		// tmp dir is on a different device from the target file.
		// Copy the tmp file across the device to the target directory and try again.
		var path string
		path, err = out.Copy(target.Dir)
		if err != nil {
			return
		}

		err = os.Rename(path, target.Fullpath())
		if err != nil {
			_ = os.Remove(path)
		}
	}

	return
}

func (target *File) runCmd(out0 *os.File, outfn string, doInfo *DoInfo) error {

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
	args = append(args, doInfo.Name, relTarget, doInfo.RelPath(doInfo.Arg2), outfn)

	target.Debug("@sh %s $3\n", strings.Join(args[0:len(args)-1], " "))

	cmd := exec.Command(shell, args...)
	cmd.Dir = doInfo.Dir
	cmd.Stdout = out0
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
