package apntool

import "github.com/GlshchnkLx/go-aospapn/pkg/apnxml"

func include(data apnxml.Array, predicate Predicate) apnxml.Array {
	if predicate == nil {
		predicate = All
	}

	var result apnxml.Array
	for groupIndex := range data {
		group := &data[groupIndex]

		if !group.HasGroup() {
			record := MaterializeRecord(nil, group)
			if predicate(record) {
				result = append(result, record)
			}
			continue
		}

		groupClone := apnxml.Object{
			ObjectRoot:     group.ObjectRoot.Clone(),
			GroupMapByType: map[apnxml.ObjectBaseType]*apnxml.Object{},
		}

		for _, apnType := range group.GroupTypes() {
			record := group.GroupMapByType[apnType]
			materializedRecord := MaterializeRecord(group, record)
			if predicate(materializedRecord) {
				groupClone.GroupMapByType[apnType] = record.Clone()
			}
		}

		if len(groupClone.GroupMapByType) > 0 {
			groupClone.Carrier = groupClone.GetCarrier()
			result = append(result, groupClone)
		}
	}

	return result
}

func exclude(data apnxml.Array, predicate Predicate) apnxml.Array {
	if predicate == nil {
		return data.Clone()
	}

	return include(data, Not(predicate))
}
