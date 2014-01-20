// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

// Dependent is the inverse of Prerequisite
type Dependent struct {
	Path string
}

func (d Dependent) File(dir string) (*File, error) {
  f, err := NewFile(dir, d.Path)
  if err != nil {
	return nil, err
  }
  return f, nil
}

func (p Prerequisite) File(dir string) (*File, error) {
  f, err := NewFile(dir, p.Path)
  if err != nil {
	return nil, err
  }
  return f, nil
}


func (f *File) DependentFiles(prefix string) ([]*File, error) {

	data, err := f.db.GetValues(prefix)
	if err != nil {
		return nil, err
	}

	files := make([]*File, len(data))

	for i, b := range data {
		if dep, err := decodeDependent(b); err != nil {
			return nil, err
		} else if item, err := dep.File(f.RootDir); err != nil {
			return nil, err
		} else {
			files[i] = item
		}
	}

	return files, nil
}

func (f *File) AllDependents() ([]*File, error) {
	return f.DependentFiles(f.makeKey(SATISFIES))
}

func (f *File) EventDependents(event Event) ([]*File, error) {
	return f.DependentFiles(f.makeKey(SATISFIES, event))
}

func (f *File) DeleteAllDependencies() (err error) {
	keys, err := f.db.GetKeys(f.makeKey(SATISFIES))
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := f.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) DeleteDependency(event Event, hash Hash) error {
	return f.Delete(f.makeKey(SATISFIES, event, hash))
}

func (f *File) PutDependency(event Event, hash Hash, dep Dependent) error {
	return f.Put(f.makeKey(SATISFIES, event, hash), dep)
}

// NotifyDependents flags dependents as out of date because target has been created, modified,  or deleted.
func (f *File) NotifyDependents(event Event) (err error) {

	dependents, err := f.EventDependents(event)
	if err != nil {
		return err
	}

	for _, dependent := range dependents {
		if err := dependent.PutMustRebuild(); err != nil {
			return err
		}
		f.Debug("@Notify %s %s -> %s\n", event, f.Path, dependent.Path)
	}

	return nil
}
