package redo

// MustRebuild returns a boolean denoting whether the target must be rebuilt.
func (f *File) MustRebuild() bool {
	_, found, err := f.db.Get(f.mustRebuildKey())
	if err != nil {
		panic(err)
	}
	return found
}

func (f *File) PutMustRebuild() error {
	return f.db.Put(f.mustRebuildKey(), []byte(nil))
}

func (f *File) DeleteMustRebuild() error {
	return f.db.Delete(f.mustRebuildKey())
}
