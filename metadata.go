package redo

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

func (m Metadata) Equal(other Metadata) bool {
    //Path is excluded from comparison. 
	return m.Size == other.Size &&
		m.ModTime == other.ModTime &&
		m.ContentHash == other.ContentHash &&
		m.DoFile == other.DoFile
}

func NewMetadata(path string) (m Metadata, found bool, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return m, false, err
	}
	m.Path = path
	m.Size = fi.Size()
	m.ModTime = fi.ModTime()

	m.ContentHash, err = ContentHash(path)

	found = true

	return
}

func (m Metadata) HasDoFile() bool {
	return len(m.DoFile) > 0
}

// DeleteMetadata removes the metadata record, if any, from the database.
func (f *File) DeleteMetadata() error {
	return f.db.Delete(f.metadataKey())
}

// putMetadata stores the provided metadata in the database.
func (f *File) putMetadata(m Metadata) error {
	b, err := m.encode()
	if err != nil {
		return err
	}

	return f.db.Put(f.metadataKey(), b)
}

// PutMetadata stores the file's metadata in the database.
func (f *File) PutMetadata() error {
	m, _, err := NewMetadata(f.Fullpath())
	if err != nil {
		return err
	}
	return f.putMetadata(m)
}

// GetMetadata returns a record as a metadata structure
// found denotes whether the record was found.
func (f *File) GetMetadata(keys ...interface{}) (Metadata, bool, error) {
	var key string

	if len(keys) > 0 {
		key = f.makeKey(keys...)
	} else {
		key = f.metadataKey()
	}

	m := Metadata{}
	found, err := f.Get(key, &m)
	return m, found, err
}
