// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

type Record struct {
	Key   string
	Value []byte
}

// DB allows allows for multiple implementations of the redo database.
type DB interface {
	IsNull() bool

	// Put stores value under key
	Put(key string, value []byte) error

	// Get returns the value stored under key and a boolean indicating
	// whether the returned value was actually found in the database.
	Get(key string) ([]byte, bool, error)

	// Delete removes the value stored under key.
	// If key does not exist, the operation is a noop.
	Delete(key string) error

	// GetKeys returns a list of keys which have the specified key as a prefix.
	GetKeys(prefix string) ([]string, error)

	// GetValues returns a list of values whose keys matching the specified key prefix.
	GetValues(prefix string) ([][]byte, error)

	// GetRecords returns a list of records (keys and data) matchign the specified key prefix.
	GetRecords(prefix string) ([]Record, error)

	Close() error
}

func WithDB(arg string, f func(DB) error) error {
	db, err := FileDbOpen(arg)
	if err != nil {
		return err
	}
	defer db.Close()
	return f(db)
}

// Delete removes a target's database records
func (f *File) DeleteRecords() error {
	//   procedure delete-file-record(file)
	if err := f.NotifyDependents(IFCREATE); err != nil {
		return err
	}

	if err := f.DeleteAllDependencies(); err != nil {
		return err
	}

	if err := f.DeleteAllPrerequisites(); err != nil {
		return err
	}

	if err := f.DeleteMetadata(); err != nil {
		return err
	}

	if err := f.DeleteMustRebuild(); err != nil {
		return err
	}

	return nil
}

func (f *File) Delete(key string) error {
	err := f.db.Delete(key)
	f.Debug("@Delete: %s -> %s\n", key, err)
	return err
}
