// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redo

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func initRoot() (root string, fn func(), err error) {

	root, err = ioutil.TempDir("", "redo-db-root-")
	fn = func() {
		os.RemoveAll(root)
	}

	if err != nil {
		return
	}

	err = InitDir(root)
	return
}

func makeDBFunc(root string) func(func(DB) error) error {
	return func(f func(DB) error) error {
		return WithDB(root, f)
	}
}

func randBytes(b []byte) (err error) {
	_, err = io.ReadFull(rand.Reader, b)
	return
}

func TestDBOpen(t *testing.T) {

	root, fn, err := initRoot()
	if err != nil {
		t.Fatal(err)
	}
	defer fn()

	f := func(db DB) error {
		return nil
	}

	if err := WithDB(root, f); err != nil {
		t.Fatal(err)
	}
}

func TestDBNullKey(t *testing.T) {
	root, fn, err := initRoot()
	if err != nil {
		t.Fatal(err)
	}
	defer fn()

	err = WithDB(root, func(db DB) error {
		return db.Put("", []byte{0})
	})

	if err != NullKeyErr {
		t.Fatal(err)
	}
}

func TestDBAction(t *testing.T) {
	root, fn, err := initRoot()
	if err != nil {
		t.Fatal(err)
	}
	defer fn()

	type KeyValue struct {
		key   string
		value []byte
	}

	sizes := []int{0, 1, 2, 3, 5, 7, 10, 100, 1001, 10002, 100003}
	data := make([]KeyValue, len(sizes))
	var prefixKeys []string
	var prefixValues [][]byte

	prefix := "random/byte/0"

	for i, size := range sizes {

		data[i].key = fmt.Sprintf("random/byte/%02d", size)

		data[i].value = make([]byte, size)
		if err := randBytes(data[i].value); err != nil {
			t.Fatal(err)
		}
		if size < 10 {
			prefixKeys = append(prefixKeys, data[i].key)
			prefixValues = append(prefixValues, data[i].value)
		}
	}

	for _, rec := range data {
		//Put
		err := WithDB(root, func(db DB) error {
			return db.Put(rec.key, rec.value)
		})

		if err != nil {
			t.Fatalf("Error: %s in db.Put(%q,...)", err, rec.key)
		}

		// Get
		var value []byte
		err = WithDB(root, func(db DB) error {
			var found bool
			var err error
			value, found, err = db.Get(rec.key)
			if err == nil && !found {
				err = fmt.Errorf("cannot find key: %s", rec.key)
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(rec.value, value) {
			t.Fatalf("Get: input and output bytes differ: %v != %v", rec.value, value)
		}

		t.Logf("%s/%s\n", root, rec.key)
	}

	//GetKeys
	var keys []string
	err = WithDB(root, func(db DB) error {
		var err error
		keys, err = db.GetKeys(prefix)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(prefixKeys, keys) {
		t.Fatalf("GetKeys: input and output prefixed keys differ: %v != %v", prefixKeys, keys)
	}

	//GetValues
	var values [][]byte
	err = WithDB(root, func(db DB) error {
		var err error
		values, err = db.GetValues(prefix)
		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(prefixValues, values) {
		t.Fatalf("GetValues: input and output prefixed values differ: %v != %v", prefixValues, values)
	}

	//Delete
	for _, rec := range data {
		err = WithDB(root, func(db DB) error {
			return db.Delete(rec.key)
		})

		if err != nil {
			t.Fatal(err)
		}
	}

}
