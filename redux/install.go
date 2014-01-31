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

    redux install links -- installs links to main binary
    redux install man   -- installs man pages
    redux install       -- installs all components 

links are installed in $BINDIR, if specified, or the same directory as the executable.
man pages are installed in $MANDIR, if specified, the first directory printed by  manpath(1), or /usr/local/man.
It may be necessary to specify either or both environment variables or use the sudo hammer. 
`,
}

var dryRun bool
var verbose bool

func init() {
	// break loop
	cmdInstall.Run = runInstall

	flg := flag.NewFlagSet("install", flag.ContinueOnError)
	flg.BoolVar(&dryRun, "n", false, "Dry run. Show actions without running them.")
	flg.BoolVar(&verbose, "v", false, "Be verbose. Show actions while running them.")
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

func installLinks() (err error) {
	oldname := os.Args[0]

	// link target must be absolute, no matter how it was invoked
	if oldname[0] == '.' {
		oldname, err = filepath.Abs(oldname)
		if err != nil {
			return
		}
		oldname = filepath.Clean(oldname)
	} else {
		oldname, err = exec.LookPath(filepath.Clean(oldname))
	}

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
	pkg, err := build.Import(pkgName, "", 0)
	if err != nil {
		return "", err
	}
	if pkg.Dir == "" {
		return "", fmt.Errorf("no directory for package: %s", pkgName)
	}
	return pkg.Dir, nil
}

func findManDir() (string, error) {
	if s := os.Getenv("MANDIR"); s != "" {
		return s, nil
	}

	cmd := exec.Command("manpath")
	b, err := cmd.Output()
	if err == nil && len(b) > 0 {
		i := bytes.Index(b, []byte{':'})
		if i == 0 {
			return "", fmt.Errorf("malformed manpath: %s", string(b))
		}

		if i > 0 {
			return string(b[0:i]), nil
		}

		return string(b), nil
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

func installMan() error {
	pkgDir, err := findPackageDir(docPkg)
	if err != nil {
		return err
	}
	//FIXME: fine in this context, but should be generalized for other uses.
	srcDir := path.Join(pkgDir, "doc")

	manDir, err := findManDir()
	if err != nil {
		return err
	}

	for _, section := range strings.Split("1", "") {
		pattern := "*." + section
		dirs, err := ioutil.ReadDir(srcDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}

		for _, info := range dirs {
			match, err := path.Match(pattern, info.Name())
			if err != nil {
				return err
			}
			if !match {
				continue
			}

			src := path.Join(srcDir, info.Name())
			dst := path.Join(manDir, "man"+section, info.Name())

			if dryRun || verbose {
				fmt.Fprintf(os.Stderr, "cp %s %s\n", src, dst)
				if dryRun {
					continue
				}
			}

			err = os.MkdirAll(path.Dir(dst), 0755)
			if err != nil {
				return err
			}

			err = copyFile(dst, src, info.Mode().Perm())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
