package redo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type Dir struct {
	root string
	path string
}

func newDir() (Dir, error) {
	s, err := ioutil.TempDir("", "redo-test-")
	return Dir{s, s}, err
}

func DirAt(dir string) (Dir, error) {
	d, err := newDir()
	if err != nil {
		return d, err
	}
	s := filepath.Join(d.path, dir)
	if err := os.Mkdir(s, 0755); err != nil {
		return d, err
	}
	d.path = s
	return d, nil
}

func (d Dir) Init() error {
	state := run(exec.Command("redo-init", d.path))
	if state.Err != nil {
		return state
	}
	return nil
}

func (d Dir) Cleanup() {
	os.RemoveAll(d.root)
}

func (d Dir) Append(s string) string {
	return filepath.Join(d.path, s)
}

type Script struct {
	Name       string
	doName     string
	In         string
	Out        string
	Command    string
	ShouldFail bool
}

var AllCaps Script

func init() {
	s := Script{
		Name: "allcaps",
		In: `When to the sessions of sweet silent thought,
  I summon up remembrance of things past,
  I sigh the lack of many a thing I sought,
  And with old woes new wail my dear time's waste:
  `}

	s.Out = strings.ToUpper(s.In)
	s.Command = fmt.Sprintf("echo -n '%s' | tr a-z A-Z", quote(s.In))
	AllCaps = s
}

func quote(s string) string {
	return strings.Replace(s, "'", "'\\''", -1)
}

func (s Script) Write(dir string) error {
	return ioutil.WriteFile(filepath.Join(dir, s.DoName()), []byte(s.Command), os.ModePerm)
}

func (s Script) DoName() string {
	if len(s.doName) > 0 {
		return s.doName
	}
	return s.Name + ".do"
}

func (s Script) CheckOutput(t *testing.T, projectDir string) {
	want := s.Out

	b, err := ioutil.ReadFile(filepath.Join(projectDir, s.Name))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)

	if want != got {
		t.Errorf("Mismatched content for %s.\nWANT:\n[%s]\nGOT:\n[%s]", s.Name, want, got)
	}
}

func (s Script) Checks(t *testing.T, dir Dir) {
	s.CheckOutput(t, dir.path)
	checkMetadata(t, dir.Append(s.DoName()))
	checkMetadata(t, dir.Append(s.Name))
	checkPrerequisites(t, dir.Append(s.Name), dir.Append(s.DoName()))
}

type Result struct {
	Command string
	Stdout  string
	Stderr  string
	Err     error
}

func (r Result) String() string {
	return fmt.Sprintf(`
Result{
  Command: %s

  Stdout:
  %s
   
  Stderr:
  %s
   
  Error:
  %s
}`, r.Command, r.Stdout, r.Stderr, r.Err)
}

func (r Result) Error() string {
	return r.String()
}

func run(cmd *exec.Cmd) Result {

	so, se := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdout = so
	cmd.Stderr = se

	result := Result{Command: fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))}

	result.Err = cmd.Run()
	result.Stdout = so.String()
	result.Stderr = se.String()

	return result
}

func (dir Dir) Command(t *testing.T, target Script, scripts ...Script) *exec.Cmd {

	err := dir.Init()
	if err != nil {
		t.Fatal(err)
	}

	err = target.Write(dir.path)
	if err != nil {
		t.Fatal(err)
	}

	for _, script := range scripts {
		err := script.Write(dir.path)
		if err != nil {
			t.Fatal(err)
		}
	}

	cmd := exec.Command("redo", target.Name)
	cmd.Dir = dir.path

	return cmd
}

func (dir Dir) Run(t *testing.T, target Script, scripts ...Script) Result {
	return run(dir.Command(t, target, scripts...))
}

func init() {
	base, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := os.Getenv("PATH")
	if err := os.Setenv("PATH", fmt.Sprintf("%s/%s:%s", base, "bin", path)); err != nil {
		panic(err)
	}
}

func checkFileMetadata(t *testing.T, path string, m0 Metadata) {
	f, err := NewFile(path)
	if err != nil {
		t.Fatal(err)
	}

	m1, found, err := f.NewMetadata()
	if err != nil {
		t.Fatal(err)
	} else if !found {
		t.Fatalf("Missing file (for metadata): " + path)
	}

	if !m1.Equal(m0) {
		t.Errorf("mismatched record and file metadata for %s", path)
	}
}

func checkMetadata(t *testing.T, path string) {
	f, err := NewFile(path)
	if err != nil {
		t.Fatal(err)
	}

	m0, found, err := f.GetMetadata()
	if err != nil {
		t.Fatal(err)
	} else if !found {
		t.Fatalf("Missing record metadata for: " + path)
	}

	checkFileMetadata(t, path, m0)
}

func checkPrerequisites(t *testing.T, source string, prerequisites ...string) {
	f, err := NewFile(source)
	if err != nil {
		t.Fatal(err)
	}

	list, err := f.Prerequisites()
	if err != nil {
		t.Fatal(err)
	}

	paths := make(map[string]bool)

	for _, p := range list {
		paths[p.Path] = true
	}

	for _, p := range prerequisites {
		_, ok := paths[p]
		if !ok {
			t.Errorf("File %s is missing prerequisite %s", source, p)
		}
	}
}

// Arguments should make no difference
func TestRedoInvocation(t *testing.T) {

	testCases := []struct {
		Name string
		Dir  string
		Path string
	}{
		{"pwd_file", "$root", "$target"},
		{"pwd_relative_file", "$root", "./$target"},
		{"pwd_relative_path", "$dir", "$base/$target"},
		{"relative_path", "$dir", "./$base/$target"},
		{"indir_fullpath", "$root", "$root/$target"},
		{"outdir_fullpath", "/tmp", "$root/$target"},
	}

	s := AllCaps

	for _, v := range testCases {

		dir, err := DirAt(v.Name)
		if err != nil {
			t.Fatal(err)
		}

		expandfn := func(name string) string {
			switch name {
			case "root":
				return dir.path
			case "target":
				return s.Name
			case "dir":
				return filepath.Dir(dir.path)
			case "base":
				return filepath.Base(dir.path)
			}
			panic("Unknown expansion variable: " + name)
		}

		cmd := dir.Command(t, s)
		cmd.Dir = os.Expand(v.Dir, expandfn)
		cmd.Args[1] = os.Expand(v.Path, expandfn)

		t.Logf("cd %s && redo %s\n", cmd.Dir, cmd.Args[1])

		result := run(cmd)
		if result.Err != nil {
			t.Errorf("%s: %s\n", v.Name, result)
			continue
		}

		s.Checks(t, dir)
		dir.Cleanup()
	}
}

// A well behaved .do script does not write to both outputs.
func TestMultipleWrites(t *testing.T) {
	dir, err := newDir()
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Cleanup()

	s := Script{
		Name: "multiple_writes",
		Command: `
echo -n "writes to stdout"
echo -n "writes to file too!" > $3
`}
	result := dir.Run(t, s)

	pattern := "Error.+wrote to stdout and to file"

	matched, err := regexp.Match(pattern, []byte(result.Stderr))
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("Expected pattern [%s] to match stderr message: [%s]", pattern, result.Stderr)
	}
}
