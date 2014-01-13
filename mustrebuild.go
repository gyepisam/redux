package redo

// MustRebuild returns a boolean denoting whether the target must be rebuilt.
func (f *File) MustRebuild() bool {
   var x string 
	found, err := f.Get(f.mustRebuildKey(), &x)
	if err != nil {
		panic(err)
	}
	return found
}

func (f *File) PutMustRebuild() error {
	return f.Put(f.mustRebuildKey(), nil)
}

func (f *File) DeleteMustRebuild() error {
	return f.Delete(f.mustRebuildKey())
}
