package redux

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const shell = "/bin/sh"

// RunDoFile executes the do file script, records the metadata for the resulting output, then
// saves the resulting output to the target file, if applicable.
func (target *File) RunDoFile() (err error) {
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
			outputs[i] = &Output{os.Stdout, false}
		} else {
			outputs[i], err = target.NewOutput(i == 1)
			if err != nil {
				return err
			}
			defer func(f *Output) {
				f.Close()
				os.Remove(f.Name())
			}(outputs[i])
		}
	}

	err = target.runCmd(outputs)
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
	//  In the correct case where one file has content and the other is empty,
	//  the former is chosen and the latter is deleted.
	//  If both are empty, the first one is chosen and the second deleted.
	//  If both are non-empty, an error is reported and both are deleted.

	// Default to the first one in case both are empty.
	out := outputs[0]

	// number of files written to
	outCount := 0

	for _, f := range outputs {
		size, err := f.Size()
		if err != nil {
			return err
		}

		if size > 0 {
			outCount++
			out = f
		}

		f.Close()
	}

	// It is an error to write to both files.
	if outCount == len(outputs) {
		return target.Errorf(".do file %s wrote to stdout and to file $3", target.DoFile)
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

func (target *File) runCmd(outputs [2]*Output) error {

	args := []string{"-e"}

	if ShellArgs != "" {
		if ShellArgs[0] != '-' {
			ShellArgs = "-" + ShellArgs
		}
		args = append(args, ShellArgs)
	}

	args = append(args, target.DoFile, target.Path, target.Basename, outputs[1].Name())

	cmd := exec.Command(shell, args...)
	cmd.Dir = filepath.Dir(target.DoFile)
	cmd.Stdout = outputs[0]
	cmd.Stderr = os.Stderr

	depth := os.Getenv("REDO_DEPTH")
    parent := os.Getenv(REDO_PARENT_ENV_NAME)

	// Add environment variables, replacing existing entries if necessary.
	cmdEnv := os.Environ()
	env := map[string]string{
		REDO_PARENT_ENV_NAME: target.Fullpath(),
		"REDO_DEPTH":         depth + " ",
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
		prefix := depth
		if parent != "" {
			prefix += target.Rel(parent) + " => "
		}
		target.Log("%s%s (%s)\n", prefix, target.Rel(target.Fullpath()), target.Rel(target.DoFile))
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
