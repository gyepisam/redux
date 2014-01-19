// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redo

import (
	"encoding/json"
	"path/filepath"
)

func decodePrerequisite(b []byte) (Prerequisite, error) {
	var p Prerequisite
	return p, json.Unmarshal(b, &p)
}

func decodeDependent(b []byte) (Dependent, error) {
	var d Dependent
	return d, json.Unmarshal(b, &d)
}

func (f *File) AsDependent(dir string) Dependent {
	relpath, err := filepath.Rel(dir, f.Fullpath())
	if err != nil {
		panic(err)
	}
	return Dependent{Path: relpath}
}

func (f *File) AsPrerequisite(dir string, m *Metadata) Prerequisite {
	relpath, err := filepath.Rel(dir, f.Fullpath())
	if err != nil {
		panic(err)
	}
	return Prerequisite{Path: relpath, Metadata: m}
}

// Get returns a database record decoded into the specified type.
func (f *File) Get(key string, obj interface{}) (bool, error) {
	data, found, err := f.db.Get(key)
	defer f.Debug("@Get %s %s %s -> %t ...\n", f.Name, f.Path, key, found)
	if err == nil && found {
		err = json.Unmarshal(data, &obj)
	}
	return found, err
}

// Put stores a database record.
func (f *File) Put(key string, obj interface{}) (err error) {
	defer f.Debug("@Put %s %s %s -> %s\n", f.Name, f.Path, key, err)
	if b, err := json.Marshal(obj); err != nil {
		return err
	} else {
		return f.db.Put(key, b)
	}
}
