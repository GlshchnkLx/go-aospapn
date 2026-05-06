package apntool

import (
	"fmt"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func forEachEntryPointer(data apnxml.Array, handler EntryMutator) error {
	if handler == nil {
		return nil
	}

	for groupIndex := range data {
		group := &data[groupIndex]
		records := group.Records()
		for recordIndex, record := range records {
			if record == nil {
				continue
			}

			if err := handler(group, record); err != nil {
				return fmt.Errorf("for each group %d record %d: %w", groupIndex, recordIndex, err)
			}
		}
	}

	return nil
}

func MaterializeRecord(group *apnxml.Object, record *apnxml.Object) apnxml.Object {
	if record == nil {
		return apnxml.Object{}
	}

	recordClone := record.Clone()
	if recordClone == nil {
		return apnxml.Object{}
	}

	if recordClone.ObjectRoot == nil && group != nil && group.ObjectRoot != nil {
		recordClone.ObjectRoot = group.ObjectRoot.Clone()
	}

	return *recordClone
}
