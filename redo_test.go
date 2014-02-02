// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type Dir struct {
	root string
	path string
	t    *testing.T
}

func newDir(t *testing.T) (Dir, error) {
	s, err := ioutil.TempDir("", "redo-test-")
	return Dir{s, s, t}, err
}

func newDirAt(t *testing.T, dir string) (Dir, error) {
	d, err := newDir(t)
	if err != nil || len(dir) == 0 {
		return d, err
	}
	s := filepath.Join(d.path, dir)
	if err := os.MkdirAll(s, 0755); err != nil {
		return d, err
	}
	d.path = s
	return d, nil
}

func (dir Dir) newDirAt(s string) (Dir, error) {
	s = filepath.Join(dir.path, s)
	if err := os.MkdirAll(s, 0755); err != nil {
		return dir, err
	}
	return Dir{s, s, dir.t}, nil
}

func (dir Dir) Init() error {
	state := run(dir.t, exec.Command("redo-init", dir.path))
	if state.Err != nil {
		return state
	}
	return nil
}

func (dir Dir) Cleanup() {
	os.RemoveAll(dir.root)
}

func (dir Dir) Append(values ...string) string {
	s := make([]string, len(values)+1)
	s[0] = dir.path
	copy(s[1:], values)
	return filepath.Join(s...)
}

type Script struct {
	Name       string
	In         string
	Out        string
	Command    string
	ShouldFail bool
	OutDir     string
	doFileName string
}

type TestCases map[string]*Script

var Scripts = make(TestCases)

func (c TestCases) Add(name string) *Script {
	if len(name) == 0 {
		panic("TestCase missing name")
	}
	s := &Script{Name: name}
	c[s.Name] = s
	return s
}

func (c TestCases) Get(v string) Script {
	if s, ok := c[v]; ok {
		return *s
	}
	panic("Unknown test case: " + v)
}

func init() {
	var s *Script

	s = Scripts.Add("simple")
	s.In = "slippers,pumps,mules,sneakers,bowling shoes"
	words := strings.Split(s.In, ",")
	sort.Strings(words)
	s.Out = strings.Join(words, ",")
	s.Command = fmt.Sprintf("echo -n '%s'", quote(s.Out))

	s = Scripts.Add("allcaps")
	s.In = `
When to the sessions of sweet silent thought,
I summon up remembrance of things past,
I sigh the lack of many a thing I sought,
And with old woes new wail my dear time's waste:
`
	s.Out = strings.ToUpper(s.In)
	s.Command = fmt.Sprintf("echo -n '%s' | tr a-z A-Z", quote(s.In))

	s = Scripts.Add("fmt.txt")
	s.In = `I returned, and saw under the sun, that the race is not to the
swift, nor the battle to the strong, neither yet bread to the wise,
nor yet riches to men of understanding, nor yet favour to men of
skill; but time and chance happeneth to them all.`
	s.Out = `I returned, and saw under the sun, that the race is not to the
swift, nor
swift, the
swift, battle
swift, to
swift, the
swift, strong,
swift, neither
swift, yet
swift, bread
swift, to
swift, the
swift, wise,
nor yet riches to men of understanding, nor yet favour to men of
skill; but time and chance happeneth to them all.
`
	s.Command = fmt.Sprintf("fmt --width 10 --prefix swift, <<EOS\n%s\nEOS", s.In)

	s = Scripts.Add("uses-default.txt")
	s.In = `
Now is the winter of our discontent
Made glorious summer by this sun of York`
	s.Out = s.In
	s.doFileName = "default.txt.do"
	s.Command = fmt.Sprintf("echo -n '%s'", quote(s.In))

	pears := strings.Split("Keiffer,Bosc,Moonglow,Bartlett,Magness,Seckel,Gorham,Anjou", ",")

	s = Scripts.Add("list")
	s.Out = strings.Join(pears, "\n")
	s.Command = fmt.Sprintf("echo -n '%s'", quote(s.Out))

	s = Scripts.Add("sorted-list")
	sort.Strings(pears)
	s.Out = strings.Join(pears, "\n") + "\n" //external sort adds a newline to every line.
	s.Command = `
redo-ifchange list
sort < list
`

	s = Scripts.Add("default-fail")
	s.doFileName = "default.do"
	s.Command = "false"

	s = Scripts.Add("default-txt-fail")
	s.doFileName = "default.txt.do"
	s.Command = "false"

	s = Scripts.Add("multiple-writes")
	s.Command = `
echo -n "writes to stdout"
echo -n "writes to file too!" > $3
`
}

func quote(s string) string {
	return strings.Replace(s, "'", "'\\''", -1)
}

func (s Script) Write(dir string) error {
	return ioutil.WriteFile(filepath.Join(dir, s.DoFileName()), []byte(s.Command), os.ModePerm)
}

func (s Script) DoFileName() string {
	if len(s.doFileName) > 0 {
		return s.doFileName
	}
	return s.Name + ".do"
}

func (s Script) OutputFileName() string {
	return filepath.Join(s.OutDir, s.Name)
}

func (s Script) CheckOutput(t *testing.T, projectDir string) {
	want := s.Out

	b, err := ioutil.ReadFile(filepath.Join(projectDir, s.OutputFileName()))
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
	checkMetadata(t, dir.Append(s.DoFileName()))
	checkMetadata(t, dir.Append(s.OutputFileName()))
	checkPrerequisites(t, dir.Append(s.OutputFileName()), s.DoFileName())
}

type Result struct {
	Command string
	Stdout  string
	Stderr  string
	Err     error
}

func (r Result) String() string {
	out := make([]string, 3)

	out = append(out, "Result: {\n")
	out = append(out, fmt.Sprintf("Command: %s\n", r.Command))

	if len(r.Stdout) > 0 {
		out = append(out, fmt.Sprintf("Stdout:\n%s\n", r.Stdout))
	}

	if len(r.Stderr) > 0 {
		out = append(out, fmt.Sprintf("Stderr:\n%s\n", r.Stderr))
	}

	if r.Err != nil {
		out = append(out, fmt.Sprintf("Error:\n%s\n", r.Err))
	}

	out = append(out, "}")

	return strings.Join(out, "")
}

func (r Result) Error() string {
	return r.String()
}

func run(t *testing.T, cmd *exec.Cmd) Result {

	so, se := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdout = so
	cmd.Stderr = se

	result := Result{Command: fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))}

	result.Err = cmd.Run()
	result.Stdout = so.String()
	result.Stderr = se.String()

	if testing.Verbose() {
		t.Log(result.String())
	}
	return result
}

func (dir Dir) Command(target Script, scripts ...Script) *exec.Cmd {

	t := dir.t

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

	cmdArgs := []string{}
	if testing.Verbose() {
		cmdArgs = append(cmdArgs, "-verbose")
	}
	cmdArgs = append(cmdArgs, target.Name)
	cmd := exec.Command("redo", cmdArgs...)

	cmd.Dir = dir.path

	return cmd
}

func (dir Dir) Run(target Script, scripts ...Script) Result {
	return run(dir.t, dir.Command(target, scripts...))
}

func checkFileMetadata(t *testing.T, path string, m0 *Metadata) {
	f, err := NewFile("", path)
	if err != nil {
		t.Fatal(err)
	}

	m1, err := f.NewMetadata()
	if err != nil {
		t.Fatal(err)
	} else if m1 == nil {
		t.Fatalf("Missing file (for metadata): " + path)
	}

	if !m1.Equal(m0) {
		t.Errorf("mismatched record and file metadata for %s", path)
	}
}

func checkMetadata(t *testing.T, path string) {
	f, err := NewFile("", path)
	if err != nil {
		t.Fatal(err)
	}

	m0, found, err := f.GetMetadata()
	if err != nil {
		t.Fatal(err)
	} else if !found {
		t.Fatalf("Missing record metadata for: " + path)
	}

	checkFileMetadata(t, path, &m0)
}

func checkPrerequisites(t *testing.T, source string, prerequisites ...string) {
	f, err := NewFile("", source)
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

	for _, path := range prerequisites {
		if _, ok := paths[path]; !ok {
			t.Errorf("File %s is missing prerequisite %s. Looked in %v", source, path, paths)
		}
	}
}

// File paths arguments should resolve correctly
func TestRedoArgs(t *testing.T) {

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

	s := Scripts.Get("allcaps")

	for _, v := range testCases {

		dir, err := newDirAt(t, v.Name)
		if err != nil {
			t.Fatal(err)
		}
		defer dir.Cleanup()

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

		cmd := dir.Command(s)
		cmd.Dir = os.Expand(v.Dir, expandfn)

		// replace the target name arg with test case.
		for i, arg := range cmd.Args[1:] {
			if arg == s.Name {
				cmd.Args[i+1] = os.Expand(v.Path, expandfn)
				break
			}
		}

		t.Logf("cd %s && %s\n", cmd.Dir, strings.Join(cmd.Args, " "))

		result := run(t, cmd)
		if result.Err != nil {
			t.Errorf("%s: %s\n", v.Name, result)
			continue
		}

		s.Checks(t, dir)
	}
}

func CheckMatch(t *testing.T, pattern string, data string) {
	matched, err := regexp.Match(pattern, []byte(data))
	if err != nil {
		t.Fatalf("%s while matching pattern: %s", err, pattern)
	}
	if !matched {
		t.Errorf("Expected pattern [%s] to match message: [%s]", pattern, data)
	}
}

// Stuff that should fail..
func TestFailures(t *testing.T) {
	tests := []struct {
		script  Script
		pattern string
	}{
		// Simple failure
		{Scripts.Get("default-fail"), "exit status 1"},
		// A well behaved .do script does not write to both outputs.
		{Scripts.Get("multiple-writes"), "Error.+wrote to stdout and to file"},
	}

	for _, test := range tests {
		dir, err := newDir(t)
		if err != nil {
			t.Fatal(err)
		}
		defer dir.Cleanup()

		result := dir.Run(test.script)

		if result.Err != nil {
			CheckMatch(t, test.pattern, result.Stderr)
		} else {
			t.Errorf("Expected script %s to fail", test.script.Name)
		}
	}
}

//build scripts in higher level directories should work.
func TestBuildScriptLevel(t *testing.T) {
	caseName := "fmt.txt"
	for _, subdir := range []string{"", "1", "1/2", "1/2/3"} {
		for _, doFile := range []string{"default.do", "default.txt.do", Scripts.Get(caseName).DoFileName()} {
			dir, err := newDir(t)
			if err != nil {
				t.Fatal(err)
			}
			defer dir.Cleanup()

			s := Scripts.Get(caseName)
			s.doFileName = doFile
			s.OutDir = subdir

			cmd := dir.Command(s)
			if len(subdir) > 0 {
				cmd.Dir = dir.Append(subdir)
				err := os.MkdirAll(cmd.Dir, 0777)
				if err != nil {
					t.Fatal(err)
				}
			}

			result := run(t, cmd)
			if result.Err != nil {
				t.Errorf("%s %s %s: %s\n", dir.path, subdir, doFile, result)
				continue
			}

			s.Checks(t, dir)

		}
	}
}

// Should choose more specific build scripts
func TestScriptSelectionOrder(t *testing.T) {

	cases := []struct {
		Pass Script
		Fail Script
	}{
		// target.ext.do over default.ext.do
		{Scripts.Get("fmt.txt"), Scripts.Get("default-txt-fail")},

		// target.ext.do over default.do
		{Scripts.Get("fmt.txt"), Scripts.Get("default-fail")},

		// target.do over default.do
		{Scripts.Get("allcaps"), Scripts.Get("default-fail")},

		// default.ext.do over default.do
		{Scripts.Get("uses-default.txt"), Scripts.Get("default-fail")},
	}

	for _, pair := range cases {
		dir, err := newDir(t)
		if err != nil {
			t.Fatal(err)
		}
		defer dir.Cleanup()

		result := dir.Run(pair.Pass, pair.Fail)
		if result.Err != nil {
			t.Errorf("%s: %s\n", dir.path, result)
			continue
		}

		pair.Pass.Checks(t, dir)
		// No need to check for failing script.
		// If it ran, Pass would fail!

	}
}

// Setup scripts and invoke the first one, which should produce expected output.
func SimpleTree(t *testing.T, scripts ...Script) {
	if len(scripts) == 0 {
		panic("SimpleTree requires at least one script argument")
	}

	dir, err := newDir(t)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Cleanup()
	first := scripts[0]
	var rest []Script
	if len(scripts) > 1 {
		rest = scripts[1:]
	}

	if result := dir.Run(first, rest...); result.Err != nil {
		t.Fatal(result)
	}
	first.Checks(t, dir)
}

// Basic Dependency -- redo-ifchange
func TestBasicDependency(t *testing.T) {
	dir, err := newDir(t)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Cleanup()

	sorted := Scripts.Get("sorted-list")
	list := Scripts.Get("list")

	for i := 0; i < 2; i++ {
		result := dir.Run(sorted, list)
		if result.Err != nil {
			t.Errorf("%s: %s\n", dir.path, result)
			break
		}

		sorted.Checks(t, dir)
		list.Checks(t, dir)

		// force source rebuilding
		if err := ioutil.WriteFile(dir.Append(list.OutputFileName()), []byte("Break checksum and timestamp!"), 0655); err != nil {
			t.Fatal(err)
		}
	}
}