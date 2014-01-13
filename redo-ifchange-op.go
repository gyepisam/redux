package redo

import (
	"fmt"
)

// RedoIfChange runs redo on the target if it is out of date or its current state
// disagrees with its dependent's version of its state.
func (target *File) RedoIfChange(dependent *File) error {

  recordRelation := func(m Metadata) error {
	p := target.AsPrerequisiteMetadata(m)
	return RecordRelation(dependent, target, IFCHANGE, p)
  }

	isCurrent, err := target.IsCurrent()
	if err != nil {
		return err
	} else if isCurrent {
		targetMetadata, exists, err := target.NewMetadata()
		if err != nil {
			return err
		} else if exists {

			// dependent's version of the target's state.
			prereq, found, err := dependent.GetPrerequisite(IFCHANGE, target.PathHash)
			if err != nil {
				return err
			} else if found {
				if prereq.Metadata.Equal(targetMetadata) {
					// target is up to date and its current state agrees with dependent's version.
					// Nothing to do here.
					return nil
				}
			} else {
				// There is no record of the dependency so this is the first time through.
				// Since the target is up to date, use its metadata for the dependency.
				return recordRelation(targetMetadata)
			}
		} else {
			/*
				Technically, this branch should be an error: a target just deemed to be current should not
				subsequently fail to exist. However, it is certainly possible for a file to be deleted
				between the two actions. Fortuitously, the file will be recreated, if possible.
			*/
		}
	}

	if err := target.Redo(); err != nil {
		return err
	}

	if targetMetadata, found, err := target.NewMetadata(); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("Cannot find recently created target: %s", target.Target)
	} else {
		return recordRelation(targetMetadata)
	}
}
