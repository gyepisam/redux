// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redo

import (
	"encoding/json"
)

func decodePrerequisite(b []byte) (Prerequisite, error) {
	var p Prerequisite
	return p, json.Unmarshal(b, &p)
}

func decodeDependent(b []byte) (Dependent, error) {
	var d Dependent
	return d, json.Unmarshal(b, &d)
}

// Get returns a database record decoded into the specified type.
func (f *File) Get(key string, obj interface{}) (bool, error) {
	data, found, err := f.db.Get(key)
	if err == nil && found {
		err = json.Unmarshal(data, &obj)
	}
	return found, err
}

// Put stores a database record.
func (f *File) Put(key string, obj interface{}) (err error) {
	defer f.Debug("@Put %s %s -> %s\n", f.Name, key, err)
	if b, err := json.Marshal(obj); err != nil {
		return err
	} else {
		return f.db.Put(key, b)
	}
}
