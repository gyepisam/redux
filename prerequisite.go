package redo

type Prerequisite struct {
	Path string
	Metadata
}

type PrerequisiteRecord struct {
	Key   string
	Value *Prerequisite
}

func (f *File) PutPrerequisite(event Event, hash Hash, prereq Prerequisite) error {
	return f.Put(f.makeKey(REQUIRES, event, hash), prereq)
}

func (f *File) GetPrerequisite(event Event, hash Hash) (Prerequisite, bool, error) {
	p := Prerequisite{}
	found, err := f.Get(f.makeKey(REQUIRES, event, hash), &p)
	return p, found, err
}

func (f *File) PrerequisiteRecords(prefix string) ([]*PrerequisiteRecord, error) {

	records, err := f.db.GetRecords(prefix)
	if err != nil {
		return nil, err
	}

	out := make([]*PrerequisiteRecord, len(records))

	for i, rec := range records {
		if decoded, err := decodePrerequisite(rec.Value); err != nil {
			return nil, err
		} else {
			out[i] = &PrerequisiteRecord{Key: rec.Key, Value: &decoded}
		}
	}

	return out, nil
}

func (f *File) Prerequisites() ([]*Prerequisite, error) {
	records, err := f.PrerequisiteRecords(f.makeKey(REQUIRES))
	if err != nil {
		return nil, err
	}

	out := make([]*Prerequisite, len(records))

	for i, rec := range records {
		out[i] = rec.Value
	}

	return out, nil
}

func (f *File) PrerequisiteFiles() ([]*File, error) {

	records, err := f.PrerequisiteRecords(f.makeKey(REQUIRES))
	if err != nil {
		return nil, err
	}

	out := make([]*File, len(records))

	for i, rec := range records {
		if file, err := NewFile(rec.Value.Path); err != nil {
			return nil, err
		} else {
			out[i] = file
		}
	}

	return out, nil
}

func (f *File) forPrerequisites(prefix string, fn func(*PrerequisiteRecord) error) error {
	records, err := f.PrerequisiteRecords(prefix)
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

func (f *File) DeletePrerequisite(event Event, hash Hash) error {
  return f.Delete(f.makeKey(REQUIRES, event, hash))
}

func (f *File) deletePrerequisites(prefix string) error {
	fn := func(rec *PrerequisiteRecord) error {
		return f.Delete(rec.Key)
	}

	return f.forPrerequisites(prefix, fn)
}

func (f *File) DeleteDoPrerequisites() error {
	return f.deletePrerequisites(f.makeKey(REQUIRES, AUTO))
}

func (f *File) DeleteAllPrerequisites() error {
	return f.deletePrerequisites(f.makeKey(REQUIRES))
}
