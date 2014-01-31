package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const docPkg = "github.com/gyepisam/redux"

var cmdInstall = &Command{
	UsageLine: "redux install [OPTIONS] [component, ...]",
	Short:     "Installs one or more components",
	Long: `The install command is an admin command used for post installation.
The command

    redux install links                   -- installs links to main binary
    redux install [--mandir MANDIR] man   -- installs man pages
    redux install                         -- installs all components 

links are installed in $BINDIR, if specified, or the same directory as the executable.

man pages are installed in the --mandir directory, $MANDIR, the first directory in $MANPATH,
the first directory in manpath's output, or /usr/local/man.
`,
}

var dryRun bool
var verbose bool
var manDir string

func init() {
	// break loop
	cmdInstall.Run = runInstall

	flg := flag.NewFlagSet("install", flag.ContinueOnError)
	flg.BoolVar(&dryRun, "n", false, "Dry run. Show actions without running them.")
	flg.BoolVar(&verbose, "v", false, "Be verbose. Show actions while running them.")
	flg.StringVar(&manDir, "mandir", "", "man page installation directory.")
	cmdInstall.Flag = flg
}

type runner struct {
	Name string
	Func func() error
}

var runners = []runner{
	{Name: "links", Func: installLinks},
	{Name: "man", Func: installMan},
}

func runInstall(args []string) error {

	var todo []runner

	if len(args) == 0 {
		todo = runners
	} else {
	TOP:
		for _, arg := range args {
			for _, r := range runners {
				if arg == r.Name {
					todo = append(todo, r)
					continue TOP
				}
			}
			return fmt.Errorf("unknown install target: %s", arg)
		}
	}

	for _, r := range todo {
		err := r.Func()
		if err != nil {
			return fmt.Errorf("error installing %s. %s", r.Name, err)
		}
	}

	return nil
}

// return fullpath to executable file.
func absExePath() (name string, err error) {
	name = os.Args[0]

	if name[0] == '.' {
		name, err = filepath.Abs(name)
		if err == nil {
			name = filepath.Clean(name)
		}
	} else {
		name, err = exec.LookPath(filepath.Clean(name))
	}
	return
}

func installLinks() (err error) {
	// link target must be absolute, no matter how it was invoked
	oldname, err := absExePath()
	if err != nil {
		return
	}

	var dirname string
	if s := os.Getenv("BINDIR"); s != "" {
		dirname = s
	} else {
		dirname = filepath.Dir(oldname)
	}

	for _, cmd := range commands {
		if cmd.LinkName == "" {
			continue
		}

		newname := filepath.Join(dirname, cmd.LinkName)
		if oldname == newname {
			continue
		}

		if dryRun || verbose {
			fmt.Fprintf(os.Stderr, "ln %s %s\n", oldname, newname)
			if dryRun {
				continue
			}
		}

		err = os.MkdirAll(filepath.Dir(newname), 0755)
		if err != nil {
			break
		}

		err = os.Remove(newname)
		if err != nil && !os.IsNotExist(err) {
			break
		}

		err = os.Link(oldname, newname)
		if err != nil {
			break
		}
	}

	return
}

func findPackageDir(pkgName string) (string, error) {
	pkg, err := build.Import(pkgName, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	if pkg.Dir == "" {
		return "", fmt.Errorf("no directory for package: %s", pkgName)
	}
	return pkg.Dir, nil
}

/*

 0 use --mandir flag value if it exists
 1 use MANDIR if it is set
 2 use first non-empty path in MANPATH if it is set.
   This is equivalent to: $(echo $MANPATH | cut -f1 -d:) except that we skip empty fields.
 3 use $(dirname redux)/../man if it is writable
 4 use first non-empty path in `manpath` if there's one.
   This is equivalent to: $(manpath 2>/dev/null | cut -f1 -d:) except that we skip empty fields.
 5 use '/usr/local/man'

See https://github.com/gyepisam/redux/issues/4 for details.

*/
func findManDir() (string, error) {

	if manDir != "" {
		return manDir, nil
	}

	if s := os.Getenv("MANDIR"); s != "" {
		return s, nil
	}

	if s := os.Getenv("MANPATH"); s != "" {
		// MANPATH values can have a colon at the start,  at the end, two in the middle or none at all.
		// Find and return the first non-empty path in the list.
		paths := strings.Split(s, ":")
		for _, path := range paths {
			if len(path) > 0 {
				return path, nil
			}
		}
	}

	// The go/.../man directory may not exist and if it does, may not be writable.
	// Test both before commiting to using it.
	path, err := func() (dirname string, err error) {
		path, err := absExePath()
		if err != nil {
			return
		}

		dir := filepath.Clean(filepath.Join(filepath.Dir(path), "..", "man"))

		// This is more of a club than a scalpel, but is a more effective and safer alternative
		// to access(2) which isn't available in Go anyway.
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return
		}

		// See if we can write to it
		tmpDir, err := ioutil.TempDir(dir, "redux-install-man-test-")
		if err != nil {
			return
		}

		os.Remove(tmpDir)

		return dir, nil
	}()

	if err == nil {
		return path, nil
	}

	cmd := exec.Command("manpath")
	b, err := cmd.Output()
	// Ignore error; either it doesn't exist or is somehow broken.
	if err == nil {
		// Find and return the first non-empty path in the list
		paths := bytes.Split(b, []byte{':'})
		for _, path := range paths {
			if len(path) > 0 {
				return string(path), nil
			}
		}
	}

	return "/usr/local/man", nil
}

func copyFile(dst, src string, perm os.FileMode) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, b, perm)
}

func installMan() (err error) {
	pkgDir, err := findPackageDir(docPkg)
	if err != nil {
		return
	}

	if err != nil {
		return
	}

	manDir, err := findManDir()
	if err != nil {
		return
	}

	for _, section := range strings.Split("1", "") {
		srcFiles, err := filepath.Glob(path.Join(pkgDir, "doc", "*."+section))
		if err != nil {
			return err
		}

		for _, src := range srcFiles {
			srcInfo, err := os.Stat(src)
			if err != nil {
				return err
			}

			dstDir := path.Join(manDir, "man"+section)
			dst := path.Join(dstDir, srcInfo.Name())
			if dryRun || verbose {
				fmt.Fprintf(os.Stderr, "cp %s %s\n", src, dst)
				if dryRun {
					continue
				}
			}

			err = os.MkdirAll(dstDir, 0755)
			if err != nil {
				return err
			}

			err = copyFile(dst, src, srcInfo.Mode()&os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
