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

func (m Metadata) Equal(other Metadata) (bool) {
  //Path and DoFiles are excluded from comparison. 
  return m.Size == other.Size &&
  m.ModTime == other.ModTime &&
  m.ContentHash == other.ContentHash
}

func NewMetadata(path string, storedPath string) (m Metadata, found bool, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return m, false, err
	}
	m.Path = storedPath
	m.Size = fi.Size()
	m.ModTime = fi.ModTime()

	m.ContentHash, err = ContentHash(path)

	found = true

	return
}

func (m Metadata) HasDoFile() bool {
	return len(m.DoFile) > 0
}

// PutMetadata stores the file's metadata in the database.
func (f *File) PutMetadata(m *Metadata) error {
  if m != nil {
	return f.Put(f.metadataKey(), *m)
  }

  if m, _, err := NewMetadata(f.Fullpath(), f.Path); err != nil {
	return err
  } else {
	return f.Put(f.metadataKey(), m)
  }
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


