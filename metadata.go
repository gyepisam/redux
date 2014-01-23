// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"os"
	"time"
)

// File Metadata.
type Metadata struct {
	Path        string //not used for comparison
	Size        int64
	ModTime     time.Time
	ContentHash Hash
	DoFile      string
}

// Equal compares metadata instances for equality.
func (m *Metadata) Equal(other *Metadata) bool {
	return other != nil && m.ContentHash == other.ContentHash
}

// IsCreated compares m to other to determine m represents a newly created file.
func (m Metadata) IsCreated(other Metadata) bool {
	return len(other.ContentHash) == 0 && len(m.ContentHash) > 0
}

// NewMetadata returns a metadata instance for the given path.
// If the file is not found, nil is returned.
func NewMetadata(path string, storedPath string) (m *Metadata, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	m = &Metadata{Path: storedPath, Size: fi.Size(), ModTime: fi.ModTime()}

	m.ContentHash, err = ContentHash(path)

	return
}

//HasDoFile returns true if the metadata has a non-empty DoField field.
func (m Metadata) HasDoFile() bool {
	return len(m.DoFile) > 0
}

// PutMetadata stores the file's metadata in the database.
func (f *File) PutMetadata(m *Metadata) error {
	if m != nil {
		return f.Put(f.metadataKey(), *m)
	}

	m, err := NewMetadata(f.Fullpath(), f.Path)
	if err != nil {
		return err
	}
	if m == nil {
		return f.ErrNotFound("PutMetadata")
	}

	return f.Put(f.metadataKey(), m)
}

// GetMetadata returns a record as a metadata structure
// found denotes whether the record was found.
func (f *File) GetMetadata() (Metadata, bool, error) {
	m := Metadata{}
	found, err := f.Get(f.metadataKey(), &m)
	return m, found, err
}

// DeleteMetadata removes the metadata record, if any, from the database.
func (f *File) DeleteMetadata() error {
	return f.Delete(f.metadataKey())
}
