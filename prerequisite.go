// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redo

// Prerequisite from a source to a target.
type Prerequisite struct {
	Path      string // path back to target of prerequisite.
	*Metadata        // target's metadata upon record creation.
}

// PutPrerequisite stores the given prerequisite using a key based on the event and hash.
func (f *File) PutPrerequisite(event Event, hash Hash, prereq Prerequisite) error {
	return f.Put(f.makeKey(REQUIRES, event, hash), prereq)
}

// GetPrerequisite returns the prerequisite for the event and hash.
// If the record does not exist, found is false and err is nil.
func (f *File) GetPrerequisite(event Event, hash Hash) (prereq Prerequisite, found bool, err error) {
	found, err = f.Get(f.makeKey(REQUIRES, event, hash), &prereq)
	return
}

type record struct {
	key string
	*Prerequisite
}

func prefixed(f *File, prefix string) ([]*record, error) {

	rows, err := f.db.GetRecords(prefix)
	if err != nil {
		return nil, err
	}

	out := make([]*record, len(rows))

	for i, row := range rows {
		if decoded, err := decodePrerequisite(row.Value); err != nil {
			return nil, err
		} else {
			out[i] = &record{row.Key, &decoded}
		}
	}

	return out, nil
}

// Prerequisites returns a slice of prerequisites for the file.
func (f *File) Prerequisites() (out []*Prerequisite, err error) {
	records, err := prefixed(f, f.makeKey(REQUIRES))
	if err != nil {
		return
	}

	out = make([]*Prerequisite, len(records))

	for i, rec := range records {
		out[i] = rec.Prerequisite
	}

	return
}

// PrerequisiteFiles returns a slice of *File objects for the file's prerequisites for the list of events.
func (f *File) PrerequisiteFiles(events ...Event) ([]*File, error) {

	var records []*record

	for _, event := range events {
		eventRecords, err := prefixed(f, f.makeKey(REQUIRES, event))
		if err != nil {
			return nil, err
		}
		records = append(records, eventRecords...)
	}

	out := make([]*File, len(records))

	for i, rec := range records {
		if file, err := rec.File(f.RootDir); err != nil {
			return nil, err
		} else {
			out[i] = file
		}
	}

	return out, nil
}

// DeletePrerequisite removes a single prerequisite.
func (f *File) DeletePrerequisite(event Event, hash Hash) error {
	return f.Delete(f.makeKey(REQUIRES, event, hash))
}

type visitor func(*record) error

func visit(f *File, prefix string, fn visitor) error {
	records, err := prefixed(f, prefix)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if err := fn(rec); err != nil {
			return err
		}
	}

	return nil
}

func destroy(f *File, prefix string) error {
	return visit(f, prefix, func(rec *record) error {
		return f.Delete(rec.key)
	})
}

// DeleteAutoPrerequisites removes all of the file's system generated prerequisites.
func (f *File) DeleteAutoPrerequisites() error {
	return destroy(f, f.makeKey(REQUIRES, AUTO))
}

// DeleteAllPrerequisites removed all of the file's prerequisites.
func (f *File) DeleteAllPrerequisites() error {
	return destroy(f, f.makeKey(REQUIRES))
}
