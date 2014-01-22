// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

// MustRebuild returns a boolean denoting whether the target must be rebuilt.
func (f *File) MustRebuild() bool {
	var x string
	found, err := f.Get(f.mustRebuildKey(), &x)
	if err != nil {
		panic(err)
	}
	return found
}

// PutMustRebuild sets the eponymous database record value.
func (f *File) PutMustRebuild() error {
	return f.Put(f.mustRebuildKey(), nil)
}

// DeleteMustRebuild removed the database record.
func (f *File) DeleteMustRebuild() error {
	return f.Delete(f.mustRebuildKey())
}
