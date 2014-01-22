// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
)

type Hash string

func MakeHash(content interface{}) Hash {

	hash := sha1.New()

	switch c := content.(type) {
	case []byte:
		hash.Write(c)
	case string:
		hash.Write([]byte(c))
	default:
		panic(fmt.Errorf("Unhandled argument: %+v", content))
	}

	return Hash(hex.EncodeToString(hash.Sum(nil)))
}

func ContentHash(path string) (hash Hash, err error) {
	b, err := ioutil.ReadFile(path)
	if err == nil {
		hash = MakeHash(b)
	}
	return
}
