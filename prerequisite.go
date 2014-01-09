package redo

type Prerequisite struct {
	Path string
	Metadata
}

type PrerequisiteRecord struct {
	Key   string
	Value *Prerequisite
}

func NewPrerequisite(path string) (p Prerequisite, err error) {

	p.Path = path

	m, _, err := NewMetadata(path)
	if err != nil {
		return
	}
	p.Metadata = m

	return
}

func (f *File) putPrerequisite(e Event, path string) error {

	p, err := NewPrerequisite(path)
	if err != nil {
		return err
	}

	b, err := p.encode()
	if err != nil {
		return err
	}

	pf, err := NewFile(path)
	if err != nil {
	  return err
	}

	return f.db.Put(f.makeKey(REQUIRES, e, pf.PathHash), b)
}

func (f *File) PutPrerequisite(e Event, path string) error {
	return f.putPrerequisite(e, path)
}

func (f *File) PutDoPrerequisite(e Event, path string) error {
	return f.putPrerequisite(e.AutoPrefix(), path)
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

func (f *File) deletePrerequisites(prefix string) error {
	fn := func(rec *PrerequisiteRecord) error {
		return f.db.Delete(rec.Key)
	}

	return f.forPrerequisites(prefix, fn)
}

func (f *File) DeleteDoPrerequisites() error {
	return f.deletePrerequisites(f.makeKey(REQUIRES, AUTO))
}

func (f *File) DeleteAllPrerequisites() error {
	return f.deletePrerequisites(f.makeKey(REQUIRES))
}
