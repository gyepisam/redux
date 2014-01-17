package redo

type Dependent struct {
	Path string
}

func (f *File) DependentFiles(prefix string) ([]*File, error) {

	data, err := f.db.GetValues(prefix)
	if err != nil {
		return nil, err
	}

	files := make([]*File, len(data))

	for i, b := range data {
		if dep, err := decodeDependent(b); err != nil {
			return nil, err
		} else if item, err := NewFile(dep.Path); err != nil {
			return nil, err
		} else {
			files[i] = item
		}
	}

	return files, nil
}

func (f *File) AllDependents() ([]*File, error) {
	return f.DependentFiles(f.makeKey(SATISFIES))
}

func (f *File) EventDependents(event Event) ([]*File, error) {
	return f.DependentFiles(f.makeKey(SATISFIES, event))
}

func (f *File) DeleteAllDependencies() (err error) {
	keys, err := f.db.GetKeys(f.makeKey(SATISFIES))
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := f.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) DeleteDependency(event Event, hash Hash) error {
	return f.Delete(f.makeKey(SATISFIES, event, hash))
}

func (f *File) PutDependency(event Event, hash Hash, path string) error {
	return f.Put(f.makeKey(SATISFIES, event, hash), Dependent{Path: path})
}

// NotifyDependents flags dependents as out of date because target has been created, modified,  or deleted.
func (f *File) NotifyDependents(event Event) (err error) {

	dependents, err := f.EventDependents(event)
	if err != nil {
		return err
	}

	for _, dependent := range dependents {
		if err := dependent.PutMustRebuild(); err != nil {
			return err
		}
		f.Log("@Notify %s %s -> %s\n", event, f.Path, dependent.Path)
	}

	return nil
}
