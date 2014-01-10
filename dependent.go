package redo

func (f *File) AllDependents() ([]*File, error) {
	return f.relatedFiles(f.makeKey(SATISFIES))
}

func (f *File) EventDependents(event Event) ([]*File, error) {
	return f.relatedFiles(f.makeKey(SATISFIES, event))
}

func (f *File) DeleteAllDependencies() (err error) {
  keys, err := f.db.GetKeys(f.makeKey(SATISFIES))
  if err != nil {
	return err
  }

  for _, key := range keys {
	if err := f.db.Delete(key); err != nil {
	  return err
	}
  }
  return nil
}

// NotifyDependents flags dependents as out of date because target has been created, modified,  or deleted.
func (f *File) NotifyDependents(e Event) (err error) {

	dependents, err := f.EventDependents(e)
	if err != nil {
		return err
	}

	for _, dependent := range dependents {
		if err := dependent.PutMustRebuild(); err != nil {
			return err
		}
	}

	return nil
}
