// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gyepisam/fileutils"
)

const (
	// Where is data kept?
	DATA_DIR = "data"
)

// FileDb is a file based DB for storing Redo relationships and metadata.
type FileDb struct {
	DataDir string
}

var NullKeyErr = errors.New("Key cannot be empty.")
var NullPrefixErr = errors.New("Prefix cannot be empty.")

// Open requires a project root argument
func FileDbOpen(rootdir string) (DB, error) {

	redodir := filepath.Join(rootdir, REDO_DIR)

	if exists, err := fileutils.DirExists(redodir); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("FileDb redo directory [%s] not found. Forget to call redo-init?", redodir)
	}

	datadir := filepath.Join(redodir, DATA_DIR)
	if err := os.Mkdir(datadir, DIR_PERM); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("FileDb cannot make data directory [%s]. %s", datadir, err)
	}

	return &FileDb{DataDir: datadir}, nil
}

func (db *FileDb) IsNull() bool { return false }

func (db *FileDb) makePath(key string) string {
	return filepath.Join(db.DataDir, key)
}

// Close is a noop for this DB type
func (db *FileDb) Close() error {
	return nil
}

func (db *FileDb) Put(key string, value []byte) error {
	if len(key) == 0 {
		return NullKeyErr
	}
	return fileutils.AtomicWrite(db.makePath(key), func(tmpFile *os.File) error {
		_, err := tmpFile.Write(value)
		return err
	})
}

func (db *FileDb) Get(key string) ([]byte, bool, error) {
	if len(key) == 0 {
		return nil, false, NullKeyErr
	}

	b, err := ioutil.ReadFile(db.makePath(key))

	var found bool
	if err == nil {
		found = true
	} else if os.IsNotExist(err) {
		found = false //explicit is better than implicit
		err = nil
	}
	return b, found, err
}

func (db *FileDb) Delete(key string) error {
	if len(key) == 0 {
		return NullKeyErr
	}

	err := os.Remove(db.makePath(key))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

func (db *FileDb) GetRecords(prefix string) ([]Record, error) {

	if len(prefix) == 0 {
		return nil, NullPrefixErr
	}

	var out []Record
	rootLen := len(db.DataDir) + 1

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		key := path[rootLen:]
		// Go 1.0.x compatible syntax for info.Mode().IsRegular()
		if isRegular := info.Mode()&os.ModeType == 0; isRegular && strings.HasPrefix(key, prefix) {
			if b, err := ioutil.ReadFile(path); err != nil {
				return err
			} else {
				out = append(out, Record{Key: key, Value: b})
			}
		}

		return nil
	}

	return out, filepath.Walk(db.DataDir+"/", walker)
}

// GetKeys returns an array of keys that are prefixes of the specified key.
func (db *FileDb) GetKeys(prefix string) ([]string, error) {

	records, err := db.GetRecords(prefix)
	if err != nil {
		return nil, err
	}

	out := make([]string, len(records))
	for i, rec := range records {
		out[i] = rec.Key
	}
	return out, nil
}

// GetValues returns an array of data values for keys with the specified prefix.
func (db *FileDb) GetValues(prefix string) ([][]byte, error) {
	records, err := db.GetRecords(prefix)
	if err != nil {
		return nil, err
	}

	out := make([][]byte, len(records))
	for i, rec := range records {
		out[i] = make([]byte, len(rec.Value))
		copy(out[i], rec.Value)
	}
	return out, nil
}
